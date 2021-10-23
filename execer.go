package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

// SQLStatement defines a SQL statement with the SQL and arguments.
type sqlStatement struct {
	sql       string
	arguments []interface{}
}

func newSQLStatement(sql string, arguments ...interface{}) *sqlStatement {
	return &sqlStatement{sql, arguments}
}

// Execer is an interface which describes objects which execute
// SQL statements.
type execer interface {
	exec(ctx context.Context, sql string, arguments ...interface{}) error
	execMultiple(ctx context.Context, stmts ...*sqlStatement) error
	close(ctx context.Context) error
}

// PGXExecer executes SQL statements using the PGX library.
type pgxExecer struct {
	conn conn
}

func newPGXExecer(conn conn) *pgxExecer {
	return &pgxExecer{conn}
}

// Exec executes the provided SQL statement using the provided arguments.
func (e *pgxExecer) exec(ctx context.Context,
	sql string,
	arguments ...interface{}) error {
	if _, err := e.conn.Exec(ctx, sql, arguments...); err != nil {
		return fmt.Errorf("execing SQL on connection: %w", err)
	}

	return nil
}

// ExecMultiple executes the provided SQL statement(s) using the provided arguments.
// If more than one statement is provided, they are executed atomically
// (i.e. in a transaction).
func (e *pgxExecer) execMultiple(ctx context.Context, stmts ...*sqlStatement) (err error) {

	tx, err := e.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting database transaction: %w", err)
	}

	// "Finally" block
	defer func(ctx context.Context, tx pgx.Tx) {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Printf("Error rolling-back database transaction: %v", rollbackErr)
			}
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			// Replace the nil-error return from the parent function with this error
			err = fmt.Errorf("committing database transaction: %w", commitErr)
		}
	}(ctx, tx)

	for i, stmt := range stmts {
		sql := stmt.sql
		args := stmt.arguments

		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("execing SQL statement %d within transaction: %w", i, err)
		}
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
