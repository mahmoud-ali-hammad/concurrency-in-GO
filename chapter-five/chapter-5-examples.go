
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// ------------------------------------------------------------
// 1. Error Propagation
// ------------------------------------------------------------
func demoErrorPropagation() {
	fmt.Println("\n=== 1. Error Propagation ===")

	// Custom error type with stack trace and misc info
	type MyError struct {
		Inner      error
		Message    string
		StackTrace string
		Misc       map[string]interface{}
	}
	wrapError := func(err error, messagef string, msgArgs ...interface{}) MyError {
		return MyError{
			Inner:      err,
			Message:    fmt.Sprintf(messagef, msgArgs...),
			StackTrace: string(debug.Stack()),
			Misc:       make(map[string]interface{}),
		}
	}
	func (err MyError) Error() string { return err.Message }

	// Low-level module
	type LowLevelErr struct{ error }
	isGloballyExec := func(path string) (bool, error) {
		// Simulate a file stat error
		return false, LowLevelErr{wrapError(fmt.Errorf("file not found"), "stat %s: no such file", path)}
	}

	// Intermediate module
	type IntermediateErr struct{ error }
	runJob := func(id string) error {
		const jobBinPath = "/bad/job/binary"
		isExec, err := isGloballyExec(jobBinPath)
		if err != nil {
			// Wrap the error with intermediate context
			return IntermediateErr{wrapError(err, "cannot run job %q: requisite binaries not available", id)}
		}
		if !isExec {
			return wrapError(nil, "cannot run job %q: requisite binaries are not executable", id)
		}
		// Simulate command execution
		return nil
	}

	// Top-level handler
	handleError := func(key int, err error, message string) {
		log.SetPrefix(fmt.Sprintf("[logID: %v]: ", key))
		log.Printf("%#v", err)
		fmt.Printf("[%v] %v\n", key, message)
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	err := runJob("1")
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug."
		if _, ok := err.(IntermediateErr); ok {
			msg = err.Error()
		}
		handleError(1, err, msg)
	}
}

// ------------------------------------------------------------
// 2. Timeouts and Cancellation
// ------------------------------------------------------------
func demoTimeoutCancellation() {
	fmt.Println("\n=== 2. Timeouts and Cancellation ===")

	// Simulate a long-running calculation that can be preempted
	reallyLongCalculation := func(done <-chan struct{}, value int) int {
		// Simulate work by sleeping in chunks
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				return 0
			default:
				time.Sleep(100 * time.Millisecond) // chunk of work
			}
		}
		return value * 2
	}

	doWork := func(done <-chan struct{}) int {
		value := 42
		// Preemptable work
		intermediate := reallyLongCalculation(done, value)
		select {
		case <-done:
			return 0
		default:
		}
		// More work...
		result := intermediate + 10
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := ctx.Done()
	resultChan := make(chan int, 1)
	go func() {
		resultChan <- doWork(done)
	}()

	select {
	case res := <-resultChan:
		fmt.Printf("  Result: %d\n", res)
	case <-done:
		fmt.Println("  Operation timed out")
	}
}

// ------------------------------------------------------------
// 3. Heartbeats
// ------------------------------------------------------------
func demoHeartbeats() {
	fmt.Println("\n=== 3. Heartbeats ===")

	// Interval-based heartbeat
	doWorkInterval := func(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
		heartbeat := make(chan interface{})
		results := make(chan time.Time)
		go func() {
			defer close(heartbeat)
			defer close(results)
			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(2 * pulseInterval)

			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}
			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}
			for {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartbeat, results
	}

	done := make(chan interface{})
	time.AfterFunc(5*time.Second, func() { close(done) })
	const timeout = 1 * time.Second
	heartbeat, results := doWorkInterval(done, timeout/2)

	// Consume heartbeats and results until timeout or done
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				fmt.Println("  Heartbeat channel closed")
				return
			}
			fmt.Println("  pulse")
		case r, ok := <-results:
			if !ok {
				fmt.Println("  Results channel closed")
				return
			}
			fmt.Printf("  result %v\n", r.Second())
		case <-time.After(timeout):
			fmt.Println("  Timeout: no signal")
			return
		}
	}
}

// Work-start heartbeat for testing
func demoHeartbeatTesting() {
	fmt.Println("\n=== 3b. Heartbeat for Testing ===")

	doWork := func(done <-chan interface{}) (<-chan interface{}, <-chan int) {
		heartbeat := make(chan interface{}, 1)
		workStream := make(chan int)
		go func() {
			defer close(heartbeat)
			defer close(workStream)
			for i := 0; i < 5; i++ {
				// pulse before work
				select {
				case heartbeat <- struct{}{}:
				default:
				}
				select {
				case <-done:
					return
				case workStream <- rand.Intn(10):
				}
			}
		}()
		return heartbeat, workStream
	}

	done := make(chan interface{})
	defer close(done)
	heartbeat, results := doWork(done)

	// Wait for first heartbeat (ensures goroutine started)
	<-heartbeat
	fmt.Println("  Goroutine started")
	for r := range results {
		fmt.Printf("  result: %d\n", r)
	}
}

// ------------------------------------------------------------
// 4. Replicated Requests
// ------------------------------------------------------------
func demoReplicatedRequests() {
	fmt.Println("\n=== 4. Replicated Requests ===")

	doWork := func(done <-chan interface{}, id int, wg *sync.WaitGroup, result chan<- int) {
		defer wg.Done()
		// Simulate random load between 1 and 5 seconds
		load := time.Duration(1+rand.Intn(5)) * time.Second
		select {
		case <-done:
			return
		case <-time.After(load):
		}
		select {
		case <-done:
			return
		case result <- id:
		}
	}

	done := make(chan interface{})
	result := make(chan int)
	var wg sync.WaitGroup
	numHandlers := 10
	wg.Add(numHandlers)
	for i := 0; i < numHandlers; i++ {
		go doWork(done, i, &wg, result)
	}

	first := <-result
	close(done)
	wg.Wait()
	fmt.Printf("  First response from handler #%d\n", first)
}

