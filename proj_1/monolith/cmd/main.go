package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/bashocode/gowallet/monolith/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
	"github.com/bashocode/gowallet/monolith/internal/email"
	ledgerHandler "github.com/bashocode/gowallet/monolith/internal/ledger/handler"
	ledgerRepository "github.com/bashocode/gowallet/monolith/internal/ledger/repository"
	ledgerService "github.com/bashocode/gowallet/monolith/internal/ledger/service"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/middleware"
	otpRepository "github.com/bashocode/gowallet/monolith/internal/otp/repository"
	"github.com/bashocode/gowallet/monolith/internal/scheduler"
	txHandler "github.com/bashocode/gowallet/monolith/internal/transaction/handler"
	txRepository "github.com/bashocode/gowallet/monolith/internal/transaction/repository"
	txService "github.com/bashocode/gowallet/monolith/internal/transaction/service"
	userHandler "github.com/bashocode/gowallet/monolith/internal/user/handler"
	userRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userService "github.com/bashocode/gowallet/monolith/internal/user/service"
	walletHandler "github.com/bashocode/gowallet/monolith/internal/wallet/handler"
	walletRepository "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	walletService "github.com/bashocode/gowallet/monolith/internal/wallet/service"
	"github.com/gin-gonic/gin"
)

// @title GoWallet Monolith API
// @version 1.0
// @description API Documentation for GoWallet Monolith
// @host localhost:8080
// @termsOfService http://swagger.io/terms/
// @contact.name Phong Luong
// @contact.email phongluong3366@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <your_token>" to authenticate.

func main() {
	//initialize the lo
	logger.InitLogger()
	logger.Log.Info("Starting Monolith Wallet Application...")

	//1. Load configuration
	cfg := config.LoadConfig()

	//2. Connect to database with retry
	db, err := database.ConnectWithRetry(cfg.DBDSN)
	if err != nil {
		logger.Log.Error("Critical Error: Could not connect to database after retries", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// connect to redis
	rdb, err := database.ConnectRedis(cfg.RedisAddr)
	if err != nil {
		logger.Log.Error("Critical Error: Could not connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// initiate email sender & otp repository
	emailSender := email.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	otpRepo := otpRepository.NewMySQLOTPRRepository(db)

	//1. initiate layer
	uRepo := userRepository.NewMySqlUserRepository(db)
	wRepo := walletRepository.NewMySQLWalletRepository(db)
	lRepo := ledgerRepository.NewMysqlLedgerRepository(db)
	tRepo := txRepository.NewMySQLTransactionRepository(db)

	//inject db to user service for transaction
	uSvc := userService.NewUserService(db, rdb, uRepo, wRepo, otpRepo, emailSender)
	wSvc := walletService.NewWalletService(wRepo, rdb)
	tSvc := txService.NewTransactionService(db, rdb, tRepo, uRepo, wRepo, lRepo, cfg.WebhookSecret)
	lSvc := ledgerService.NewLedgerService(lRepo, wRepo)

	// handler layer
	uHandler := userHandler.NewUserHandler(uSvc)
	wHandler := walletHandler.NewWalletHandler(wSvc)
	tHandler := txHandler.NewTransactionHandler(tSvc)
	lHandler := ledgerHandler.NewLedgerHandler(lSvc)

	// start conjob
	cronSched := scheduler.NewScheduler(db, wRepo, lRepo)
	cronSched.Start() //Kích hoạt một tiến trình chạy ngầm (background goroutine). Từ khoảnh khắc này, các mốc thời gian bạn đã cấu hình sẽ bắt đầu đếm ngược và tự động thực thi đúng giờ.

	//2. Setup gin router
	// r := gin.Default()
	r := gin.New()
	r.Use(middleware.CorrelationID())
	r.Use(gin.Recovery()) // recover from panic, return 500 status

	//Register global error handling middlware
	r.Use(middleware.ErrorHandler())

	// apply global rate limiter max 60 request per minutes per ip
	r.Use(middleware.RateLimiter(rdb, 60, time.Minute))

	//register the swagger api
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Route grouping
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/users/register", uHandler.Register)
		v1.POST("/users/login", uHandler.Login)

		v1.POST("/users/forgot-password", uHandler.ForgotPassword)
		v1.POST("/users/verify-password-reset", uHandler.VerifyPasswordReset)

		v1.GET("/auth/google/login", uHandler.GoogleLogin)
		v1.GET("/auth/google/callback", uHandler.GoogleCallback)

		internal := v1.Group("")
		internal.Use(middleware.APIKeyMiddleware(cfg.WebhookSecret))
		{
			internal.POST("/wallets/inquiry", wHandler.EmailInquiry)
			internal.POST("/transfers/external", tHandler.ReceiveExternalTransfer)
			internal.GET("/transfers/external/:id/status", tHandler.GetExternalTransferStatus)
		}

		// Protected routes (requires valid JWT token)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(rdb))
		{
			protected.GET("/users/me", uHandler.GetProfileMe)
			protected.POST("/users/avatar", uHandler.UploadAvatar)
			protected.PUT("/users/:id", uHandler.UpdateProfile)
			protected.GET("/users/:id", uHandler.GetProfile) //use redis (rq1: 41ms, rq2: 3ms)
			protected.DELETE("/users/me", uHandler.DeleteAccount)
			protected.POST("/users/logout", uHandler.Logout)
			protected.POST("/users/verify-email", uHandler.VerifyEmail)

			protected.POST("/auth/refresh", uHandler.RefreshToken)

			protected.GET("/wallets/me", wHandler.GetMyWallet)

			protected.POST("/transactions/transfer", tHandler.Transfer)
			protected.POST("/transactions/topup", tHandler.TopUp)
			protected.GET("/transactions/history", tHandler.GetHistory)

			protected.GET("/ledger/mutations", lHandler.GetMutations)
			protected.GET("/ledger/reconcile", lHandler.Reconcile)

			// only admin that can access
			adminOnly := protected.Group("/admin")
			adminOnly.Use(middleware.RequireRole("admin")) // RBAC Protection
			{
				adminOnly.GET("/users", uHandler.AdminGetUsers)

			}
		}
	}

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	// run server in separate goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server failed to run", "error", err)
		}
	}()

	// start server
	//Graceful Shutdown (Tắt ứng dụng an toàn). Mục đích là để server không bị ngắt đột ngột làm hỏng dữ liệu hoặc cắt đứt ngang các request mà người dùng đang gửi lên.
	logger.Log.Info("Server running on Port 8000...")
	// graceful shutdown - wait for signal from os
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)s
	//Lệnh <-quit có nhiệm vụ chặn (block) luồng chính tại đây. Nhờ có nó, ứng dụng cứ tiếp tục chạy bình thường phục vụ người dùng cho đến khi có tín hiệu dừng được gửi tới.
	<-quit

	logger.Log.Info("Server shutting down gracefully...")

	// give 10 seconds to complet in-flight requests
	//Gọi hàm này thay vì tắt ngang để server ngừng nhận request mới nhưng vẫn cố gắng xử lý cho xong xuôi các request đang chạy dở (in-flight requests) trong giới hạn 10 giây vừa cấp.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", "error", err)

	}

	// stop scheduler after http server shutdown
	//Sau khi HTTP server đã đóng xong, tiến hành dừng luôn các luồng cron job đang chạy ngầm để giải phóng hoàn toàn tài nguyên hệ thống.
	cronSched.Stop()

	logger.Log.Info("Server exited gracefully")

}
