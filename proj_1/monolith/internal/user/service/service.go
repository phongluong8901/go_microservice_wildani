package service

import (
	"context"
	"net/http"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Khai báo các nghiệp vụ người dùng có thể thực hiện
type UserService interface {
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
}

// Struct ẩn chứa dependency để giao tiếp với cơ sở dữ liệu.
type userService struct {
	repo repository.UserRepository
}

// Hàm khởi tạo trả về interface UserService, áp dụng mô hình Dependency Injection.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// (Đăng ký tài khoản
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	//1. check if the email already registered
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
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

	//3. store it to the database
	//Lưu user xuống database, sau đó truy vấn lại để lấy đầy đủ thông tin vừa tạo
	if err := s.repo.Create(ctx, user); err != nil {
		//return internal server error
		return nil, customError.ErrInternalServer
	}

	return s.repo.GetByID(ctx, user.ID)

}

// Lấy thông tin
func (s *userService) GetProfile(ctx context.Context, id string) (*model.User, error) {
	//Chuyển tiếp yêu cầu lấy thông tin trực tiếp xuống tầng repository thông qua GetByID
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	return u, nil
}

// Cập nhật thông tin
func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	//Kiểm tra xem user cần cập nhật có tồn tại hay không.
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	//Thay đổi tên mới theo dữ liệu client gửi lên
	user.FullName = req.FullName
	//Lưu thay đổi xuống database và trả về thông tin mới nhất của user.
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, customError.ErrInternalServer
	}
	return s.repo.GetByID(ctx, id)
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// find by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
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
