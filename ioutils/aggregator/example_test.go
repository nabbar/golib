/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package aggregator_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nabbar/golib/ioutils/aggregator"
)

// ExampleNew demonstrates basic aggregator creation and usage.
func ExampleNew() {
	ctx := context.Background()

	// Create a simple in-memory writer
	var mu sync.Mutex
	var data [][]byte

	writer := func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		// Make a copy to avoid data races
		buf := make([]byte, len(p))
		copy(buf, p)
		data = append(data, buf)
		return len(p), nil
	}

	// Configure the aggregator
	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: writer,
	}

	// Create and start aggregator
	agg, err := aggregator.New(ctx, cfg, nil)
	if err != nil {
		fmt.Printf("Error creating aggregator: %v\n", err)
		return
	}

	if err := agg.Start(ctx); err != nil {
		fmt.Printf("Error starting aggregator: %v\n", err)
		return
	}
	defer agg.Close()

	// Write some data
	agg.Write([]byte("Hello"))
	agg.Write([]byte("World"))

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Check results
	mu.Lock()
	fmt.Printf("Received %d writes\n", len(data))
	mu.Unlock()

	// Output:
	// Received 2 writes
}

// ExampleNew_fileWriter demonstrates using aggregator to write to a file.
func ExampleNew_fileWriter() {
	ctx := context.Background()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "aggregator-example-*.txt")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure aggregator to write to file
	cfg := aggregator.Config{
		BufWriter: 100,
		FctWriter: tmpFile.Write,
	}

	// Create and start
	agg, err := aggregator.New(ctx, cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := agg.Start(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer agg.Close()

	// Multiple concurrent writes (thread-safe)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := fmt.Sprintf("Line from goroutine %d\n", id)
			agg.Write([]byte(data))
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("File written successfully")
	// Output:
	// File written successfully
}

// ExampleConfig_asyncCallback demonstrates periodic async callbacks.
func ExampleConfig_asyncCallback() {
	ctx := context.Background()

	var count int
	var mu sync.Mutex

	// Configure with async callback
	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			return len(p), nil
		},
		AsyncTimer: 50 * time.Millisecond,
		AsyncMax:   2, // Max 2 concurrent async calls
		AsyncFct: func(ctx context.Context) {
			mu.Lock()
			count++
			mu.Unlock()
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)
	defer agg.Close()

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	fmt.Printf("Async function called %d times\n", count)
	mu.Unlock()

	// Output varies, but should be called multiple times
}

// ExampleConfig_syncCallback demonstrates periodic sync callbacks.
func ExampleConfig_syncCallback() {
	ctx := context.Background()

	var flushCount int

	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			return len(p), nil
		},
		SyncTimer: 100 * time.Millisecond,
		SyncFct: func(ctx context.Context) {
			// Simulating a flush operation
			flushCount++
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)
	defer agg.Close()

	// Write some data and wait
	agg.Write([]byte("data1"))
	agg.Write([]byte("data2"))
	time.Sleep(250 * time.Millisecond)

	fmt.Printf("Flush called %d times\n", flushCount)
	// Output varies based on timing
}

// ExampleAggregator_IsRunning demonstrates checking aggregator status.
func ExampleAggregator_IsRunning() {
	ctx := context.Background()

	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			return len(p), nil
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)

	fmt.Printf("Before Start: %v\n", agg.IsRunning())

	agg.Start(ctx)
	time.Sleep(50 * time.Millisecond) // Give it time to start

	fmt.Printf("After Start: %v\n", agg.IsRunning())

	agg.Close()
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("After Close: %v\n", agg.IsRunning())

	// Output:
	// Before Start: false
	// After Start: true
	// After Close: false
}

// ExampleAggregator_Restart demonstrates restarting the aggregator.
func ExampleAggregator_Restart() {
	ctx := context.Background()

	var writeCount int
	var mu sync.Mutex

	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			mu.Lock()
			writeCount++
			mu.Unlock()
			return len(p), nil
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)

	// Write some data
	agg.Write([]byte("before restart"))
	time.Sleep(50 * time.Millisecond)

	// Restart
	agg.Restart(ctx)
	time.Sleep(50 * time.Millisecond)

	// Write after restart
	agg.Write([]byte("after restart"))
	time.Sleep(50 * time.Millisecond)

	agg.Close()

	mu.Lock()
	fmt.Printf("Total writes: %d\n", writeCount)
	mu.Unlock()

	// Output:
	// Total writes: 2
}

