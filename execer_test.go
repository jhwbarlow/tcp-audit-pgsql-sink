package main

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

var errNotImplemented error = errors.New("interface method not implemented")

type mockConn struct {
	txToReturn         pgx.Tx
	beginErrorToReturn error

	execCalled    bool
	closeCalled   bool
	beginCalled   bool
	configCalled  bool
	prepareCalled bool
}

func newMockConn(txToReturn pgx.Tx, beginErrorToReturn error) *mockConn {
	return &mockConn{
		txToReturn:         txToReturn,
		beginErrorToReturn: beginErrorToReturn,
	}
}

func (mc *mockConn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	mc.execCalled = true
	return nil, nil
}

func (mc *mockConn) Begin(ctx context.Context) (pgx.Tx, error) {
	mc.beginCalled = true

	if mc.beginErrorToReturn != nil {
		return nil, mc.beginErrorToReturn
	}

	return mc.txToReturn, nil
}

func (mc *mockConn) Config() *pgx.ConnConfig {
	mc.configCalled = true
	return nil
}

func (mc *mockConn) Close(ctx context.Context) error {
	mc.closeCalled = true
	return nil
}

func (mc *mockConn) Prepare(ctx context.Context, name, sql string) (sd *pgconn.StatementDescription, err error) {
	mc.prepareCalled = true
	return nil, nil
}

type mockTx struct {
	execErrorToReturn   error
	commitErrorToReturn error

	execCalled     bool
	commitCalled   bool
	rollbackCalled bool
}

func newMockTx(execErrorToReturn, commitErrorToReturn error) *mockTx {
	return &mockTx{
		execErrorToReturn:   execErrorToReturn,
		commitErrorToReturn: commitErrorToReturn,
	}
}

func (mt *mockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	panic(errNotImplemented)
}

func (mt *mockTx) BeginFunc(ctx context.Context, f func(pgx.Tx) error) (err error) {
	panic(errNotImplemented)
}

func (mt *mockTx) Commit(ctx context.Context) error {
	mt.commitCalled = true

	if mt.commitErrorToReturn != nil {
		return mt.commitErrorToReturn
	}

	return nil
}

func (mt *mockTx) Rollback(ctx context.Context) error {
	mt.rollbackCalled = true
	return nil
}

func (mt *mockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	mt.execCalled = true

	if mt.execErrorToReturn != nil {
		return nil, mt.execErrorToReturn
	}

	return nil, nil
}

func (mt *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	panic(errNotImplemented)
}

func (mt *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic(errNotImplemented)
}

func (mt *mockTx) LargeObjects() pgx.LargeObjects {
	panic(errNotImplemented)
}

func (mt *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	panic(errNotImplemented)
}

func (mt *mockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	panic(errNotImplemented)
}

func (mt *mockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	panic(errNotImplemented)
}

func (mt *mockTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	panic(errNotImplemented)
}

func (mt *mockTx) Conn() *pgx.Conn {
	panic(errNotImplemented)
}

func TestExecerCommitTxOnNoExecError(t *testing.T) {
	mockTx := newMockTx(nil, nil)
	mockConn := newMockConn(mockTx, nil)
	mockStmt1 := newSQLStatement("INSERT INTO foo (foo, bar) VALUES ($1, $2)", "foo", "bar")
	mockStmt2 := newSQLStatement("INSERT INTO bar (baz, bosh) VALUES ($1, $2)", "baz", "qux")

	execer := newPGXExecer(mockConn)

	if err := execer.execMultiple(context.TODO(), mockStmt1, mockStmt2); err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	if !mockConn.beginCalled {
		t.Error("expected conn Begin() to be called, but was not")
	}

	if !mockTx.commitCalled {
		t.Error("expected Tx Commit() to be called, but was not")
	}
}

func TestExecerRollbackTxOnExecError(t *testing.T) {
	mockError := errors.New("mock tx exec error")
	mockTx := newMockTx(mockError, nil)
	mockConn := newMockConn(mockTx, nil)
	mockStmt1 := newSQLStatement("INSERT INTO foo (foo, bar) VALUES ($1, $2)", "foo", "bar")
	mockStmt2 := newSQLStatement("INSERT INTO bar (baz, bosh) VALUES ($1, $2)", "baz", "qux")

	execer := newPGXExecer(mockConn)

	err := execer.execMultiple(context.TODO(), mockStmt1, mockStmt2)
	if err == nil {
		t.Error("expected error, got nil")
	}
	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}

	if !mockConn.beginCalled {
		t.Error("expected conn Begin() to be called, but was not")
	}

	if !mockTx.rollbackCalled {
		t.Error("expected Tx Rollback() to be called, but was not")
	}

	if mockTx.commitCalled {
		t.Error("expected Tx Commit() to not be called, but was")
	}
}

func TestExecerErrorOnCommitError(t *testing.T) {
	mockError := errors.New("mock tx commit error")
	mockTx := newMockTx(nil, mockError)
	mockConn := newMockConn(mockTx, nil)
	mockStmt1 := newSQLStatement("INSERT INTO foo (foo, bar) VALUES ($1, $2)", "foo", "bar")
	mockStmt2 := newSQLStatement("INSERT INTO bar (baz, bosh) VALUES ($1, $2)", "baz", "qux")

	execer := newPGXExecer(mockConn)

	err := execer.execMultiple(context.TODO(), mockStmt1, mockStmt2)
	if err == nil {
		t.Error("expected error, got nil")
	}
	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}

	if !mockConn.beginCalled {
		t.Error("expected conn Begin() to be called, but was not")
	}

	if !mockTx.commitCalled {
		t.Error("expected Tx Commit() to be called, but was not")
	}

	if mockTx.rollbackCalled {
		t.Error("expected Tx Rollback() to not be called, but was")
	}
}

func TestExecerErrorOnTxBeginError(t *testing.T) {
	mockError := errors.New("mock conn begin error")
	mockConn := newMockConn(nil, mockError)
	mockStmt1 := newSQLStatement("INSERT INTO foo (foo, bar) VALUES ($1, $2)", "foo", "bar")
	mockStmt2 := newSQLStatement("INSERT INTO bar (baz, bosh) VALUES ($1, $2)", "baz", "qux")

	execer := newPGXExecer(mockConn)

	err := execer.execMultiple(context.TODO(), mockStmt1, mockStmt2)
	if err == nil {
		t.Error("expected error, got nil")
	}
	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error chain to include %q, but did not", mockError)
	}

	if !mockConn.beginCalled {
		t.Error("expected conn Begin() to be called, but was not")
	}
}
