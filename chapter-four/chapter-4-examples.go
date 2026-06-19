import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ------------------------------------------------------------
// 1. Confinement (Lexical)
// ------------------------------------------------------------
func demoConfinement() {
	fmt.Println("\n=== 1. Confinement ===")
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()
		var buff bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}
	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")
	go printData(&wg, data[:3]) // only first 3 bytes
	go printData(&wg, data[3:]) // only last 3 bytes
	wg.Wait()
}

// ------------------------------------------------------------
// 2. for-select loop
// ------------------------------------------------------------
func demoForSelect() {
	fmt.Println("\n=== 2. for-select loop ===")
	done := make(chan interface{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(done)
	}()
	stringStream := make(chan string)
	go func() {
		for _, s := range []string{"a", "b", "c"} {
			select {
			case <-done:
				return
			case stringStream <- s:
			}
		}
		close(stringStream)
	}()
	for s := range stringStream {
		fmt.Println("Received:", s)
	}
}

// ------------------------------------------------------------
// 3. Preventing Goroutine Leaks
// ------------------------------------------------------------
func demoGoroutineLeak() {
	fmt.Println("\n=== 3. Preventing Goroutine Leaks ===")
	doWork := func(done <-chan interface{}, strings <-chan string) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("  doWork exited")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					fmt.Println("  Received:", s)
				case <-done:
					return
				}
			}
		}()
		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("  Canceling doWork...")
		close(done)
	}()
	<-terminated
	fmt.Println("  Done.")
}

// ------------------------------------------------------------
// 4. or-channel
// ------------------------------------------------------------
func demoOrChannel() {
	fmt.Println("\n=== 4. or-channel ===")
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
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
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("  Done after %v\n", time.Since(start))
}

// ------------------------------------------------------------
// 5. Error Handling
// ------------------------------------------------------------
func demoErrorHandling() {
	fmt.Println("\n=== 5. Error Handling ===")
	type Result struct {
		Error    error
		Response *http.Response
	}
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}
	done := make(chan interface{})
	defer close(done)
	urls := []string{"https://www.google.com", "https://badhost"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("  error: %v\n", result.Error)
			continue
		}
		fmt.Printf("  Response: %v\n", result.Response.Status)
	}
}

// ------------------------------------------------------------
// 6. Pipelines (generators, fan-out/fan-in, or-done, tee, bridge)
// ------------------------------------------------------------
func demoPipelines() {
	fmt.Println("\n=== 6. Pipelines ===")

	// Generators: repeat, repeatFn, take, toInt, primeFinder
	repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}
	repeatFn := func(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}
	take := func(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
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
	toInt := func(done <-chan interface{}, valueStream <-chan interface{}) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()
		return intStream
	}
	primeFinder := func(done <-chan interface{}, intStream <-chan int) <-chan interface{} {
		primeStream := make(chan interface{})
		go func() {
			defer close(primeStream)
			for i := range intStream {
				// Naive prime test
				isPrime := true
				for j := 2; j*j <= i; j++ {
					if i%j == 0 {
						isPrime = false
						break
					}
				}
				if isPrime {
					select {
					case <-done:
						return
					case primeStream <- i:
					}
				}
			}
		}()
		return primeStream
	}
	// Fan-in function
	fanIn := func(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
		var wg sync.WaitGroup
		multiplexedStream := make(chan interface{})
		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}
		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()
		return multiplexedStream
	}
	// or-done-channel
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
	// tee-channel
	tee := func(done <-chan interface{}, in <-chan interface{}) (_, _ <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})
		go func() {
			defer close(out1)
			defer close(out2)
			for val := range orDone(done, in) {
				var out1, out2 = out1, out2
				for i := 0; i < 2; i++ {
					select {
					case <-done:
					case out1 <- val:
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()
		return out1, out2
	}
	// bridge-channel
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

	// Demonstrate generators and take
	done := make(chan interface{})
	defer close(done)
	fmt.Println("  repeat + take:")
	for num := range take(done, repeat(done, 1), 5) {
		fmt.Printf("    %v ", num)
	}
	fmt.Println()

	// Demonstrate fan-out/fan-in prime finder
	fmt.Println("  Fan-out/fan-in prime finder:")
	randInt := func() interface{} { return rand.Intn(50000000) }
	randIntStream := toInt(done, repeatFn(done, randInt))
	numFinders := runtime.NumCPU()
	finders := make([]<-chan interface{}, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}
	start := time.Now()
	primeStream := fanIn(done, finders...)
	for prime := range take(done, primeStream, 5) {
		fmt.Printf("    %d ", prime)
	}
	fmt.Printf("\n  Found 5 primes in %v\n", time.Since(start))

	// Demonstrate tee
	fmt.Println("  tee-channel:")
	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))
	for v1 := range out1 {
		fmt.Printf("    out1: %v, out2: %v\n", v1, <-out2)
	}

	// Demonstrate bridge
	fmt.Println("  bridge-channel:")
	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 5; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}
	for v := range bridge(done, genVals()) {
		fmt.Printf("    %v ", v)
	}
	fmt.Println()
}

