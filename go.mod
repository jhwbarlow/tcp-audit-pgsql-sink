module github.com/jhwbarlow/tcp-audit-pgsql-sink

go 1.15

replace github.com/jhwbarlow/tcp-audit => ../tcp-audit/src

require (
	github.com/google/uuid v1.3.0
	github.com/jackc/pgconn v1.10.0
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jackc/pgx/v4 v4.13.0
	github.com/jhwbarlow/tcp-audit v0.0.0
)
