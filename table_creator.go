package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
)

const (
	tableCreateSQL = `
CREATE TABLE tcp_events (
	uid         TEXT PRIMARY KEY,
	timestamp   TIMESTAMP,
	pid_on_cpu  INTEGER,
	comm_on_cpu TEXT,
	src_ip      INET,
	dst_ip      INET,
	src_port    INTEGER,
	dst_port    INTEGER,
	old_state   TEXT,
	new_state   TEXT
)`
)

type tableCreator interface {
	createTable(ctx context.Context) error
}

type pgxTableCreator struct {
	conn *pgx.Conn
}

func newPGXTableCreator(conn *pgx.Conn) *pgxTableCreator {
	return &pgxTableCreator{conn}
}

func (tc *pgxTableCreator) createTable(ctx context.Context) error {
	if _, err := tc.conn.Exec(ctx, tableCreateSQL); err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.DuplicateTable {
			// Table already created - nothing to do!
			return nil
		}
		
		return fmt.Errorf("creating tcp_events table: %w", err)
	}

	return nil
}
