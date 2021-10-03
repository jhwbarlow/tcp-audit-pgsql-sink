package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

const (
	insertTCPEventsTableSQL = `
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

	insertTCPEventsTableSQLStmtName = "tcp_events_insert"

	insertSocketInfoTableSQL = `
INSERT INTO tcp_events_socket_info (
	uid,
	tcp_event_uid,
	id,
	inode,
	user_id,
	group_id,
	state
) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	insertSocketInfoTableSQLStmtName = "tcp_events_socket_info_insert"
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
		newState string,
		socketInfo *socketInfo) error
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

// Prepare prepares the SQL insert statements for future use in the insert
// method.
func (i *preparedStatementInserter) prepare(ctx context.Context) error {
	if err := i.stmtPreparer.prepareStatement(ctx,
		insertTCPEventsTableSQL,
		insertTCPEventsTableSQLStmtName); err != nil {
		return fmt.Errorf("preparing insert tcp_events statement: %w", err)
	}

	if err := i.stmtPreparer.prepareStatement(ctx,
		insertSocketInfoTableSQL,
		insertSocketInfoTableSQLStmtName); err != nil {
		return fmt.Errorf("preparing insert tcp_events_socket_info statement: %w", err)
	}

	return nil
}

// Insert uses the prepared SQL insert statements created in the prepare
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
	newState string,
	socketInfo *socketInfo) error {
	if socketInfo == nil {
		if err := i.execer.exec(context.TODO(),
			insertTCPEventsTableSQLStmtName,
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

	tcpEventsSQLStatement := newSQLStatement(insertTCPEventsTableSQLStmtName,
		uid,
		time,
		pid,
		comm,
		srcIP.To4(),
		dstIP.To4(),
		srcPort,
		dstPort,
		oldState,
		newState)
	socketInfoSQLStatement := newSQLStatement(insertSocketInfoTableSQLStmtName,
		socketInfo.uid,
		uid,
		socketInfo.id,
		socketInfo.iNode,
		socketInfo.userID,
		socketInfo.groupID,
		socketInfo.state)

	if err := i.execer.execMultiple(context.TODO(),
		tcpEventsSQLStatement,
		socketInfoSQLStatement); err != nil {
		return fmt.Errorf("inserting into tcp_events or tcp_events_socket_info: %w", err)
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
