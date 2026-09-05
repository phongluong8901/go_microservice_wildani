package model

import "time"

type User struct {
	ID           string     `json:"id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // don't expose hash password to json //Dấu "-" bảo vệ an toàn bằng cách ẩn trường này đi, không bao giờ trả về password hash trong phản hồi JSON.
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	IsVerified   bool       `json:"is_verified"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateUserRequest struct {
	FullName string `json:"full_name" binding:"required" example:"test"`             //Họ tên bắt buộc phải có
	Email    string `json:"email" binding:"required,email" example:"test@gmail.com"` //Email bắt buộc và phải đúng định dạng email chuẩn.
	Password string `json:"password" binding:"required,min=6" example:"12346"`       //Password bắt buộc và phải có ít nhất 6 ký tự.
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" biding:"required"` //Họ tên bắt buộc phải có khi update
}
