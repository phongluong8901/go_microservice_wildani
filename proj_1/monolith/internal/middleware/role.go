package middleware

import (
	"net/http"

	customErr "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/gin-gonic/gin"
)

// RequireRole tạo một Gin Middleware kiểm tra phân quyền RBAC dựa trên danh sách các vai trò được phép truy cập (allowedRoles).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the role from AuthMiddleware
		// 1. Lấy thông tin 'role' của người dùng đã được lưu vào context từ AuthMiddleware chạy trước đó.
		userRole, exists := c.Get("role")

		if !exists {
			c.Error(customErr.NewAppError(http.StatusForbidden, "ACCESS_DENIED", "You don't have access to this api."))
			c.Abort()
			return
		}

		// Check whether user role is registered in allowedRoles
		// 3. Ép kiểu giá trị role lấy từ context sang chuỗi (string) để so sánh.
		roleStr := userRole.(string)
		isAllowed := false
		// 4. Duyệt qua mảng các vai trò được phép (allowedRoles)
		for _, role := range allowedRoles {
			if roleStr == role {
				isAllowed = true
				break
			}
		}
		// 5. Nếu role của user không nằm trong danh sách được phép, từ chối truy cập.
		if !isAllowed {
			c.Error(customErr.NewAppError(http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You don't have permission to this api."))
			c.Abort()
			return
		}

		c.Next()
	}
}
