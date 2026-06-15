# Chapter 2: Modeling Your Code — Communicating Sequential Processes

This chapter explains an important idea in Go concurrency: **concurrency is not the same as parallelism**.

## Concurrency vs. parallelism

- **Concurrency** is a property of the code.
- **Parallelism** is a property of the running program.

You can write concurrent code even if it runs on only one CPU core. In that case, the runtime may still make it _look_ like tasks are happening at the same time by switching between them.

The key point is:

- You write **concurrent** code.
- The runtime and hardware may execute it **in parallel**.

## Why this distinction matters

This distinction helps you model problems more naturally.

For example:

- A web server can treat each request as an independent concurrent task.
- A calculator app or other isolated process can be reasoned about at the process level.
- Go lets you model these tasks with goroutines and channels instead of manually managing OS threads.

## What is CSP?

CSP stands for **Communicating Sequential Processes**.

It is a programming model introduced by Tony Hoare that emphasizes:

- independent sequential processes
- communication between processes
- synchronization through message passing

Go takes inspiration from CSP and builds concurrency around:

- **goroutines**
- **channels**
- **select**

## Why Go’s model is powerful

Go adds a layer of abstraction below OS threads:

- goroutines are lightweight
- the runtime schedules goroutines onto OS threads
- you usually do not need to think about thread pools directly

This makes concurrent code easier to:

- write
- read
- reason about
- scale

## Channels vs. shared memory

The chapter explains that Go supports two main approaches to concurrency:

### 1. Communicating by sharing

This means using shared memory and synchronization tools like mutexes.

This is useful when:

- protecting internal state
- defining atomic sections inside a type

### 2. Sharing by communicating

This means using channels to transfer ownership of data between goroutines.

This is useful when:

- moving data from one part of a system to another
- coordinating multiple concurrent tasks
- composing concurrent workflows

## When to use what

A simple decision guide from the chapter:

- **Transfer ownership of data?** Use channels.
- **Protect internal state of a struct?** Use mutexes or other sync primitives.
- **Coordinate multiple pieces of logic?** Channels are often the better choice.
- **Need performance in a small critical section?** Mutexes may be appropriate.

## Main takeaway

Go encourages you to:

- model the problem naturally
- use goroutines freely
- prefer channels for communication
- use shared memory synchronization when it is the clearest choice

The chapter’s philosophy can be summarized as:

> Aim for simplicity, use channels when possible, and treat goroutines like a free resource.

---

## Example file

The `chapter2_examples.go` file demonstrates:

- concurrency vs. parallelism
- goroutines
- channels
- a simple worker example
- mutex-protected state
- `select` for composing concurrent logic
