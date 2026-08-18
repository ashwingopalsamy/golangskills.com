# Pools and query lifetimes

## Pool sizing

`sql.DB` creates and reuses connections. `SetMaxOpenConns` bounds database concurrency but also creates a wait queue. Size it as part of a global database connection budget across replicas, jobs, migrations, and administrative clients.

Observe:

- `OpenConnections`, `InUse`, and `Idle`;
- `WaitCount` and `WaitDuration`;
- query latency separated from pool wait;
- database saturation and lock waits.

Increasing the pool can reduce client wait while overloading the database. Reducing it can stabilize the database while increasing request timeouts. Tune the whole queueing system.

## Rows

Close `Rows` promptly. Iteration can end because the context was canceled or the driver failed after earlier rows succeeded, so check `Rows.Err()`. `QueryRow` defers errors until `Scan`.

Avoid `defer rows.Close()` inside an unbounded loop at the outer function scope; extract one iteration so the defer runs per query.

## Prepared statements

Prepared-statement behavior differs across drivers and pool connections. Use it for semantics or measured performance, not as a ritual. A statement prepared on `DB` may prepare per underlying connection; a statement on `Tx` is transaction-bound.

## Dedicated connections

Use `DB.Conn` only when a sequence must use the same physical connection and a transaction is not the right contract. Close it to return it to the pool. Session state can leak across pooled use unless reset according to the driver contract.
