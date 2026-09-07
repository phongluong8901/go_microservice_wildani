package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cho phép mọi domain (*) gọi tới API
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// Cho phép gửi kèm thông tin xác thực/cookie
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Liệt kê các HTTP header được phép gửi lên
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Correlation-ID")
		// Giới hạn các phương thức RESTful HTTP được gọi
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
