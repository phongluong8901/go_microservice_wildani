package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bashocode/gowallet/monolith/internal/otp/model"
)

type OTPRepository interface {
	Create(ctx context.Context, o *model.OTP) error
	GetActiveOTP(ctx context.Context, userID string, code string, otpType string) (*model.OTP, error)
	MarkAsUsed(ctx context.Context, id string) error
	MarkAsUsedTx(ctx context.Context, tx *sql.Tx, id string) error
}

type mysqlOTPRepository struct {
	db *sql.DB
}

func NewMySQLOTPRRepository(db *sql.DB) OTPRepository {
	return &mysqlOTPRepository{db: db}
}

// Create thực hiện câu lệnh SQL INSERT để lưu một bản ghi mã OTP mới vào bảng otp_codes.
func (r *mysqlOTPRepository) Create(ctx context.Context, o *model.OTP) error {
	query := `INSERT INTO otp_codes
				(id, user_id, code, type, expires_at, used) VALUES
				(?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, o.ID, o.UserID, o.Code, o.Type, o.ExpiresAt, o.Used)
	return err
}

// GetActiveOTP truy vấn lấy mã OTP chưa sử dụng, chưa hết hạn dựa vào user_id, mã code và phân loại type.
func (r *mysqlOTPRepository) GetActiveOTP(ctx context.Context, userID string, code string, otpType string) (*model.OTP, error) {
	query := `SELECT id, user_id, code, type, expires_at, used 
				FROM otp_codes WHERE user_id = ? AND code = ? 
				AND type = ? AND expires_at > NOW() AND used = 0`
	o := &model.OTP{}
	err := r.db.QueryRowContext(ctx, query, userID, code, otpType).
		Scan(&o.ID, &o.UserID, &o.Code, &o.Type, &o.ExpiresAt, &o.Used)
	if err != nil {
		// Nếu không tìm thấy dòng dữ liệu khớp, trả về lỗi chuẩn hóa "active OTP not found or expired".
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("active OTP not found or expired")
		}
		return nil, err
	}
	return o, nil
}

// MarkAsUsed cập nhật trạng thái cột used thành 1 (đã sử dụng) cho mã OTP theo id thông thường (ngoài transaction).
func (r *mysqlOTPRepository) MarkAsUsed(ctx context.Context, id string) error {
	query := `UPDATE otp_codes SET used = 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

// MarkAsUsedTx cập nhật trạng thái used thành 1 (đã sử dụng) cho mã OTP bên trong một transaction có sẵn (tx).
func (r *mysqlOTPRepository) MarkAsUsedTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := `UPDATE otp_codes SET used = 1 WHERE id = ?`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}
