package middleware

import (
	"fmt"
	"net/http"
	"time"

	customErr "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter nhận vào Redis client, giới hạn số request (limit) trong một khoảng thời gian (window)
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy địa chỉ IP của client đang thực hiện request gọi tới server
		ip := c.ClientIP()

		// key format: rate_limit:ip_address:minute_timestamp
		// Tạo cấu trúc khóa (key) định danh trên Redis theo IP và mốc thời gian hiện tại chia theo khoảng thời gian cửa sổ (window).
		currentTime := time.Now().Unix() / int64(window.Seconds())
		key := fmt.Sprintf("rate_limit:%s:%d", ip, currentTime)

		// Lấy context hiện tại từ request của Gin.
		ctx := c.Request.Context()

		// use multi/exec to atomically increment counter & set TTL
		// Sử dụng Pipeline của Redis để gom nhiều lệnh chạy gộp chung một lần gọi mạng
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window*2) // store longer for safety margin// Thiết lập thời gian hết hạn gấp đôi khoảng thời gian cửa sổ để đảm bảo an toàn bộ đếm

		// Thực thi chuỗi lệnh trong pipeline gửi lên Redis
		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Error(customErr.ErrInternalServer)
			c.Abort()
			return
		}

		// Lấy giá trị số lượng request hiện tại đã được đếm từ Redis
		count := incr.Val()
		// Kiểm tra nếu số lượng request vượt quá giới hạn cho phép (limit).
		if count > int64(limit) {
			c.Error(customErr.NewAppError(
				http.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many request. Please try again later.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
