# Chapter 4: Concurrency Patterns in Go

This chapter builds on the fundamental building blocks from Chapter 3 and introduces **composition patterns** that make concurrent code scalable, maintainable, and robust. We cover:

- Confinement (ad‑hoc and lexical)
- The `for‑select` loop
- Preventing goroutine leaks
- The `or`‑channel
- Error handling in concurrent pipelines
- Pipelines (stages, generators, fan‑out/fan‑in)
- The `or‑done`‑channel
- The `tee`‑channel
- The `bridge`‑channel
- Queuing and Little’s Law
- The `context` package (cancellation and request‑scoped data)

All examples are collected in a single Go file: `concurrency_patterns_examples.go`.

---

## Detailed Explanations

### 1. Confinement

**Confinement** ensures that data is only ever accessed from one concurrent process, eliminating the need for synchronization.

- **Ad‑hoc confinement**: achieved by convention (e.g., only one goroutine touches a slice). Risky.
- **Lexical confinement**: enforced by the compiler using lexical scope (e.g., exposing only a read‑only channel, or passing a sub‑slice).

**Example**: `demoConfinement()` shows passing sub‑slices to separate goroutines.

---

### 2. The `for‑select` Loop

A ubiquitous pattern for looping while waiting on channels:

- Sending iteration variables on a channel.
- Looping infinitely until a `done` channel is closed (with or without a `default` branch).

**Example**: `demoForSelect()`.

---

### 3. Preventing Goroutine Leaks

Goroutines are not garbage‑collected; they must be explicitly terminated. The standard approach is to pass a **`done` channel** from parent to child. The parent closes it to signal cancellation.

- If a goroutine blocks on a read or write, the `done` channel allows it to exit cleanly.
- The convention: _If a goroutine creates a goroutine, it is responsible for ensuring it can be stopped._

**Example**: `demoGoroutineLeak()` shows a safe producer that respects a `done` channel.

---

### 4. The `or`‑channel

Combine multiple `done`‑style channels into a single channel that closes when **any** of its components close. Useful when you cannot know the number of channels at compile time.

The `or` function recursively creates a tree of goroutines to multiplex the signals.

**Example**: `demoOrChannel()`.

---

### 5. Error Handling

In concurrent code, errors should not be swallowed inside goroutines. Instead, they should be returned as part of a result type, coupled with the successful value, and passed along the same communication lines.

**Example**: `demoErrorHandling()` shows a `checkStatus` function that returns a `Result` struct containing either a response or an error.

---

### 6. Pipelines

A **pipeline** is a series of stages that consume, transform, and emit data. Stages can be **batch** or **stream** oriented. In Go, pipelines are built with channels.

**Key components**:

- **Generators**: convert discrete values into a channel stream (e.g., `repeat`, `repeatFn`).
- **Stages**: take a channel, process each value, and return a new channel.
- **Fan‑out / Fan‑in**: parallelize a slow stage by running multiple copies, then merge their results.

**Additional utility patterns**:

- **or‑done‑channel**: wraps a channel to respect a `done` channel without cluttering loops.
- **tee‑channel**: duplicates a stream into two separate channels.
- **bridge‑channel**: flattens a channel of channels into a single channel.

**Examples**: `demoPipelines()` demonstrates `repeat`, `take`, and a prime‑finding pipeline with fan‑out/fan‑in.

---

### 7. Queuing

Queues (buffered channels) are used to decouple stages, batch requests, or absorb bursts. They **do not** reduce total runtime (per Little’s Law). They can help avoid negative feedback loops (death‑spirals) and allow chunking for performance.

Little’s Law: `L = λW`

- `L`: average number of units in the system
- `λ`: average arrival rate
- `W`: average time a unit spends in the system

Queues should be placed **at the entrance** or **before batching stages**.

**Example**: `demoQueuing()` compares buffered vs. unbuffered writes.

---

### 8. The `context` Package

The `context` package provides a standard way to:

- Cancel branches of a call‑graph (`WithCancel`, `WithTimeout`, `WithDeadline`).
- Pass request‑scoped data (`WithValue`).

It replaces the `done` channel pattern with a richer API that includes deadlines and error reasons.

**Guidelines for data stored in a `Context`**:

- Must be request‑scoped and cross API boundaries.
- Should be immutable and simple (e.g., request ID, auth token).
- Should not be used for optional parameters.

**Examples**: `demoContextCancellation()` shows timeout and parent cancellation; `demoContextValues()` shows type‑safe accessors.

---

## Running the Examples

```bash
go run concurrency_patterns_examples.go
```
