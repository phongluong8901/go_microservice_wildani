package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/repository"
	userRepo "github.com/bashocode/gowallet/monolith/internal/user/repository"
	walletModel "github.com/bashocode/gowallet/monolith/internal/wallet/model"
	walletRepo "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/google/uuid"
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
}

// Struct ẩn chứa dependency để giao tiếp với cơ sở dữ liệu.
type userService struct {
	db         *sql.DB
	userRepo   userRepo.UserRepository
	walletRepo walletRepo.WalletRepository
}

// Hàm khởi tạo trả về interface UserService, áp dụng mô hình Dependency Injection.
func NewUserService(db *sql.DB, uRepo repository.UserRepository, wRepo walletRepo.WalletRepository) UserService {
	return &userService{
		db:         db,
		userRepo:   uRepo,
		walletRepo: wRepo,
	}
}

// (Đăng ký tài khoản
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	//1. check if the email already registered
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		// return custom AppError
		return nil, customError.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "this emial already registed")
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
