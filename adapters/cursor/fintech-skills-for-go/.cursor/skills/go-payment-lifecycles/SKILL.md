---
name: go-payment-lifecycles
description: "Use for Go payment authorization, capture, refund, disputes, webhooks, and ambiguity. Do not use for ledger arithmetic."
license: Apache-2.0
compatibility: "Go 1.24 or newer; provider, rail, and scheme state semantics are version- and region-specific."
---

# Go payment lifecycles

Model provider evidence as events and local payment state as a constrained projection. A network result is not automatically a financial result.

## Define the machine

List states, legal commands, provider requests, provider evidence, terminality, amount constraints, expiry, and late-event behavior. Keep authorization, capture, refund, reversal, dispute, and settlement distinct even if the UI says “paid.”

## Handle ambiguous outcomes

A timeout or lost response after request transmission can mean success, failure, or still processing. Persist the attempt and stable provider identity, replay only under the provider's idempotency contract, query authoritative state, and reconcile asynchronous evidence. Never create a new payment identity merely because the response was missing.

## Apply events safely

Verify authenticity before parsing into trusted events. Deduplicate by provider event or operation identity, but make handlers tolerant of out-of-order evidence. Apply transitions conditionally against current version/state. Quarantine impossible transitions with full non-sensitive evidence rather than silently dropping them.

## Amount and terminality

Enforce cumulative capture and refund limits with exact money. Partial capture/refund and multiple disputes can coexist. “Succeeded” may be locally terminal for fulfillment while chargebacks and settlement adjustments remain possible.

Read [references/payment-machine.md](references/payment-machine.md) for transition and ambiguity rules.

## Output contract

Provide the state/event table, authoritative evidence, ambiguity policy, and illegal-transition handling. Do not compress the design into booleans or infer provider success from HTTP status alone.
