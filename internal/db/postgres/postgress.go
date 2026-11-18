package postgres

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgres opens the pool. maxOpenConns is per replica, so replicas times
// maxOpenConns must stay under the server's max_connections minus the ~15
// Azure reserves (B1ms allows 50, leaving about 35).
func NewPostgres(dsn string, maxOpenConns int) *sqlx.DB {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to open DB:", err)
	}

	// Verify connection early
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping DB:", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(min(10, maxOpenConns))
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
