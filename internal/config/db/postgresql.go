package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
)

func InitDB(config config.Config) *sql.DB {
	db, err := sql.Open("pgx", config.DatabaseDsn)
	if err != nil {
		panic(err)
	}

	return db
}
