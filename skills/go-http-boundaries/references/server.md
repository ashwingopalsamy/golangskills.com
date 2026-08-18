# HTTP server decisions

## Timeout scope

`ReadHeaderTimeout` limits header reading. `ReadTimeout` covers reading the request, including the body, and can be too blunt for routes with different upload contracts. `WriteTimeout` bounds response writes but interacts poorly with legitimate streaming. `IdleTimeout` controls keep-alive idle time.

There is no universal set of values. Derive limits from endpoint contracts, proxy behavior, deployment budgets, and observed latency. Prefer route-level body and operation budgets when global server fields cannot express different semantics safely.

## Body limits

Apply a maximum before decoding. `http.MaxBytesReader` can bound server reads and signal oversized bodies. JSON decoding should also define unknown-field and trailing-data policy according to compatibility needs; those are API decisions, not universal hardening rules.

## Response state

Once headers or body bytes are written, the status is committed. Middleware that wants uniform error envelopes must not assume it can replace a partially written response. Buffering can restore that control for small responses but increases memory and breaks streaming.

## Shutdown and upgrades

`Server.Shutdown` does not close hijacked connections. Track WebSockets and other upgrades in an application-owned registry or protocol-specific server. Signal them before waiting, and bound the wait with the platform deadline.

`RegisterOnShutdown` starts registered functions but does not wait for them to finish. Own their completion separately if shutdown correctness depends on it.
