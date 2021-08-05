package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
)

type tableCreator interface {
	createTable(ctx context.Context, conn *pgx.Conn, sql string) error
}

type pgxTableCreator struct{}

func (*pgxTableCreator) createTable(ctx context.Context, conn *pgx.Conn, sql string) error {
	if _, err := conn.Exec(ctx, sql); err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.DuplicateTable {
			// Table already created - nothing to do!
			return nil
		}

		return fmt.Errorf("creating table: %w", err)
	}

	return nil
}
