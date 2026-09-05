package model

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"phongluong3366@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
