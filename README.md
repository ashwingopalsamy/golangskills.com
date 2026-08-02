# Go Engineering Skills by Ashwin Gopalsamy

Evidence-backed Agent Skills for production Go, distributed systems, and fintech—authored by [Ashwin Gopalsamy](https://ashwingopalsamy.in).

The ecosystem ships three non-overlapping installable collections:

| Collection | Skills | Owns |
|---|---:|---|
| [Engineering Skills for Go](docs/collections/engineering.md) | 8 | Language, APIs, service boundaries, tests, performance, security, operations, general review |
| [Distributed Systems Skills for Go](docs/collections/distributed-systems.md) | 6 | Concurrency, consistency, messaging, resilience, coordination, distributed review |
| [Fintech Skills for Go](docs/collections/fintech.md) | 6 | Money, ledgers, payment lifecycles, idempotency, settlement, reconciliation, compliance, fintech review |

Install one, two, or all three. No skill appears in more than one collection.

## Why this is different

- 20 focused skills with progressive one-hop references, not a monolithic prompt.
- 7,746 conservative discovery characters for all three collections, including a 160-character path allowance per skill; every collection is below 4,000.
- 80 development scenarios covering routing, confusion negatives, and deterministic quality criteria, including 12 executable failure-mode fixtures.
- 21 primary-evidence claims with scope, qualifications, counterexamples, Go versions, owners, and provenance.
- All six local Go reference snapshots locked by remote, commit, license, 729 file hashes, 98 canonical skill entrypoints, and 10,165 material-item dispositions.
- One canonical corpus generates Codex, Claude Code, Cursor, OpenCode, catalog, search, site, and `llms.txt` artifacts.
- Published skills contain no executable resources or portable tool grants.

These are verified structural properties, not a claim of empirical superiority. `catalog/benchmark-status.json` remains `evidence-pending`, and `skillctl release check` refuses a category-leadership release until the benchmark gates pass.

## Install

See [docs/install.md](docs/install.md) for client-specific layouts. The repository marketplace contains:

- `engineering-skills-for-go`
- `distributed-systems-skills-for-go`
- `fintech-skills-for-go`

The canonical root also supports selected-skill installation through open Agent Skills tooling:

```text
npx skills add ashwingopalsamy/golangskills.com
```

## Engineer the corpus

```text
go run ./cmd/skillctl audit refs -refs /path/to/go-refs
go run ./cmd/skillctl check
go run ./cmd/skillctl generate
go run ./cmd/skillctl package
go run ./cmd/skillctl eval preflight
go run ./cmd/skillctl eval run -runner codex -arm ours -kind routing -output evaluations/runs/ours-routing.jsonl
go run ./cmd/skillctl eval score -input evaluations/runs/ours-routing.jsonl -output evaluations/scores/ours-routing.jsonl
go run ./cmd/skillctl eval report -input evaluations/scores/ours-routing.jsonl
go run ./cmd/skillctl eval report -input evaluations/scores/ours-routing.jsonl -against evaluations/scores/competitor-routing.jsonl
go run ./cmd/skillctl release check
```

`check` validates schema v2, claim ownership and exposed evidence, complete reference disposition coverage, source freshness, discovery budgets, activation overlap, fixture paths, links, relations, licenses, and generated freshness. The eval harness creates a new isolated directory and ephemeral session per cell, randomizes order, uses opaque arm labels, resumes JSONL artifacts, scores deterministic graders first, and reports Wilson confidence intervals, routing macro-F1, unrelated-prompt false activation, collection scores, token efficiency, paired bootstrap intervals, and Pareto relations. Fixture cells preserve edited Go sources and allowlisted `go test` output. Competitor routing requires a committed `-routing-map`; explicit competitor runs require a committed `-skill-map`.

Frozen local competitor mappings live under `evaluations/arms/`. A per-case routing map can override the canonical-to-arm skill map; otherwise the harness derives accepted arm-local IDs and accepts their isolated plugin namespace prefix.

## Evidence boundary

Codex is the primary behaviorally validated client. Other generated adapters have structural compatibility evidence; cross-client behavioral claims require equivalent authenticated runners. See [docs/compatibility.md](docs/compatibility.md).

## Attribution and independence

Canonical prose is original writing derived from primary sources. Competitor text and code are not copied. Reference licenses and policies are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) and [research/corpus-lock.json](research/corpus-lock.json).

This independent project is not affiliated with or endorsed by Google or the Go project. “Go” and the Go gopher are trademarks of Google LLC; no Go logo is used.

## License

Apache-2.0.
