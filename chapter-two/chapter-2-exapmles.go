package main

import (
	"fmt"
	"sync"
	"time"
)

// ------------------------------------------------------------
// Concurrency vs parallelism
// ------------------------------------------------------------

// concurrentWork simulates concurrent tasks.
// Whether they run in parallel depends on the runtime and hardware.
func concurrentWork() {
	var wg sync.WaitGroup

	task := func(name string, delay time.Duration) {
		defer wg.Done()
		fmt.Printf("%s started\n", name)
		time.Sleep(delay)
		fmt.Printf("%s finished\n", name)
	}

	wg.Add(2)
	go task("task-1", 500*time.Millisecond)
	go task("task-2", 500*time.Millisecond)

	wg.Wait()
}

// ------------------------------------------------------------
// Goroutines and channels
// ------------------------------------------------------------

// pingPong shows basic channel communication between goroutines.
func pingPong() {
	ch := make(chan string)

	go func() {
		ch <- "ping"
	}()

	msg := <-ch
	fmt.Println(msg)
}

// workerPoolLikeExample demonstrates how goroutines can be used
// to handle work concurrently without explicitly managing OS threads.
func workerPoolLikeExample() {
	jobs := make(chan int)
	results := make(chan int)

	worker := func(id int) {
		for job := range jobs {
			fmt.Printf("worker %d got job %d\n", id, job)
			time.Sleep(100 * time.Millisecond)
			results <- job * 2
		}
	}

	go worker(1)
	go worker(2)

	go func() {
		for j := 1; j <= 4; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	for i := 1; i <= 4; i++ {
		fmt.Println("result:", <-results)
	}
}

// ------------------------------------------------------------
// Channels for ownership transfer
// ------------------------------------------------------------

// ownershipTransferExample sends data from one goroutine to another.
// The sender no longer needs to touch the value after sending it.
func ownershipTransferExample() {
	type payload struct {
		value string
	}

	ch := make(chan payload)

	go func() {
		ch <- payload{value: "hello from producer"}
	}()

	msg := <-ch
	fmt.Println(msg.value)
}

// ------------------------------------------------------------
// Shared memory with mutexes
// ------------------------------------------------------------

// Counter protects its internal state with a mutex.
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// mutexExample shows how internal state can be guarded safely.
func mutexExample() {
	var c Counter

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}

	wg.Wait()
	fmt.Println("counter:", c.Value())
}

// ------------------------------------------------------------
// select for composing concurrent logic
// ------------------------------------------------------------

// selectExample demonstrates waiting on multiple channel events.
func selectExample() {
	fast := make(chan string)
	slow := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fast <- "fast result"
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		slow <- "slow result"
	}()

	select {
	case msg := <-fast:
		fmt.Println("received:", msg)
	case msg := <-slow:
		fmt.Println("received:", msg)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("timed out")
	}
}

// ------------------------------------------------------------
// Main
// ------------------------------------------------------------

func main() {
	// Uncomment the examples you want to run.

	// concurrentWork()
	// pingPong()
	// workerPoolLikeExample()
	// ownershipTransferExample()
	// mutexExample()
	// selectExample()
}