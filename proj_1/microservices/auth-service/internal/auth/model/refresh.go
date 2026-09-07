package model

import "time"

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"` // Thời điểm chính xác token này hết hiệu lực sử dụng
	Revoked   bool      `json:"revoked"`    // Trạng thái đánh dấu token đã bị thu hồi/vô hiệu hóa hay chưa
	CreatedAt time.Time `json:"created_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
