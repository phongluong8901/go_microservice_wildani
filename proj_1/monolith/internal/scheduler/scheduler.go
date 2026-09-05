package scheduler

import (
	"database/sql"

	ledgerRepo "github.com/bashocode/gowallet/monolith/internal/ledger/repository"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	walletRepo "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron       *cron.Cron
	db         *sql.DB
	walletRepo walletRepo.WalletRepository
	ledgerRepo ledgerRepo.LedgerRepository
}

func NewScheduler(
	db *sql.DB,
	wRepo walletRepo.WalletRepository,
	lRepo ledgerRepo.LedgerRepository,
) *Scheduler {
	c := cron.New(cron.WithSeconds())

	s := &Scheduler{
		cron:       c,
		db:         db,
		walletRepo: wRepo,
		ledgerRepo: lRepo,
	}

	return s
}

func (s *Scheduler) Start() {
	// clean expired OTP tokens every 30 minutes
	s.cron.AddFunc("0 */30 * * * *", s.CleanupExpiredOTPs)
	// daily balance reconciliation at 2 AM
	s.cron.AddFunc("0 0 2 * * *", s.ReconcileAllBalances)
	// clean expired refresh token daily at 3 AM
	s.cron.AddFunc("0 0 3 * * *", s.CleanupExpiredRefreshTokens)
	// export daily transaction to csv at 23.59 PM
	s.cron.AddFunc("0 59 23 * * *", s.ExportDailyTransactions)

	// debug: testing every 10 seconds
	s.cron.AddFunc("*/10 * * * * *", s.TestingOnly)

	s.cron.Start()
	logger.Log.Info("Background scheduler successfully started!")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	logger.Log.Info("Background scheduler stopped.")
}

// Đúng vậy, khi bạn khởi động server và gọi hàm scheduler.Start(), hệ thống sẽ tự động chạy các tác vụ nền (background jobs) theo đúng lịch trình (cron expression) đã đăng ký:

// Mỗi 10 giây (*/10 * * * * *): Chạy hàm TestingOnly (dùng để debug/kiểm tra).

// Mỗi 30 phút (0 */30 * * * *): Xóa các mã OTP đã hết hạn (CleanupExpiredOTPs).

// Lúc 02:00 AM hằng ngày (0 0 2 * * *): Đối soát lại toàn bộ số dư ví (ReconcileAllBalances).

// Lúc 03:00 AM hằng ngày (0 0 3 * * *): Xóa token làm mới (refresh token) đã hết hạn (CleanupExpiredRefreshTokens).

// Lúc 23:59 PM hằng ngày (0 59 23 * * *): Xuất file CSV danh sách giao dịch trong ngày (ExportDailyTransactions).

// Lệnh s.cron.Start() sẽ chạy ngầm bằng goroutine riêng biệt nên sẽ không làm block luồng xử lý HTTP request chính của server. Bạn chỉ cần đảm bảo ở file cmd/main.go có khởi tạo và gọi scheduler.Start() khi ứng dụng boot lên.
