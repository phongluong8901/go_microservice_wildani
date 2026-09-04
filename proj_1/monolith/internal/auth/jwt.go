package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Lấy khóa bí mật ký token
func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		//only for local
		return []byte("fallback-local-development-secret-key")
	}

	return []byte(secret)
}

// Tạo JWT mới
func GenerateToken(userID string, email string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),       //Thiết lập thời điểm token hết hạn
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     //Thiết lập thời điểm token bắt đầu được phát hành
			ID:        userID + "-" + time.Now().Format("20060102150405"), //Tạo ID duy nhất cho token
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //Tạo mới JWT và gán các claim vào đó

	//Ký token bằng khóa bí mật
	return token.SignedString(getSecretKey())
}

// Xác thực JWT
func ValidateToken(tokenString string) (*JWTClaims, error) {
	//Phân tích chuỗi token và xác thực chữ ký
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
