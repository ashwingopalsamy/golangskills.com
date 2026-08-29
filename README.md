# Go skills for serious systems

<p align="center">
  <img src="https://raw.githubusercontent.com/ashwingopalsamy/golangskills.com/main/plugins/engineering-skills-for-go/assets/logo.png" width="150" alt="Go: Production Engineering gopher" />
  <img src="https://raw.githubusercontent.com/ashwingopalsamy/golangskills.com/main/plugins/distributed-systems-skills-for-go/assets/logo.png" width="150" alt="Go: Distributed Systems gopher" />
  <img src="https://raw.githubusercontent.com/ashwingopalsamy/golangskills.com/main/plugins/fintech-skills-for-go/assets/logo.png" width="150" alt="Go: Fintech gopher" />
</p>

<p align="center">
  <strong>20 Go skills · 3 collections · 1 clean source</strong><br />
  Built by <a href="https://ashwingopalsamy.in">Ashwin Gopalsamy</a> for AI coding agents.
</p>

<p align="center">
  <a href="https://github.com/ashwingopalsamy/golangskills.com/releases/tag/v0.4.1"><img src="https://img.shields.io/github/v/release/ashwingopalsamy/golangskills.com?label=release&color=00ADD8" alt="v0.4.1 release" /></a>
  <a href="https://github.com/ashwingopalsamy/golangskills.com/actions"><img src="https://img.shields.io/github/actions/workflow/status/ashwingopalsamy/golangskills.com/validate.yml?label=checks&color=2ea44f" alt="checks" /></a>
  <a href="https://github.com/ashwingopalsamy/golangskills.com/blob/main/LICENSE"><img src="https://img.shields.io/github/license/ashwingopalsamy/golangskills.com?color=blue" alt="Apache-2.0" /></a>
  <img src="https://img.shields.io/badge/Go-1.24%20%7C%201.25%20%7C%201.26%20%7C%201.27-00ADD8" alt="Go versions" />
</p>

Three focused Go collections for ChatGPT, Codex, Claude Code, Cursor, OpenCode, and `npx skills`.

| Collection | Best for | Skills |
| --- | --- | ---: |
| **[Go: Production Engineering](docs/collections/engineering.md)** | APIs, services, testing, performance, security, operations, and general review | 8 |
| **[Go: Distributed Systems](docs/collections/distributed-systems.md)** | Concurrency, consistency, messaging, resilience, coordination, and partial failure | 6 |
| **[Go: Fintech](docs/collections/fintech.md)** | Money, ledgers, payments, idempotency, settlement, reconciliation, and compliance | 6 |

## Install

Install the whole family from the canonical source:

```sh
npx skills add ashwingopalsamy/golangskills.com
```

Or use a generated collection package in the Codex/ChatGPT plugin directory:

| Plugin | Package ID | Directory |
| --- | --- | --- |
| Go: Production Engineering | `engineering-skills-for-go` | [open](https://chatgpt.com/plugins/plugins_6a92c6b7e7948191ab7802aa05afc6f7) |
| Go: Distributed Systems | `distributed-systems-skills-for-go` | [open](https://chatgpt.com/plugins/plugins_6a931917afc48191a2ce571d737eb104) |
| Go: Fintech | `fintech-skills-for-go` | [open](https://chatgpt.com/plugins/plugins_6a9319929b74819193f3f276f98627c4) |

Client layouts and package commands are documented in [docs/install.md](docs/install.md).

## The point

- **Sharp boundaries:** payment integrity goes to Fintech. Cross-process failure goes to Distributed Systems. The rest goes to Production Engineering.
- **Small entrypoints:** each `SKILL.md` gets to the decision quickly; deeper material sits in nearby references.
- **One source, every client:** Codex, Claude Code, Cursor, OpenCode, catalogs, search data, and `llms.txt` are generated from the same corpus.
- **Easy to inspect:** 121 sources, 27 claims, locked reference snapshots, provenance, checksums, and 146 development cases.

All three together use **7,774 discovery characters**; each collection stays below 4,000. Benchmark status is tracked in [`catalog/benchmark-status.json`](catalog/benchmark-status.json).

## Maintainer workflow

```sh
go run ./cmd/skillctl check
go run ./cmd/skillctl generate
go run ./cmd/skillctl package -version 0.4.1
go run ./cmd/skillctl npm package -version 0.4.1
```

Read [the continuation handoff](docs/continuation-handoff.md) for benchmark status, reference locks, release history, and publication records.

## Scope and attribution

Canonical prose is independently written from primary sources. Reference licenses and competitor attribution are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). This project is independent and is not endorsed by Google or the Go project.

## License

Apache-2.0.
