# Chapter 1: An Introduction to Concurrency in Go

This chapter introduces the core ideas behind concurrency and explains why it matters in modern software development.

## What concurrency means

Concurrency means multiple tasks are in progress at the same time. They may not all be executing simultaneously on different CPU cores, but they are making progress in overlapping time periods.

In Go, concurrency is a first-class concept. The language makes it easier to model concurrent problems clearly and safely.

## Why concurrency became important

The chapter explains the historical reasons concurrency became so important:

- **Moore’s Law slowed down**: CPU performance gains stopped coming from raw clock speed increases.
- **Multicore processors became common**: Instead of making one CPU faster, hardware started adding more cores.
- **Amdahl’s Law**: Even if part of a program can run in parallel, the sequential part limits overall speedup.
- **Cloud computing and web scale**: Applications began running across many machines and regions, increasing the need for concurrency.

## Why concurrency is hard

Concurrency is difficult because timing is unpredictable. Problems may only appear under load or after subtle changes in runtime conditions.

The chapter covers several major concurrency hazards:

### 1. Race conditions

A race condition happens when the order of operations matters, but the program does not guarantee that order.

A common form is a **data race**, where two goroutines access the same memory at the same time and at least one of them writes to it.

### 2. Atomicity

An operation is atomic if it cannot be interrupted within a given context.

A simple statement like `i++` may look like one operation, but it actually consists of:

- reading `i`
- incrementing the value
- writing it back

That means it is not necessarily atomic in a concurrent program.

### 3. Memory access synchronization

One way to protect shared memory is to synchronize access, often with a mutex.

This ensures only one goroutine at a time can access a critical section of code.

However, synchronization has tradeoffs:

- it can slow programs down
- it can be abused or ignored
- it does not automatically solve logical ordering problems

### 4. Deadlocks

A deadlock occurs when goroutines are waiting on one another forever.

The chapter explains the Coffman conditions, which are the four conditions that must all exist for a deadlock to happen:

- mutual exclusion
- wait-for condition
- no preemption
- circular wait

If you break any one of those conditions, you can prevent deadlocks.

### 5. Livelocks

A livelock is similar to a deadlock, but the goroutines are still active.

They keep doing work, but they never make real progress.

The hallway example in the chapter shows how two people can keep stepping aside forever without ever passing each other.

### 6. Starvation

Starvation happens when a goroutine cannot get the resources it needs to make progress.

Unlike deadlock, the system may still be working overall, but one goroutine is being unfairly blocked or delayed.

## What Go gives you

Go helps make concurrency easier by providing:

- lightweight goroutines
- channels for communication
- a runtime that multiplexes goroutines onto OS threads
- garbage collection, which reduces memory-management burden

This lets developers focus more on the problem domain and less on thread management.

## Main takeaway

The chapter’s core message is:

- concurrency is powerful
- concurrency is difficult
- correct concurrency requires careful thinking about ordering, atomicity, and shared state
- Go provides tools that make concurrent programming clearer and safer

---

## Example code

The `concurrency_examples.go` file contains small examples based on the chapter, including:

- a race condition example
- a mutex-protected version
- a deadlock example
- livelock-style behavior
- a starvation demonstration
