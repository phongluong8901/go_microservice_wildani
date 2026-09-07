package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	"github.com/bashocode/gowallet/monolith/internal/email"
	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/otp/generator"
	otpModel "github.com/bashocode/gowallet/monolith/internal/otp/model"
	otpRepository "github.com/bashocode/gowallet/monolith/internal/otp/repository"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/repository"
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
	GetGoogleLoginURL() string
	HandleGoogleCallback(ctx context.Context, code string) (*model.LoginResponse, error)
}

// Struct ẩn chứa dependency để giao tiếp với cơ sở dữ liệu.
type userService struct {
	db          *sql.DB
	rdb         *redis.Client
	userRepo    userRepo.UserRepository
	walletRepo  walletRepo.WalletRepository
	otpRepo     otpRepository.OTPRepository
	emailSender email.EmailSender
}

// Hàm khởi tạo trả về interface UserService, áp dụng mô hình Dependency Injection.
func NewUserService(db *sql.DB, rdb *redis.Client, uRepo repository.UserRepository, wRepo walletRepo.WalletRepository, otpRepo otpRepository.OTPRepository, emailSender email.EmailSender) UserService {
	return &userService{
		db:          db,
		rdb:         rdb,
		userRepo:    uRepo,
		walletRepo:  wRepo,
		otpRepo:     otpRepo,
		emailSender: emailSender,
	}
}

// (Đăng ký tài khoản
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	//1. check if the email already registered
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		// return custom AppError
		return nil, customError.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "this email already registered.")
	}

	//hahs the password with bcrypt
	hashsedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, customError.ErrInternalServer
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
		return nil, customError.ErrInternalServer
	}

	//we should rollback if anything error or panic in the middleware
	//save to 2 table, users and wallet, if open of them lets say wallet if failed..
	defer tx.Rollback()

	//store user to db with a tx connection
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		return nil, customError.ErrInternalServer
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
		return nil, customError.ErrInternalServer
	}

	// commit the transaction if all of the step is success
	if err := tx.Commit(); err != nil {
		return nil, customError.ErrInternalServer
	}

	// // Generate OTP
	// // Tạo số ngẫu nhiên an toàn mật mã từ rand.Reader để làm mã OTP gồm 6 chữ số.
	// n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	// if err != nil {
	// 	return nil, customError.ErrInternalServer
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
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	return u, nil
}

// Cập nhật thông tin
func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	//Kiểm tra xem user cần cập nhật có tồn tại hay không.
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	//Thay đổi tên mới theo dữ liệu client gửi lên
	user.FullName = req.FullName
	//Lưu thay đổi xuống database và trả về thông tin mới nhất của user.
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, customError.ErrInternalServer
	}
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// find by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password.")
	}

	// verify the hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password.")
	}

	// generate access token 15 minutes
	accessToken, err := auth.GenerateToken(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	// generate refresh token 7 days
	refreshToken, err := auth.GenerateToken(user.ID, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, customError.ErrInternalServer
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
		return customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	if err := s.userRepo.SoftDelete(ctx, user.ID); err != nil {
		return customError.ErrInternalServer
	}

	return nil
}

// Logout thực hiện chức năng đăng xuất bằng cách đưa JWT token hiện tại vào danh sách đen (blacklist) trên Redis.
func (s *userService) Logout(ctx context.Context, tokenString string) error {
	// validate token
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return customError.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid or expired.")
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
		return customError.ErrInternalServer
	}

	return nil
}

// / VerifyEmail xử lý logic nghiệp vụ xác thực email của người dùng thông qua mã OTP, đảm bảo tính toàn vẹn dữ liệu bằng Database Transaction
func (s *userService) VerifyEmail(ctx context.Context, userID string, code string) error {
	// 1. Get active OTP
	otp, err := s.otpRepo.GetActiveOTP(ctx, userID, code, "email_verification")
	if err != nil {
		// Custom AppError: OTP not found or expired
		return customError.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 2. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return customError.ErrInternalServer
	}
	defer tx.Rollback()

	// 3. Mark user as verified
	if err := s.userRepo.UpdateVerificationStatusTx(ctx, tx, userID, true); err != nil {
		return customError.ErrInternalServer
	}

	// 4. Mark OTP as used
	if err := s.otpRepo.MarkAsUsedTx(ctx, tx, otp.ID); err != nil {
		return customError.ErrInternalServer
	}

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		return customError.ErrInternalServer
	}

	return nil
}

