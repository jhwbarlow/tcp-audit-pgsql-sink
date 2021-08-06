package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type statementPreparer interface {
	prepareStatement(ctx context.Context, sql, name string) error
}

type pgxStatementPreparer struct {
	conn *pgx.Conn
}

func newPGXStatementPreparer(conn *pgx.Conn) *pgxStatementPreparer {
	return &pgxStatementPreparer{conn}
}

func (sp *pgxStatementPreparer) prepareStatement(ctx context.Context,
	sql string,
	name string) error {
	if _, err := sp.conn.Prepare(ctx, name, sql); err != nil {
		return fmt.Errorf("preparing statement on connection: %w", err)
	}

	return nil
}
