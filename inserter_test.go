package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

const expectedNumberOfPreparedStmts = 2

type mockStatementPreparer struct {
	errorToReturn    error
	errorReturnDelay int

	callCount                 int
	prepareStatementCalled    bool
	receivedSQLs              []string
	receivedPreparedStmtNames []string
}

func newMockStatementPreparer(errorToReturn error, errorReturnDelay int) *mockStatementPreparer {
	return &mockStatementPreparer{
		errorToReturn:             errorToReturn,
		errorReturnDelay:          errorReturnDelay,
		receivedSQLs:              []string{},
		receivedPreparedStmtNames: []string{},
	}
}

func (msp *mockStatementPreparer) prepareStatement(ctx context.Context,
	sql string,
	name string) error {
	defer func() {
		msp.callCount++
	}()

	msp.prepareStatementCalled = true
	msp.receivedSQLs = append(msp.receivedSQLs, sql)
	msp.receivedPreparedStmtNames = append(msp.receivedPreparedStmtNames, name)

	if msp.errorToReturn != nil && msp.callCount >= msp.errorReturnDelay {
		return msp.errorToReturn
	}

	return nil
}

type mockExecer struct {
	errorToReturn error

	execCalled         bool
	execMultipleCalled bool
	closeCalled        bool
	receivedSQL        string
	receivedArgs       []interface{}
	receivedStmts      []*sqlStatement
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
	me.execMultipleCalled = true
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

func TestInsertNoSocketInfo(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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

func TestInsertWithSocketInfo(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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
	mockSocketInfo := &socketInfo{
		id:      "mock-socket-id",
		iNode:   0xF00DF00D,
		userID:  0xCAFEBABE,
		groupID: 0xDEADBEEF,
		state:   "mock-socket-state",
	}

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

	if !mockExecer.execMultipleCalled {
		t.Error("expected execer execMultiple() to be called, but was not")
	}

	if mockExecer.receivedStmts == nil || len(mockExecer.receivedStmts) == 0 {
		t.Error("expected execer to receive non-empty statements list, but was empty")
	}

	if len(mockExecer.receivedStmts) != expectedNumberOfPreparedStmts {
		t.Errorf("expected execer to receive %d statements in list, but received %d",
			expectedNumberOfPreparedStmts,
			len(mockExecer.receivedStmts))
	}

	for i, receivedStmt := range mockExecer.receivedStmts {
		j := i + 1

		if receivedStmt.sql == "" {
			t.Errorf("expected execer statement %d to receive non-empty prepared statement name, but was empty", j)
		}
		t.Logf("execer statement %d received prepared statement name: %q", j, receivedStmt.sql)

		if receivedStmt.arguments == nil || len(receivedStmt.arguments) == 0 {
			t.Errorf("expected execer statement %d to receive non-empty non-empty arguments, but was empty", j)
		}
	}
}

func TestInsertErrorNoSocketInfo(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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

func TestInsertErrorWithSocketInfo(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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
	mockSocketInfo := &socketInfo{
		id:      "mock-socket-id",
		iNode:   0xF00DF00D,
		userID:  0xCAFEBABE,
		groupID: 0xDEADBEEF,
		state:   "mock-socket-state",
	}

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
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
	mockExecer := newMockExecer(nil)

	inserter := newPreparedStatementInserter(mockStmtPreparer, mockExecer)
	if err := inserter.prepare(context.TODO()); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockStmtPreparer.prepareStatementCalled {
		t.Error("expected statementPreparer to be called, but was not")
	}

	if mockStmtPreparer.receivedPreparedStmtNames == nil || len(mockStmtPreparer.receivedPreparedStmtNames) == 0 {
		t.Error("expected statementPreparer to receive non-empty prepared statement names, but was empty")
	}

	if mockStmtPreparer.receivedSQLs == nil || len(mockStmtPreparer.receivedSQLs) == 0 {
		t.Error("expected statementPreparer to receive non-empty SQL statements, but was empty")
	}

	if len(mockStmtPreparer.receivedPreparedStmtNames) != len(mockStmtPreparer.receivedSQLs) {
		t.Error("expected statementPreparer to receive same number of SQL statements as prepared statement names, but did not")
	}

	if len(mockStmtPreparer.receivedPreparedStmtNames) != expectedNumberOfPreparedStmts &&
		len(mockStmtPreparer.receivedSQLs) != expectedNumberOfPreparedStmts {
		t.Errorf("expected execer to receive %d statements to prepare, but received %d",
			expectedNumberOfPreparedStmts,
			len(mockStmtPreparer.receivedPreparedStmtNames))
	}

	for i, receivedStmtName := range mockStmtPreparer.receivedPreparedStmtNames {
		j := i + 1

		if receivedStmtName == "" {
			t.Errorf("expected statementPreparer to receive non-empty prepared statement name for statement %d, but was empty", j)
		}

		t.Logf("statementPreparer received prepared statement name for statement %d: %q", j, receivedStmtName)
	}

	for i, receivedSQL := range mockStmtPreparer.receivedSQLs {
		j := i + 1

		if receivedSQL == "" {
			t.Errorf("expected statementPreparer to receive non-empty SQL for statement %d, but was empty", j)
		}

		t.Logf("statementPreparer received SQL for statement %d: %q", j, receivedSQL)
	}
}

func TestInserterPrepareStatementPreparerError(t *testing.T) {
	mockError := errors.New("mock statement preparer error")

	// Ensure the error return is tested for all prepared statements
	for i := 0; i < expectedNumberOfPreparedStmts; i++ {
		mockStmtPreparer := newMockStatementPreparer(mockError, i)
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
}

func TestInserterClose(t *testing.T) {
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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
	mockStmtPreparer := newMockStatementPreparer(nil, 0)
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
