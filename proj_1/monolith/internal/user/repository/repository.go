package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bashocode/gowallet/monolith/internal/user/model"
)

// Khai báo tập hợp các phương thức mà tầng repository phải cung cấp
// giúp code dễ dàng viết unit test hoặc đổi database sau này.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
	CreateTx(ctx context.Context, tx *sql.Tx, u *model.User) error
	UpdateAvatar(ctx context.Context, id string, path string) error
	SoftDelete(ctx context.Context, id string) error
}

// Struct chứa kết nối cơ sở dữ liệu db *sql.DB
type mysqlUserRepository struct {
	db *sql.DB
}

// Hàm khởi tạo (constructor) trả về một thể hiện cụ thể triển khai interface UserRepository
func NewMySqlUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

// Thêm mới User
func (r *mysqlUserRepository) Create(ctx context.Context, u *model.User) error {
	//Thực thi câu lệnh SQL INSERT để lưu thông tin người dùng
	query := `INSERT INTO users (id, full_name, email, password_hash) VALUES (?, ?, ?, ?)`
	//Sử dụng ExecContext để hỗ trợ truyền context và chống SQL Injection thông qua dấu hỏi chấm ?
	_, err := r.db.ExecContext(ctx, query, u.ID, u.FullName, u.Email, u.PasswordHash)
	return err
}

// Lấy User theo ID
func (r *mysqlUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	//Chọn dữ liệu từ bảng users chỉ lấy những dòng chưa bị xóa (deleted_at IS NULL)
	query := `SELECT id, full_name, email, password_hash, created_at, updated_at, 
		deleted_at FROM users WHERE id = ? AND deleted_at IS NULL`
	u := &model.User{}

	//Ánh xạ các cột trong kết quả trả về của database vào các trường tương ứng của struct u
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.FullName, &u.Email, &u.PasswordHash,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u, nil
}

// Lấy User theo Email
func (r *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	//Chọn dữ liệu từ bảng users chỉ lấy những dòng chưa bị xóa (deleted_at IS NULL)
	query := `SELECT id, full_name, email, password_hash, created_at, updated_at, 
		deleted_at FROM users WHERE email = ? AND deleted_at IS NULL`
	u := &model.User{}

	//Ánh xạ các cột trong kết quả trả về của database vào các trường tương ứng của struct u
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.FullName, &u.Email, &u.PasswordHash,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return u, nil
}

// Cập nhật thông tin User
func (r *mysqlUserRepository) Update(ctx context.Context, u *model.User) error {
	//để cập nhật tên mới cho người dùng dựa theo ID của họ
	query := `UPDATE users SET full_name = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, u.FullName, u.ID)
	return err
}

func (r *mysqlUserRepository) CreateTx(ctx context.Context, tx *sql.Tx, u *model.User) error {
	query := `INSERT INTO users (id, full_name, email, password_hash) VALUES (?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, u.ID, u.FullName, u.Email, u.PasswordHash)
	return err
}

func (r *mysqlUserRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *mysqlUserRepository) UpdateAvatar(ctx context.Context, id string, path string) error {
	query := `UPDATE users SET avatar_url = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, path, id)
	return err
}
