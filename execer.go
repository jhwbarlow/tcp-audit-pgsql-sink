package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

type execer interface {
	exec(ctx context.Context,
		sql string,
		arguments ...interface{}) error
	close(ctx context.Context) error
}

type pgxExecer struct {
	conn *pgx.Conn
}

func newPGXExecer(conn *pgx.Conn) *pgxExecer {
	return &pgxExecer{conn}
}

func (e *pgxExecer) exec(ctx context.Context,
	sql string,
	arguments ...interface{}) error {
	if _, err := e.conn.Exec(ctx, sql, arguments...); err != nil {
		return fmt.Errorf("execing SQL on connection: %w", err)
	}

	return nil
}

func (e *pgxExecer) close(ctx context.Context) error {
	log.Printf("Closing database connection: %s:%d",
		e.conn.Config().Host,
		e.conn.Config().Port)
	if err := e.conn.Close(ctx); err != nil {
		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}
