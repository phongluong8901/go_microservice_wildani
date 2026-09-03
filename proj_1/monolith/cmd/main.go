package main

import (
	"log"
	"os"

	"github.com/bashocode/gowallet/monolith/internal/config"
	"github.com/bashocode/gowallet/monolith/internal/database"
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

	//3. Set
}
