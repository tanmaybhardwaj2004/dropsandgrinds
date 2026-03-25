package config

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(conString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), conString)
	if err != nil {
		return nil, err
	}
	return conn, nil

}
