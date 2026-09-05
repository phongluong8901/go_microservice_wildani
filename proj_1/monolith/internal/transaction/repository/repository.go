package repository

import (
	"context"
	"database/sql"

	"github.com/bashocode/gowallet/monolith/internal/transaction/model"
)

type TransactionRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	GetHistory(ctx context.Context, walletID string, params model.PaginationParams) ([]model.Transaction, int64, error)
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

// GetHistory thực hiện truy vấn lịch sử giao dịch của một ví
func (r *mysqlTransactionRepository) GetHistory(ctx context.Context, walletID string, params model.PaginationParams) ([]model.Transaction, int64, error) {
	// counting total data for pagination meta
	countQuery := `SELECT COUNT(*) FROM transactions WHERE (sender_wallet_id) = ? OR receiver_wallet_id = ?`
	var total int64
	var err error

	// Nếu có truyền thêm bộ lọc trạng thái (status), nối thêm điều kiện vào câu lệnh đếm.
	if params.Status != "" {
		countQuery += " AND status = ?"
		// Thực thi câu lệnh đếm với tham số trạng thái kèm theo.
		err = r.db.QueryRowContext(ctx, countQuery, walletID, walletID, params.Status).Scan(&total)
	} else {
		// Thực thi câu lệnh đếm không có tham số trạng thái.
		err = r.db.QueryRowContext(ctx, countQuery, walletID, walletID).Scan(&total)
	}

	if err != nil {
		return nil, 0, err
	}

	// get the paginated data, use sort and order
	// important, use whitelist for sort and order to prevent sql injection
	sortColumn := "created_at"
	if params.Sort == "amount" {
		sortColumn = "amount"
	}

	// Thiết lập thứ tự sắp xếp mặc định là giảm dần (DESC).
	sortOrder := "DESC"
	if params.Order == "asc" {
		sortOrder = "ASC"
	}

	// Khai báo câu lệnh SQL lấy dữ liệu phân trang cho các giao dịch liên quan đến ví.
	query := `SELECT id, sender_wallet_id, receiver_wallet_id,
				amount, description, idempotency_key, status, created_at
			FROM transactions WHERE (sender_wallet_id = ? OR
			receiver_wallet_id = ?)`

	var rows *sql.Rows
	if params.Status != "" {
		query += " AND status = ? ORDER BY " + sortColumn + " " + sortOrder + " LIMIT ? OFFSET ?"
		rows, err = r.db.QueryContext(ctx, query, walletID, walletID, params.Status, params.Limit, params.Offset())
	} else {
		query += " ORDER BY " + sortColumn + " " + sortOrder + " LIMIT ? OFFSET ?"
		rows, err = r.db.QueryContext(ctx, query, walletID, walletID, params.Limit, params.Offset())
	}

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var txs []model.Transaction
	for rows.Next() {
		var t model.Transaction
		var sender sql.NullString
		err := rows.Scan(
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
			return nil, 0, err
		}
		if sender.Valid {
			t.SenderWalletID = &sender.String
		}
		txs = append(txs, t)
	}

	return txs, total, nil
}