// ------------------------------------------------------------
// 5. Rate Limiting (simple token bucket)
// ------------------------------------------------------------
// Simple token bucket implementation
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	rate     float64   // tokens per second
	capacity float64   // max tokens
	last     time.Time
}

func NewLimiter(rate float64, capacity int) *Limiter {
	return &Limiter{
		tokens:   float64(capacity),
		rate:     rate,
		capacity: float64(capacity),
		last:     time.Now(),
	}
}

// Wait blocks until a token is available
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		// add tokens since last check
		elapsed := now.Sub(l.last).Seconds()
		l.tokens += elapsed * l.rate
		if l.tokens > l.capacity {
			l.tokens = l.capacity
		}
		l.last = now
		if l.tokens >= 1.0 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// need to wait
		waitTime := time.Duration((1.0-l.tokens)/l.rate*1000) * time.Millisecond
		l.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}

// MultiLimiter combines multiple limiters
type MultiLimiter struct {
	limiters []*Limiter
}

func NewMultiLimiter(limiters ...*Limiter) *MultiLimiter {
	return &MultiLimiter{limiters: limiters}
}
func (m *MultiLimiter) Wait(ctx context.Context) error {
	for _, l := range m.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// API connection with rate limiters
type APIConnection struct {
	apiLimiter    *MultiLimiter
	diskLimiter   *Limiter
	networkLimiter *Limiter
}

func Open() *APIConnection {
	// Per-second: 2 tokens, burst 2
	secLimiter := NewLimiter(2.0, 2)
	// Per-minute: 10 tokens, burst 10
	minLimiter := NewLimiter(10.0/60.0, 10)
	apiMulti := NewMultiLimiter(secLimiter, minLimiter)

	diskLimiter := NewLimiter(1.0, 1) // 1 read/sec
	netLimiter := NewLimiter(3.0, 3)  // 3 requests/sec

	return &APIConnection{
		apiLimiter:    apiMulti,
		diskLimiter:   diskLimiter,
		networkLimiter: netLimiter,
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := NewMultiLimiter(a.apiLimiter, a.diskLimiter).Wait(ctx); err != nil {
		return err
	}
	// Simulate work
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := NewMultiLimiter(a.apiLimiter, a.networkLimiter).Wait(ctx); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

func demoRateLimiting() {
	fmt.Println("\n=== 5. Rate Limiting ===")

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConn := Open()
	var wg sync.WaitGroup
	const total = 10
	wg.Add(total * 2)

	ctx := context.Background()

	// Read files
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			if err := apiConn.ReadFile(ctx); err != nil {
				log.Printf("ReadFile error: %v", err)
				return
			}
			log.Println("ReadFile")
		}()
	}

	// Resolve addresses
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			if err := apiConn.ResolveAddress(ctx); err != nil {
				log.Printf("ResolveAddress error: %v", err)
				return
			}
			log.Println("ResolveAddress")
		}()
	}

	wg.Wait()
	log.Println("Done.")
}

// ------------------------------------------------------------
// 6. Healing Unhealthy Goroutines
// ------------------------------------------------------------
func demoHealing() {
	fmt.Println("\n=== 6. Healing Unhealthy Goroutines ===")

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	// or-channel (from chapter 4)
	or := func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}
		orDone := make(chan interface{})
		go func() {
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}

	type startGoroutineFn func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{}

	newSteward := func(timeout time.Duration, startGoroutine startGoroutineFn) startGoroutineFn {
		return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
			heartbeat := make(chan interface{})
			go func() {
				defer close(heartbeat)
				var wardDone chan interface{}
				var wardHeartbeat <-chan interface{}
				startWard := func() {
					wardDone = make(chan interface{})
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)
			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)
					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat:
							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()
			return heartbeat
		}
	}

	// Ward that simulates an unhealthy behaviour (logs negative values and exits)
	doWorkFn := func(done <-chan interface{}, intList ...int) (startGoroutineFn, <-chan interface{}) {
		intChanStream := make(chan (<-chan interface{}))
		intStream := bridge(done, intChanStream)

		doWork := func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
			intStream := make(chan interface{})
			heartbeat := make(chan interface{})
			go func() {
				defer close(intStream)
				select {
				case intChanStream <- intStream:
				case <-done:
					return
				}
				pulse := time.Tick(pulseInterval)
				for {
				valueLoop:
					for _, intVal := range intList {
						if intVal < 0 {
							log.Printf("negative value: %v", intVal)
							return
						}
						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}:
								default:
								}
							case intStream <- intVal:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()
			return heartbeat
		}
		return doWork, intStream
	}

	// bridge-channel from chapter 4
	bridge := func(done <-chan interface{}, chanStream <-chan <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				var stream <-chan interface{}
				select {
				case maybeStream, ok := <-chanStream:
					if !ok {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valStream <- val:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}
	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	done := make(chan interface{})
	defer close(done)

	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for val := range take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", val)
	}
}

// Helper: take from pipeline
func take(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
	takeStream := make(chan interface{})
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()
	return takeStream
}

// ------------------------------------------------------------
// main – run all demos
// ------------------------------------------------------------
func main() {
	fmt.Println("Concurrency at Scale – Examples from Chapter 5")
	demoErrorPropagation()
	demoTimeoutCancellation()
	demoHeartbeats()
	demoHeartbeatTesting()
	demoReplicatedRequests()
	demoRateLimiting()
	demoHealing()
	fmt.Println("\nAll examples completed.")
}