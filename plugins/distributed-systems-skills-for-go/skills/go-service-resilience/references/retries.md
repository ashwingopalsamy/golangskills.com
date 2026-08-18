# Retry decisions

## Failure classification

Classify from stable protocol or typed errors. Examples that may be transient include connection establishment failure, a documented unavailable response, or a serialization conflict. Authentication failure, validation failure, and a business rejection generally do not become valid with unchanged input.

Timeout is ambiguous: the dependency may have completed after the caller stopped waiting. Replay safety must come from the operation contract.

## Backoff and jitter

Cap both attempts and elapsed time. Exponential backoff grows delay; jitter prevents clients from waking in lockstep. Full jitter is a common starting point:

```text
sleep = random(0, min(cap, base * 2^attempt))
```

Use a randomness source appropriate for scheduling, not cryptographic identity. Make the clock and random source controllable in deterministic tests when the repository’s abstractions support it.

## Retry-After

Honor a trusted dependency’s retry hint only within the caller budget and local cap. Treat malformed or extreme values defensively. A hint does not override replay safety.

## Hedging

Hedged requests deliberately issue a second attempt before the first fails. They can reduce tail latency at direct capacity cost. Use only for replay-safe reads, with a measured tail distribution, a strict hedge budget, and cancellation of losing attempts. Never hedge non-idempotent mutations.

## Singleflight

Coalescing identical in-flight reads reduces duplicate load within one process. It does not cache results, coordinate across replicas, or make mutations idempotent. Ensure one slow leader does not force callers beyond their individual deadlines.
