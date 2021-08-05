package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

var connString = "postgresql://postgres:mysecretpassword@127.0.0.1/postgres"

type connector interface {
	connect(ctx context.Context, connString string) (*pgx.Conn, error)
}

type pgxConnector struct{}

func (*pgxConnector) connect(ctx context.Context, connString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("establishing connection to database: %w", err)
	}

	return conn, nil
}
