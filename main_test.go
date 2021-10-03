package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jhwbarlow/tcp-audit-common/pkg/event"
	"github.com/jhwbarlow/tcp-audit-common/pkg/tcpstate"
)

type mockInserter struct {
	errorToReturnOnPrepare error
	errorToReturnOnInsert  error
	errorToReturnOnClose   error

	prepareCalled bool
	insertCalled  bool
	closeCalled   bool

	receivedTime       time.Time
	receivedPID        int
	receivedComm       string
	receivedSrcIP      net.IP
	receivedDstIP      net.IP
	receivedSrcPort    uint16
	receivedDstPort    uint16
	receivedOldState   string
	receivedNewState   string
	receivedSocketInfo *socketInfo
}

func newMockInserter(errorToReturnOnPrepare error,
	errorToReturnOnInsert error,
	errorToReturnOnClose error) *mockInserter {
	return &mockInserter{
		errorToReturnOnPrepare: errorToReturnOnPrepare,
		errorToReturnOnInsert:  errorToReturnOnInsert,
		errorToReturnOnClose:   errorToReturnOnClose,
	}
}

func (mi *mockInserter) prepare(ctx context.Context) error {
	mi.prepareCalled = true

	if mi.errorToReturnOnPrepare != nil {
		return mi.errorToReturnOnPrepare
	}

	return nil
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
	newState string,
	socketInfo *socketInfo) error {
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
	mi.receivedSocketInfo = socketInfo

	if mi.errorToReturnOnInsert != nil {
		return mi.errorToReturnOnInsert
	}

	return nil
}

func (mi *mockInserter) close(ctx context.Context) error {
	mi.closeCalled = true

	if mi.errorToReturnOnClose != nil {
		return mi.errorToReturnOnClose
	}

	return nil
}

type mockTableCreator struct {
	errorToReturn error

	createTablesCalled bool
}

func newMockTableCreator(errorToReturn error) *mockTableCreator {
	return &mockTableCreator{errorToReturn: errorToReturn}
}

func (mtc *mockTableCreator) createTables(ctx context.Context) error {
	mtc.createTablesCalled = true

	if mtc.errorToReturn != nil {
		return mtc.errorToReturn
	}

	return nil
}

func TestSinkerConstructor(t *testing.T) {
	mockTableCreator := newMockTableCreator(nil)
	mockInserter := newMockInserter(nil, nil, nil)
	_, err := newSinker(mockTableCreator, mockInserter)
	if err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockTableCreator.createTablesCalled {
		t.Error("expected table creator to be called, but was not")
	}

	if !mockInserter.prepareCalled {
		t.Error("expected inserter prepare() to be called, but was not")
	}
}

func TestSinkerConstructorTableCreatorError(t *testing.T) {
	mockError := errors.New("mock table creator error")
	mockTableCreator := newMockTableCreator(mockError)
	mockInserter := newMockInserter(nil, nil, nil)
	_, err := newSinker(mockTableCreator, mockInserter)
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestSinkerConstructorInserterPrepareError(t *testing.T) {
	mockTableCreator := newMockTableCreator(nil)
	mockError := errors.New("mock inserter prepare error")
	mockInserter := newMockInserter(mockError, nil, nil)
	_, err := newSinker(mockTableCreator, mockInserter)
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestSink(t *testing.T) {
	mockTableCreator := newMockTableCreator(nil)
	mockInserter := newMockInserter(nil, nil, nil)
	sinker, err := newSinker(mockTableCreator, mockInserter)
	if err != nil {
		t.Errorf("expected nil Sinker construction error, got %q (of type %T)", err, err)
	}

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
	mockTableCreator := newMockTableCreator(nil)
	mockError := errors.New("mock inserter insert error")
	mockInserter := newMockInserter(nil, mockError, nil)
	sinker, err := newSinker(mockTableCreator, mockInserter)
	if err != nil {
		t.Errorf("expected nil Sinker construction error, got %q (of type %T)", err, err)
	}

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

	err = sinker.Sink(mockEvent)
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestClose(t *testing.T) {
	mockTableCreator := newMockTableCreator(nil)
	mockInserter := new(mockInserter)
	sinker, err := newSinker(mockTableCreator, mockInserter)
	if err != nil {
		t.Errorf("expected nil Sinker construction error, got %q (of type %T)", err, err)
	}

	if err := sinker.Close(); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockInserter.closeCalled {
		t.Error("expected inserter to be called closed, but was not")
	}
}

func TestCloseInserterError(t *testing.T) {
	mockTableCreator := newMockTableCreator(nil)
	mockError := errors.New("mock inserter close error")
	mockInserter := newMockInserter(nil, nil, mockError)
	sinker, err := newSinker(mockTableCreator, mockInserter)
	if err != nil {
		t.Errorf("expected nil Sinker construction error, got %q (of type %T)", err, err)
	}

	err = sinker.Close()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}
