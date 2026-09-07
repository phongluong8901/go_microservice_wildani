package model

import "time"

type RefreshToken struct {
	ID        string    `json:"id"`         // Mã định danh duy nhất của bản ghi token (thường dùng UUID).
	UserID    string    `json:"user_id"`    // Mã định danh người dùng (Foreign Key liên kết tới bảng users).
	Token     string    `json:"token"`      // Chuỗi giá trị thực tế của Refresh Token.
	ExpiresAt time.Time `json:"expires_at"` // Thời điểm mốc thời gian token chính thức hết hạn.
	Revoked   bool      `json:"revoked"`    // Cờ trạng thái: true nếu token đã bị thu hồi hoặc đã được sử dụng (phục vụ Reuse Detection).
	CreatedAt time.Time `json:"created_at"` // Mốc thời điểm bản ghi token được khởi tạo trong hệ thống.
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
