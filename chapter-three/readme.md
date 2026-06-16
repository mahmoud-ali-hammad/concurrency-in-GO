# Chapter 3: Go’s Concurrency Building Blocks – Detailed Explanation

This chapter introduces the fundamental concurrency primitives that Go provides. Understanding these building blocks is essential for writing correct, performant, and maintainable concurrent programs in Go.

## Goroutines

A **goroutine** is a lightweight thread of execution managed by the Go runtime. You start one by placing the `go` keyword before a function call.

- Goroutines are **not OS threads**; they are **coroutines** with deep runtime integration.
- They run in the same address space, so shared memory must be synchronized.
- Stacks start at a few KB and grow/shrink automatically.
- You can create hundreds of thousands of goroutines without exhausting system resources.

### The Fork‑Join Model

Go follows the **fork‑join** concurrency model:

- **Fork** – a child branch of execution splits off (using `go`).
- **Join** – the branches reunite at a **join point** (using synchronization like `sync.WaitGroup` or channels).

### Closure Capture

Goroutines close over the variables in their lexical scope **by reference**, not by value. This can lead to surprising results inside loops. Always pass loop variables as arguments to the anonymous function to capture the correct value.

---

## The `sync` Package

Low‑level memory access synchronization primitives.

### `sync.WaitGroup`

- Waits for a set of goroutines to finish.
- `Add(int)` increments the internal counter.
- `Done()` decrements it (usually called with `defer`).
- `Wait()` blocks until the counter becomes zero.

### `sync.Mutex` and `sync.RWMutex`

- **Mutex**: mutual exclusion lock (`Lock()` / `Unlock()`). Guards critical sections.
- **RWMutex**: multiple readers, single writer.
  - `RLock()` / `RUnlock()` for read access.
  - `Lock()` / `Unlock()` for write access.
  - Allows higher concurrency when reads vastly outnumber writes.

### `sync.Cond`

- Implements a **rendezvous point** for goroutines waiting for or announcing an event.
- `Wait()` suspends the calling goroutine until signaled.
- `Signal()` wakes one waiting goroutine.
- `Broadcast()` wakes all waiting goroutines.
- Important: `Wait()` atomically unlocks the associated `Locker` and suspends; upon return it relocks.

### `sync.Once`

- Ensures a function is executed **only once**, even when called from multiple goroutines.
- `Do(f func())` calls `f` exactly once.

### `sync.Pool`

- A concurrent‑safe **object pool**.
- Reduces garbage collector pressure by reusing temporary objects.
- `Get()` returns an interface{} (or creates one using the `New` function if none available).
- `Put(x)` returns the object to the pool.
- Objects in the pool may be removed automatically by the GC at any time.

---

## Channels

Channels are Go’s implementation of **CSP (Communicating Sequential Processes)**. They allow goroutines to communicate by passing values.

### Basic Operations

- Create: `ch := make(chan T)` (unbuffered) or `ch := make(chan T, capacity)` (buffered).
- Send: `ch <- value`
- Receive: `value := <-ch` or `value, ok := <-ch` (ok indicates if value was actually sent vs. zero from closed channel).
- Close: `close(ch)` – indicates no more values will be sent.

### Blocking Behaviour

| Operation | nil channel | open, empty | open, not empty | closed      |
| --------- | ----------- | ----------- | --------------- | ----------- |
| Read      | block       | block       | value           | zero, false |
| Write     | block       | write       | block (if full) | panic       |
| Close     | panic       | success     | success         | panic       |

### Channel Ownership

To write robust programs, clarify **channel ownership**:

- **Owner** goroutine:

  1. Instantiates the channel.
  2. Performs writes (or passes ownership).
  3. Closes the channel.
  4. Exposes only a read‑only view (`<-chan T`).

- **Consumer** goroutines:
  - Only read (receive) from the channel.
  - Handle blocking and channel closure.

This pattern prevents panics (write to closed/nil channel) and reduces deadlocks.

---

## The `select` Statement

`select` allows a goroutine to wait on **multiple channel operations** simultaneously.

- All cases are evaluated simultaneously (not sequentially).
- If one or more channels are ready, a **pseudo‑random** uniform choice is made among the ready ones.
- If none are ready, `select` blocks until one becomes ready.
- A `default` clause makes the select non‑blocking.
- An empty `select{}` blocks forever.

Common patterns:

- Timeout: `case <-time.After(duration):`
- Non‑blocking send/receive: `default:`
- For‑select loop with done channel.

---

## `GOMAXPROCS`

`runtime.GOMAXPROCS(n int) int` controls the number of **OS threads** that run goroutine work queues.

- Prior to Go 1.5 it was 1; now it defaults to `runtime.NumCPU()`.
- Changing it is rarely needed and can harm performance or portability.
- It can be temporarily increased to trigger race conditions more often during testing.

---

## Running the Examples

All examples are combined in a single Go file: `concurrency_examples.go`.

```bash
go run concurrency_examples.go
```
