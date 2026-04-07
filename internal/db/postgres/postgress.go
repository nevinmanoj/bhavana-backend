package postgres

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) *sqlx.DB {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to open DB:", err)
	}

	// Verify connection early
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping DB:", err)
	}

	// Connection pool tuning (still works the same)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
