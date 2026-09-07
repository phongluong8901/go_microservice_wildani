package model

import "time"

type User struct {
	ID           string     `json:"id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	Role         string     `json:"role"` // Phân quyền của người dùng trong hệ thống (ví dụ: admin, user)
	PasswordHash string     `json:"-"`    // Chuỗi mật khẩu đã được mã hóa băm (dấu "-" giúp ẩn trường này khỏi kết quả JSON trả về cho client)
	AvatarURL    *string    `json:"avatar_url"`
	IsVerified   bool       `json:"is_verified"` // Cờ đánh dấu tài khoản đã được xác thực email hay chưa
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}
