# Chapter 6: Goroutines and the Go Runtime

This chapter delves into the internals of how the Go runtime manages goroutines, focusing on the work‑stealing scheduler, the difference between task and continuation stealing, and the roles of **G**, **M**, and **P**. Understanding these concepts helps you write more efficient concurrent code and debug performance issues.

---

## Key Concepts

### Work‑Stealing Scheduler

Go uses a **work‑stealing** strategy to distribute goroutines across OS threads. Instead of a single global queue (which becomes a bottleneck), each processor (P) has its own **double‑ended queue (deque)** of goroutines.

**Rules**:

- At a **fork** (`go` statement), the goroutine is added to the tail of the current P’s deque.
- At a **join** (e.g., channel receive, `WaitGroup.Wait`), if the join is not yet satisfied, the thread pops work from the **tail** of its own deque (LIFO) to keep cache locality.
- Idle threads **steal** work from the **head** of another random P’s deque (FIFO) to balance load.

This design minimises contention, improves cache locality, and keeps CPUs busy.

---

### Continuation Stealing vs Task Stealing

|                      | Continuation Stealing (Go)         | Task Stealing (others)          |
| -------------------- | ---------------------------------- | ------------------------------- |
| **Queued work**      | The continuation after a `go` call | The goroutine itself (the task) |
| **Queue size**       | Bounded (LIFO tail)                | Unbounded                       |
| **Order**            | Serial (like function calls)       | Out‑of‑order                    |
| **Join stalls**      | Less frequent                      | More frequent                   |
| **Compiler support** | Required                           | Not required                    |

Go steals **continuations**, which means it pushes the continuation (the code after the goroutine call) onto the deque and immediately runs the goroutine. This makes goroutines behave almost like function calls and reduces stalls at join points.

---

### The Scheduler’s Trio: G, M, P

- **G** – a goroutine (contains stack, instruction pointer, and state).
- **M** – an OS thread (machine).
- **P** – a processor (context) that holds a runqueue of Gs.

`GOMAXPROCS` sets the number of Ps. The runtime ensures at least one M per P, but can have more Ms for blocking syscalls.

**Blocking optimisations**:

- When a G blocks on I/O or a syscall, its M blocks too.
- The P is **detached** from that M and handed to another idle M, so the P can keep scheduling other Gs.
- When the blocking G unblocks, its M tries to steal a P; if none are available, the G is placed on a **global runqueue** and the M goes to sleep.

---

## Important Takeaways

- You rarely need to tweak `GOMAXPROCS`; the default (number of CPU cores) is optimal for most workloads.
- Goroutines are **preemptible** only at function calls – tight loops without function calls can starve the scheduler.
- The runtime’s design makes `go` incredibly cheap and scalable.

---

## Running the Examples

```bash
go run runtime_examples.go
```
