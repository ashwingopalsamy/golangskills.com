# Benchmark and acceptance criteria

## Baselines

The initial audit covered the supplied local repositories plus the strongest current public Go-specific collection found during research.

| Collection | Skills | Discovery pressure | Main strength | Material weakness |
| --- | ---: | --- | --- | --- |
| samber/cc-skills-golang | 46 | about 24,000 description characters | Breadth and scenarios | Disabled schema CI, overlapping activation, preference-heavy assertions |
| muratmirgun/gophers | 26 | about 10,800 description characters | Generated adapters and validation | Exceeds Codex discovery cap; style-heavy taxonomy |
| spf13/go-skills | 6 | about 3,100 description characters | Workflow and release judgment | One 774-line core skill; little validation or provenance |
| cxuu/golang-skills | 20 | about 7,200 description characters | Rule ownership and source snapshots | Near discovery cap; eval definitions are not agent executions |
| eduardo-sl/go-agent-skills | 33 | reported about 4,726 discovery tokens | Honest paired-eval harness and broad operations coverage | Far beyond Codex cap; only 5 eval suites; client-only frontmatter in portable skills |

Counts were measured from the repositories as inspected on 2026-08-18. They are snapshots, not permanent claims.

## “Better” is testable

The v1 repository must satisfy all structural gates:

1. every canonical skill passes the open Agent Skills name and description constraints;
2. total discovery text stays below 6,000 characters, leaving headroom under Codex’s 8,000-character cap;
3. every skill has current provenance, compatibility metadata, routing positives, routing negatives, and quality assertions;
4. all generated adapters are fresh and byte-equivalent to canonical content;
5. no published skill contains executable code or a tool permission grant;
6. all cross-skill relations resolve and no description pair exceeds the configured similarity ceiling;
7. the repository tool builds and its deterministic validators pass on the declared minimum Go version.

Quality gates are paired and scenario-based:

- run the same fixture with and without the selected skill;
- hold agent, model, repository commit, prompt, and verifier constant;
- score externally observable decisions, not preferred wording;
- include at least one counterexample that should suppress the skill’s default recommendation;
- reject a skill whose median quality delta is non-positive or whose token cost is disproportionate to its gain.

Published results must record date, agent, model, versions, trials, failures, and raw outputs. This repository will not claim a percentage advantage from regex-only grading.

## V1 differentiators

- **Context efficiency:** the six names and descriptions total 2,016 characters. The official reference serializer produced 3,127 characters including absolute paths on the inspected checkout; the repository's conservative location-padded metric is 3,519 characters, leaving substantial Codex headroom.
- **Engineering scope:** message delivery, commit ambiguity, overload, and retry amplification are first-class rather than afterthoughts.
- **Authority honesty:** Google and Uber style guides remain organization-specific unless a repository adopts them.
- **Evaluation coverage:** 36 checked cases cover every skill: 12 positive routes, 12 negative routes, and 12 semantic quality rubrics.
- **Provenance:** 25 classified sources identify the claims they support and when they were verified.
- **Supply-chain restraint:** no skill scripts, network fetches, or portable tool grants.
- **Adapter discipline:** portable content is canonical; client packages are generated.

## Post-implementation re-evaluation

The same dimensions produce a mixed result rather than a universal win:

| Dimension | V1 result against the strongest alternatives |
| --- | --- |
| Coverage | **Weaker.** Six skills cannot match Gophers' 26 or samber's 46-topic breadth. V1 deliberately omits basic language, tooling, cloud-native, security, diagnostics, and performance workflows. |
| Depth and failure modes | **Stronger inside the six owned boundaries.** The skills start from ownership, atomicity, replay, ambiguity, overload, and recovery rather than API inventories or style rules. This does not establish superiority on uncovered tasks. |
| Activation and overlap | **Structurally stronger.** Names and descriptions total 2,016 characters, the conservative serialized discovery estimate is 3,519, every description has an exclusion boundary, and maximum pairwise description Jaccard overlap is 0.125 under the checked metric. |
| Progressive disclosure | **Stronger than monolithic competitors.** Core files are 133-163 lines and link one-level decision references; no skill requires loading a router or the whole corpus. |
| Correctness and provenance | **Stronger auditability.** Every skill has claim-scoped, authority-classified, dated sources. Correct metadata cannot prove every engineering statement is right. |
| Validation | **Stronger corpus invariants.** The Go tool checks strict JSON, budgets, relationships, links, source age, eval balance, overlap, and byte-fresh generation; CI also invokes the pinned official Agent Skills validator. Gophers remains strong in adapter validation against a wider set of client outputs. |
| Agent behavior evidence | **Not yet a decisive win.** V1 has complete routing and quality fixtures plus initial paired forward tests, but it lacks repeated trials across multiple public agent/model combinations. Eduardo's collection has fewer suites but a more automated paired execution harness. |
| Portability | **Cleaner canonical model, narrower packaging.** Canonical skills follow the open specification and generate Codex and Claude packages. Gophers publishes more explicit client adapters, which can be easier for users whose installer is unavailable. |
| Supply-chain risk | **Stronger default.** Published skills contain no executables, network fetches, or tool grants. This trades away task-specific automation that could be valuable with a separate threat model. |
| Discoverability | **Stronger machine surface, immature human surface.** The catalog is deterministic and website-ready, but competitors with mature marketplace presence and broader READMEs currently have greater real-world visibility. |
| Maintainability | **Stronger drift controls, smaller maintainer burden.** One canonical corpus generates adapters and dates sources. The 400-day source gate intentionally creates recurring maintenance work. |

The defensible preference claim is therefore narrow: choose this repository when production failure-boundary reasoning, activation precision, provenance, and context cost matter more than topic breadth. It has not yet earned a claim of better task success across Go engineering generally.
