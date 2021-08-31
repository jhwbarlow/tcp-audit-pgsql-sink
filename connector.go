package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

// PGXConnector creates a connection to a PostgreSQL database using the
// PGX library.
type pgxConnector struct {
	configGetter configGetter
}

func newPGXConnector(configGetter configGetter) *pgxConnector {
	return &pgxConnector{configGetter}
}

// Connect creates a connection to the PostgreSQL database described by the
// configuration returned by the ConfigGetter supplied in the constructor.
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
