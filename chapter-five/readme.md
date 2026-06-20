# Chapter 5: Concurrency at Scale

This chapter moves beyond individual patterns and focuses on **composing** them to build large, maintainable, and resilient concurrent systems. We cover:

- **Error Propagation** – crafting rich errors that include context, stack traces, and user‑friendly messages.
- **Timeouts and Cancellation** – strategies for preempting work, handling cleanup, and avoiding death‑spirals.
- **Heartbeats** – two types: interval‑based and work‑start pulses; used for health checks and deterministic testing.
- **Replicated Requests** – dispatching the same request to multiple handlers and taking the fastest response.
- **Rate Limiting** – the token bucket algorithm, composing multiple limits (per‑second, per‑minute), and applying them to API, disk, and network resources.
- **Healing Unhealthy Goroutines** – using stewards to monitor and restart goroutines that fail or become unresponsive.

All examples are collected in a single Go file: `concurrency_at_scale_examples.go`.

---

## Detailed Explanations

### 1. Error Propagation

A well‑formed error should include:

- **What happened** – a clear, contextual message.
- **When and where** – a stack trace, timestamp, and machine identifier.
- **User‑friendly message** – concise, human‑centric text.
- **How to get more info** – a log ID or reference.

We distinguish between **bugs** (unwrapped, raw errors) and **known edge cases** (wrapped, well‑crafted errors). At module boundaries, we wrap incoming errors with our own type, adding context. User‑facing code checks the error type and either displays a friendly message or a generic “unexpected issue” message with a log ID.

**Example**: `demoErrorPropagation()` shows a multi‑package simulation with custom error types and wrapping.

---

### 2. Timeouts and Cancellation

Timeouts prevent system saturation, stale data, and deadlocks. Cancellation can originate from timeouts, user intervention, parent cancellation, or replicated requests.

Key considerations:

- Ensure your goroutines are **preemptable** – break long operations into smaller, interruptible chunks.
- Handle **rollback** of partial state changes – modify state as late as possible.
- Avoid **duplicate messages** by using bidirectional communication or heartbeats.

**Example**: `demoTimeoutCancellation()` demonstrates a goroutine that can be canceled via `context.WithTimeout` and handles partial state modifications.

---

### 3. Heartbeats

Heartbeats signal that a goroutine is alive and making progress. Two types:

- **Interval‑based** – pulses at a fixed frequency; used to detect stalls.
- **Work‑start** – pulses before each unit of work; used for deterministic testing.

Heartbeats allow us to write tests without timeouts, making them reliable and fast. They also help avoid deadlocks when a goroutine is unresponsive.

**Examples**: `demoHeartbeats()` shows interval‑based and work‑start pulses; `demoHeartbeatTesting()` demonstrates deterministic testing with heartbeats.

---

### 4. Replicated Requests

When speed is critical, dispatch a request to multiple handlers (goroutines) and return the first response. Cancel the remaining handlers.

This technique improves latency at the cost of resource replication. It works well when handlers have different runtime conditions (e.g., different machines, network paths).

**Example**: `demoReplicatedRequests()` spins up 10 handlers with random delays, takes the fastest, and cancels the rest.

---

### 5. Rate Limiting

Rate limiting protects systems from overload and malicious use. The **token bucket** algorithm is common: a bucket holds `b` tokens, refilled at rate `r` tokens/second. Requests consume tokens; if none are available, they block or fail.

We can compose multiple limiters (e.g., per‑second and per‑minute) and apply them to different resources (API, disk, network). The `MultiLimiter` type aggregates them and waits for all to grant access.

**Example**: `demoRateLimiting()` creates a composite limiter and uses it to throttle API calls.

---

### 6. Healing Unhealthy Goroutines

Long‑lived goroutines can become stuck. A **steward** monitors a **ward** goroutine using heartbeats. If the ward stops sending pulses within a timeout, the steward closes its `done` channel, starts a new ward, and continues monitoring.

Stewards themselves are monitorable (return a `startGoroutineFn`). Wards can be stateful or stateless; closures allow passing parameters.

**Example**: `demoHealing()` simulates an unhealthy ward (that logs errors) and restarts it repeatedly.

---

## Running the Examples

```bash
go run concurrency_at_scale_examples.go
```
