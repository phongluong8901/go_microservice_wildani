package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	"github.com/bashocode/gowallet/monolith/internal/email"
	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/logger"
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
)

// Khai báo các nghiệp vụ người dùng có thể thực hiện
type UserService interface {
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	UpdateAvatar(ctx context.Context, id string, path string) error
	DeleteAccount(ctx context.Context, id string) error
	Logout(ctx context.Context, tokenString string) error
	VerifyEmail(ctx context.Context, userID string, code string) error
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

	// Generate OTP
	// Tạo số ngẫu nhiên an toàn mật mã từ rand.Reader để làm mã OTP gồm 6 chữ số.
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, customError.ErrInternalServer
	}
	otpCode := fmt.Sprintf("%06d", n.Int64())
	fmt.Println("otp codes", otpCode)

	// Khởi tạo đối tượng struct OTP chuẩn bị lưu trữ vào cơ sở dữ liệu.
	otpModel := &otpModel.OTP{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Code:      otpCode,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// save to db
	if err := s.otpRepo.Create(ctx, otpModel); err != nil {
		logger.Log.Error("failed to save otp", "error", err)
	}

	// Chạy tiến trình nền (goroutine) để gửi email chứa mã OTP mà không làm nghẽn/chậm response trả về cho client.
	go func() {
		// Tạo context mới độc lập với timeout là 10 giây cho tác vụ gửi email bất đồng bộ.
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		subject := "GoWallet - Verify Your Email"
		body := fmt.Sprintf("Hello %s,\n\nYour verification code is %s\n\nThis code will expire in 15 minutes.\n\nThank you!", user.FullName, otpCode)

		// Gọi service gửi email đi thông qua giao diện SMTPEmailSender đã cấu hình.
		s.emailSender.SendEmail(bgCtx, user.Email, subject, body)
	}()

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
