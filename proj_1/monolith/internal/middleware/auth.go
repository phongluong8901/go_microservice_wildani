package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bashocode/gowallet/monolith/internal/auth"
	customErr "github.com/bashocode/gowallet/monolith/internal/errors"
	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware là hàm nhận vào Redis client và trả về một Gin middleware
func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy giá trị của header "Authorization" từ request
		authHeader := c.GetHeader("Authorization")
		// Nếu header trống (không có token), trả về lỗi 401
		if authHeader == "" {
			c.Error(customErr.NewAppError(http.StatusUnauthorized, "MISSING_TOKEN", "auth token is missing."))
			c.Abort()
			return
		}

		// split Bearer token
		// Cắt chuỗi Authorization theo khoảng trắng để tách "Bearer"
		parts := strings.Split(authHeader, " ")
		// Nếu độ dài không phải là 2 hoặc không bắt đầu bằng "Bearer", trả về lỗi 401
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(customErr.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid, should be Bearer <token>."))
			c.Abort()
			return
		}

		// Lấy ra phần thứ hai chính là chuỗi JWT token.
		tokenString := parts[1]

		// check if token is in redis blacklist
		// Tạo khóa kiểm tra trong Redis dựa trên token (dùng để kiểm soát token đã bị thu hồi/đăng xuất chưa).
		blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
		// Kiểm tra xem token này có đang nằm trong blacklist của Redis không.
		exists, err := rdb.Exists(c.Request.Context(), blacklistKey).Result()
		// Nếu không có lỗi và tồn tại (exists > 0), tức là token đã bị thu hồi.
		if err == nil && exists > 0 {
			c.Error(customError.NewAppError(
				http.StatusUnauthorized,
				"TOKEN_REVOKED",
				"Login session has ended. Please login again.",
			))
			c.Abort()
			return
		}

		// validate token
		// Gọi hàm giải mã và kiểm tra tính hợp lệ của chữ ký token
		claims, err := auth.ValidateToken(tokenString)
		// Nếu token không hợp lệ hoặc đã hết hạn, trả về lỗi 401 và chặn request.
		if err != nil {
			c.Error(customErr.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid or expired."))
			c.Abort()
			return
		}

		// save to context
		// Lưu các thông tin giải mã được từ token (UserID, Email, TokenString) vào context của Gin
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("token_string", tokenString) // store for logout needs

		c.Next()
	}
}
