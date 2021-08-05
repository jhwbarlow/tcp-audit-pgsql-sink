package main

import (
	"context"

	"github.com/jackc/pgconn"
)

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type conn interface {
	execer
}

type pgxConn struct {

}
