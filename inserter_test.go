package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type mockStatementPreparer struct {
	errorToReturn error

	prepareStatementCalled   bool
	receivedSQL              string // TODO: Should be a slice
	receivedPreparedStmtName string // TODO: Should be a slice
}

func newMockStatementPreparer(errorToReturn error) *mockStatementPreparer {
	return &mockStatementPreparer{errorToReturn: errorToReturn}
}

func (msp *mockStatementPreparer) prepareStatement(ctx context.Context,
	sql string,
	name string) error {
	msp.prepareStatementCalled = true
	msp.receivedSQL = sql
	msp.receivedPreparedStmtName = name

	if msp.errorToReturn != nil {
		return msp.errorToReturn
	}

	return nil
}

type mockExecer struct {
	errorToReturn error

	execCalled    bool
	closeCalled   bool
	receivedSQL   string
	receivedArgs  []interface{}
	receivedStmts []*sqlStatement
}

func newMockExecer(errorToReturn error) *mockExecer {
	return &mockExecer{errorToReturn: errorToReturn}
}

func (me *mockExecer) exec(ctx context.Context,
	sql string,
	arguments ...interface{}) error {

	me.execCalled = true
	me.receivedSQL = sql
	me.receivedArgs = arguments

	if me.errorToReturn != nil {
		return me.errorToReturn
	}

	return nil
}

func (me *mockExecer) execMultiple(ctx context.Context, stmts ...*sqlStatement) error {
	me.execCalled = true
	me.receivedStmts = stmts

	if me.errorToReturn != nil {
		return me.errorToReturn
	}

	return nil
}

func (me *mockExecer) close(ctx context.Context) error {
	me.closeCalled = true

	if me.errorToReturn != nil {
		return me.errorToReturn
	}

	return nil
}

func TestInsert(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil)
	mockExecer := newMockExecer(nil)
	mockUID := "mock-uid"
	mockTime := time.Now()
	mockPID := 7337
	mockComm := "mock-command"
	mockSrcIP := net.ParseIP("1.2.3.4")
	mockDstIP := net.ParseIP("7.3.3.7")
	mockSrcPort := uint16(1234)
	mockDstPort := uint16(7337)
	mockOldState := "mock-old-state"
	mockNewState := "mock-new-state"
	var mockSocketInfo *socketInfo

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)

	if err := inserter.insert(context.TODO(),
		mockUID,
		mockTime,
		mockPID,
		mockComm,
		mockSrcIP,
		mockDstIP,
		mockSrcPort,
		mockDstPort,
		mockOldState,
		mockNewState,
		mockSocketInfo); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockExecer.execCalled {
		t.Error("expected execer exec() to be called, but was not")
	}

	if mockExecer.receivedSQL == "" {
		t.Error("expected execer to receive non-empty prepared statement name, but was empty")
	}
	t.Logf("execer received prepared statement name: %q", mockExecer.receivedSQL)

	if mockExecer.receivedArgs == nil || len(mockExecer.receivedArgs) == 0 {
		t.Error("expected execer to receive non-empty arguments, but was empty")
	}
}

func TestInsertError(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil)
	mockError := errors.New("mock exec error")
	mockExecer := newMockExecer(mockError)
	mockUID := "mock-uid"
	mockTime := time.Now()
	mockPID := 7337
	mockComm := "mock-command"
	mockSrcIP := net.ParseIP("1.2.3.4")
	mockDstIP := net.ParseIP("7.3.3.7")
	mockSrcPort := uint16(1234)
	mockDstPort := uint16(7337)
	mockOldState := "mock-old-state"
	mockNewState := "mock-new-state"
	var mockSocketInfo *socketInfo

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)

	err := inserter.insert(context.TODO(),
		mockUID,
		mockTime,
		mockPID,
		mockComm,
		mockSrcIP,
		mockDstIP,
		mockSrcPort,
		mockDstPort,
		mockOldState,
		mockNewState,
		mockSocketInfo)
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestInserterPrepare(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil)
	mockExecer := newMockExecer(nil)

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)
	if err := inserter.prepare(context.TODO()); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockStmtPreparer.prepareStatementCalled {
		t.Error("expected statementPreparer to be called, but was not")
	}

	if mockStmtPreparer.receivedPreparedStmtName == "" {
		t.Error("expected statementPreparer to receive non-empty prepared statement name, but was empty")
	}
	t.Logf("statementPreparer received prepared statement name: %q", mockStmtPreparer.receivedPreparedStmtName)

	if mockStmtPreparer.receivedSQL == "" {
		t.Error("expected statementPreparer to receive non-empty SQL, but was empty")
	}
	t.Logf("statementPreparer received SQL: %q", mockStmtPreparer.receivedSQL)
}

func TestInserterPrepareStatementPreparerError(t *testing.T) {
	mockError := errors.New("mock statement preparer error")
	mockStmtPreparer := newMockStatementPreparer(mockError)
	mockExecer := newMockExecer(nil)

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)
	err := inserter.prepare(context.TODO())
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}

func TestInserterClose(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil)
	mockExecer := newMockExecer(nil)

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)

	if err := inserter.close(context.TODO()); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockExecer.closeCalled {
		t.Error("expected execer to be closed, but was not")
	}
}

func TestInserterCloseError(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil)
	mockError := errors.New("mock exec close error")
	mockExecer := newMockExecer(mockError)

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)

	err := inserter.close(context.TODO())
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}
}
