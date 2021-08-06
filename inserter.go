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

type inserter interface {
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

type preparedStatementInserter struct {
	execer execer
}

func newPreparedStatementInserter(stmtPreparer statementPreparer,
	execer execer) (*preparedStatementInserter, error) {
	if err := stmtPreparer.prepareStatement(context.TODO(),
		insertSQL,
		insertSQLStmtName); err != nil {
		return nil, fmt.Errorf("preparing insert statement: %w", err)
	}

	return &preparedStatementInserter{
		execer: execer,
	}, nil
}

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
		srcIP,
		dstIP,
		srcPort,
		dstPort,
		oldState,
		newState); err != nil {
		return fmt.Errorf("inserting into tcp_events: %w", err)
	}

	return nil
}

func (i *preparedStatementInserter) close(ctx context.Context) error {
	if err := i.execer.close(ctx); err != nil {
		return fmt.Errorf("closing execer: %w", err)
	}

	return nil
}
