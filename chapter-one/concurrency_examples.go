package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// This file contains small examples inspired by Chapter 1.
// Some examples are intentionally flawed to demonstrate concurrency problems.

// ------------------------------------------------------------
// Race condition example
// ------------------------------------------------------------

// raceConditionExample demonstrates a data race.
// The result is nondeterministic because the goroutine and the if statement
// may access data in different orders.
func raceConditionExample() {
	var data int

	go func() {
		data++
	}()

	if data == 0 {
		fmt.Printf("raceConditionExample: the value is %v\n", data)
	}
}

// raceConditionWithSleep is still incorrect.
// Sleeping does not fix the race; it only makes it less likely to appear.
func raceConditionWithSleep() {
	var data int

	go func() {
		data++
	}()

	time.Sleep(1 * time.Second)

	if data == 0 {
		fmt.Printf("raceConditionWithSleep: the value is %v\n", data)
	}
}

// ------------------------------------------------------------
// Synchronization example
// ------------------------------------------------------------

// synchronizedAccessExample uses a mutex to protect shared memory.
// This avoids concurrent access to the critical section.
func synchronizedAccessExample() {
	var memoryAccess sync.Mutex
	var value int

	go func() {
		memoryAccess.Lock()
		value++
		memoryAccess.Unlock()
	}()

	memoryAccess.Lock()
	if value == 0 {
		fmt.Printf("synchronizedAccessExample: the value is %v\n", value)
	} else {
		fmt.Printf("synchronizedAccessExample: the value is %v\n", value)
	}
	memoryAccess.Unlock()
}

// ------------------------------------------------------------
// Deadlock example
// ------------------------------------------------------------

type value struct {
	mu    sync.Mutex
	value int
}

// deadlockExample can deadlock if each goroutine locks one value and waits
// for the other in opposite order.
func deadlockExample() {
	var wg sync.WaitGroup

	printSum := func(v1, v2 *value) {
		defer wg.Done()

		v1.mu.Lock()
		defer v1.mu.Unlock()

		time.Sleep(2 * time.Second)

		v2.mu.Lock()
		defer v2.mu.Unlock()

		fmt.Printf("deadlockExample sum=%v\n", v1.value+v2.value)
	}

	var a, b value

	wg.Add(2)
	go printSum(&a, &b)
	go printSum(&b, &a)

	wg.Wait()
}

// ------------------------------------------------------------
// Livelock example
// ------------------------------------------------------------

// livelockExample demonstrates active work without progress.
// The goroutines keep trying to move in opposite directions.
func livelockExample() {
	cadence := sync.NewCond(&sync.Mutex{})

	go func() {
		for range time.Tick(1 * time.Millisecond) {
			cadence.Broadcast()
		}
	}()

	takeStep := func() {
		cadence.L.Lock()
		cadence.Wait()
		cadence.L.Unlock()
	}

	tryDir := func(dirName string, dir *int32, out *bytes.Buffer) bool {
		fmt.Fprintf(out, " %v", dirName)
		atomic.AddInt32(dir, 1)
		takeStep()

		if atomic.LoadInt32(dir) == 1 {
			fmt.Fprint(out, ". Success!")
			return true
		}

		takeStep()
		atomic.AddInt32(dir, -1)
		return false
	}

	var left, right int32

	tryLeft := func(out *bytes.Buffer) bool {
		return tryDir("left", &left, out)
	}
	tryRight := func(out *bytes.Buffer) bool {
		return tryDir("right", &right, out)
	}

	walk := func(walking *sync.WaitGroup, name string) {
		var out bytes.Buffer
		defer func() { fmt.Println(out.String()) }()
		defer walking.Done()

		fmt.Fprintf(&out, "%v is trying to scoot:", name)
		for i := 0; i < 5; i++ {
			if tryLeft(&out) || tryRight(&out) {
				return
			}
		}

		fmt.Fprintf(&out, "\n%v tosses her hands up in exasperation!", name)
	}

	var peopleInHallway sync.WaitGroup
	peopleInHallway.Add(2)

	go walk(&peopleInHallway, "Alice")
	go walk(&peopleInHallway, "Barbara")

	peopleInHallway.Wait()
}

// ------------------------------------------------------------
// Starvation example
// ------------------------------------------------------------

// starvationExample compares a greedy worker and a polite worker.
// The greedy worker may complete more loops because it holds the lock
// for a larger portion of its work.
func starvationExample() {
	var wg sync.WaitGroup
	var sharedLock sync.Mutex
	const runtime = 1 * time.Second

	greedyWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(3 * time.Nanosecond)
			sharedLock.Unlock()
			count++
		}
		fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
	}

	politeWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			count++
		}
		fmt.Printf("Polite worker was able to execute %v work loops\n", count)
	}

	wg.Add(2)
	go greedyWorker()
	go politeWorker()
	wg.Wait()
}

// ------------------------------------------------------------
// Main
// ------------------------------------------------------------

func main() {
	// Uncomment the one you want to observe.
	// raceConditionExample()
	// raceConditionWithSleep()
	// synchronizedAccessExample()
	// deadlockExample()
	// livelockExample()
	// starvationExample()
}