package main

import (
	"os"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
	"github.com/bashocode/gowallet/monolith/internal/logger"
	userHandler "github.com/bashocode/gowallet/monolith/internal/user/handler"
	userRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userServer "github.com/bashocode/gowallet/monolith/internal/user/service"
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
	uSvc := userServer.NewUserService(uRepo)
	uHandler := userHandler.NewUserHandler(uSvc)

	//2. Setup gin router
	r := gin.Default()

	// routes
	r.POST("/api.v1.users", uHandler.Register)
	r.GET("/api.v1.users/:id", uHandler.GetProfile)
	r.PUT("/api.v1.users/:id", uHandler.UpdateProfile)

	// start server
	logger.Log.Info("Server running on Port 8080...")
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Server failed to run", "error", err)
		os.Exit(1)
	}
}
