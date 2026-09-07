package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	"github.com/bashocode/gowallet/monolith/internal/email"
	customErr "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/otp/generator"
	otpModel "github.com/bashocode/gowallet/monolith/internal/otp/model"
	otpRepository "github.com/bashocode/gowallet/monolith/internal/otp/repository"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/repository"
	refreshRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userRepo "github.com/bashocode/gowallet/monolith/internal/user/repository"
	walletModel "github.com/bashocode/gowallet/monolith/internal/wallet/model"
	walletRepo "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Khai báo các nghiệp vụ người dùng có thể thực hiện
type UserService interface {
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	GenerateAndSendOTP(ctx context.Context, userID string, email string, otpType string) error
	UpdateAvatar(ctx context.Context, id string, path string) error
	DeleteAccount(ctx context.Context, id string) error
	Logout(ctx context.Context, tokenString string) error
	VerifyEmail(ctx context.Context, userID string, code string) error
	RequestPasswordReset(ctx context.Context, email string) error
	VerifyPasswordReset(ctx context.Context, email string, code string) (string, error)
	ResetPassword(ctx context.Context, email string, newPassword string) error
	GetGoogleLoginURL(ctx context.Context) (string, error)
	HandleGoogleCallback(ctx context.Context, code string, state string) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, oldTokenString string) (*model.LoginResponse, error)
	GetAllUsers(ctx context.Context, params model.PaginationParams) ([]*model.User, *model.PaginationMeta, error)
}

// Struct ẩn chứa dependency để giao tiếp với cơ sở dữ liệu.
type userService struct {
	db          *sql.DB
	rdb         *redis.Client
	userRepo    userRepo.UserRepository
	walletRepo  walletRepo.WalletRepository
	otpRepo     otpRepository.OTPRepository
	rtRepo      refreshRepository.RefreshTokenRepository
	emailSender email.EmailSender
}

// Hàm khởi tạo trả về interface UserService, áp dụng mô hình Dependency Injection.
func NewUserService(
	db *sql.DB,
	rdb *redis.Client,
	uRepo repository.UserRepository,
	wRepo walletRepo.WalletRepository,
	otpRepo otpRepository.OTPRepository,
	emailSender email.EmailSender,
) UserService {
	return &userService{
		db:          db,
		rdb:         rdb,
		userRepo:    uRepo,
		walletRepo:  wRepo,
		otpRepo:     otpRepo,
		rtRepo:      repository.NewMySQLRefreshTokenRepository(db),
		emailSender: emailSender,
	}
}

