package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	customError "github.com/bashocode/gowallet/monolith/internal/errors"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/wallet/model"
	"github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/redis/go-redis/v9"
)

type WalletService interface {
	GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error)
}

type walletService struct {
	repo repository.WalletRepository
	rdb  *redis.Client
}

func NewWalletService(repo repository.WalletRepository, rdb *redis.Client) WalletService {
	return &walletService{
		repo: repo,
		rdb:  rdb,
	}
}

// GetWalletByUserID thực hiện lấy thông tin ví của người dùng, áp dụng chiến lược Cache-Aside kết hợp Redis và MySQL để tối ưu hiệu năng
func (s *walletService) GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	// Tạo khóa (cache key) định danh trên Redis dựa theo user_id.
	cacheKey := fmt.Sprintf("wallet:user:%s", userID)

	// check if data exist in redis
	// Truy vấn dữ liệu từ Redis xem ví đã được cache lại hay chưa
	cachedVal, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if err != redis.Nil {
			// redis is down or has an issues, don't fail the request
			logger.Warn(
				ctx,
				"Redis error during cache lookup, falling back to MySQL",
				"error", err.Error(), "user_id", userID,
			)
		}
		// cache miss or redis down
		logger.Info(
			ctx,
			"Cache miss for wallet, fetching from MySQL...",
			"user_id", userID,
		)
	} else {
		// cache hit, Deserialize JSON string to model.Wallet struct
		// Nếu tìm thấy dữ liệu trên Redis (Cache hit), khởi tạo struct Wallet trống để chứa dữ liệu
		wallet := &model.Wallet{}
		if err := json.Unmarshal([]byte(cachedVal), wallet); err == nil {
			logger.Info(
				ctx,
				"Cache hit for wallet, returning from Redis",
				"user_id", userID,
			)

			return wallet, nil
		}

		// Nếu dữ liệu trong cache bị lỗi định dạng không deserialize
		logger.Warn(
			ctx,
			"Failed to deserialize cached wallet data, falling back to MySQL",
			"error", err.Error(), "user_id", userID,
		)
	}

	// Nếu không có cache hoặc cache lỗi, gọi tầng repository để lấy dữ liệu trực tiếp từ cơ sở dữ liệu MySQL.
	wallet, err := s.repo.GetByUserID(ctx, userID)

	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "Wallet Not Found", "wallet not found")
	}

	// save to redis for 5 minutes TTL
	// Chuyển đổi struct wallet mới lấy từ MySQL thành chuỗi JSON (Marshal) để chuẩn bị lưu đệm vào Redis.
	walletBytes, err := json.Marshal(wallet)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, walletBytes, 5*time.Minute)
	}

	return wallet, nil

}
