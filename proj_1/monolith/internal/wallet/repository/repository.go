package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bashocode/gowallet/monolith/internal/wallet/model"
)

// Định nghĩa các phương thức thao tác database cho v
type WalletRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error
	GetByUserID(ctx context.Context, userID string) (*model.Wallet, error)
}

// Struct lưu giữ kết nối db *sql.DB để thực thi câu lệnh SQL.
type mysqlWalletRepository struct {
	db *sql.DB
}

func NewMySQLWalletRepository(db *sql.DB) WalletRepository {
	return &mysqlWalletRepository{db: db}
}

// Tạo ví trong Transaction
func (r *mysqlWalletRepository) CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error {
	//Thực thi câu lệnh INSERT để tạo bản ghi ví mới. Điểm đặc biệt là hàm này nhận vào tx *sql.Tx thay vì dùng r.db trực tiếp, giúp gộp thao tác tạo ví chung một transaction với các tiến trình khác (ví dụ: tạo ví tự động ngay khi đăng ký tài khoản user) nhằm đảm bảo tính nguyên tử (Atomicity).
	query := "INSERT INTO wallets (id, user_id, balance, currency, status) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, w.ID, w.UserID, w.Balance, w.Currency, w.Status)
	return err
}

// Lấy thông tin ví theo User ID
func (r *mysqlWalletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	//Truy vấn thông tin chi tiết của ví từ bảng wallets dựa vào user_id với điều kiện deleted_at IS NULL.
	query := `SELECT id, user_id, balance, currency, status, version, created_at, updated_at FROM wallets WHERE user_id = ? AND deleted_at IS NULL`
	w := &model.Wallet{}
	//Scan(...): Đổ dữ liệu từ các cột trong database vào các trường của struct model.Wallet (bao gồm cả trường version dùng cho xử lý đồng thời).
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&w.ID,
		&w.UserID,
		&w.Balance,
		&w.Currency,
		&w.Status,
		&w.Version,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		//Xử lý trường hợp không tìm thấy ví, trả về lỗi "Wallet not found" thân thiện hơn.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Wallet not found")
		}
		return nil, err
	}
	return w, nil
}
