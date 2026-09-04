package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type LedgerEntry struct {
	ID            string          `json:"id"`
	WalletID      string          `json:"wallet_id"`      //Mã định danh chiếc ví chịu tác động của bút toán này.
	TransactionID string          `json:"transaction_id"` //Mã giao dịch gốc liên kết (giúp gom nhóm các bút toán thuộc cùng một giao dịch chuyển tiền hoặc nạp/rút).
	EntryType     string          `json:"entry_type"`     // credit + or debit -
	Amount        decimal.Decimal `json:"amount"`
	CreatedAt     time.Time       `json:"created_at"`
}