// ------------------------------------------------------------
// 7. Queuing (buffered vs unbuffered)
// ------------------------------------------------------------
func demoQueuing() {
	fmt.Println("\n=== 7. Queuing ===")
	// Simulate buffered write vs unbuffered write
	// We'll use a simple benchmark-like function
	performWrite := func(writer io.Writer, n int) {
		done := make(chan interface{})
		defer close(done)
		for bt := range take(done, repeat(done, byte(0)), n) {
			writer.Write([]byte{bt.(byte)})
		}
	}
	// Unbuffered: write directly to a bytes.Buffer (but that's already buffered)
	// So we simulate by writing to a no-op writer that does nothing (to isolate overhead)
	noopWriter := ioutil.Discard
	start := time.Now()
	performWrite(noopWriter, 10000)
	unbufferedTime := time.Since(start)

	// Buffered: use bufio.Writer
	bufWriter := &bytes.Buffer{}
	start = time.Now()
	performWrite(bufWriter, 10000)
	bufferedTime := time.Since(start)
	fmt.Printf("  Unbuffered write time: %v\n", unbufferedTime)
	fmt.Printf("  Buffered write time:   %v\n", bufferedTime)
	// Note: In this simple example, buffered may be slower due to allocation overhead.
	// The point is to illustrate the concept.
}

// ------------------------------------------------------------
// 8. context package
// ------------------------------------------------------------
func demoContextCancellation() {
	fmt.Println("\n=== 8. context (cancellation) ===")
	// Simulate a scenario with timeout and parent cancellation
	// Functions from the book example
	locale := func(ctx context.Context) (string, error) {
		if deadline, ok := ctx.Deadline(); ok {
			if deadline.Sub(time.Now().Add(1*time.Minute)) <= 0 {
				return "", context.DeadlineExceeded
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Minute):
		}
		return "EN/US", nil
	}
	genGreeting := func(ctx context.Context) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "hello", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}
	genFarewell := func(ctx context.Context) (string, error) {
		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "goodbye", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}
	printGreeting := func(ctx context.Context) error {
		greeting, err := genGreeting(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("  %s world!\n", greeting)
		return nil
	}
	printFarewell := func(ctx context.Context) error {
		farewell, err := genFarewell(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("  %s world!\n", farewell)
		return nil
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("  cannot print greeting: %v\n", err)
			cancel()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("  cannot print farewell: %v\n", err)
		}
	}()
	wg.Wait()
}

// Context values with type-safe accessors
func demoContextValues() {
	fmt.Println("\n=== 9. context (values) ===")
	type ctxKey int
	const (
		ctxUserID ctxKey = iota
		ctxAuthToken
	)
	userID := func(ctx context.Context) string { return ctx.Value(ctxUserID).(string) }
	authToken := func(ctx context.Context) string { return ctx.Value(ctxAuthToken).(string) }

	handleResponse := func(ctx context.Context) {
		fmt.Printf("  handling response for %v (auth: %v)\n", userID(ctx), authToken(ctx))
	}
	processRequest := func(userID, authToken string) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxAuthToken, authToken)
		handleResponse(ctx)
	}
	processRequest("jane", "abc123")
}

// ------------------------------------------------------------
// main – run all demos
// ------------------------------------------------------------
func main() {
	fmt.Println("Concurrency Patterns in Go – Examples from Chapter 4")
	demoConfinement()
	demoForSelect()
	demoGoroutineLeak()
	demoOrChannel()
	demoErrorHandling()
	demoPipelines()
	demoQueuing()
	demoContextCancellation()
	demoContextValues()
	fmt.Println("\nAll examples completed.")
}
