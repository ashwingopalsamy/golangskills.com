# Release-candidate benchmark freeze

The release-candidate lock turns “freeze before treatment” into an executable protocol. It does not make the current evidence release-eligible: no real RC lock or private holdout has been committed yet, and the large comparison batch remains deferred.

## What the public lock binds

`skillctl eval freeze` records:

- the committed source revision and SHA-256 digests of canonical skills/evals, fixtures, hidden oracles, arm maps, corpus lock, and all runner/scorer code;
- baseline, this project, and every competitor repository commit, repository-relative skill root, exact installed-skill digest, and mapping digest;
- every public development case key, collection, kind, split, and canonical digest;
- Codex and Go versions, treatment and judge models, native/explicit modes, timeouts, scorer/rubric versions, repetitions, and one treatment/judge seed per repetition;
- a private-holdout ID, case count, coverage strata, and HMAC-SHA-256 commitment when a holdout is supplied.

The public lock never contains private prompts, case IDs, reasons, or the HMAC key. Repository-relative competitor roots make the commitment independent of a developer's checkout path.

## Prepare the private holdout

Keep pre-release holdouts under the ignored `.private/` tree or outside the repository. Both files must be regular and owner-only when the freeze is created. Generate at least 256 bits of key material as lowercase hexadecimal, for example:

```sh
mkdir -p .private/evaluations/rc1
chmod 700 .private/evaluations/rc1
openssl rand -hex 32 > .private/evaluations/rc1/holdout.key
chmod 600 .private/evaluations/rc1/holdout.key
chmod 600 .private/evaluations/rc1/holdout.json
```

The holdout is a strict JSON overlay. Every case names a canonical owning skill and one task type (`implementation`, `diagnosis`, `design`, or `review`); its embedded eval uses `split: "holdout"`. A negative unrelated routing case still has an owning skill for collection stratification, but uses `should_activate: false` with no confusion route, so its correct route is `NONE`.

```json
{
  "schema_version": 1,
  "id": "rc1-private",
  "cases": [
    {
      "skill": "go-service-resilience",
      "task_type": "diagnosis",
      "case": {
        "id": "opaque-private-id",
        "kind": "routing",
        "split": "holdout",
        "prompt": "private prompt",
        "should_activate": true,
        "reason": "private adjudication reason"
      }
    }
  ]
}
```

The current private overlay supports routing and non-executable semantic-quality cases. Executable holdout fixtures remain public-input fixtures; do not imply a private executable-fixture protocol until private fixture/oracle loading is implemented and threat-modeled.

## Freeze and commit before treatment

All public benchmark inputs must already be committed and clean. Freeze creation fails when relevant paths are modified or untracked, when a competitor checkout differs from its audited commit, or when an in-repository secret is not ignored.

If the locked competitors are installed somewhere other than the committed portable root `/reference-checkouts`, point the evaluator at their common absolute parent. Do not edit the corpus lock or arm manifest for a host-specific checkout:

```sh
export GOLANGSKILLS_REFERENCE_ROOT=/absolute/path/to/go-refs
```

The override changes runtime path resolution only. Freeze creation still verifies the locked repository identity and commit, rejects dirty checkouts, validates mappings and containment, and records the exact installed-skill digest.

```sh
go run ./cmd/skillctl eval freeze \
  -id rc1 \
  -model gpt-5.6-sol \
  -judge-model gpt-5.6-sol \
  -repetitions 3 \
  -seed-base 2026082401 \
  -judge-seed-base 2026083401 \
  -private-holdout .private/evaluations/rc1/holdout.json \
  -holdout-key .private/evaluations/rc1/holdout.key \
  -output evaluations/releases/rc1.lock.json

git add evaluations/releases/rc1.lock.json
git commit -m "eval: freeze rc1 benchmark protocol"
```

Anyone can verify disclosed inputs without the private material:

```sh
go run ./cmd/skillctl eval verify-freeze \
  -freeze evaluations/releases/rc1.lock.json \
  -public-only
```

The operator performs full verification, including the live client/toolchain and private commitment:

```sh
go run ./cmd/skillctl eval verify-freeze \
  -freeze evaluations/releases/rc1.lock.json \
  -private-holdout .private/evaluations/rc1/holdout.json \
  -holdout-key .private/evaluations/rc1/holdout.key \
  -check-environment
```

## Execute resumable cells

Read the model, timeout, repetition index, and corresponding seed from the lock. Run native and explicit development arms separately; use the same output path to resume. A targeted `-kind` or `-case` subset is allowed for recovery, but `-limit` and `-fixtures-only` are rejected because they can silently change a frozen sample.

```sh
go run ./cmd/skillctl eval matrix \
  -freeze evaluations/releases/rc1.lock.json \
  -split development \
  -model gpt-5.6-sol \
  -repetition 0 \
  -seed 2026082401 \
  -timeout 5m \
  -output evaluations/runs/rc1-development-native-r0.jsonl
```

For explicit mode add `-explicit`. For the private split add all three of `-split holdout`, `-private-holdout`, and `-holdout-key`. Each frozen result stores `freeze_id`, `freeze_digest`, `treatment_seed`, client/model, fixture commit, and grader version. The scorer rejects foreign or drifted cells.

```sh
go run ./cmd/skillctl eval score \
  -freeze evaluations/releases/rc1.lock.json \
  -public-only \
  -input evaluations/runs/rc1-development-native-r0.jsonl \
  -judgments evaluations/judgments/rc1-development-native-r0.jsonl \
  -judge-model gpt-5.6-sol \
  -judge-seed 2026083401 \
  -judge-timeout 5m \
  -output evaluations/scores/rc1-development-native-r0.jsonl
```

Omit `-public-only` and provide the holdout plus key when scoring holdout cells. Judgment artifacts explicitly retain the same freeze ID and digest.

## Publish and rotate

After release scoring is final, publish the holdout JSON, commitment key, raw treatment and judgment JSONL, scores, reports, client/model versions, and scorer revision. Full verification must reproduce the public commitment. Then create a new private holdout and key for the next release; never reuse a disclosed key or scored holdout as a future hidden set.

An incomplete batch stays incomplete. Preserve error and partial cells, document missing repetitions, and do not promote its numbers into the leadership gates.
