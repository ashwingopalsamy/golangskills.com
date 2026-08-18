# Coverage and next decisions

V1 is intentionally a production-failure-boundary set, not a complete Go curriculum. The collection currently covers concurrent execution, HTTP resource ownership, SQL transaction semantics, asynchronous message effects, distributed-call resilience, and risk-driven change review.

## Highest-leverage candidate skills

A candidate is implemented only after paired agent evaluations show positive marginal value over the base model and existing skills.

1. **Go production diagnostics:** correlate logs, metrics, traces, profiles, runtime signals, and recent changes into a falsifiable incident investigation. This is the clearest missing workflow because no v1 skill owns evidence-driven diagnosis.
2. **Go security boundaries:** review trust boundaries, authorization, injection, SSRF, path handling, secrets, cryptographic API use, and dependency exposure. This needs focused threat-model evals to avoid becoming a generic checklist.
3. **Go schema and migration safety:** expand beyond transaction-local reasoning into online schema evolution, compatibility windows, backfills, and rollback. It should remain separate from `go-sql-transactions` only if activation tests distinguish them reliably.
4. **Go performance investigation:** measurement-first profiling across CPU, allocation, contention, blocking, and GC. It should teach causal diagnosis rather than micro-optimization patterns.

Cloud-native deployment, gRPC, observability instrumentation, testing syntax, package layout, and language-feature summaries remain deferred. Most are either repository-specific, handled adequately by base models, or too broad to justify discovery cost without stronger eval evidence.

## Promotion and release

- `beta` to `stable`: paired evaluations on at least two current agent/model combinations, no negative median quality delta, and no unresolved high-severity guidance defect.
- source refresh: re-verify source claims within 400 days and whenever a referenced Go release family leaves support.
- collection release: regenerate catalogs and plugin packages, run the full CI matrix, record eval conditions and raw artifacts, then tag the repository with the collection semantic version.

The future golangskills.com experience should consume `catalog/catalog.json`; it should not scrape Markdown or introduce a second authored taxonomy.
