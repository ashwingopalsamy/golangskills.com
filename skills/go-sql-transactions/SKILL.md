---
name: go-sql-transactions
description: "Design, implement, or review Go SQL transaction boundaries, isolation, retries, connection pools, query cancellation, and ambiguous commits. Use for database/sql, pgx, repositories, migrations, or persisted idempotency. Do not use for query tuning alone, non-SQL stores, or broker acknowledgment semantics."
license: Apache-2.0
compatibility: "Go 1.24 or newer. Standard-library contracts are primary; driver and database behavior must be verified against repository versions."
---

# Go SQL transactions

Model the database boundary around durable invariants. A transaction is not a repository convenience: it is the unit in which concurrent executions must preserve state, and its externally visible effects determine whether retries are safe.

## Establish the invariant and boundary

Before editing, inspect the schema, constraints, transaction call path, driver, database version, migrations, and retry ownership. State:

1. the invariant that must survive concurrent execution;
2. the reads whose results justify writes;
3. the statements and side effects that must commit together;
4. the required isolation behavior and acceptable anomalies;
5. the idempotency identity for an externally retried operation;
6. what the caller may observe if commit outcome is unknown.

Keep the transaction at the use-case layer that knows this set. A repository method may execute statements, but splitting one invariant across independently committed repository calls is not atomic.

## Use the transaction handle exclusively

After `BeginTx`, execute transaction work through `*sql.Tx` (or the driver’s transaction handle), not the parent pool. Mixing `DB` calls into the transaction path can run them on another connection and outside the atomic unit.

Use a shape that makes conclusion explicit:

```go
tx, err := db.BeginTx(ctx, opts)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}
defer func() { _ = tx.Rollback() }()

if err := apply(ctx, tx, command); err != nil {
    return err
}
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit transaction: %w", err)
}
return nil
```

The deferred rollback is cleanup; it does not replace a commit error. Do not log and return the same error at this layer.

Read [references/transactions.md](references/transactions.md) for isolation, retries, savepoints, and commit ambiguity.

## Let the database enforce concurrency

Prefer database constraints and conditional writes over check-then-act logic in application memory:

- uniqueness for durable deduplication;
- foreign keys and checks for relational invariants when operationally acceptable;
- `UPDATE ... WHERE version = ?` for optimistic concurrency;
- row or advisory locks when serialized ownership is actually required;
- transaction isolation selected from documented database behavior.

Do not assume a transaction at the default isolation level prevents every write skew or lost-update schedule. Name the anomaly being prevented and verify the database’s semantics.

## Classify retries

Retry only when all of the following hold:

- the database identifies a transient/retryable class using a stable code or typed error;
- the whole transaction function can be replayed from fresh reads;
- non-database side effects are outside the retry loop or made idempotent/transactional;
- a bounded attempt and time budget exists;
- backoff and admission do not amplify overload.

Never retry an arbitrary `error`, string-match an error message, or retry only the last statement after a serialization failure. The database may require replay of the entire transaction.

Use `go-service-resilience` for the attempt budget and overload interaction.

## Treat commit errors as ambiguous when appropriate

If the connection fails while committing, the client may not know whether the server committed. The exact driver/database error taxonomy matters, but the general rule is:

- do not report “definitely not committed” unless the contract proves it;
- do not blindly replay a non-idempotent operation;
- reconcile by durable operation identity or read-after-reconnect;
- design externally retried commands so the same identity returns the original outcome.

This is especially important for payments, inventory, and message-driven state transitions.

## Keep external side effects out of the transaction

Do not hold a database transaction open across an HTTP/RPC call, broker publish, email, or slow computation unless a documented protocol requires it. Doing so extends locks and pool occupancy while still failing to make the external effect atomic.

For database state plus message publication, use a transactional outbox or another explicit consistency protocol. For inbound messages, combine durable deduplication and state mutation in one local transaction. Use `go-message-processing` for that workflow.

## Bound query and pool behavior

- Propagate the caller’s context through `BeginTx`, `QueryContext`, `ExecContext`, and related driver APIs.
- Call cancellation functions when deriving shorter deadlines.
- Close `Rows` and check `Rows.Err()` after iteration.
- Scan before advancing again; copy driver-owned byte slices if retaining them.
- Treat `sql.DB` as a concurrent pool handle, not a single connection.
- Tune pool limits from database capacity and service concurrency; observe `DB.Stats()` rather than copying constants.
- Account for the fact that `SetMaxOpenConns` is a semaphore: callers can wait, time out, or deadlock when code needs an additional connection while holding one.

Read [references/pools-and-queries.md](references/pools-and-queries.md) for pool and row-lifetime details.

## Migrations and compatibility

For schema changes, reason across mixed application versions and partial rollout:

1. expand with backward-compatible schema;
2. deploy code that tolerates both forms;
3. backfill in bounded, restartable batches;
4. switch reads/writes with observability;
5. contract only after old code and data paths are gone.

Do not place a large unbounded backfill in a startup migration or one long transaction. Database-specific online DDL behavior must come from current vendor documentation.

## Finish with evidence

Within the request’s authority, exercise:

- concurrent attempts at the invariant;
- duplicate idempotency identities with equal and conflicting payloads;
- serialization/deadlock retry from a fresh transaction;
- context cancellation while waiting for the pool and while executing;
- commit failure or ambiguous-result handling through a controllable seam;
- rollback and row-close paths.

Tests against mocks cannot prove database isolation or constraint behavior. Use the actual database/version for those claims, and report when that validation was not performed.

## Output contract

For implementation, make the durable invariant and transaction owner apparent in code. For review, cite the concrete interleaving or failure point, its data consequence, and the smallest correction. Do not prescribe a repository layer or ORM as a universal architecture.
