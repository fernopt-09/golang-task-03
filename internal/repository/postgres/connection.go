package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection settings.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN builds the connection string.
func (c Config) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, sslMode,
	)
}

// NewPool opens a connection pool and checks it with ping.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := cfg.DSN()
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pgxpool config: %w", err)
	}

	// pool limits, they are enough for this service
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// pgxpool connects lazily, so check the connection right now
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
