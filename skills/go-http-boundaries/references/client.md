# HTTP client decisions

## Reuse and pool limits

Reuse `http.Client` and `http.Transport`. Configure pool limits from expected concurrency and downstream capacity. A low `MaxConnsPerHost` becomes a client-side queue; an unbounded effective pool can overload the peer. Observe wait time and reuse before tuning.

## Response bodies

After `Do` returns a response, close `Body` on every path. Reading to EOF can enable connection reuse, but draining an attacker-controlled or unexpectedly large response can waste bandwidth and memory. Bound first; decide whether reuse is worth draining the remaining bounded bytes.

Some errors can return a non-nil response. Follow the documented method contract rather than assuming `err != nil` always means `resp == nil`.

## Redirects

Redirects can cross hosts and change methods. Review credential forwarding, cookie policy, body replay through `GetBody`, maximum hops, and destination allowlists. A client used for server-side URL fetching also needs DNS/IP and redirect checks to resist SSRF; URL syntax validation alone is insufficient.

## Attempts

An HTTP request may be replayable at the protocol level while the application operation is not. `GET` can still trigger a broken side effect; `POST` can be safely retryable with an application idempotency contract. Decide from operation semantics, not method folklore alone.