// (Đăng ký tài khoản
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	//1. check if the email already registered
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		// return custom AppError
		return nil, customErr.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "this email already registered.")
	}

	//hahs the password with bcrypt
	hashsedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	//2. create new user object
	user := &model.User{
		ID:           uuid.New().String(), //Khởi tạo một mã UUID v4 ngẫu nhiên làm khóa chính cho user mới.
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashsedBytes),
	}

	//begin transaction database
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	//we should rollback if anything error or panic in the middleware
	//save to 2 table, users and wallet, if open of them lets say wallet if failed..
	defer tx.Rollback()

	//store user to db with a tx connection
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		return nil, customErr.ErrInternalServer
	}

	//create wallet for the user
	wallet := &walletModel.Wallet{
		ID:       uuid.New().String(),
		UserID:   user.ID,
		Balance:  decimal.NewFromInt(0),
		Currency: "IDR",
		Status:   "active",
	}

	if err := s.walletRepo.CreateTx(ctx, tx, wallet); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// commit the transaction if all of the step is success
	if err := tx.Commit(); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// // Generate OTP
	// // Tạo số ngẫu nhiên an toàn mật mã từ rand.Reader để làm mã OTP gồm 6 chữ số.
	// n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	// if err != nil {
	// 	return nil, customErr.ErrInternalServer
	// }
	// otpCode := fmt.Sprintf("%06d", n.Int64())
	// fmt.Println("otp codes", otpCode)

	// // Khởi tạo đối tượng struct OTP chuẩn bị lưu trữ vào cơ sở dữ liệu.
	// otpModel := &otpModel.OTP{
	// 	ID:        uuid.New().String(),
	// 	UserID:    user.ID,
	// 	Code:      otpCode,
	// 	Type:      "email_verification",
	// 	ExpiresAt: time.Now().Add(15 * time.Minute),
	// 	Used:      false,
	// }

	// Generate and send OTP
	if err := s.GenerateAndSendOTP(ctx, user.ID, user.Email, "email_verification"); err != nil {
		logger.Log.Error("failed to generate and send otp during registration", "error", err)
	}

	// // save to db
	// if err := s.otpRepo.Create(ctx, otpModel); err != nil {
	// 	logger.Log.Error("failed to save otp", "error", err)
	// }

	// // Chạy tiến trình nền (goroutine) để gửi email chứa mã OTP mà không làm nghẽn/chậm response trả về cho client.
	// go func() {
	// 	// Tạo context mới độc lập với timeout là 10 giây cho tác vụ gửi email bất đồng bộ.
	// 	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()

	// 	subject := "GoWallet - Verify Your Email"
	// 	body := fmt.Sprintf("Hello %s,\n\nYour verification code is %s\n\nThis code will expire in 15 minutes.\n\nThank you!", user.FullName, otpCode)

	// 	// Gọi service gửi email đi thông qua giao diện SMTPEmailSender đã cấu hình.
	// 	s.emailSender.SendEmail(bgCtx, user.Email, subject, body)
	// }()

	// return the new user
	return s.userRepo.GetByID(ctx, user.ID)

}

// Lấy thông tin
func (s *userService) GetProfile(ctx context.Context, id string) (*model.User, error) {
	//Chuyển tiếp yêu cầu lấy thông tin trực tiếp xuống tầng repository thông qua GetByID
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	return u, nil
}

// Cập nhật thông tin
func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	//Kiểm tra xem user cần cập nhật có tồn tại hay không.
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	//Thay đổi tên mới theo dữ liệu client gửi lên
	user.FullName = req.FullName
	//Lưu thay đổi xuống database và trả về thông tin mới nhất của user.
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, customErr.ErrInternalServer
	}
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// find by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password.")
	}

	// verify the hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, customErr.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password.")
	}

	// generate access token 15 minutes
	accessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role, 15*time.Minute)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// generate refresh token 7 days
	// / Gọi hàm sinh token định danh cho người dùng với thời hạn sống kéo dài 7 ngày dùng làm Refresh Token.
	refreshToken, err := auth.GenerateToken(user.ID, user.Email, user.Role, 7*24*time.Hour)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// save token to db
	rt := &model.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}
	if err := s.rtRepo.Create(ctx, rt); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// return the tokens
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) UpdateAvatar(ctx context.Context, id string, path string) error {
	return s.userRepo.UpdateAvatar(ctx, id, path)
}

func (s *userService) DeleteAccount(ctx context.Context, id string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return customErr.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	if err := s.userRepo.SoftDelete(ctx, user.ID); err != nil {
		return customErr.ErrInternalServer
	}

	return nil
}

