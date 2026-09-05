package repository

import (
	"context"
	"database/sql"

	"github.com/bashocode/gowallet/monolith/internal/user/model"
	"github.com/stretchr/testify/mock"
)

// type MockUserRepository struct { mock.Mock }: Định nghĩa struct mô phỏng, nhúng mock.Mock để kế thừa toàn bộ các tính năng tracking và cấu hình kết quả giả lập từ thư viện Testify.
type MockUserRepository struct {
	mock.Mock
}

// : Hàm giả lập tạo mới user, nhận context và thông tin user.
func (m *MockUserRepository) Create(ctx context.Context, u *model.User) error {
	//Ghi lại các tham số đầu vào vừa được gọi để kiểm tra trong unit test.
	args := m.Called(ctx, u)
	//Trả về giá trị lỗi (error) đã được người viết test cấu hình sẵn từ trước cho hàm Create.
	return args.Error(0)
}

// Hàm giả lập tạo user chạy trong transaction, nhận thêm tham số tx *sql.Tx.
func (m *MockUserRepository) CreateTx(ctx context.Context, tx *sql.Tx, u *model.User) error {
	args := m.Called(ctx, tx, u)
	return args.Error(0)
}

// Hàm giả lập tìm kiếm user theo ID, trả về con trỏ user và lỗi.
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// Hàm giả lập tìm kiếm user theo email, trả về con trỏ user và lỗi.
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *model.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, id string, path string) error {
	args := m.Called(ctx, id, path)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
