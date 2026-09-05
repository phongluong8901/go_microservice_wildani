package main

import (
	"os"

	_ "github.com/bashocode/gowallet/monolith/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
	ledgerRepository "github.com/bashocode/gowallet/monolith/internal/ledger/repository"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/middleware"
	txHandler "github.com/bashocode/gowallet/monolith/internal/transaction/handler"
	txRepository "github.com/bashocode/gowallet/monolith/internal/transaction/repository"
	txService "github.com/bashocode/gowallet/monolith/internal/transaction/service"
	userHandler "github.com/bashocode/gowallet/monolith/internal/user/handler"
	userRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userServer "github.com/bashocode/gowallet/monolith/internal/user/service"
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
	db, err := database.ConnectWithRetry(cfg.DBSN)
	if err != nil {
		logger.Log.Error("Critical Error: Could not connect to database after retries", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// connect to redis
	rdb, err := database.ConnectRedis(cfg.RedisAddr)
	if err != nil {
		logger.Log.Error("Critical Error: Could not connect to Redis", "error", err)
	}
	defer rdb.Close()

	//1. initiate layer
	uRepo := userRepository.NewMySqlUserRepository(db)
	wRepo := walletRepository.NewMySQLWalletRepository(db)
	lRepo := ledgerRepository.NewMysqlLedgerRepository(db)
	tRepo := txRepository.NewMySQLTransactionRepository(db)

	//inject db to user service for transaction
	uSvc := userServer.NewUserService(db, uRepo, wRepo)
	wSvc := walletService.NewWalletService(wRepo, rdb)
	tSvc := txService.NewTransactionService(db, rdb, tRepo, uRepo, wRepo, lRepo)

	// handler layer
	uHandler := userHandler.NewUserHandler(uSvc)
	wHandler := walletHandler.NewWalletHandler(wSvc)
	tHandler := txHandler.NewTransactionHandler(tSvc)

	//2. Setup gin router
	// r := gin.Default()
	r := gin.New()
	r.Use(gin.Recovery()) // recover from panic, return 500 status

	//Register global error handling middlware
	r.Use(middleware.ErrorHandler())

	//register the swagger api
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Route grouping
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/users/register", uHandler.Register)
		v1.POST("/users/login", uHandler.Login)

		// Protected routes (requires valid JWT token)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/users/me", uHandler.GetProfileMe)
			protected.POST("/users/avatar", uHandler.UploadAvatar)
			protected.PUT("/users/:id", uHandler.UpdateProfile)
			protected.GET("/users/:id", uHandler.GetProfile) //use redis (rq1: 41ms, rq2: 3ms)
			protected.DELETE("/users/me", uHandler.DeleteAccount)

			protected.GET("/wallets/me", wHandler.GetMyWallet)

			protected.POST("/transactions/transfer", tHandler.Transfer)
			protected.GET("/transactions/history", tHandler.GetHistory)
		}
	}

	// start server
	logger.Log.Info("Server running on Port 8080...")
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
		os.Exit(1)
	}
}
