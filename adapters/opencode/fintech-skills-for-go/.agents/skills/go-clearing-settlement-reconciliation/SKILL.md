---
name: go-clearing-settlement-reconciliation
description: "Use for Go clearing, settlement, payouts, statements, and item reconciliation. Do not use for payment authorization."
license: Apache-2.0
compatibility: "Go 1.24 or newer; rail cutoffs, calendars, reports, and finality come from current provider and scheme contracts."
---

# Go clearing, settlement, and reconciliation

Do not collapse payment acceptance, clearing, settlement, payout, and bank movement into one success flag. They are separate evidence domains with different timing and finality.

## Preserve evidence

Ingest external reports and statements immutably with source identity, version, checksum, retrieval time, business date, and parser version. Make re-ingestion idempotent. Preserve raw evidence under access and retention policy; derive normalized items separately.

## Match at item level

Match stable external/internal identity first, then amount, currency, lifecycle, account, and time window. Each item must match exactly once or remain in an explicit exception state: missing internal, missing external, duplicate, amount/currency mismatch, status mismatch, or timing break.

Aggregate totals are controls, not sufficient reconciliation. Equal and opposite item errors can net to zero.

## Adjust deliberately

Age timing breaks by rail calendars and cutoffs. Separate automated detection from authorized remediation. Adjust with linked immutable ledger entries; never rewrite the original payment or report. Reruns must not duplicate adjustments or erase prior decisions.

## Recovery and operations

Use restartable batches, checkpoints, deterministic matching, bounded concurrency, and work queues with ownership. Publish completeness, unmatched amount/count, age, duplicate count, and source freshness without leaking sensitive data.

Read [references/reconciliation-model.md](references/reconciliation-model.md) for stages and controls.

## Output contract

State source evidence, matching key, timing policy, exception taxonomy, adjustment authorization, and replay behavior. Missing or silently auto-matched money is a critical defect.
