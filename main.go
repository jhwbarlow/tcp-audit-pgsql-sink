package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jhwbarlow/tcp-audit-common/pkg/event"
	"github.com/jhwbarlow/tcp-audit-common/pkg/sink"
)

type Sinker struct {
	inserter inserter
}

func New() (sink.Sinker, error) {
	configGetter := new(envVarConfigGetter)
	connector := newPGXConnector(configGetter)
	conn, err := connector.connect(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	tableCreator := newPGXTableCreator(conn)
	stmtPreparer := newPGXStatementPreparer(conn)
	execer := newPGXExecer(conn)
	inserter := newPreparedStatementInserter(stmtPreparer, execer)

	return newSinker(tableCreator, inserter)
}

func newSinker(tableCreator tableCreator, inserter inserter) (*Sinker, error) {
	if err := tableCreator.createTables(context.TODO()); err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}

	if err := inserter.prepare(context.TODO()); err != nil {
		return nil, fmt.Errorf("preparing inserter: %w", err)
	}

	return &Sinker{
		inserter: inserter,
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

	var sockInfo *socketInfo
	if event.SocketInfo != nil {
		sockInfo = &socketInfo{
			uid:     uuid.NewString(),
			id:      event.SocketInfo.ID,
			iNode:   event.SocketInfo.INode,
			userID:  event.SocketInfo.UID,
			groupID: event.SocketInfo.GID,
			state:   event.SocketInfo.SocketState.String(),
		}
	}

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
		newState,
		sockInfo); err != nil {
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
