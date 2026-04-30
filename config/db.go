package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(conString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func ConnectReadReplica(conString string) (*pgxpool.Pool, error) {
	if conString == "" {
		return nil, nil // No read replica configured
	}
	pool, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
