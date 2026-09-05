package config

import (
	"os"

	"github.com/bashocode/gowallet/monolith/internal/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	DBSN      string
	RedisAddr string
}

func LoadConfig() *Config {
	//load file .env if there is any
	if err := godotenv.Load(); err != nil {
		logger.Log.Warn(".env file not found, using environment variables")
	}

	dsn := os.Getenv("DB_USER") + ":" +
		os.Getenv("DB_PASSWORD") + "@tcp(" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?parseTime=true"

	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")

	return &Config{
		DBSN:      dsn,
		RedisAddr: redisAddr,
	}
}
