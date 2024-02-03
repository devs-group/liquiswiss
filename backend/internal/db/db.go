// Location: /internal/db/db.go

package db

import (
	"database/sql"
	"fmt"
	"liquiswiss/config"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (db *sql.DB, err error) {
	cfg := config.GetConfig()

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName))
	if err != nil {
		return
	}

	err = db.Ping()
	if err != nil {
		return
	}

	return
}
