# Go: Distributed Systems

This collection owns failure semantics that cross goroutines, processes, stores, brokers, and coordination services.

It contains:

- `go-concurrency-lifecycle`
- `go-data-consistency`
- `go-message-processing`
- `go-service-resilience`
- `go-distributed-coordination`
- `review-go-distributed-change`

Its unifying contract is schedule-based: each recommendation must identify ownership, capacity, ordering, atomicity, retry, ambiguity, crash, or recovery behavior. It does not restate Go syntax or turn fintech consequences into generic distributed-systems advice.

Use this collection alongside Go: Production Engineering for production service work. Add Go: Fintech when an invariant can create, lose, duplicate, misstate, or conceal money.
