package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

const (
	insertSQL = `
INSERT INTO tcp_events (
	uid,
	timestamp,  
	pid_on_cpu,
	comm_on_cpu,
	src_ip,
	dst_ip,
	src_port,
	dst_port,
	old_state,
	new_state
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	insertSQLStmtName = "tcp_events_insert"
)

// Inserter is an interface which describes objects which inserts
// TCP state-change data into the backing store.
type inserter interface {
	prepare(ctx context.Context) error
	insert(ctx context.Context,
		uid string,
		time time.Time,
		pid int,
		comm string,
		srcIP net.IP,
		dstIP net.IP,
		srcPort uint16,
		dstPort uint16,
		oldState string,
		newState string) error
	close(ctx context.Context) error
}

// PreparedStatementInserter inserts TCP state-change data into the
// database using a SQL prepared statement.
type preparedStatementInserter struct {
	execer       execer
	stmtPreparer statementPreparer
}

func newPreparedStatementInserter(stmtPreparer statementPreparer,
	execer execer) *preparedStatementInserter {
	return &preparedStatementInserter{
		stmtPreparer: stmtPreparer,
		execer:       execer,
	}
}

// Prepare prepares the SQL insert statement for future use in the insert
// method.
func (i *preparedStatementInserter) prepare(ctx context.Context) error {
	if err := i.stmtPreparer.prepareStatement(ctx,
		insertSQL,
		insertSQLStmtName); err != nil {
		return fmt.Errorf("preparing insert statement: %w", err)
	}

	return nil
}

// Insert uses the prepared SQL insert statement created in the prepare
// method to insert TCP state-change data into the database.
func (i *preparedStatementInserter) insert(ctx context.Context,
	uid string,
	time time.Time,
	pid int,
	comm string,
	srcIP net.IP,
	dstIP net.IP,
	srcPort uint16,
	dstPort uint16,
	oldState string,
	newState string) error {
	if err := i.execer.exec(context.TODO(),
		insertSQLStmtName,
		uid,
		time,
		pid,
		comm,
		srcIP.To4(),
		dstIP.To4(),
		srcPort,
		dstPort,
		oldState,
		newState); err != nil {
		return fmt.Errorf("inserting into tcp_events: %w", err)
	}

	return nil
}

// Close releases the resources held by this Inserter.
func (i *preparedStatementInserter) close(ctx context.Context) error {
	if err := i.execer.close(ctx); err != nil {
		return fmt.Errorf("closing execer: %w", err)
	}

	return nil
}
