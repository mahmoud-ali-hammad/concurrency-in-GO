package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

// ------------------------------------------------------------
// 1. Goroutines basics
// ------------------------------------------------------------
func demoGoroutines() {
	fmt.Println("\n=== 1. Goroutines ===")

	// Basic goroutine with WaitGroup
	var wg sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("  Hello from goroutine")
	}
	wg.Add(1)
	go sayHello()
	wg.Wait()

	// Loop variable capture – correct version (pass as argument)
	wg.Add(3)
	for _, sal := range []string{"hello", "greetings", "good day"} {
		go func(s string) {
			defer wg.Done()
			fmt.Println("  Captured:", s)
		}(sal)
	}
	wg.Wait()
}

// ------------------------------------------------------------
// 2. sync.WaitGroup
// ------------------------------------------------------------
func demoWaitGroup() {
	fmt.Println("\n=== 2. sync.WaitGroup ===")
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id) * 100 * time.Millisecond)
			fmt.Printf("  Worker %d done\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("  All workers finished")
}

// ------------------------------------------------------------
// 3. sync.Mutex
// ------------------------------------------------------------
func demoMutex() {
	fmt.Println("\n=== 3. sync.Mutex ===")
	var count int
	var lock sync.Mutex
	var wg sync.WaitGroup

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			increment()
		}()
	}
	wg.Wait()
	fmt.Printf("  Final count: %d (expected 1000)\n", count)
}

// ------------------------------------------------------------
// 4. sync.RWMutex
// ------------------------------------------------------------
func demoRWMutex() {
	fmt.Println("\n=== 4. sync.RWMutex ===")
	var rw sync.RWMutex
	var data int
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rw.Lock()
			data = id
			fmt.Printf("  Writer %d set data = %d\n", id, id)
			time.Sleep(50 * time.Millisecond)
			rw.Unlock()
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rw.RLock()
			fmt.Printf("  Reader %d read data = %d\n", id, data)
			rw.RUnlock()
		}(i)
	}
	wg.Wait()
}

// ------------------------------------------------------------
// 5. sync.Cond
// ------------------------------------------------------------
func demoCond() {
	fmt.Println("\n=== 5. sync.Cond ===")
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	queue := make([]int, 0, 10)

	remove := func(delay time.Duration) {
		time.Sleep(delay)
		mu.Lock()
		queue = queue[1:]
		fmt.Println("  Removed from queue")
		mu.Unlock()
		cond.Signal()
	}

	for i := 0; i < 10; i++ {
		mu.Lock()
		for len(queue) == 2 {
			cond.Wait()
		}
		fmt.Println("  Adding to queue")
		queue = append(queue, i)
		go remove(100 * time.Millisecond)
		mu.Unlock()
	}
	time.Sleep(500 * time.Millisecond) // let final removals finish
}

// ------------------------------------------------------------
// 6. sync.Once
// ------------------------------------------------------------
func demoOnce() {
	fmt.Println("\n=== 6. sync.Once ===")
	var once sync.Once
	var wg sync.WaitGroup

	init := func() {
		fmt.Println("  Initialized only once")
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(init)
		}()
	}
	wg.Wait()
}

// ------------------------------------------------------------
// 7. sync.Pool
// ------------------------------------------------------------
func demoPool() {
	fmt.Println("\n=== 7. sync.Pool ===")
	pool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("  Creating new instance")
			return make([]byte, 1024)
		},
	}

	// Pre‑warm the pool
	for i := 0; i < 4; i++ {
		pool.Put(pool.New())
	}

	var wg sync.WaitGroup
	const workers = 10000
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := pool.Get().([]byte)
			defer pool.Put(buf)
			// Simulate work
			_ = buf[0]
		}()
	}
	wg.Wait()
	fmt.Printf("  %d workers used pooled objects\n", workers)
}

// ------------------------------------------------------------
// 8. Channels (unbuffered, buffered, ownership)
// ------------------------------------------------------------
func demoChannels() {
	fmt.Println("\n=== 8. Channels ===")

	// Unbuffered channel
	stringStream := make(chan string)
	go func() {
		stringStream <- "Hello from unbuffered channel"
	}()
	fmt.Println("  Unbuffered:", <-stringStream)

	// Buffered channel
	intStream := make(chan int, 4)
	go func() {
		defer close(intStream)
		for i := 1; i <= 5; i++ {
			intStream <- i
			fmt.Printf("  Sent: %d\n", i)
		}
	}()
	for v := range intStream {
		fmt.Printf("  Received: %d\n", v)
	}

	// Channel ownership pattern
	chanOwner := func() <-chan int {
		ch := make(chan int, 2)
		go func() {
			defer close(ch)
			for i := 0; i < 2; i++ {
				ch <- i
			}
		}()
		return ch
	}
	for v := range chanOwner() {
		fmt.Printf("  Owned channel value: %d\n", v)
	}
}

// ------------------------------------------------------------
// 9. select statement (timeout, default, for-select)
// ------------------------------------------------------------
func demoSelect() {
	fmt.Println("\n=== 9. select statement ===")

	// Timeout
	ch := make(chan int)
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timeout triggered")
	}

	// Non‑blocking with default
	select {
	case <-ch:
	default:
		fmt.Println("  Default case – no channel ready")
	}

	// for-select loop with done signal
	done := make(chan interface{})
	go func() {
		time.Sleep(500 * time.Millisecond)
		close(done)
	}()
	workCounter := 0
loop:
	for {
		select {
		case <-done:
			break loop
		default:
			workCounter++
			time.Sleep(100 * time.Millisecond)
		}
	}
	fmt.Printf("  Did %d units of work before done signal\n", workCounter)

	// Simultaneous ready channels – pseudo‑random selection
	c1, c2 := make(chan int), make(chan int)
	close(c1)
	close(c2)
	c1Count, c2Count := 0, 0
	for i := 0; i < 100; i++ {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}
	fmt.Printf("  Random selection: c1=%d, c2=%d\n", c1Count, c2Count)
}

// ------------------------------------------------------------
// 10. GOMAXPROCS
// ------------------------------------------------------------
func demoGOMAXPROCS() {
	fmt.Println("\n=== 10. GOMAXPROCS ===")
	fmt.Printf("  Current GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	// Changing it (only for demonstration; rarely needed)
	runtime.GOMAXPROCS(2)
	fmt.Printf("  After setting to 2: %d\n", runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("  Restored to %d (NumCPU)\n", runtime.GOMAXPROCS(0))
}

// ------------------------------------------------------------
// Extra: channel broadcast (closing to unblock many)
// ------------------------------------------------------------
func demoBroadcastViaClose() {
	fmt.Println("\n=== Bonus: Broadcast by closing channel ===")
	begin := make(chan interface{})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-begin // wait until closed
			fmt.Printf("  Goroutine %d started\n", i)
		}(i)
	}
	fmt.Println("  Unblocking all goroutines...")
	close(begin)
	wg.Wait()
}

// ------------------------------------------------------------
// main – run all demos
// ------------------------------------------------------------
func main() {
	fmt.Println("Go Concurrency Building Blocks – Examples from Chapter 3")
	// Optionally, comment out any demo to skip it
	demoGoroutines()
	demoWaitGroup()
	demoMutex()
	demoRWMutex()
	demoCond()
	demoOnce()
	demoPool()
	demoChannels()
	demoSelect()
	demoGOMAXPROCS()
	demoBroadcastViaClose()

	fmt.Println("\nAll examples completed.")
}