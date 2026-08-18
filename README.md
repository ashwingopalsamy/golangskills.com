# Engineering Skills for Go

Focused Agent Skills for production Go systems. The initial collection targets failure boundaries where generic coding advice is least reliable: goroutine lifecycles, HTTP boundaries, SQL transactions, message delivery, distributed-call resilience, and production change review.

This project is independent and is not affiliated with or endorsed by the Go project or Google. “Go” refers to the Go programming language; no Go logo is used.

## Why this collection exists

More skills are not automatically better. Every installed skill spends discovery context, every activated skill can conflict with repository evidence, and executable resources expand the trust boundary. This repository therefore optimizes for measurable marginal value:

- six narrow skills instead of an always-on style encyclopedia;
- under 4,000 total discovery characters, including names and paths;
- source claims classified as normative, primary, operational, or organization-specific;
- one eval suite per skill, including negative-routing cases;
- no executable code inside published skills;
- generated catalogs and plugin packages checked against canonical source;
- explicit compatibility levels instead of “works everywhere” claims.

The design and comparative baseline are recorded in [docs/architecture.md](docs/architecture.md) and [docs/benchmark.md](docs/benchmark.md).

## Skills

| Skill | Use it for |
| --- | --- |
| `go-concurrency-lifecycle` | Goroutine ownership, cancellation, bounded work, synchronization, races, and leaks |
| `go-http-boundaries` | `net/http` server/client timeouts, body ownership, shutdown, middleware, and overload behavior |
| `go-sql-transactions` | Transaction boundaries, isolation, retries, connection pools, and commit ambiguity |
| `go-message-processing` | At-least-once consumers, idempotency, ordering, acknowledgments, poison messages, and outbox/inbox design |
| `go-service-resilience` | Deadlines, retry budgets, backoff, jitter, load shedding, and partial failure across remote calls |
| `review-go-production-change` | Risk-driven review of a Go change that touches concurrency, persistence, remote calls, or operability |

The generated machine catalog is [catalog/catalog.json](catalog/catalog.json).

The [agent-evaluation protocol and exploratory results](docs/evaluations.md) separate structural quality from observed model improvement. The initial single-model run found precise routing and only a small quality delta, so the collection does not claim universal task-performance superiority.

## Install

### Codex plugin

```text
codex plugin marketplace add ashwingopalsamy/golangskills.com
codex plugin add engineering-skills-for-go@golangskills
```

The Codex package includes per-skill `agents/openai.yaml` metadata and otherwise uses the same portable skill content.

### Skills CLI

Use the open ecosystem installer and select only the skills relevant to the repository:

```text
npx skills add ashwingopalsamy/golangskills.com
```

The installer maps canonical skills to the supported location for Codex, Claude Code, GitHub Copilot, Gemini CLI, Cursor, OpenCode, Windsurf, Cline, and Roo Code. Platform behavior and evidence are documented in [docs/compatibility.md](docs/compatibility.md).

### Manual installation

Copy individual directories from `skills/` to the agent’s documented project skill directory. Avoid installing the entire collection when one or two skills cover the work.

## Repository checks

```text
go run ./cmd/skillctl check
go run ./cmd/skillctl generate
go test ./...
go vet ./...
```

`check` validates the open specification subset, metadata budgets, source provenance, routing eval coverage, relative links, cross-skill relations, and generated output freshness. `generate` updates OpenAI UI metadata, the public catalog, and the Codex/Claude plugin package.

## Contributing

Start with [CONTRIBUTING.md](CONTRIBUTING.md). A new skill must identify a failure mode that existing agents handle inconsistently, own a non-overlapping job, cite primary evidence, stay within the repository budgets, and include paired evaluation scenarios. Generic style preferences do not qualify as language requirements. The [naming decision](docs/naming.md) and [coverage roadmap](docs/roadmap.md) record the public identity and concrete gaps.

## License

Apache-2.0. External sources are linked, not copied. Their trademarks and copyrights remain with their owners.