// Logout thực hiện chức năng đăng xuất bằng cách đưa JWT token hiện tại vào danh sách đen (blacklist) trên Redis.
func (s *userService) Logout(ctx context.Context, tokenString string) error {
	// validate token
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return customErr.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid or expired.")
	}

	// revoke all refresh token from users that logged out
	if err := s.rtRepo.RevokeAllByUserID(ctx, claims.UserID); err != nil {
		return customErr.NewAppError(http.StatusUnauthorized, "REVOKE_FAILED", "Failed to revoke refresh token.")
	}

	// calculate the remaining active token
	// Lấy ra thời điểm token sẽ hết hạn từ thông tin claims của JWT.
	expirationTime := claims.ExpiresAt.Time
	// Tính toán khoảng thời gian còn lại trước khi token tự hết hạn.
	timeLeft := time.Until(expirationTime)

	// Nếu thời gian còn lại nhỏ hơn hoặc bằng 0 (nghĩa là token đã hết hạn rồi).
	if timeLeft <= 0 {
		return nil // token already expired, no need to blacklist
	}

	// insert into redis blacklist
	// Tạo khóa (key) định danh trên Redis cho token cần đưa vào blacklist
	blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
	// Lưu token vào Redis với trạng thái "logged_out" và đặt thời gian tồn tại bằng với thời gian còn lại của token (timeLeft).
	// Việc này giúp Redis tự động xóa key khi token hết hạn tự nhiên, tiết kiệm bộ nhớ
	err = s.rdb.Set(ctx, blacklistKey, "logged_out", timeLeft).Err()
	if err != nil {
		return customErr.ErrInternalServer
	}

	return nil
}

// / VerifyEmail xử lý logic nghiệp vụ xác thực email của người dùng thông qua mã OTP, đảm bảo tính toàn vẹn dữ liệu bằng Database Transaction

func (s *userService) VerifyEmail(ctx context.Context, userID string, code string) error {
	// 1. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return customErr.ErrInternalServer
	}
	defer tx.Rollback()

	// 2. Get active OTP inside the transaction (acquires FOR UPDATE lock)
	otp, err := s.otpRepo.GetActiveOTPTx(ctx, tx, userID, code, "email_verification")
	if err != nil {
		// Custom AppError: OTP not found or expired
		return customErr.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 3. Mark user as verified
	if err := s.userRepo.UpdateVerificationStatusTx(ctx, tx, userID, true); err != nil {
		return customErr.ErrInternalServer
	}

	// 4. Mark OTP as used
	if err := s.otpRepo.MarkAsUsedTx(ctx, tx, otp.ID); err != nil {
		return customErr.ErrInternalServer
	}

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		return customErr.ErrInternalServer
	}

	return nil
}

// GenerateAndSendOTP tạo mã OTP ngẫu nhiên an toàn, lưu trữ vào cơ sở dữ liệu và kích hoạt tiến trình nền gửi email thông báo cho người dùng tùy theo loại nghiệp vụ
func (s *userService) GenerateAndSendOTP(ctx context.Context, userID string, emailAddr string, otpType string) error {
	// Gọi generator tạo một mã OTP dạng chuỗi gồm 6 chữ số an toàn mật mã
	otpCode, err := generator.GenerateOTP(6)
	if err != nil {
		return customErr.ErrInternalServer
	}
	// Khởi tạo đối tượng model OTP mới với thời gian sống hiệu lực là 15 phút.
	otpModel := &otpModel.OTP{
		ID:        uuid.New().String(),
		UserID:    userID,
		Code:      otpCode,
		Type:      otpType,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// Lưu bản ghi OTP vào cơ sở dữ liệu
	if err := s.otpRepo.Create(ctx, otpModel); err != nil {
		logger.Log.Error("failed to save otp", "error", err)
		return customErr.ErrInternalServer
	}

	// Chạy tiến trình nền (goroutine) bất đồng bộ để gửi email chứa mã OTP mà không làm block request của client.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var subject string
		var body string

		// Phân nhánh nội dung tiêu đề và thông điệp email dựa theo loại OTP (xác thực email hoặc khôi phục mật khẩu).
		switch otpType {
		case "email_verification":
			subject = "GoWallet - Verify Your Email"
			body = fmt.Sprintf("Your verification code is %s\n\nThis code will expire in 15 minutes.\n\nThank you!", otpCode)
		case "password_reset":
			subject = "GoWallet - Reset Your Password"
			body = fmt.Sprintf("Your password reset code is %s\n\nThis code will expire in 15 minutes.\n\nThank you!", otpCode)
		default:
			subject = "GoWallet - Security Code"
			body = fmt.Sprintf("Your code is %s\n\nThis code will expire in 15 minutes.\n\nThank you!", otpCode)
		}

		s.emailSender.SendEmail(bgCtx, emailAddr, subject, body)
	}()

	return nil
}

