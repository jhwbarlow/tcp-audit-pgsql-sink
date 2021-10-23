# tcp-audit-pgsql-sink

This module implements a `tcp-audit` Sinker plugin which stores TCP state change events in PostgreSQL database.

One table is used to store the TCP state change events, and another stores related socket information if the Eventer plugin being used supports supplying this information.

## Database schema

The schema of the main state change events table is:

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

The schema of the socket information table is:

```sql
TABLE tcp_events_socket_info (
	uid           TEXT PRIMARY KEY,
	tcp_event_uid TEXT,
	id            TEXT,
	inode         INTEGER,
	user_id       INTEGER,
	group_id      INTEGER,
	state         TEXT,	
	CONSTRAINT fk_tcp_events FOREIGN KEY(tcp_event_uid) 
		REFERENCES tcp_events(uid) ON DELETE CASCADE
)
```

For example:

```
                 uid                  |            tcp_event_uid             |        id        |  inode  | user_id | group_id |    state    
--------------------------------------+--------------------------------------+------------------+---------+---------+----------+-------------
 4c149711-fb0f-41ea-a323-14d276f92988 | 0f2cbe68-099a-49af-a3d4-a938c22c7a37 | ffff9e45710b3d40 | 1718963 |       0 |        0 | UNCONNECTED
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