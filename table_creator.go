package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

const (
	eventsTableCreateSQL = `
CREATE TABLE IF NOT EXISTS tcp_events (
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

	socketInfoTableCreateSQL = `
CREATE TABLE IF NOT EXISTS tcp_events_socket_info (
	uid           TEXT PRIMARY KEY,
	tcp_event_uid TEXT,
	id            TEXT,
	inode         INTEGER,
	user_id       INTEGER,
	group_id      INTEGER,
	state         TEXT,	
	CONSTRAINT fk_tcp_events FOREIGN KEY(tcp_event_uid) 
		REFERENCES tcp_events(uid) ON DELETE CASCADE
)`
)

// TableCreator is an interface which describes objects which create
// the database tables required to store TCP state-change events.
type tableCreator interface {
	createTables(ctx context.Context) error
}

// PGXTableCreator creates the database tables required to store TCP
// state-change events using the PGX library.
type pgxTableCreator struct {
	conn conn
}

func newPGXTableCreator(conn conn) *pgxTableCreator {
	return &pgxTableCreator{conn}
}

// CreateTables creates the tables in the database if they do not already exist.
func (tc *pgxTableCreator) createTables(ctx context.Context) error {
	if _, err := tc.conn.Exec(ctx, eventsTableCreateSQL); err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.DuplicateTable {
			// Table already created - nothing to do!
			// This should not happen as we use CREATE TABLE IF NOT EXISTS, but it is a easy check to do.
			return nil
		}

		return fmt.Errorf("creating tcp_events table: %w", err)
	}

	if _, err := tc.conn.Exec(ctx, socketInfoTableCreateSQL); err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.DuplicateTable {
			// Table already created - nothing to do!
			// This should not happen as we use CREATE TABLE IF NOT EXISTS, but it is a easy check to do.
			return nil
		}

		return fmt.Errorf("creating tcp_events_socket_info table: %w", err)
	}

	return nil
}