// RequestPasswordReset xử lý yêu cầu cấp lại mật khẩu từ người dùng qua email, đồng thời chống tấn công liệt kê tài khoản (email enumeration).
func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	// find by email // Tìm kiếm người dùng dựa trên email
	user, err := s.userRepo.GetByEmailNoErrorNotFound(ctx, email)
	if err != nil {
		return customErr.ErrInternalServer
	}
	if user == nil {
		// return nil to prevent email enumeration attacks
		return nil
	}
	// Nếu email tồn tại, tiến hành tạo và gửi mã OTP với loại "password_reset".
	return s.GenerateAndSendOTP(ctx, user.ID, user.Email, "password_reset")
}

// VerifyPasswordReset kiểm tra tính hợp lệ của mã OTP khôi phục mật khẩu và đánh dấu đã sử dụng, trả về userID nếu thành công.
func (s *userService) VerifyPasswordReset(ctx context.Context, email string, code string) (string, error) {
	// 1. Get user by email // 1. Lấy thông tin người dùng theo email
	user, err := s.userRepo.GetByEmailNoErrorNotFound(ctx, email)
	if err != nil || user == nil {
		return "", customErr.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 2. Truy vấn lấy mã OTP hoạt động (chưa hết hạn và chưa dùng) dành riêng cho việc reset password.
	// 2. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", customErr.ErrInternalServer
	}
	defer tx.Rollback()

	// 3. Get active OTP inside transaction (locks the row)
	otp, err := s.otpRepo.GetActiveOTPTx(ctx, tx, user.ID, code, "password_reset")
	if err != nil {
		return "", customErr.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 3. Đánh dấu mã OTP này đã được sử dụng (used = 1) để tránh việc tái sử dụng lại mã cũ.
	// 4. Mark OTP as used inside transaction
	if err := s.otpRepo.MarkAsUsedTx(ctx, tx, otp.ID); err != nil {
		return "", customErr.ErrInternalServer
	}

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		return "", customErr.ErrInternalServer
	}

	return user.ID, nil
}

// ResetPassword mã hóa mật khẩu mới của người dùng bằng thuật toán bcrypt và cập nhật vào cơ sở dữ liệu.
func (s *userService) ResetPassword(ctx context.Context, id string, newPassword string) error {
	// hash password using bcrypt // Mã hóa chuỗi mật khẩu mới thành chuỗi hash sử dụng bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return customErr.ErrInternalServer
	}

	// update password // Cập nhật mật khẩu mới đã mã hóa vào cơ sở dữ liệu cho người dùng tương ứng.
	if err := s.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return customErr.ErrInternalServer
	}

	// revoke all refresh token from user that reset the password
	// Thu hồi toàn bộ các Refresh Token của người dùng khi họ thực hiện đổi mật khẩu thành công.
	if err := s.rtRepo.RevokeAllByUserID(ctx, id); err != nil {
		return customErr.ErrInternalServer
	}

	return nil
}

// Helper struct for Google UserInfo API response
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// getOAuthConfig khởi tạo và trả về cấu hình OAuth2 kết nối với Google dựa trên các biến môi trường.
func (s *userService) getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (s *userService) GetGoogleLoginURL(ctx context.Context) (string, error) {
	// Generate secure random state token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state in Redis with 10 minute expiry
	stateKey := fmt.Sprintf("oauth:state:%s", state)
	err := s.rdb.Set(ctx, stateKey, "valid", 10*time.Minute).Err()
	if err != nil {
		return "", err
	}

	// Generate OAuth URL with random state
	url := s.getOAuthConfig().AuthCodeURL(state)
	return url, nil
}

