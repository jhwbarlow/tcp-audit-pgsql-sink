package main

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Conn is an interface which is a wrapper around the *pgx.Conn struct.
type conn interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Config() *pgx.ConnConfig
	Close(ctx context.Context) error
	Prepare(ctx context.Context, name, sql string) (sd *pgconn.StatementDescription, err error)
}