// GenerateAndSendOTP tạo mã OTP ngẫu nhiên an toàn, lưu trữ vào cơ sở dữ liệu và kích hoạt tiến trình nền gửi email thông báo cho người dùng tùy theo loại nghiệp vụ
func (s *userService) GenerateAndSendOTP(ctx context.Context, userID string, emailAddr string, otpType string) error {
	// Gọi generator tạo một mã OTP dạng chuỗi gồm 6 chữ số an toàn mật mã
	otpCode, err := generator.GenerateOTP(6)
	if err != nil {
		return customError.ErrInternalServer
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
		return customError.ErrInternalServer
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
		return customError.ErrInternalServer
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
		return "", customError.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 2. Get active OTP // 2. Truy vấn lấy mã OTP hoạt động (chưa hết hạn và chưa dùng) dành riêng cho việc reset password.
	otp, err := s.otpRepo.GetActiveOTP(ctx, user.ID, code, "password_reset")
	if err != nil {
		return "", customError.NewAppError(http.StatusBadRequest, "INVALID_OTP", "invalid or expired verification code.")
	}

	// 3. Mark OTP as used  // 3. Đánh dấu mã OTP này đã được sử dụng (used = 1) để tránh việc tái sử dụng lại mã cũ.
	if err := s.otpRepo.MarkAsUsed(ctx, otp.ID); err != nil {
		return "", customError.ErrInternalServer
	}

	return user.ID, nil
}

// ResetPassword mã hóa mật khẩu mới của người dùng bằng thuật toán bcrypt và cập nhật vào cơ sở dữ liệu.
func (s *userService) ResetPassword(ctx context.Context, id string, newPassword string) error {
	// hash password using bcrypt // Mã hóa chuỗi mật khẩu mới thành chuỗi hash sử dụng bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return customError.ErrInternalServer
	}

	// update password // Cập nhật mật khẩu mới đã mã hóa vào cơ sở dữ liệu cho người dùng tương ứng.
	if err := s.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return customError.ErrInternalServer
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

func (s *userService) GetGoogleLoginURL() string {
	oauthStateString := "random-state-string" // In production, use a secure random token stored in Redis/session to prevent CSRF
	return s.getOAuthConfig().AuthCodeURL(oauthStateString)
}

func (s *userService) HandleGoogleCallback(ctx context.Context, code string) (*model.LoginResponse, error) {
	config := s.getOAuthConfig()

	// 1. Exchange the auth code for a token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. Fetch userinfo using access token
	client := config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// 3. Find or Create user in database
	var user *model.User
	user, err = s.userRepo.GetByOAuth(ctx, "google", googleUser.ID)
	if err != nil {
		// If user not found, register new user
		if err.Error() == "user not found" {
			// Check if email already registered via normal email/password
			existingUser, _ := s.userRepo.GetByEmail(ctx, googleUser.Email)
			if existingUser != nil {
				return nil, customError.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "This email is registered using password credentials. Please sign in with email and password.")
			}

			// Begin database transaction to guarantee user & wallet creation consistency
			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, customError.ErrInternalServer
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

			// Save user
			if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
				return nil, customError.ErrInternalServer
			}

			// Create wallet for the new user
			wallet := &walletModel.Wallet{
				ID:       uuid.New().String(),
				UserID:   user.ID,
				Balance:  decimal.NewFromInt(0),
				Currency: "IDR",
				Status:   "active",
				Version:  1,
			}
			if err := s.walletRepo.CreateTx(ctx, tx, wallet); err != nil {
				return nil, customError.ErrInternalServer
			}

			if err := tx.Commit(); err != nil {
				return nil, customError.ErrInternalServer
			}
		} else {
			return nil, customError.ErrInternalServer
		}
	}

	// 4. Generate JWT token
	accessToken, err := auth.GenerateToken(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	refreshToken, err := auth.GenerateToken(user.ID, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	err = s.rdb.Set(ctx, "refresh_token:"+user.ID, refreshToken, 7*24*time.Hour).Err()
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
