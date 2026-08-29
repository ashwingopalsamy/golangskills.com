# OpenAI plugin submission dossier

Status: prepared for the OpenAI Platform submission portal. These are skills-only plugins with no MCP server, authentication, remote tools, or publisher-operated data processing.

Publisher: Ashwin Gopalsamy  
Developer website: https://ashwingopalsamy.in  
Product website: https://github.com/ashwingopalsamy/golangskills.com  
Support: https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/support.md  
Privacy: https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/privacy.md  
Terms: https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/terms.md  
Category: Developer Tools  
Availability: Worldwide, subject to OpenAI platform availability and policy.

## Engineering Skills for Go

Short description: Engineer and review production Go software with evidence-backed guidance.

Long description: Focused Go skills for language semantics, project and API design, service boundaries, testing, diagnostics, security hardening, production operations, and systematic change review. Guidance is conditional on explicit invariants and links claims to primary sources.

Starter prompts:

- Review this Go change for correctness, API compatibility, lifecycle ownership, and missing verification evidence.
- Design a cancellable Go HTTP service boundary with bounded request bodies and graceful shutdown.
- Diagnose this Go latency regression using measurements before recommending optimizations.

Positive test cases:

1. Ask for a review of an exported Go API change; expect compatibility and ownership analysis.
2. Ask for an HTTP service design; expect request, response, deadline, and shutdown boundaries.
3. Ask how to verify concurrent Go code; expect race, fuzz, deterministic-time, and leak evidence chosen by risk.
4. Provide a measured allocation regression; expect a profiling-led diagnostic procedure.
5. Ask for production shutdown design; expect admission stop, drain, bounded flush, and forced termination handling.

Negative test cases:

1. Ask for a React component; no Go skill should activate.
2. Ask for legal interpretation of a contract; the plugin should not claim legal authority.
3. Ask for speculative micro-optimization without measurements; it should request evidence rather than invent a bottleneck.

## Distributed Systems Skills for Go

Short description: Design failure-safe distributed Go systems from explicit invariants.

Long description: Focused Go skills for concurrency lifecycle, data consistency, message processing, resilience, coordination, and distributed-change review. They reason about retries, ambiguity, ordering, fencing, overload, recovery, and transport cutovers without promising impossible exactly-once behavior.

Starter prompts:

- Review this Go consumer for redelivery, ordering, offset-commit ambiguity, and retry amplification.
- Design a lease-based worker whose stale owner cannot commit after losing ownership.
- Diagnose a live transport cutover that leaks or reuses obsolete idle connections.

Positive test cases:

1. Ask for a bounded worker pool; expect cancellation, admission, draining, and goroutine ownership.
2. Ask about a transaction plus message publish; expect atomicity-gap and outbox/inbox reasoning.
3. Ask for retry design; expect deadline budget, backoff, jitter, amplification, and terminal classification.
4. Ask for leader leases; expect fencing and authoritative commit validation.
5. Ask for runtime transport reload; expect atomic cutover plus retirement of obsolete idle connections.

Negative test cases:

1. Ask for CSS layout help; no distributed Go skill should activate.
2. Ask for an unconditional exactly-once guarantee; the plugin should reject the premise and define observable invariants.
3. Ask for unbounded retry-until-success; the plugin should identify overload and deadline failure modes.

## Fintech Skills for Go

Short description: Build financially correct Go payment and ledger systems.

Long description: Focused Go skills for exact money representation, double-entry ledgers, payment lifecycles, replay-safe financial operations, clearing, settlement, reconciliation, security boundaries, compliance-aware design, and financial-change review.

Starter prompts:

- Design a Go ledger posting path that preserves balance, immutability, and replay safety.
- Review this authorization, capture, refund, and reversal state machine for ambiguous outcomes.
- Design settlement reconciliation that detects missing, duplicate, late, and conflicting records.

Positive test cases:

1. Ask how to represent multi-currency amounts; expect explicit currency, scale, rounding, and overflow rules.
2. Ask for ledger posting design; expect balanced entries, immutable posting, unique business identity, and correction transactions.
3. Ask for payment idempotency; expect parameter binding, atomic ownership, stored outcome replay, and ambiguity handling.
4. Ask for capture/refund lifecycle design; expect state-machine eligibility and provider reconciliation.
5. Ask for settlement close; expect source manifests, control totals, unmatched-item queues, and auditable close/reopen policy.

Negative test cases:

1. Ask for investment advice; the plugin should not present itself as an investment adviser.
2. Ask to store card verification values after authorization; the plugin should reject retention of sensitive authentication data.
3. Ask to silently delete ledger history to fix a balance; the plugin should require compensating or correcting entries.

## Evidence boundary

The public listing must not claim category leadership, universal correctness, or cross-agent behavioral superiority. Structural, provenance, packaging, reference-inventory, and discovery-budget evidence is public. The locked release-candidate superiority benchmark remains pending.
