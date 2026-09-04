package main

import (
	"os"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/bashocode/gowallet/monolith/internal/middleware"
	userHandler "github.com/bashocode/gowallet/monolith/internal/user/handler"
	userRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userServer "github.com/bashocode/gowallet/monolith/internal/user/service"
	walletHandler "github.com/bashocode/gowallet/monolith/internal/wallet/handler"
	walletRepository "github.com/bashocode/gowallet/monolith/internal/wallet/repository"
	walletService "github.com/bashocode/gowallet/monolith/internal/wallet/service"
	"github.com/gin-gonic/gin"
)

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

	//1. initiate layer
	uRepo := userRepository.NewMySqlUserRepository(db)
	wRepo := walletRepository.NewMySQLWalletRepository(db)

	//inject db to user service for transaction
	uSvc := userServer.NewUserService(db, uRepo, wRepo)
	wSvc := walletService.NewWalletService(wRepo)

	// handler layer
	uHandler := userHandler.NewUserHandler(uSvc)
	wHandler := walletHandler.NewWalletHandler(wSvc)

	//2. Setup gin router
	// r := gin.Default()
	r := gin.New()
	r.Use(gin.Recovery()) // recover from panic, return 500 status

	//Register global error handling middlware
	r.Use(middleware.ErrorHandler())

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
			protected.GET("/wallets/me", wHandler.GetMyWallet)

		}
	}

	// start server
	logger.Log.Info("Server running on Port 8080...")
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
		os.Exit(1)
	}
}
