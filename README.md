# Production Go skills for AI coding agents

Current implementation status and benchmark evidence are tracked in the public catalog and evaluation reports.

**Go: Production Engineering**, **Go: Distributed Systems**, and **Go: Fintech** are evidence-backed Agent Skills authored by [Ashwin Gopalsamy](https://ashwingopalsamy.in). Install them in ChatGPT, Codex, Claude Code, Cursor, OpenCode, or through the open `npx skills` workflow.

The ecosystem ships three non-overlapping installable collections:

| Collection | Skills | Owns |
|---|---:|---|
| [Go: Production Engineering](docs/collections/engineering.md) | 8 | Language, APIs, service boundaries, tests, performance, security, operations, general review |
| [Go: Distributed Systems](docs/collections/distributed-systems.md) | 6 | Concurrency, consistency, messaging, resilience, coordination, distributed review |
| [Go: Fintech](docs/collections/fintech.md) | 6 | Money, ledgers, payment lifecycles, idempotency, settlement, reconciliation, compliance, fintech review |

Install one, two, or all three. No skill appears in more than one collection.

## Why this is different

- 20 focused skills with progressive one-hop references.
- 7,774 conservative discovery characters for all three collections, including a 160-character path allowance per skill; every collection is below 4,000.
- 133 development scenarios covering routing, confusion negatives, and deterministic quality criteria, including 15 executable failure-mode fixtures.
- 27 primary-evidence claims and 121 source records with scope, qualifications, counterexamples, Go versions, owners, and provenance.
- All seven local Go reference snapshots locked by remote, commit, license, 759 file hashes, 103 canonical skill entrypoints, and 10,239 material-item dispositions.
- One canonical corpus generates Codex, Claude Code, Cursor, OpenCode, catalog, search, site, and `llms.txt` artifacts.
- Published skills contain no executable resources or portable tool grants.

Structural verification covers the properties listed here. Empirical superiority remains pending. `catalog/benchmark-status.json` remains `evidence-pending`, and `skillctl release check` refuses a category-leadership release until the benchmark gates pass.

## Install

See [docs/install.md](docs/install.md) for client-specific layouts and [docs/npm-publishing.md](docs/npm-publishing.md) for release setup. The repository marketplace contains:

- `engineering-skills-for-go`
- `distributed-systems-skills-for-go`
- `fintech-skills-for-go`

The planned public npm organization is `@golangskills`. The data-only packages are staged and validated locally but have not been bootstrap-published, so the following commands become available only after that release step:

```sh
npm install @golangskills/engineering
npm install @golangskills/distributed-systems
npm install @golangskills/fintech
```

Each npm package is data-only and contains the generated client layouts. See the package README for the selected collection and [docs/install.md](docs/install.md) for copy and plugin paths.

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
go run ./cmd/skillctl npm package -version 0.4.0
go run ./cmd/skillctl npm check -version 0.4.0
go run ./cmd/skillctl eval preflight
go run ./cmd/skillctl eval verify-freeze -freeze evaluations/releases/rc1.lock.json -public-only
go run ./cmd/skillctl eval run -runner codex -arm ours -kind routing -output evaluations/runs/ours-routing.jsonl
go run ./cmd/skillctl eval matrix -runner codex -model gpt-5.6-sol -fixtures-only -output evaluations/runs/fixture-matrix.jsonl
go run ./cmd/skillctl eval score -input evaluations/runs/ours-routing.jsonl -output evaluations/scores/ours-routing.jsonl
go run ./cmd/skillctl eval score -input evaluations/runs/quality-matrix.jsonl -judgments evaluations/judgments/quality-matrix-codex.jsonl -judge-model gpt-5.6-sol -output evaluations/scores/quality-matrix-semantic.jsonl
go run ./cmd/skillctl eval report -input evaluations/scores/ours-routing.jsonl
go run ./cmd/skillctl eval report -input evaluations/scores/ours-routing.jsonl -against evaluations/scores/competitor-routing.jsonl
go run ./cmd/skillctl release check
```

`check` validates schema v2, claim ownership and exposed evidence, complete reference disposition coverage, source freshness, discovery budgets, activation overlap, fixture/oracle paths, links, relations, licenses, and generated freshness. The eval harness creates a new isolated directory and ephemeral session per cell, globally randomizes mixed-arm matrices, uses distinct opaque arm labels, resumes by mode and repetition, scores deterministic graders first, and reports Wilson confidence intervals, routing macro-F1, unrelated-prompt false activation, collection scores, token efficiency, paired bootstrap intervals, and Pareto relations. Fixture cells hide the decisive test until the agent exits, then preserve edited Go sources, the injected oracle, and allowlisted offline `go test` output. Non-executable quality cases can use schema-constrained, identity-redacted, resumable semantic judgments; Codex-on-Codex results are labeled same-platform rather than independent. Competitor routing requires a committed `-routing-map`; explicit competitor runs require a committed `-skill-map`.

Frozen local competitor mappings live under `evaluations/arms/`. A per-case routing map can override the canonical-to-arm skill map; otherwise the harness derives accepted arm-local IDs and accepts their isolated plugin namespace prefix.

Release-candidate treatment and scoring must use a committed benchmark lock. The lock binds exact public cases, skill and fixture bytes, hidden oracles, arm commits and content, scorer code, client/toolchain versions, models, modes, timeouts, repetitions, and randomization seeds. Private holdouts are committed with HMAC-SHA-256 while their prompts and key remain outside Git until scoring. See [docs/release-candidate-freeze.md](docs/release-candidate-freeze.md).

## Evidence boundary

Codex is the primary behaviorally validated client. Other generated adapters have structural compatibility evidence; cross-client behavioral claims require equivalent authenticated runners. See [docs/compatibility.md](docs/compatibility.md).

## Attribution and independence

Canonical prose is original writing derived from primary sources. Competitor text and code are not copied. Reference licenses and policies are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) and [research/corpus-lock.json](research/corpus-lock.json).

This independent project is not affiliated with or endorsed by Google or the Go project. “Go” and the Go gopher are trademarks of Google LLC; no Go logo is used.

## License

Apache-2.0.
