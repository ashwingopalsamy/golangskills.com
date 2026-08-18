---
name: go-financial-idempotency
description: "Use for Go financial idempotency, key scope, fingerprints, replay, retention, and ambiguity. Do not use for read caching."
license: Apache-2.0
compatibility: "Go 1.24 or newer; external idempotency guarantees must match the active provider contract."
---

# Go financial idempotency

Idempotency means equivalent retries under one authenticated operation identity produce at most one committed financial effect and a consistent outcome.

## Define identity and scope

Scope the key by tenant/principal, operation type, and target resource. Bind it to a canonical fingerprint of semantic input. Reject same-key different-input reuse. Never put PII or secrets in the key.

## Serialize first use

Use a durable unique constraint or equivalent atomic claim. Store state such as processing, succeeded, failed-final, and ambiguous; include response/result evidence and canonical fingerprint. Check-then-insert without a constraint races. A process-local mutex or cache cannot survive replicas and crashes.

## Couple effect and record

When local state is the effect, write idempotency record, domain transition, ledger journal, and stored result in one transaction. For an external provider, persist the attempt before calling, reuse the same provider identity, and reconcile ambiguity before issuing a new operation.

## Retention and replay

Retain records for the maximum credible client, webhook, and operational replay window. Define behavior after expiry explicitly. Do not prune while in-flight or ambiguous operations can still resolve. Return the stored outcome only if authorization and current disclosure policy allow it.

Read [references/idempotency-record.md](references/idempotency-record.md) for a state model.

## Output contract

State key scope, fingerprint, concurrency control, atomic boundary, result replay, ambiguity, and retention. Treat mismatched reuse and duplicate financial effects as critical failures.
