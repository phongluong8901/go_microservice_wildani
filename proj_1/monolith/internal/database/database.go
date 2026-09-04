package database

import (
	"database/sql"
	"time"

	"github.com/bashocode/gowallet/monolith/internal/logger"
	_ "github.com/go-sql-driver/mysql"
)

func ConnectWithRetry(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	maxRetries := 5
	backoff := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		logger.Log.Info("Connecting to database...", "attempt", i, "max_retries", maxRetries)

		db, err = sql.Open("mysql", dsn)
		if err == nil {
			// do ping for make sure connetion is alive
			err = db.Ping()
			if err == nil {
				logger.Log.Info("Database connected successfully")

				//setup connection pool progression
				db.SetMaxOpenConns(25)                 //Giới hạn tối đa 25 kết nối đồng thời được mở tới cơ sở dữ liệu cùng một lúc
				db.SetMaxIdleConns(25)                 //Số lượng kết nối tối đa mà “idle pool” (bộ đệm kết nối chờ) được phép giữ lại.
				db.SetConnMaxLifetime(5 * time.Minute) //Thời gian tối đa mà một kết nối được phép tồn tại

				return db, nil
			}
		}

		logger.Log.Warn("Database connection failed, retrying...", "error", err, "backoff", backoff)
		time.Sleep(backoff)

		//double backoff for waiting to next retry, or exponential backoff
		backoff *= 2
	}

	return nil, err
}
