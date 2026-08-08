---
name: review-go-fintech-change
description: "Use as lead for fintech Go diff/PR review: money, payments, settlement, payment data, and audit. Do not use for general work."
license: Apache-2.0
compatibility: "Go 1.24 or newer; product, provider, rail, and compliance contracts control the review."
---

# Review a Go fintech change

Treat any path that can invent, duplicate, lose, misstate, or conceal money as financial-integrity sensitive.

## Trace the financial effect

Follow authenticated command, money representation, idempotency identity, state transition, provider attempt, ledger posting, event, settlement evidence, reconciliation, response, and audit record. Identify exact transaction boundaries and unknown-outcome windows.

For every caller, payment, provider, or replay identity, verify its authority scope and canonical payload binding. State all three behaviors explicitly: equivalent input replays the same outcome; different input is rejected before a new effect or prior-result disclosure; wrong authority learns nothing. Saying an identity is “bound to a fingerprint” is not an enforceable finding unless the mismatch path is defined.

Load the focused skills that own the changed invariant. Common compositions are payment lifecycle + financial idempotency + data consistency for money-moving mutations; settlement + money/ledger + message processing for replayed adjustments; and fintech security/compliance + security hardening for payment-data code with an implementation exploit.

## Critical schedules

- concurrent same-key and different-payload requests;
- timeout after provider or database may have committed;
- duplicate, delayed, and out-of-order webhooks;
- partial capture/refund and cumulative amount races;
- rounding residual across allocations and currencies;
- unbalanced or mutable journal history;
- report re-ingestion and duplicate adjustment;
- late chargeback or settlement event after local terminal state;
- cross-tenant replay, excessive privilege, or sensitive data in telemetry;
- mixed versions during schema and state-machine rollout.

## Findings

Zero tolerance: silent precision loss, unbalanced committed entries, duplicate financial effects, illegal state transitions presented as success, missing reconciliation evidence, or sensitive authentication data leakage. Tie every finding to a reachable schedule and authoritative invariant. Do not treat provider documentation as universal across rails.

Read [references/financial-finding-standard.md](references/financial-finding-standard.md) before finalizing.

## Output contract

Lead with critical financial-integrity findings, then security/compliance and operational risks. State when scheme, legal, or compliance interpretation requires a qualified owner; never infer certification.