// ExampleAggregator_contextCancellation demonstrates context cancellation.
func ExampleAggregator_contextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			return len(p), nil
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)

	// Write some data
	_, err := agg.Write([]byte("data1"))
	fmt.Printf("Write 1 error: %v\n", err)

	// Cancel context
	cancel()
	time.Sleep(50 * time.Millisecond)

	// Try to write after cancellation
	_, err = agg.Write([]byte("data2"))
	fmt.Printf("Write 2 error: %v\n", err != nil)

	// Output:
	// Write 1 error: <nil>
	// Write 2 error: true
}

// ExampleAggregator_errorHandling demonstrates error handling.
func ExampleAggregator_errorHandling() {
	ctx := context.Background()

	// Writer that fails on specific data
	cfg := aggregator.Config{
		BufWriter: 10,
		FctWriter: func(p []byte) (int, error) {
			if string(p) == "fail" {
				return 0, fmt.Errorf("write failed")
			}
			return len(p), nil
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)
	defer agg.Close()

	// Successful write
	_, err := agg.Write([]byte("success"))
	fmt.Printf("Success write error: %v\n", err)

	// Failed write (error is logged internally but Write returns nil)
	_, err = agg.Write([]byte("fail"))
	fmt.Printf("Fail write error: %v\n", err)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Errors are logged internally")

	// Output:
	// Success write error: <nil>
	// Fail write error: <nil>
	// Errors are logged internally
}

// ExampleAggregator_monitoring demonstrates using NbWaiting and NbProcessing
// to monitor the aggregator's buffer state.
func ExampleAggregator_monitoring() {
	ctx := context.Background()

	var processedCount int
	var mu sync.Mutex

	// Slow writer to create backpressure
	cfg := aggregator.Config{
		BufWriter: 5, // Small buffer to demonstrate monitoring
		FctWriter: func(p []byte) (int, error) {
			time.Sleep(50 * time.Millisecond) // Simulate slow I/O
			mu.Lock()
			processedCount++
			mu.Unlock()
			return len(p), nil
		},
	}

	agg, _ := aggregator.New(ctx, cfg, nil)
	agg.Start(ctx)
	defer agg.Close()

	// Start writing in background
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := fmt.Sprintf("message-%d", id)
			agg.Write([]byte(data))
		}(i)
	}

	// Monitor the aggregator state
	time.Sleep(25 * time.Millisecond)

	waiting := agg.NbWaiting()
	processing := agg.NbProcessing()
	sizeWaiting := agg.SizeWaiting()
	sizeProcessing := agg.SizeProcessing()

	total := waiting + processing
	totalMemory := sizeWaiting + sizeProcessing
	avgSize := int64(0)
	if processing > 0 {
		avgSize = sizeProcessing / processing
	}

	fmt.Printf("Monitoring aggregator:\n")
	fmt.Printf("Count metrics:\n")
	fmt.Printf("  - Waiting writes: %d\n", waiting)
	fmt.Printf("  - Processing in buffer: %d\n", processing)
	fmt.Printf("  - Total in flight: %d\n", total)
	fmt.Printf("Size metrics:\n")
	fmt.Printf("  - Bytes waiting: %d\n", sizeWaiting)
	fmt.Printf("  - Bytes processing: %d\n", sizeProcessing)
	fmt.Printf("  - Total memory: %d bytes\n", totalMemory)
	fmt.Printf("  - Avg message size: %d bytes\n", avgSize)
	fmt.Printf("Status:\n")
	fmt.Printf("  - Buffer under load: %v\n", waiting > 0 || processing > 3)
	fmt.Printf("  - Memory pressure: %v\n", totalMemory > 1024)

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// After processing completes
	fmt.Printf("After processing:\n")
	fmt.Printf("  - Waiting: %d (size: %d bytes)\n", agg.NbWaiting(), agg.SizeWaiting())
	fmt.Printf("  - Processing: %d (size: %d bytes)\n", agg.NbProcessing(), agg.SizeProcessing())

	// Output varies due to timing, but demonstrates monitoring
}

// Example_socketToFile demonstrates aggregating socket data to a file.
// This is a common use case with github.com/nabbar/golib/socket/server.
func Example_socketToFile() {
	ctx := context.Background()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "socket-data-*.tmp")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create aggregator to serialize writes to file
	cfg := aggregator.Config{
		BufWriter: 1000, // Buffer up to 1000 socket reads
		FctWriter: tmpFile.Write,
		SyncTimer: 5 * time.Second,
		SyncFct: func(ctx context.Context) {
			// Periodic file sync
			tmpFile.Sync()
		},
	}

	agg, err := aggregator.New(ctx, cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := agg.Start(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer agg.Close()

	// Simulate multiple socket connections writing concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			// Simulate reading from socket
			data := fmt.Sprintf("Data from connection %d\n", connID)
			agg.Write([]byte(data))
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Socket data aggregated to file")
	// Output:
	// Socket data aggregated to file
}
