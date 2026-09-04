package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Balance   decimal.Decimal `json:"balance"`  //Số dư hiện tại trong ví
	Currency  string          `json:"currency"` //Loại tiền tệ của ví (ví dụ: "USD", "VND").
	Status    string          `json:"status"`   // active, frozen
	Version   int             `json:"-"`        // used for concurrency control
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type WalletInquiry struct {
	Valid     bool   `json:"valid"`
	AccountID string `json:"account_id,omitempty"` //Mã tài khoản/ví, thuộc tính omitempty giúp tự động ẩn trường này khỏi kết quả JSON nếu nó rỗng hoặc không có giá trị.
	Name      string `json:"name,omitempty"`       //Họ và tên của chủ sở hữu ví.//Tên của chủ tài khoản được tra cứu (ẩn đi nếu không có).
	Email     string `json:"email,omitempty"`
}

type EmailInquiryRequest struct {
	Email string `json:"email" binding:"required,email" example:"recipient@example.com"` //đúng định dạng email (email),
}
