# Continuation handoff: category-leading Go agent skills

Status date: 2026-08-29. This is the operational handoff for continuing the category-leadership goal. It records public state and private-artifact metadata only; it intentionally contains no holdout prompts, case IDs, commitment value, or key material.

## Version 0.4.0 publication work

The public family is now branded **Go: Production Engineering**, **Go: Distributed Systems**, and **Go: Fintech** while retaining the stable package IDs `engineering-skills-for-go`, `distributed-systems-skills-for-go`, and `fintech-skills-for-go`. Canonical skill IDs and collection ownership are unchanged.

Canonical generation targets version `0.4.0` and carries collection-specific names, subtitles, descriptions, capabilities, keywords, starter prompts, and CC0 Gopher artwork across Codex/ChatGPT, Claude Code, Cursor, OpenCode, npm, catalogs, site data, and `llms.txt`. The source artwork is recorded in `THIRD_PARTY_NOTICES.md`. Thirteen fresh routing cases were added, bringing the public corpus to 146 cases without changing the deferred release-candidate comparison plan.

The fresh v0.4.0 development routing screen retained its initial 8/13 artifact, corrected five overly broad expected review-skill ownership decisions to the precise canonical domain skills Codex had selected, and then scored 13/13 on the final one-pass screen. The final report records macro-F1 1.0, zero false activations across two unrelated prompts, zero critical failures, and 1.0 collection pass rates. This is metadata-development evidence only and does not satisfy the release-candidate superiority contract.

All three version 0.4.0 plugins are published. The Production Engineering release updated the existing plugin in place and preserved its plugin ID. The user completed the final specialist publication actions manually in the OpenAI portal.

| Public plugin | Stable package ID | OpenAI plugin ID | Submission record | Public URL | State |
| --- | --- | --- | --- | --- | --- |
| Go: Production Engineering | `engineering-skills-for-go` | `plugins_6a92c6b7e7948191ab7802aa05afc6f7` | `appsub_6a931763413081918cf2d44be988d8e5` | <https://chatgpt.com/plugins/plugins_6a92c6b7e7948191ab7802aa05afc6f7> | Published |
| Go: Distributed Systems | `distributed-systems-skills-for-go` | `plugins_6a931917afc48191a2ce571d737eb104` | `appsub_6a931917c48c8191850dc67a76f9a10f` | <https://chatgpt.com/plugins/plugins_6a931917afc48191a2ce571d737eb104> | Published |
| Go: Fintech | `fintech-skills-for-go` | `plugins_6a9319929b74819193f3f276f98627c4` | `appsub_6a931992ae24819184c6e722a64b7939` | <https://chatgpt.com/plugins/plugins_6a9319929b74819193f3f276f98627c4> | Published |

The three public endpoints resolve. Their unauthenticated HTML shell does not expose listing metadata, so the names and published states above are the owner-confirmed portal record rather than an inference from the public page body. No moderation feedback was reported at publication time.

