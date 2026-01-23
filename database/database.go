package kyugo

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"

	cfg "github.com/go-kyugo/kyugo/config"
)

type DB struct {
	SQL *sql.DB
}

// ConnectFromConfig opens a database connection from the provided config.
// It accepts the `config.DatabaseConfig` type so callers can pass
// `cfg.ConfigVar.Database` directly.
func ConnectFromConfig(c cfg.DatabaseConfig) (*DB, error) {
	if c.Type != "postgres" {
		return nil, fmt.Errorf("unsupported database type: %s", c.Type)
	}

	password := url.QueryEscape(c.Password)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, password, c.DBName, c.SSLMode)

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &DB{SQL: sqlDB}, nil
}

var defaultDB *DB

func SetDefault(db *DB) {
	defaultDB = db
}
