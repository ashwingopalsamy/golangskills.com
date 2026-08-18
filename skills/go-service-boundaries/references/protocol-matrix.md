# Protocol matrix

| Concern | HTTP | gRPC | GraphQL |
|---|---|---|---|
| Client correction | 4xx semantics | canonical status code | typed field/error contract |
| Cancellation | request context | RPC/stream context | request context and resolver work |
| Capacity | body, handler, connection | message, stream, RPC | complexity, depth, pagination |
| Compatibility | methods, fields, status | protobuf field numbers and methods | schema fields, nullability, enums |
| Long-lived work | streaming/hijacked connection | stream send/receive loops | subscriptions |

Exercise malformed and oversized input, cancellation during effects, partial response writes, active-stream shutdown, and mixed client/server versions where applicable.
