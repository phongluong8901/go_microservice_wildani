package repository

import (
	"context"
	"database/sql"

	"github.com/bashocode/gowallet/monolith/internal/transaction/model"
)

type TransactionRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
}

type mysqlTransactionRepository struct {
	db *sql.DB
}

func NewMySQLTransactionRepository(db *sql.DB) TransactionRepository {
	return &mysqlTransactionRepository{db: db}
}

// Tạo giao dịch trong Database Transaction
// tx *sql.Tx thay vì dùng trực tiếp r.db, nghĩa là nó bắt buộc phải chạy bên trong một database transaction lớn hơn (ví dụ dùng chung transaction với việc trừ/cộng tiền trong ví) để đảm bảo tính toàn vẹn (ACID).
func (r *mysqlTransactionRepository) CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error {
	//Thực hiện câu lệnh INSERT vào bảng transactions.
	query := `INSERT INTO transactions (id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, t.ID, t.SenderWalletID, t.ReceiverWalletID, t.Amount, t.Description, t.IdempotencyKey, t.Status)
	return err
}

// Truy vấn theo Khóa chống trùng lặp
// Dùng để tìm lại một giao dịch đã tồn tại dựa vào idempotency_key nhằm ngăn chặn việc xử lý trùng lặp request.
func (r *mysqlTransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	query := `SELECT id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status, created_at FROM transactions WHERE idempotency_key = ?`
	t := &model.Transaction{}
	var sender sql.NullString
	//Truy vấn một dòng duy nhất từ cơ sở dữ liệu.
	//.Scan(...): Đọc các cột trả về từ database gán ngược lại vào các trường của struct transaction t.
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&t.ID,
		&sender,
		&t.ReceiverWalletID,
		&t.Amount,
		&t.Description,
		&t.IdempotencyKey,
		&t.Status,
		&t.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if sender.Valid {
		t.SenderWalletID = &sender.String
	}

	return t, nil
}
