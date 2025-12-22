# runtime package

Runtime is a blocking event loop that:
- starts EventSource
- consumes Event stream
- dispatches events sequentially
- stops on context cancel or fatal error

## Lifecycle
Start(ctx) -> blocking loop -> Stop()

## Guarantees
- Only one Runtime can run at a time
- Dispatch is sequential
- Context cancellation stops everything

---

# Runtime Architecture

## High-Level Flow
EventSource -> Runtime -> Dispatcher -> Engine -> Alerts

## Responsibilities
- EventSource: produce Event stream
- Runtime: lifecycle & orchestration
- Dispatcher: routing Event to Engine/Rules
- Engine: pure evaluation logic (non-concurrent)

## Concurrency Model
- Runtime is single-threaded by default
- Engine is NOT concurrency-safe
- Parallelism is future work (Phase 2+)

## Error Handling
- Any Dispatch error stops the Runtime
- Caller is responsible for restart strategy
