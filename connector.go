package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type connector interface {
	connect(ctx context.Context) (*pgx.Conn, error)
}

type pgxConnector struct {
	configGetter configGetter
}

func newPGXConnector(configGetter configGetter) *pgxConnector {
	return &pgxConnector{configGetter}
}

func (c *pgxConnector) connect(ctx context.Context) (*pgx.Conn, error) {
	connString, err := c.configGetter.config()
	if err != nil {
		return nil, fmt.Errorf("getting connection string from config: %w", err)
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("establishing connection to database: %w", err)
	}

	return conn, nil
}
