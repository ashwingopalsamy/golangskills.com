---
name: go-http-boundaries
description: "Design, implement, or review Go HTTP client and server boundaries: timeouts, body ownership, cancellation, middleware, graceful shutdown, streaming, and overload behavior. Use for net/http services and outbound calls. Do not use for domain logic, generic API schema design, or non-HTTP transports."
license: Apache-2.0
compatibility: "Go 1.24 or newer. Guidance targets the stable Go 1.25 and 1.26 families and standard net/http behavior."
---

# Go HTTP boundaries

Treat HTTP as a resource and failure boundary. Correctness depends on more than handler output: connection lifetime, request cancellation, body ownership, timeout scope, protocol upgrades, and overload behavior determine whether the service remains safe under partial failure.

## Map the boundary first

Inspect existing server construction, transports, middleware, deployment shutdown budget, proxies, and streaming endpoints. Record:

1. who creates and owns each `http.Server`, `http.Client`, and `Transport`;
2. which layer sets the end-to-end deadline;
3. maximum request and response body sizes;
4. whether endpoints stream, hijack, upgrade, or hold long polls;
5. the concurrency and queue limit before expensive work;
6. the shutdown sequence and the platform’s termination deadline.

Do not paste a “secure server” template without reconciling these facts.

## Server invariants

- Construct an explicit `http.Server`; package-level convenience serving hides boundary configuration.
- Bound header parsing with `ReadHeaderTimeout` when exposed to untrusted or slow peers.
- Choose `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` from actual endpoint semantics. A global write timeout can break streaming responses.
- Bound request bodies at the edge before decoding when the route has a finite contract.
- Propagate `r.Context()` through request-scoped calls; do not replace it with a background context.
- Return protocol-safe errors without leaking internal details, secrets, or untrusted error text.
- Place panic recovery only at an intentional isolation boundary and record the stack; recovery does not make corrupted shared state safe.

Read [references/server.md](references/server.md) for timeout and shutdown decisions.

## Client invariants

- Reuse clients and transports; they own connection pools and are safe for concurrent use.
- Set a caller-visible end-to-end deadline or `Client.Timeout` according to the operation contract. Transport phase timeouts do not replace an overall budget.
- Build outbound requests with the caller’s context.
- On a successful `Do`, close the response body. Read or drain only when the protocol and reuse benefit justify the bytes; never read an unbounded body merely to reuse a connection.
- Bound response bodies before decoding when the peer is not fully trusted.
- Treat redirects as new trust-boundary decisions: credentials, hosts, methods, and replayable bodies can change.
- Do not retry in the transport wrapper unless replay safety, retry ownership, and budget are explicit.

Read [references/client.md](references/client.md) when configuring transports, body handling, redirects, or retries.

## Timeout composition

A timeout is a budget allocation, not a magic constant. Account for:

```text
caller budget
  = queue/admission
  + request read
  + application work
  + downstream attempts and backoff
  + response write
  + safety margin
```

Avoid stacking unrelated defaults that make the smallest hidden timeout win. Preserve the parent’s earlier deadline. If the remaining budget is insufficient for a useful attempt, fail before starting work.

For streaming, use protocol-aware idle or per-operation deadlines rather than a total timeout that terminates a healthy long-lived stream.

## Middleware ownership

Keep middleware ordered by causal need. A common shape is request identity and telemetry outside recovery, then admission/authentication, then domain handlers, but repository semantics decide the exact order.

For each middleware, verify:

- whether it reads or replaces the body;
- whether it buffers a streaming response;
- whether it writes headers before downstream code;
- whether its metrics label unbounded values;
- whether it trusts proxy headers only from configured proxies;
- whether it changes cancellation or error ownership.

Do not store mutable request state in package globals. Context values carry only request-scoped cross-cutting data, with collision-safe keys; ordinary parameters remain clearer for domain inputs.

## Admission and overload

Bound expensive work before allocating large bodies, spawning goroutines, or calling dependencies. Choose an observable overload policy:

- reject cheaply with a stable response and retry semantics;
- queue within a measured latency and memory bound;
- shed lower-priority work;
- degrade a documented optional feature.

An unbounded server queue turns overload into latency, memory pressure, missed deadlines, and retry amplification. Health endpoints must remain cheap enough to distinguish overload from process death, and readiness must match the deployment’s traffic-drain behavior.

## Graceful shutdown

`Server.Shutdown` closes listeners and idle connections, then waits for active connections until its context expires. It does not wait for hijacked connections such as WebSockets.

Order shutdown from ownership:

1. mark the instance unready or stop admission according to platform behavior;
2. stop accepting new connections;
3. signal application-owned streams and background work;
4. wait within the deployment budget;
5. close dependencies after their users stop;
6. surface timeout or forced-close outcomes.

Do not let `main` return as soon as `ListenAndServe` reports `http.ErrServerClosed`; it must wait for the shutdown path it initiated.

## Error and retry semantics

Classify errors by what the caller can safely do:

- malformed or unauthorized request: do not retry unchanged;
- overload or transient dependency failure: retry only if the operation is replay-safe and the caller owns a budget;
- ambiguous outcome after bytes may have reached the peer: do not infer failure from a missing response;
- context cancellation: preserve the caller’s reason rather than rewriting it as an internal server error.

Use `go-service-resilience` for multi-attempt policy. Use `go-message-processing` when HTTP only enqueues work whose durable delivery semantics dominate correctness.

## Finish with evidence

Within the request’s authority, exercise:

- slow headers and oversized bodies;
- client cancellation during dependency work;
- response-body close paths;
- overload rejection;
- shutdown with active requests and long-lived connections;
- streaming behavior under configured deadlines.

Use focused `httptest` tests where they model the boundary. Use a real server when connection reuse, shutdown, HTTP/2, or socket timing is the behavior under test. State which protocol and deployment behaviors remain unverified.

## Output contract

For implementation, change the smallest owning layer and keep timeout/body/shutdown rationale visible. For review, report the exact resource leak, trust-boundary error, or failure schedule—not generic advice to “set all timeouts.”
