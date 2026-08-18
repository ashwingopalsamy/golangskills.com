# Transaction decisions

## Isolation

Map the invariant to prohibited histories before selecting isolation. Names such as “repeatable read” and “serializable” have database-specific behavior. Verify the deployed database documentation and driver support.

Common mechanisms:

- a unique constraint turns concurrent duplicate insertion into one winner;
- optimistic version columns reject stale updates without holding a lock during user work;
- `SELECT ... FOR UPDATE` protects rows that exist, but does not necessarily protect absence or arbitrary predicates;
- serializable isolation can protect broader invariants but requires whole-transaction retry on serialization failure.

## Transaction function shape

Pass a narrow executor interface only when it reflects existing repository design; do not invent one solely to make `DB` and `Tx` interchangeable. The calling use case should own `Begin`, `Commit`, and retry. Helpers inside it should accept the transaction handle or an established query interface.

## Savepoints

Savepoints provide partial rollback inside a transaction, not independent durability. They are database- and driver-sensitive and do not make external side effects transactional. Use only when the partial-failure semantics are part of the business operation.

## Commit ambiguity

A client can lose the connection after the database made a commit durable but before the acknowledgment arrived. Retrying with a new operation identity can duplicate effects. A durable idempotency record should bind one operation identity to a request fingerprint and terminal outcome, inside the same transaction as the effect.

If the same key arrives with a different semantic payload, reject it rather than returning the first result for a different command.
