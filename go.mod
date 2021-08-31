module github.com/jhwbarlow/tcp-audit-pgsql-sink

go 1.15

//replace github.com/jhwbarlow/tcp-audit-common => ../tcp-audit-common

require (
	github.com/google/uuid v1.3.0
	github.com/jackc/pgconn v1.10.0
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jackc/pgx/v4 v4.13.0
	github.com/jhwbarlow/tcp-audit-common v0.0.0-20210831195703-56b4e4c3ea54
)
