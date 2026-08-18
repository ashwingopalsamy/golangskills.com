---
name: go-money-and-ledgers
description: "Use for Go money, currency, rounding, allocation, journals, balances, and reversals. Do not use for provider states."
license: Apache-2.0
compatibility: "Go 1.24 or newer; currency and accounting rules require current product and jurisdiction evidence."
---

# Go money and ledgers

Money is a typed exact value. A ledger is immutable evidence, not a mutable balance table with an audit log added later.

## Represent money

Carry amount and currency together. Use integer minor units when the product contract fits one exponent; use exact decimal with explicit scale when it does not. Never use binary floating point for postings or settlement amounts. Validate currency, scale, range, and sign at the boundary.

Rounding is a named business operation: specify mode, scale, stage, and residual allocation. Do not round intermediate values merely for display. Allocate remainders deterministically so parts sum exactly to the source amount.

## Post a journal

Persist immutable entries with journal ID, account, currency, signed side/amount, effective and recorded time, operation identity, and source evidence. Enforce per-journal, per-currency balance in the same transaction. Use reversal or adjustment journals rather than editing history.

Distinguish pending, posted, and available balances if the product needs them. Derive balances from entries or maintain a projection transactionally coupled to entries and rebuildable from them.

## Concurrency and audit

Serialize or conditionally update account/version state where overdraft, limits, or sequence matter. A balanced journal can still be unauthorized or economically wrong; preserve the business command and approval evidence. Avoid sensitive payloads while keeping immutable actor, time, reason, and correlation.

Read [references/money-ledger-invariants.md](references/money-ledger-invariants.md) for arithmetic and schema checks.

## Output contract

State exact representation, rounding, journal balance, mutation policy, and concurrency invariant. Zero tolerance for silent precision loss or unbalanced committed entries.
