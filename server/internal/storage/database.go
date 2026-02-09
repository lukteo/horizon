package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseService handles interactions with PostgreSQL
type DatabaseService struct {
	Pool *pgxpool.Pool
}

// NewDatabaseService creates a new instance of DatabaseService
func NewDatabaseService(databaseURL string) (*DatabaseService, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseService{
		Pool: pool,
	}, nil
}

// Close closes the database connection pool
func (db *DatabaseService) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// HealthCheck checks if the database is accessible
func (db *DatabaseService) HealthCheck() error {
	return db.Pool.Ping(context.Background())
}

// GetPool returns the underlying connection pool
func (db *DatabaseService) GetPool() interface{} {
	return db.Pool
}