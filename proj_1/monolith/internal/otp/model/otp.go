package model

import "time"

type OTP struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"` // Mã định danh của người dùng sở hữu mã OTP này.
	Code      string    `json:"code"`
	Type      string    `json:"type"`       // email_verification, password_reset
	ExpiresAt time.Time `json:"expires_at"` // Thời điểm mốc thời gian mã OTP hết hiệu lực.
	Used      bool      `json:"used"`       // Trạng thái đánh dấu mã đã được sử dụng chưa (true: rồi, false: chưa).
	CreatedAt time.Time `json:"created_at"`
}
