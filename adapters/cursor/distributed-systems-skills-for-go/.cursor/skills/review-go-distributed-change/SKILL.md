---
name: review-go-distributed-change
description: "Use for distributed Go diff/PR review. Do not use when fintech correctness dominates."
license: Apache-2.0
compatibility: "Go 1.24 or newer; repository and deployed-system guarantees control the review."
---

# Review a Go distributed change

Find the failure schedule that violates a system invariant.

## Build the state/effect graph

Trace admission, local reads, decisions, durable writes, remote calls, publication, acknowledgement, response, and recovery. Mark transaction boundaries, goroutine owners, retry owners, ordering keys, leases, queues, and version transitions.

Load the focused distributed skills for the changed path; this review skill coordinates findings rather than replacing their protocol and storage references.

## Test causal schedules

- concurrent same/conflicting identities;
- cancellation before admission, during work, and after a durable effect;
- crash immediately before and after commit, publish, and acknowledgement;
- timeout with unknown remote result;
- redelivery, duplication, delay, and reordering;
- retry at several layers under dependency overload;
- lease expiry while an owner is paused;
- mixed versions during deploy and rollback;
- recovery with cold caches, reconnects, and queued retries.

Use only schedules reachable in the changed path. Quantify maximum goroutines, queue entries, attempts, held connections, and duplicate effects where possible.

## Close every effect boundary

Before finalizing, account separately for database commit, remote call, publication, and acknowledgement when they appear. For each, state whether a lost response leaves the outcome unknown and name the stable identity or authoritative read that makes replay safe. A transactional outbox does not by itself explain a failed or unknown consumer acknowledgement.

When retries or redeliveries exist at several layers, compute the configured maximum leaf attempts as their product; if a bound is unknown, leave it symbolic rather than inventing one. Assign retry ownership to one bounded layer, enforce one end-to-end deadline, and distinguish transient faults from permanent or ambiguous outcomes.

## Completeness gate

Do not emit the final review while either applicable disposition is only implied:

- If consumer acknowledgement appears, state when it becomes safe, how a failed or unknown acknowledgement causes replay, and which durable outcome makes that replay harmless. Scope any exactly-once statement to the boundary that actually provides it.
- If two or more layers can retry or redeliver, state the product bound and a repair that selects one retry owner, one end-to-end deadline, retryable error classes, and reconciliation for ambiguous outcomes. A numeric bound without the ownership repair is incomplete.

## Findings

Tie each finding to a changed line and include invariant, trigger schedule, state consequence, and smallest correction. Do not recommend retries without replay safety, locks without an authoritative boundary, or “exactly once” without naming its scope.

Read [references/schedule-catalog.md](references/schedule-catalog.md) before finalizing.

## Output contract

Lead with correctness and availability findings. Route monetary, ledger, settlement, or compliance consequences to `review-go-fintech-change` for domain adjudication.

Audit the final prose—not only the analysis—against every applicable completeness-gate clause. Add any missing clause to the nearest causal finding; do not rely on an inbox/outbox recommendation or attempt count to imply acknowledgement replay, retry ownership, deadline, or error classification.
