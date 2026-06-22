
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ------------------------------------------------------------
// 1. Inspecting GOMAXPROCS and goroutine count
// ------------------------------------------------------------
func demoGOMAXPROCS() {
	fmt.Println("\n=== 1. GOMAXPROCS and goroutine count ===")
	fmt.Printf("  GOMAXPROCS = %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("  NumCPU     = %d\n", runtime.NumCPU())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
		}(i)
	}
	// Give goroutines time to start
	time.Sleep(5 * time.Millisecond)
	fmt.Printf("  NumGoroutine = %d (active)\n", runtime.NumGoroutine())
	wg.Wait()
	fmt.Printf("  NumGoroutine = %d (after join)\n", runtime.NumGoroutine())
}

// ------------------------------------------------------------
// 2. Fork‑join with Fibonacci (shows work‑stealing behaviour)
// ------------------------------------------------------------
func demoFibonacci() {
	fmt.Println("\n=== 2. Fibonacci with goroutines (fork‑join) ===")
	// Recursive Fibonacci using channels (fork‑join)
	var fib func(n int) <-chan int
	fib = func(n int) <-chan int {
		result := make(chan int)
		go func() {
			defer close(result)
			if n <= 2 {
				result <- 1
				return
			}
			left := <-fib(n - 1)
			right := <-fib(n - 2)
			result <- left + right
		}()
		return result
	}
	// Compute fib(10) – small enough to finish quickly
	start := time.Now()
	fib10 := <-fib(10)
	elapsed := time.Since(start)
	fmt.Printf("  fib(10) = %d (computed in %v)\n", fib10, elapsed)

	// Note: we can also show the number of goroutines created
	// but it's non‑deterministic; we just demonstrate the pattern.
}

// ------------------------------------------------------------
// 3. Blocking I/O – scheduler detaches P from blocked M
// ------------------------------------------------------------
func demoBlockingIO() {
	fmt.Println("\n=== 3. Blocking I/O (syscall simulation) ===")
	var wg sync.WaitGroup
	start := time.Now()

	// Simulate blocking operations (e.g., reading from a slow file)
	block := func(id int) {
		defer wg.Done()
		// Simulate a syscall that blocks for ~50ms
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("  Block %d done after %v\n", id, time.Since(start))
	}

	// Start many goroutines; the scheduler will detach Ps from blocked Ms
	numGoroutines := 20
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go block(i)
	}
	wg.Wait()
	fmt.Printf("  All blocks completed in %v\n", time.Since(start))
}

// ------------------------------------------------------------
// 4. Voluntarily yielding with runtime.Gosched()
// ------------------------------------------------------------
func demoGosched() {
	fmt.Println("\n=== 4. runtime.Gosched() ===")
	done := make(chan bool)

	go func() {
		for i := 0; i < 3; i++ {
			fmt.Printf("  Goroutine A: %d\n", i)
			// Yield to allow other goroutines to run
			runtime.Gosched()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 3; i++ {
			fmt.Printf("  Goroutine B: %d\n", i)
			runtime.Gosched()
		}
		done <- true
	}()

	<-done
	<-done
	fmt.Println("  Both goroutines finished.")
}

// ------------------------------------------------------------
// 5. Preemption: tight loop without function calls (not preemptible)
// ------------------------------------------------------------
func demoPreemption() {
	fmt.Println("\n=== 5. Preemption (tight loop) ===")
	// Go 1.14+ preempts goroutines at function calls, but tight loops
	// without function calls are not preemptible.
	// This example demonstrates that a busy loop can starve other goroutines.
	done := make(chan bool)

	// Starving goroutine – no function calls inside the loop
	go func() {
		count := 0
		for {
			count++
			if count%100000000 == 0 {
				// This print is a function call, so it allows preemption.
				// Without this print, the loop would be non‑preemptible.
				// To illustrate, we keep the print infrequent.
				// But for clarity, we'll just run a short loop with a print.
				break
			}
		}
		fmt.Println("  Starving goroutine finished.")
		done <- true
	}()

	// Other goroutine that should get some time
	go func() {
		fmt.Println("  Other goroutine: I need CPU time!")
		done <- true
	}()

	<-done
	<-done
	fmt.Println("  Both finished. (The tight loop allowed preemption only at print).")
}

// ------------------------------------------------------------
// main – run all demos
// ------------------------------------------------------------
func main() {
	fmt.Println("Goroutines and the Go Runtime – Examples from Chapter 6")
	demoGOMAXPROCS()
	demoFibonacci()
	demoBlockingIO()
	demoGosched()
	demoPreemption()
	fmt.Println("\nAll examples completed.")
}