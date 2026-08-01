---
name: go-service-boundaries
description: "Use for Go HTTP, gRPC, GraphQL, or OpenAPI boundaries and streams. Do not use for domain or retry policy."
license: Apache-2.0
compatibility: "Go 1.24 or newer; protocol and library behavior must be verified against deployed versions."
---

# Go service boundaries

Treat every service adapter as a trust, resource, and compatibility boundary. Keep domain behavior independent of transport encoding while preserving protocol-specific semantics.

## Define the boundary

Record authentication and authorization ownership, request and response size limits, deadline source, streaming behavior, concurrency admission, error mapping, versioning, and shutdown policy. Validate untrusted input once before constructing a domain command.

## HTTP

Use explicit `http.Server`, client, and transport ownership. Bound headers and finite bodies. Reuse clients and close response bodies. Propagate `r.Context()`. Choose total, phase, or idle timeouts from endpoint semantics; a global write timeout can break streaming. Treat redirects and proxy headers as new trust decisions.

## gRPC

Preserve status codes, cancellation, metadata trust, message limits, and stream lifecycle. Reuse connections. Put authentication, tracing, and recovery in ordered interceptors, but keep domain decisions out of them. A retry policy must respect method idempotency and the caller's deadline.

## GraphQL and OpenAPI

Keep resolvers and generated handlers thin. Bound query complexity and pagination. Batch only within request lifetime and authorization scope. Treat schema nullability, enums, field removal, and generated client expectations as public contracts. Generated OpenAPI is evidence only when it matches runtime behavior.

## Shutdown and ambiguity

Stop admission, drain accepted work within the platform budget, terminate application-owned streams, then close dependencies. A missing response after a request may mean an unknown remote outcome; do not convert transport uncertainty into a known business failure.

Read [references/protocol-matrix.md](references/protocol-matrix.md) for protocol-specific failure and test cases.

## Output contract

Change the smallest owning adapter. Report exact protocol, resource, trust, or compatibility failures, including the triggering request and observable result.