// HandleGoogleCallback xử lý mã code từ Google, xác thực tài khoản, tự động đăng ký mới nếu chưa có và cấp phát JWT token nội bộ.
func (s *userService) HandleGoogleCallback(ctx context.Context, code string, state string) (*model.LoginResponse, error) {
	// Validate state token exists in Redis
	stateKey := fmt.Sprintf("oauth:state:%s", state)
	val, err := s.rdb.Get(ctx, stateKey).Result()
	if err != nil || val != "valid" {
		return nil, customErr.NewAppError(http.StatusBadRequest, "INVALID_STATE", "invalid or expired OAuth state - possible CSRF attack")
	}

	// Delete state token (one-time use)
	s.rdb.Del(ctx, stateKey)

	config := s.getOAuthConfig()

	// 1. Exchange the auth code for a token // 1. Thực hiện đổi mã code nhận được từ Google để lấy token.
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. Fetch userinfo using access token // Dùng token vừa đổi được tạo HTTP client gọi đến Google API để lấy thông tin profile người dùng.
	client := config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	// Giải mã dữ liệu JSON trả về từ Google vào struct GoogleUserInfo.
	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// 3. Find or Create user in database
	// Tìm kiếm trong database xem user đã đăng nhập bằng Google ID này trước đó chưa.
	var user *model.User
	user, err = s.userRepo.GetByOAuth(ctx, "google", googleUser.ID)
	if err != nil {
		// If user not found, register new user
		// Nếu không tìm thấy user theo OAuth ID, tiến hành quy trình đăng ký tài khoản mới.
		if err.Error() == "user not found" {
			// Check if email already registered via normal email/password
			// Kiểm tra xem email này đã từng được đăng ký bằng phương thức password thông thường chưa.
			existingUser, _ := s.userRepo.GetByEmail(ctx, googleUser.Email)
			if existingUser != nil {
				return nil, customErr.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "This email is registered using password credentials. Please sign in with email and password.")
			}

			// Begin database transaction to guarantee user & wallet creation consistency
			// Bắt đầu một Database Transaction để đảm bảo tính toàn vẹn (tạo user và tạo ví phải thành công cùng lúc).
			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, customErr.ErrInternalServer
			}
			defer tx.Rollback()

			provider := "google"
			user = &model.User{
				ID:            uuid.New().String(),
				FullName:      googleUser.Name,
				Email:         googleUser.Email,
				OAuthProvider: &provider,
				OAuthID:       &googleUser.ID,
				PasswordHash:  "", // Null password
				AvatarURL:     &googleUser.Picture,
				IsVerified:    true, // Google verified the email
			}

			// Save user // Lưu thông tin user mới vào database thông qua transaction.
			if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
				return nil, customErr.ErrInternalServer
			}

			// Create wallet for the new user // Khởi tạo ví tiền mặc định cho user vừa đăng ký
			wallet := &walletModel.Wallet{
				ID:       uuid.New().String(),
				UserID:   user.ID,
				Balance:  decimal.NewFromFloat(0.0),
				Currency: "IDR",
				Status:   "active",
				Version:  1,
			}
			if err := s.walletRepo.CreateTx(ctx, tx, wallet); err != nil {
				return nil, customErr.ErrInternalServer
			}

			// Hoàn tất transaction lưu dữ liệu xuống database.
			if err := tx.Commit(); err != nil {
				return nil, customErr.ErrInternalServer
			}
		} else {
			return nil, customErr.ErrInternalServer
		}
	}

	// 4. Generate JWT token
	// Tạo JWT Access Token nội bộ có thời hạn ngắn (15 phút) dùng cho các request được bảo vệ.
	accessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role, 15*time.Minute)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// Tạo JWT Refresh Token nội bộ có thời hạn dài (7 ngày).
	refreshToken, err := auth.GenerateToken(user.ID, user.Email, user.Role, 7*24*time.Hour)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// Lưu Refresh Token lên Redis để quản lý trạng thái phiên đăng nhập.
	// Save new Refresh Token to Database (align with normal login)
	newRT := &model.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}
	if err := s.rtRepo.Create(ctx, newRT); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// Trả về cặp token cho client.
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken xử lý logic làm mới Access Token bằng cơ chế Refresh Token Rotation và Reuse Detection.
func (s *userService) RefreshToken(ctx context.Context, oldTokenString string) (*model.LoginResponse, error) {
	// 1. look token in db
	// 1. Truy vấn cơ sở dữ liệu để tìm thông tin Refresh Token cũ do client gửi lên.
	rt, err := s.rtRepo.GetByToken(ctx, oldTokenString)
	if err != nil {
		return nil, customErr.NewAppError(http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Refresh token invalid.")
	}

	// 2. TOKEN REUSE DETECTION: if revoked token reused -> HACKER DETECTED!
	// 2. TOKEN REUSE DETECTION: Nếu token này đã bị đánh dấu thu hồi trước đó mà vẫn tiếp tục được gửi lên -> Phát hiện kẻ gian đánh cắp token!
	if rt.Revoked {
		// revoke all active session for this user
		// Lập tức vô hiệu hóa toàn bộ các session/token còn lại của người dùng
		_ = s.rtRepo.RevokeAllByUserID(ctx, rt.UserID)
		return nil, customErr.NewAppError(http.StatusUnauthorized, "TOKEN_BREACH_DETECTED", "Token breach detected. Please login again.")
	}

	// 3. check if token is expired // 3. Kiểm tra xem thời hạn của Refresh Token đã vượt quá mốc ExpiresAt so với thời điểm hiện tại hay chưa.
	if time.Now().After(rt.ExpiresAt) {
		return nil, customErr.NewAppError(http.StatusUnauthorized, "EXPIRED_REFRESH_TOKEN", "Refresh token expired. Please login again.")
	}

	// 4. Revoke old token
	// 4. Vô hiệu hóa (thu hồi) Refresh Token cũ ngay lập tức để thực hiện quy trình xoay vòng (Rotation).
	if err := s.rtRepo.Revoke(ctx, oldTokenString); err != nil {
		return nil, customErr.ErrInternalServer
	}

	// 5. Get user detail to generate new JWT
	// 5. Truy vấn thông tin chi tiết của user dựa trên UserID liên kết với token
	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// 6. Generate Access Token & New Refresh Token (Rotation)
	// 6. Sinh ra Access Token mới (thời gian sống ngắn, ví dụ 15 phút) và Refresh Token mới hoàn toàn (thời gian sống dài, ví dụ 7 ngày).
	newAccessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role, 15*time.Minute)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	newRefreshTokenString, err := auth.GenerateToken(user.ID, user.Email, user.Role, 7*24*time.Hour)
	if err != nil {
		return nil, customErr.ErrInternalServer
	}

	// 7. Save new Refresh Token to Database
	// 7. Đóng gói và lưu bản ghi Refresh Token mới vào database với trạng thái ban đầu là Revoked = false.
	newRT := &model.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}
	if err := s.rtRepo.Create(ctx, newRT); err != nil {
		return nil, customErr.ErrInternalServer
	}

	return &model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
	}, nil
}

func (s *userService) GetAllUsers(ctx context.Context, params model.PaginationParams) ([]*model.User, *model.PaginationMeta, error) {
	// Validate pagination parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	users, total, err := s.userRepo.GetAll(ctx, params)
	if err != nil {
		logger.Log.Error("Failed to fetch all users", "error", err)
		return nil, nil, customErr.ErrInternalServer
	}

	totalPage := int(math.Ceil(float64(total) / float64(params.Limit)))
	if totalPage == 0 {
		totalPage = 1
	}

	meta := &model.PaginationMeta{
		Page:      params.Page,
		Limit:     params.Limit,
		Total:     total,
		TotalPage: totalPage,
	}

	return users, meta, nil
}
