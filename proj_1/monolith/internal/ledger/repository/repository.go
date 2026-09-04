package repository

import (
	"context"
	"database/sql"

	"github.com/bashocode/gowallet/monolith/internal/ledger/model"
	"github.com/shopspring/decimal"
)

type LedgerRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, entry *model.LedgerEntry) error
	GetBalanceByWalletID(ctx context.Context, walletID string) (decimal.Decimal, error)
	GetEntriesByWalletID(ctx context.Context, walletID string) ([]model.LedgerEntry, error)
}

type mysqlLedgerRepository struct {
	db *sql.DB
}

func NewMysqlLedgerRepository(db *sql.DB) LedgerRepository {
	return &mysqlLedgerRepository{db: db}
}

// Ghi bút toán mới trong Transaction
func (r *mysqlLedgerRepository) CreateTx(ctx context.Context, tx *sql.Tx, entry *model.LedgerEntry) error {
	//để ghi nhận một bút toán (cộng hoặc trừ tiền). Hàm nhận vào tx *sql.Tx để cho phép chạy chung transaction với các tiến trình chuyển khoản hoặc nạp rút, đảm bảo tính nguyên tử (Atomicity).
	query := `INSERT INTO ledger_entries (id, wallet_id, transaction_id, entry_type, amount) VALUES (?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, entry.ID, entry.WalletID, entry.TransactionID, entry.EntryType, entry.Amount)
	return err
}

// Tính số dư từ Sổ cái
func (r *mysqlLedgerRepository) GetBalanceByWalletID(ctx context.Context, walletID string) (decimal.Decimal, error) {
	// balance = sum(credit) - sum(debit)
	//Tính toán số dư động dựa trên công thức sổ cái: Tổng tiền ghi có (credit) trừ đi tổng tiền ghi nợ (debit).
	//COALESCE(..., 0): Hàm SQL giúp trả về số 0 nếu kết quả tổng hợp bị trả về NULL (ví dụ trường hợp ví hoàn toàn mới chưa có giao dịch nào).
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN entry_type = 'debit' THEN amount ELSE 0 END), 0)
		FROM ledger_entries
		WHERE wallet_id = ?`

	var balance decimal.Decimal
	//Scan(&balance): Đọc trực tiếp kết quả trả về từ câu lệnh tổng hợp vào biến kiểu decimal.Decimal.
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&balance)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (r *mysqlLedgerRepository) GetEntriesByWalletID(ctx context.Context, walletID string) ([]model.LedgerEntry, error) {
	//Truy vấn danh sách tất cả các bút toán thuộc về một wallet_id cụ thể, sắp xếp theo thứ tự thời gian mới nhất lên trên (ORDER BY created_at DESC)
	query := `SELECT id, wallet_id, transaction_id, entry_type, amount, created_at FROM ledger_entries WHERE wallet_id = ? ORDER BY created_at DESC`

	//Thực hiện truy vấn trả về tập hợp nhiều dòng dữ liệu (result set).
	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	//: Đảm bảo giải phóng tài nguyên rows ngay sau khi hàm thực thi xong để tránh rò rỉ kết nối database.
	defer rows.Close()

	var entries []model.LedgerEntry

	//Vòng lặp duyệt qua từng dòng kết quả, dùng phương thức Scan để ánh xạ dữ liệu vào struct model.LedgerEntry rồi gom vào mảng entries để trả về cho tầng service.
	for rows.Next() {
		var e model.LedgerEntry
		if err := rows.Scan(&e.ID, &e.WalletID, &e.TransactionID, &e.EntryType, &e.Amount, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
