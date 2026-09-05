package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID               string          `json:"id"`
	SenderWalletID   *string         `json:"sender_wallet_id"`   // nullable if top up //ID ví của người gửi.
	ReceiverWalletID string          `json:"receiver_wallet_id"` //ID ví của người nhận.
	Amount           decimal.Decimal `json:"amount"`             //Số tiền giao dịch.
	Description      string          `json:"description"`        //Mô tả giao dịch.
	IdempotencyKey   string          `json:"idempotency_key"`    //Khóa để ngăn chặn giao dịch trùng lặp.
	Status           string          `json:"status"`             //Trạng thái giao dịch.
	CreatedAt        time.Time       `json:"created_at"`         //Thời gian tạo giao dịch.
}

// Cấu trúc TransferRequest (Dữ liệu đầu vào khi chuyển khoản)
type TransferRequest struct {
	ReceiverEmail  string          `json:"receiver_email" binding:"required,email" example:"receiver@example.com"` //Email người nhận, có validate bắt buộc (required) và đúng định dạng email (email)
	Amount         decimal.Decimal `json:"amount" binding:"required" example:"50000"`                              //Số tiền chuyển, bắt buộc và phải lớn hơn 0 (gt=0).
	Description    string          `json:"description" example:"Dinner split"`                                     //Mô tả giao dịch, có thể để trống.
	IdempotencyKey string          `json:"idempotency_key" binding:"required" example:"unique-uuid-key-123"`       //Khóa để ngăn chặn giao dịch trùng lặp, bắt buộc.
}

// Cấu trúc TopUpRequest (Dữ liệu đầu vào khi nạp tiền)
type TopUpRequest struct {
	Amount         decimal.Decimal `json:"amount" binding:"required" example:"100000"`                       //Số tiền nạp, bắt buộc và phải lớn hơn 0 (gt=0).
	IdempotencyKey string          `json:"idempotency_key" binding:"required" example:"unique-uuid-key-abc"` //Khóa để ngăn chặn giao dịch trùng lặp, bắt buộc.
}

// Cấu trúc ExternalTransferRequest (Yêu cầu chuyển tiền qua hệ thống ngoài)
type ExternalTransferRequest struct {
	TransferID     string          `json:"transfer_id" binding:"required" example:"external-transfer-id"`          //Mã giao dịch bên ngoài hệ thống
	ReceiverEmail  string          `json:"receiver_email" binding:"required,email" example:"receiver@example.com"` //Email người nhận, có validate bắt buộc (required) và đúng định dạng email (email)
	Amount         decimal.Decimal `json:"amount" binding:"required" example:"50000"`                              //Số tiền chuyển, bắt buộc và phải lớn hơn 0 (gt=0).
	Currency       string          `json:"currency" example:"IDR"`                                                 //Loại tiền tệ
	IdempotencyKey string          `json:"idempotency_key" binding:"required" example:"unique-uuid-key-123"`       //Khóa để ngăn chặn giao dịch trùng lặp, bắt buộc.
	SenderUserID   string          `json:"sender_user_id" binding:"required" example:"gowallet-user-id"`           //ID người gửi
	CallbackURL    string          `json:"callback_url,omitempty"`
}

// Cấu trúc ExternalTransferStatus (Trạng thái giao dịch chuyển tiền qua hệ thống ngoài)
type ExternalTransferStatus struct {
	TransferID     string `json:"transfer_id"`     //Mã giao dịch bên ngoài hệ thống
	Status         string `json:"status"`          //Trạng thái giao dịch: pending, success, failed, timeout
	IdempotencyKey string `json:"idempotency_key"` //Khóa để ngăn chặn giao dịch trùng lặp, bắt buộc.
}
