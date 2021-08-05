package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jhwbarlow/tcp-audit/pkg/event"
	"github.com/jhwbarlow/tcp-audit/pkg/sink"
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

type Sinker struct {
	configGetter configGetter

	conn       *pgx.Conn
	insertStmt *pgconn.StatementDescription
}

func New() (sink.Sinker, error) {
	connector := new(pgxConnector)
	tableCreator := new(pgxTableCreator)
	stmtPreparer := new(pgxStatementPreparer)

	return construct(connector, tableCreator, stmtPreparer)
}

func construct(connector connector,
	tableCreator tableCreator,
	stmtPreparer statementPreparer) (sink.Sinker, error) {
	conn, err := connector.connect(context.TODO(), connString)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := tableCreator.createTable(context.TODO(), conn, tableCreateSQL); err != nil {
		return nil, fmt.Errorf("creating tcp_events table: %w", err)
	}

	insertStmt, err := stmtPreparer.prepareStatement(context.TODO(), conn, insertSQL, insertSQLStmtName)
	if err != nil {
		return nil, fmt.Errorf("preparing insert statement: %w", err)
	}

	return &Sinker{
		conn:       conn,
		insertStmt: insertStmt,
	}, nil
}

func (s *Sinker) Sink(event *event.Event) error {
	uid := uuid.NewString()
	time := event.Time
	pid := event.PIDOnCPU
	comm := event.CommandOnCPU
	srcIP := event.SourceIP
	dstIP := event.DestIP
	srcPort := event.SourcePort
	dstPort := event.DestPort
	oldState := event.OldState.String()
	newState := event.NewState.String()

	if _, err := s.conn.Exec(context.TODO(),
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
		return fmt.Errorf("inserting event: %w", err)
	}

	return nil
}

func (s *Sinker) Close() error {
	log.Printf("Closing database connection: %s:%d",
		s.conn.Config().Host,
		s.conn.Config().Port)
	if err := s.conn.Close(context.TODO()); err != nil {
		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}
