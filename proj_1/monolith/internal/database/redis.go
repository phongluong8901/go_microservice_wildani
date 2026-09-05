package database

import (
	"context"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/redis/go-redis/v9"
)

// Định nghĩa hàm ConnectRedis nhận vào địa chỉ (addr) dạng string và trả về con trỏ *redis.Client cùng lỗi (error) nếu có.
func ConnectRedis(addr string) (*redis.Client, error) {
	// Khởi tạo một đối tượng client Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	// check connection
	// Tạo một context có thời gian timeout là 3 giây để tránh việc kết nối bị treo vô thời hạn nếu Redis chết.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Đảm bảo hàm cancel() luôn được gọi khi hàm ConnectRedis kết thúc
	defer cancel()

	// Gửi lệnh "PING" tới server Redis thông qua lệnh PingContext để kiểm tra xem server có thực sự phản hồi không
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logger.Log.Info("Successfully connected to Redis!")
	return rdb, nil
}
