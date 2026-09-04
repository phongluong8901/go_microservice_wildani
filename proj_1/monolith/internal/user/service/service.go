package service

import (
	"context"
	"net/http"

	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/bashocode/gowallet/monolith/internal/user/repository"
	"github.com/google/uuid"
)

// Khai báo các nghiệp vụ người dùng có thể thực hiện
type UserService interface {
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
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

	//2. create new user object
	user := &model.User{
		ID:           uuid.New().String(), //Khởi tạo một mã UUID v4 ngẫu nhiên làm khóa chính cho user mới.
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: req.Password,
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
