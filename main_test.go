package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jhwbarlow/tcp-audit/pkg/event"
	"github.com/jhwbarlow/tcp-audit/pkg/tcpstate"
)

type mockInserter struct {
	errorToReturn error
	insertCalled  bool
	closeCalled   bool

	receivedTime     time.Time
	receivedPID      int
	receivedComm     string
	receivedSrcIP    net.IP
	receivedDstIP    net.IP
	receivedSrcPort  uint16
	receivedDstPort  uint16
	receivedOldState string
	receivedNewState string
}

func newMockInserter(errorToReturn error) *mockInserter {
	return &mockInserter{errorToReturn: errorToReturn}
}

func (mi *mockInserter) insert(ctx context.Context,
	_ string,
	time time.Time,
	pid int,
	comm string,
	srcIP net.IP,
	dstIP net.IP,
	srcPort uint16,
	dstPort uint16,
	oldState string,
	newState string) error {
	mi.insertCalled = true

	mi.receivedTime = time
	mi.receivedPID = pid
	mi.receivedComm = comm
	mi.receivedSrcIP = srcIP
	mi.receivedDstIP = dstIP
	mi.receivedSrcPort = srcPort
	mi.receivedDstPort = dstPort
	mi.receivedOldState = oldState
	mi.receivedNewState = newState

	if mi.errorToReturn != nil {
		return mi.errorToReturn
	}

	return nil
}

func (mi *mockInserter) close(ctx context.Context) error {
	mi.closeCalled = true

	if mi.errorToReturn != nil {
		return mi.errorToReturn
	}

	return nil
}

func TestSink(t *testing.T) {
	mockInserter := new(mockInserter)
	sinker := construct(mockInserter)

	mockEvent := &event.Event{
		Time:         time.Now(),
		PIDOnCPU:     7337,
		CommandOnCPU: "test",
		SourceIP:     net.ParseIP("1.2.3.4"),
		DestIP:       net.ParseIP("7.3.3.7"),
		SourcePort:   1234,
		DestPort:     7337,
		OldState:     tcpstate.StateClosed,
		NewState:     tcpstate.StateSynReceived,
	}

	if err := sinker.Sink(mockEvent); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockInserter.insertCalled {
		t.Error("expected inserter insert() to be called, but was not")
	}

	if !mockInserter.receivedTime.Equal(mockEvent.Time) {
		t.Errorf("expected inserter received time to be %q, but was %q",
			mockEvent.Time,
			mockInserter.receivedTime)
	}

	if mockInserter.receivedPID != mockEvent.PIDOnCPU {
		t.Errorf("expected inserter received PID to be %d, but was %d",
			mockEvent.PIDOnCPU,
			mockInserter.receivedPID)
	}

	if mockInserter.receivedComm != mockEvent.CommandOnCPU {
		t.Errorf("expected inserter received command to be %q, but was %q",
			mockEvent.CommandOnCPU,
			mockInserter.receivedComm)
	}

	if !mockInserter.receivedSrcIP.Equal(mockEvent.SourceIP) {
		t.Errorf("expected inserter received source IP to be %q, but was %q",
			mockEvent.SourceIP,
			mockInserter.receivedSrcIP)
	}

	if !mockInserter.receivedDstIP.Equal(mockEvent.DestIP) {
		t.Errorf("expected inserter received destination IP to be %q, but was %q",
			mockEvent.DestIP,
			mockInserter.receivedDstIP)
	}

	if mockInserter.receivedSrcPort != mockEvent.SourcePort {
		t.Errorf("expected inserter received source port to be %d, but was %d",
			mockEvent.SourcePort,
			mockInserter.receivedSrcPort)
	}

	if mockInserter.receivedDstPort != mockEvent.DestPort {
		t.Errorf("expected inserter received destination port to be %d, but was %d",
			mockEvent.DestPort,
			mockInserter.receivedDstPort)
	}

	if mockInserter.receivedOldState != mockEvent.OldState.String() {
		t.Errorf("expected inserter received old state to be %q, but was %q",
			mockEvent.OldState,
			mockInserter.receivedOldState)
	}

	if mockInserter.receivedNewState != mockEvent.NewState.String() {
		t.Errorf("expected inserter received old state to be %q, but was %q",
			mockEvent.NewState,
			mockInserter.receivedNewState)
	}
}

func TestSinkInserterError(t *testing.T) {
	mockError := errors.New("mock inserter error")
	mockInserter := newMockInserter(mockError)
	sinker := construct(mockInserter)

	mockEvent := &event.Event{
		Time:         time.Now(),
		PIDOnCPU:     7337,
		CommandOnCPU: "test",
		SourceIP:     net.ParseIP("1.2.3.4"),
		DestIP:       net.ParseIP("7.3.3.7"),
		SourcePort:   1234,
		DestPort:     7337,
		OldState:     tcpstate.StateClosed,
		NewState:     tcpstate.StateSynReceived,
	}

	err := sinker.Sink(mockEvent)
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestClose(t *testing.T) {
	mockInserter := new(mockInserter)
	sinker := construct(mockInserter)

	if err := sinker.Close(); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockInserter.closeCalled {
		t.Error("expected inserter to be called closed, but was not")
	}
}

func TestCloseInserterError(t *testing.T) {
	mockError := errors.New("mock inserter close error")
	mockInserter := newMockInserter(mockError)
	sinker := construct(mockInserter)

	err := sinker.Close()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}