GitHub release [`v0.4.0`](https://github.com/ashwingopalsamy/golangskills.com/releases/tag/v0.4.0) was published from `acadc48` with all three OpenAI ZIPs, npm tarballs, portable archives, catalogs, provenance, and separate checksum manifests. The coordinated implementation commits are `072dbdc`, `be1b392`, and `acadc48`.

## Goal and evidence boundary

Build **Go Engineering Skills by Ashwin Gopalsamy** as three focused collections—Engineering, Distributed Systems, and Fintech—that lead current Go agent-skill alternatives on correctness, reference coverage, routing, behavioral quality, critical-error avoidance, provenance, and context efficiency.

“Category-leading” is a release contract, not current marketing language. Do not make a number-one claim or create `v1.0.0-rc.1` until `skillctl release check` passes every structural and behavioral gate. Codex has behavioral evidence. Claude Code, Cursor, and OpenCode remain structurally compatible but behaviorally unbenchmarked until authenticated runners exist.

Do not rerun the public runtime-reload/cutover case three times. Its completed and partial artifacts are intentionally retained as an incomplete campaign, and repeated comparison is deferred to the release-candidate batch. New work must use fresh, independently frozen scenarios rather than tuning against published prompts.

## Repository and Git state

- Public repository: `ashwingopalsamy/golangskills.com`; use `gh` and git, never the in-app browser, for GitHub operations.
- Working branch: `codex/category-leader`; both it and remote `main` contain the v0.4.0 publication commit `acadc48`.
- Portable-reference milestone: `97ec8ff feat: support portable benchmark reference roots`.
- Latest canonical source milestone: `043ce19 feat: deepen process hedging and business-time invariants`.
- Latest provenance-bound package milestone: `c3388ac chore: package process hedging and business-time corpus`.
- Preserve local `main` at `46b2551`. It contains the unsquashed 127-commit development history and has no common ancestry with the consolidated remote history. Do not reset, delete, rebase, merge, or force-push it.
- Published tags are `v0.1.0`, `v0.2.0`, `v0.3.0`, and `v0.4.0`. There is intentionally no leadership or release-candidate tag.
- Safe publication from the work branch is a normal fast-forward of remote `main`; the local historical `main` need not move.

## Completed product and knowledge work

- 20 non-overlapping canonical skills: 8 Engineering, 6 Distributed Systems, and 6 Fintech.
- Combined conservative discovery footprint: 7,774 characters; every collection is below 4,000.
- Canonical corpus: 146 development cases, including 15 executable hidden-oracle fixtures and the fresh v0.4.0 routing screen.
- Evidence model: 121 source records, 27 adjudicated canonical claims, and all seven reference snapshots inventoried with 759 file hashes, 103 skill entrypoints, and 10,239 material-item dispositions.
- Generated Codex, Claude Code, Cursor, and OpenCode layouts; catalogs, search/site data, `llms.txt`, archives, checksums, and provenance.
- Three npm staging packages validate at `0.3.0-rc.1`: `@golangskills/engineering`, `@golangskills/distributed-systems`, and `@golangskills/fintech`. Registry checks on the status date returned 404 for all three; they are not published and must not be advertised as currently installable from npm.
- The transport-cutover finding was generalized into canonical lifecycle guidance: successful reload requires atomic publication of the new transport, routing new work only to it, allowing already-started work to drain, and actively retiring obsolete idle connections after cutover.
- The interrupted runtime-reload benchmark is documented in `docs/evaluations.md`; corrected repetition 0 is complete, repetition 1 treatment is preserved but unjudged/unscored, and repetition 2 was never started.

## Reference checkouts

The committed manifest remains portable at `/reference-checkouts`. On this host, use:

```sh
export GOLANGSKILLS_REFERENCE_ROOT=/Users/ashwin/Desktop/ref/go-refs
```

The runtime override rebases placement only. Repository identity, locked commit, dirtiness, inventory, containment, skill mappings, and installed bytes still fail closed. The current audited commits are:

| Repository | Commit |
| --- | --- |
| `cc-skills-golang` | `30cdf15cde8db8730c42a2918d7cdb4505f5ff54` |
| `cxuu-golang-skills` | `91f0c2eef559a3168f9d3c38f5f99936d472b508` |
| `go-skills` | `e67851cfcca008592c7c4965b8220c7cb37e2f1c` |
| `gophers` | `d4bc22a03ccb4550fcd148b7ad56250624799c72` |
| `golang-ddd-skills` | `0cf1ee4d7facd50d1f25cce956dcc2fe252bf29d` |

## Private holdout readiness

Ignored owner-only files exist at `.private/evaluations/rc1/holdout.json` and `.private/evaluations/rc1/holdout.key`; both are mode `0600`. Never add either file to Git, print the key, or publish prompts before release scoring is final.

The overlay has 60 unique cases:

- 20 positive routing cases, one for every canonical skill;
- 16 unrelated negative routing cases whose correct route is `NONE`;
- 24 semantic-quality cases, eight per collection and exactly two per task type per collection.

A disposable lock outside the repository successfully froze the then-current 130 public cases, seven arms, and the 60-case private commitment. Public-only verification, full private/environment verification, and a wrong-key fail-closed check all succeeded. No treatment or evaluator process was launched. Three public cases were added afterward, so that disposable lock is intentionally stale and cannot become the real RC lock. This proves tooling readiness only; it provides no routing or quality score.

Do not commit the real `evaluations/releases/rc1.lock.json` until the corpus and fresh scenario set are declared ready. Once committed, do not edit benchmark inputs before running its cells.

## Verification at this checkpoint

The following completed successfully after the portable-reference change:

```sh
go test ./...
go vet ./...
go run ./cmd/skillctl check
go run ./cmd/skillctl npm check -version 0.3.0-rc.1
git diff --check
```

`skillctl check` reports 20 skills, 133 cases, 121 source records, 27 claims, and 7,774 discovery characters. The private freeze verification used Codex CLI `0.149.1`, Go `1.24.2`, model and judge `gpt-5.6-sol`, native plus explicit modes, and three planned repetitions.

## Current release gates

Passing in the gate report: complete reference inventory, primary claim evidence, structural corpus validation, discovery budgets, generation, and `0.3.0-rc.1` staging validation. `skillctl release check` currently stops earlier at the npm packaging gate because `dist/npm` is staged as `0.3.0-rc.1` while the collection catalog targets stable `0.3.0`; do not describe the repository as release-eligible until the correctly timed stable staging is reproduced.

Still blocked: hidden routing scores, zero critical failures across the full matrix, discriminatory fixture superiority, paired superiority with confidence bounds, domain non-inferiority, statistically positive baseline improvement, global token-efficiency leadership, and complete RC traceability. `catalog/benchmark-status.json` must remain `evidence-pending`.

## Next execution order

1. Continue broad canonical depth and fresh, diverse cross-domain eval scenarios. Prefer incomplete-observability, crash-schedule, or authority-boundary problems that do not state the entire repair and that have not appeared in prior public cases.
2. For every canonical skill change, update its routing and quality cases, then run `go run ./cmd/skillctl generate` and `go run ./cmd/skillctl npm package -version 0.3.0-rc.1` before checks. Never hand-edit generated adapters, catalogs, or `dist/npm`.
3. When the corpus is ready, run `eval preflight`, create the real RC lock with the private holdout and runtime reference-root override, inspect only public metadata, then commit the lock before treatment.
4. Run a cost-bounded infrastructure pilot. Only after it is healthy, execute the locked Codex native and explicit arms across baseline, ours, and every frozen competitor. Use the exact repetitions, seeds, timeouts, model, and output paths from the lock; preserve partial JSONL and resume rather than restart.
5. Apply deterministic graders first, then blinded same-platform semantic judging. Generate per-arm, paired, domain, baseline, critical-error, and token-efficiency reports with raw-artifact links.
6. Run `skillctl release check`. If any gate fails, publish honest evidence without a leadership claim and improve only through newly frozen cases. If every gate passes, reproduce from a clean clone before tagging.
7. Keep npm bootstrap publication separate from GitHub source publication. It requires interactive npm authentication and should use the documented `bootstrap` dist-tag before Trusted Publishing is configured.

## Real freeze command template

Run this only after step 3 is authorized by corpus readiness:

```sh
export GOLANGSKILLS_REFERENCE_ROOT=/Users/ashwin/Desktop/ref/go-refs
go run ./cmd/skillctl eval freeze \
  -id rc1 \
  -model gpt-5.6-sol \
  -judge-model gpt-5.6-sol \
  -repetitions 3 \
  -seed-base 2026082901 \
  -judge-seed-base 2026083901 \
  -private-holdout .private/evaluations/rc1/holdout.json \
  -holdout-key .private/evaluations/rc1/holdout.key \
  -output evaluations/releases/rc1.lock.json
```

Immediately verify it both public-only and privately as documented in `docs/release-candidate-freeze.md`, then commit the lock before running `eval matrix`.
