package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

// Execer is an interface which describes objects which execute
// SQL statements.
type execer interface {
	exec(ctx context.Context,
		sql string,
		arguments ...interface{}) error
	close(ctx context.Context) error
}

// PGXExecer executes SQL statements using the PGX library.
type pgxExecer struct {
	conn *pgx.Conn
}

func newPGXExecer(conn *pgx.Conn) *pgxExecer {
	return &pgxExecer{conn}
}

// Exec executes the provided SQL using the provided arguments.
func (e *pgxExecer) exec(ctx context.Context,
	sql string,
	arguments ...interface{}) error {
	if _, err := e.conn.Exec(ctx, sql, arguments...); err != nil {
		return fmt.Errorf("execing SQL on connection: %w", err)
	}

	return nil
}

// Close releases the resources held by this Execer, namely the
// database connection.
func (e *pgxExecer) close(ctx context.Context) error {
	log.Printf("Closing database connection: %s:%d",
		e.conn.Config().Host,
		e.conn.Config().Port)
	if err := e.conn.Close(ctx); err != nil {
		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}
