# tcp-audit-pgsql-sink

This module implements a `tcp-audit` Sinker plugin which stores TCP state change events in PostgreSQL database, using a single table.

## Database schema

The schema of the table is:

```sql
TABLE tcp_events (
	uid         TEXT PRIMARY KEY,
	timestamp   TIMESTAMP,
	pid_on_cpu  INTEGER,
	comm_on_cpu TEXT,
	src_ip      INET,
	dst_ip      INET,
	src_port    INTEGER,
	dst_port    INTEGER,
	old_state   TEXT,
	new_state   TEXT
)
```

For example:

```
                 uid                  |         timestamp          | pid_on_cpu |  comm_on_cpu   |   src_ip    |     dst_ip      | src_port | dst_port |  old_state  |  new_state  
--------------------------------------+----------------------------+------------+----------------+-------------+-----------------+----------+----------+-------------+-------------
 5411c87d-5096-494b-ba3f-a006f69f1897 | 2021-08-31 22:46:05.428515 |      31615 | kworker/u8:2   | 192.168.1.3 | 172.217.16.225  |    58248 |      443 | FIN-WAIT-2  | CLOSED
```

## Configuration

This module requires configuration via environment variables in order to connect to the database.

The following environment variables are supported:

- `PGHOST` (required)
- `PGPORT` (optional, defaults to 5432)
- `PGDATABASE` (required)
- `PGUSER` (required)
- `PGPASSWORD` (required)

See the [PostgreSQL documentation](https://www.postgresql.org/docs/current/libpq-envars.html) for an explanation of these variables.