package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jhwbarlow/tcp-audit/pkg/event"
	"github.com/jhwbarlow/tcp-audit/pkg/sink"
)

type Sinker struct {
	inserter inserter
}

func New() (sink.Sinker, error) {
	configGetter := new(fixedStringConfigGetter)
	connector := newPGXConnector(configGetter)
	conn, err := connect(connector)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	tableCreator := newPGXTableCreator(conn)
	if err := createTable(tableCreator); err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}

	stmtPreparer := newPGXStatementPreparer(conn)
	execer := newPGXExecer(conn)

	inserter, err := newPreparedStatementInserter(stmtPreparer, execer)
	if err != nil {
		return nil, fmt.Errorf("constructing inserter: %w", err)
	}

	return construct(inserter), nil
}

func construct(inserter inserter) *Sinker {
	return &Sinker{
		inserter: inserter,
	}
}

func connect(connector connector) (*pgx.Conn, error) {
	return connector.connect(context.TODO())
}

func createTable(tableCreator tableCreator) error {
	return tableCreator.createTable(context.TODO())
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

	if err := s.inserter.insert(context.TODO(),
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
	if err := s.inserter.close(context.TODO()); err != nil {
		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}
