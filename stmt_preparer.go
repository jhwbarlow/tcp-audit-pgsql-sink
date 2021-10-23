package main

import (
	"context"
	"fmt"
)

// StatementPreparer is an interface which describes objects which prepare
// named SQL statements for future use.
type statementPreparer interface {
	prepareStatement(ctx context.Context, sql, name string) error
}

// PGXStatementPreparer prepares named SQL statements for future use using
// the PGX library.
type pgxStatementPreparer struct {
	conn conn
}

func newPGXStatementPreparer(conn conn) *pgxStatementPreparer {
	return &pgxStatementPreparer{conn}
}

// PrepareStatement prepares the SQL statement with the given name.
func (sp *pgxStatementPreparer) prepareStatement(ctx context.Context,
	sql string,
	name string) error {
	if _, err := sp.conn.Prepare(ctx, name, sql); err != nil {
		return fmt.Errorf("preparing statement on connection: %w", err)
	}

	return nil
}
