package main

import (
	"log"
	"os"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
	userHandler "github.com/bashocode/gowallet/monolith/internal/user/handler"
	userRepository "github.com/bashocode/gowallet/monolith/internal/user/repository"
	userServer "github.com/bashocode/gowallet/monolith/internal/user/service"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting Monolith Wallet Application...")

	//1. Load configuration
	cfg := config.LoadConfig()

	//2. Connect to database with retry
	db, err := database.ConnectWithRetry(cfg.DBSN)
	if err != nil {
		log.Fatal("Critical Error: Could not connect to database after retries", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Println("Application successfully initialized ...")

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

	// strat server
	log.Println("Server running on Port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
		os.Exit(1)
	}
}
