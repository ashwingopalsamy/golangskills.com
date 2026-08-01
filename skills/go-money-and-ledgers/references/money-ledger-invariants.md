# Money and ledger invariants

- Currency is explicit and validated against the product's supported set.
- Scale and rounding mode are named at every lossy boundary.
- Allocation parts sum exactly to the input and residual assignment is deterministic.
- A committed journal nets to zero per currency under the ledger's sign convention.
- Posted history is immutable; corrections are new linked entries.
- Operation identity and source evidence permit replay detection and audit.
- Balance projections reconcile to journals and can be rebuilt.

ISO 4217 minor-unit metadata is necessary but may not be sufficient for product-specific precision or cash rounding.
