package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type statementPreparer interface {
	prepareStatement(ctx context.Context,
		conn *pgx.Conn,
		sql string,
		name string) (*pgconn.StatementDescription, error)
}

type pgxStatementPreparer struct{}

func (*pgxStatementPreparer) prepareStatement(ctx context.Context,
	conn *pgx.Conn,
	sql string,
	name string) (*pgconn.StatementDescription, error) {
	stmt, err := conn.Prepare(ctx, name, sql)
	if err != nil {
		return nil, fmt.Errorf("preparing statement: %w", err)
	}

	return stmt, nil
}
