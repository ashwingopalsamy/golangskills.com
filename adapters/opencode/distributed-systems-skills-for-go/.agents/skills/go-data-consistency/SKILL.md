---
name: go-data-consistency
description: "Use for Go transactions, isolation, constraints, ambiguous commits, caches, and migrations. Do not use for brokers."
license: Apache-2.0
compatibility: "Go 1.24 or newer; database, driver, and cache semantics must be verified against deployed versions."
---

# Go data consistency

Start from the invariant and legal concurrent outcomes. A transaction is a correctness boundary, not a repository convenience.

## Define the durable unit

Identify reads that justify writes, constraints, isolation behavior, transaction owner, external effects, idempotency identity, and what a caller sees when commit outcome is unknown. Keep all statements protecting one invariant on the same transaction handle.

## Enforce close to state

Prefer unique, foreign-key, check, exclusion, or conditional-write constraints where the database can enforce the invariant atomically. Use row locks, advisory locks, optimistic versions, or serializable isolation only after naming the anomaly they prevent. Isolation labels differ across engines.

After `BeginTx`, use the transaction handle exclusively. Defer rollback as cleanup, return commit errors, close rows, check iteration errors, propagate context, and remember `sql.DB` is a concurrent pool whose limits can queue and time out callers.

## Retry and ambiguity

Replay the whole transaction from fresh reads only for stable retryable error classes, within a bounded budget, with external side effects excluded or independently idempotent. A connection failure during commit can leave the outcome unknown. Resolve by durable operation identity or reconciliation; never report a definite rollback without evidence.

## Cache consistency

State whether the cache is authoritative, derived, or optional. Define invalidation/order behavior, stale-read tolerance, stampede control, and recovery after cache loss. Do not create a correctness dependency on best-effort invalidation without version or write-through semantics.

## Migrations

Use expand, mixed-version deployment, bounded restartable backfill, switchover, and delayed contract. Verify old writer/new reader, new writer/old reader, rollback, and partial progress.

Read [references/isolation-and-ambiguity.md](references/isolation-and-ambiguity.md) for counterexamples.

## Output contract

Name the invariant, transaction boundary, permitted anomalies, and ambiguous-outcome policy. For review, show the violating interleaving and smallest repair.
