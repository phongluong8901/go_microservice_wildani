package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware(allowedOrigins ...string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		allowedMap[o] = true
	}
	
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		origin := c.Request.Header.Get("Origin")

		if origin != "" && (len(allowedMap) == 0 || allowedMap[origin]) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
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
