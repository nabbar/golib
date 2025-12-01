/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package mapCloser_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nabbar/golib/ioutils/mapCloser"
)

// mockResource is a simple closer for examples that prints when closed.
type mockResource struct {
	name   string
	closed bool
	silent bool // If true, don't print on close
}

func (m *mockResource) Close() error {
	if m.closed {
		return fmt.Errorf("already closed: %s", m.name)
	}
	m.closed = true
	if !m.silent {
		fmt.Printf("Closing resource: %s\n", m.name)
	}
	return nil
}

// Example_basic demonstrates the simplest use case: creating a closer,
// adding resources, and manually closing them.
//
// This is useful when you want explicit control over cleanup timing.
func Example_basic() {
	// Create a closer with a background context
	ctx := context.Background()
	closer := mapCloser.New(ctx)

	// Add a resource
	res := &mockResource{name: "database"}
	closer.Add(res)

	// Check how many resources are registered
	fmt.Printf("Registered resources: %d\n", closer.Len())

	// Manually close all resources
	if err := closer.Close(); err != nil {
		fmt.Printf("Error closing: %v\n", err)
	}

	fmt.Println("Done")

	// Output:
	// Registered resources: 1
	// Closing resource: database
	// Done
}

// Example_withTimeout demonstrates using a closer with a timeout context.
//
// The closer can be used with timeout contexts for controlled resource lifecycle.
func Example_withTimeout() {
	// Create a context that times out after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	closer := mapCloser.New(ctx)
	defer closer.Close()

	// Add resources
	res := &mockResource{name: "connection"}
	closer.Add(res)

	fmt.Println("Closer created with timeout context")
	fmt.Printf("Resources registered: %d\n", closer.Len())

	// Resources will be closed when closer.Close() is called by defer
	fmt.Println("Resources will be cleaned up")

	// Output:
	// Closer created with timeout context
	// Resources registered: 1
	// Resources will be cleaned up
	// Closing resource: connection
}

// Example_withCancellation demonstrates using a closer with a cancellable context.
//
// Useful for graceful shutdown scenarios where you want coordinated cleanup.
func Example_withCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	closer := mapCloser.New(ctx)
	res := &mockResource{name: "server"}
	closer.Add(res)

	fmt.Println("Server running")
	fmt.Printf("Registered: %d resource\n", closer.Len())

	// Simulate work
	time.Sleep(10 * time.Millisecond)

	// Manual cleanup before canceling
	closer.Close()
	cancel()

	// Verify cleanup occurred
	if res.closed {
		fmt.Println("Resource closed")
	}

	// Output:
	// Server running
	// Registered: 1 resource
	// Closing resource: server
	// Resource closed
}

// Example_multipleResources demonstrates managing multiple different resource types.
//
// This shows how the closer works as a central cleanup coordinator.
func Example_multipleResources() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	closer := mapCloser.New(ctx)

	// Add a single resource for deterministic output
	res := &mockResource{name: "database"}
	closer.Add(res)

	fmt.Printf("Managing %d resource\n", closer.Len())

	// Get all resources
	resources := closer.Get()
	fmt.Printf("Retrieved %d active resource\n", len(resources))

	// Clean up
	closer.Close()
	fmt.Println("All resources closed")

	// Output:
	// Managing 1 resource
	// Retrieved 1 active resource
	// Closing resource: database
	// All resources closed
}

// Example_errorHandling demonstrates how errors are aggregated when closers fail.
//
// The closer continues closing all resources even when some fail.
func Example_errorHandling() {
	ctx := context.Background()
	closer := mapCloser.New(ctx)

	// Create a resource that will fail
	res := &mockResource{name: "failing-resource", closed: true} // Already closed

	// Add the failing closer
	closer.Add(res)

	// Close will return error but continue
	err := closer.Close()
	if err != nil {
		fmt.Println("Closer failed (as expected)")
	}
	fmt.Println("Cleanup completed")

	// Output:
	// Closer failed (as expected)
	// Cleanup completed
}

// Example_clone demonstrates creating independent copies of a closer.
//
// Useful for hierarchical resource management where you want to group
// resources but maintain independent lifecycle control.
func Example_clone() {
	ctx := context.Background()
	parentCloser := mapCloser.New(ctx)
	defer parentCloser.Close()

	// Add parent resources (silent to avoid non-deterministic output)
	parentRes := &mockResource{name: "parent", silent: true}
	parentCloser.Add(parentRes)

	// Clone for a subtask
	childCloser := parentCloser.Clone()
	if childCloser != nil {
		childRes := &mockResource{name: "child", silent: true}
		childCloser.Add(childRes)

		fmt.Printf("Parent count: %d\n", parentCloser.Len())
		fmt.Printf("Child count: %d\n", childCloser.Len())

		// Close child independently
		childCloser.Close()
		fmt.Printf("Child closed: %v\n", childRes.closed)
		fmt.Println("Clone cleanup done")
	}

	// Output:
	// Parent count: 1
	// Child count: 2
	// Child closed: true
	// Clone cleanup done
}

// Example_clean demonstrates removing resources without closing them.
//
// Useful when you want to transfer ownership of resources elsewhere.
func Example_clean() {
	ctx := context.Background()
	closer := mapCloser.New(ctx)

	// Add resources
	res := &mockResource{name: "transferable"}
	closer.Add(res)

	fmt.Printf("Before clean: %d resources\n", closer.Len())

	// Remove without closing
	closer.Clean()

	fmt.Printf("After clean: %d resources\n", closer.Len())

	// Resource is still open, close it manually
	res.Close()

	// Output:
	// Before clean: 1 resources
	// After clean: 0 resources
	// Closing resource: transferable
}

// Example_fileHandling demonstrates a real-world use case: managing file handles.
//
// This shows practical integration with actual io.Closer implementations.
func Example_fileHandling() {
	ctx := context.Background()
	closer := mapCloser.New(ctx)
	defer closer.Close()

	// Create temporary files
	file1, err := os.CreateTemp("", "example1-*.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	file2, err := os.CreateTemp("", "example2-*.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Register them with the closer
	closer.Add(file1, file2)

	// Write to files
	file1.WriteString("data1")
	file2.WriteString("data2")

	fmt.Printf("Managing %d file handles\n", closer.Len())

	// Files will be closed automatically when closer.Close() is called by defer

	// Output:
	// Managing 2 file handles
}

// Example_webServer demonstrates a complex use case: HTTP server with multiple
// components that need coordinated shutdown.
//
// This shows how the closer integrates into a larger application architecture.
func Example_webServer() {
	// Create a context that can be cancelled for shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the main closer
	closer := mapCloser.New(ctx)

	// Simulate server component
	database := &mockResource{name: "database"}

	// Register component
	closer.Add(database)

	fmt.Println("Server starting with", closer.Len(), "component")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Graceful shutdown
	fmt.Println("Initiating graceful shutdown...")
	if err := closer.Close(); err != nil {
		fmt.Printf("Shutdown errors: %v\n", err)
	}

	fmt.Println("Server stopped cleanly")

	// Output:
	// Server starting with 1 component
	// Initiating graceful shutdown...
	// Closing resource: database
	// Server stopped cleanly
}

// Example_nilHandling demonstrates safe handling of nil closers.
//
// The closer accepts nil values but filters them out during operations.
func Example_nilHandling() {
	ctx := context.Background()
	closer := mapCloser.New(ctx)

	// Add mix of real and nil closers
	res := &mockResource{name: "resource"}
	closer.Add(
		res,
		nil, // Nil is accepted
		nil, // Another nil
	)

	// Len includes nils
	fmt.Printf("Total registered: %d\n", closer.Len())

	// Get filters out nils
	active := closer.Get()
	fmt.Printf("Active resources: %d\n", len(active))

	// Close only affects real closers
	closer.Close()
	fmt.Println("Done")

	// Output:
	// Total registered: 3
	// Active resources: 1
	// Closing resource: resource
	// Done
}

// Example_concurrentUsage demonstrates thread-safe concurrent operations.
//
// All methods can be called safely from multiple goroutines.
func Example_concurrentUsage() {
	ctx := context.Background()
	closer := mapCloser.New(ctx)

	// Simulate concurrent adds from multiple goroutines
	done := make(chan bool, 3)

	for i := 1; i <= 3; i++ {
		go func(id int) {
			// Silent to avoid non-deterministic output order
			res := &mockResource{name: fmt.Sprintf("resource-%d", id), silent: true}
			closer.Add(res)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	fmt.Printf("Concurrently added: %d resources\n", closer.Len())
	err := closer.Close()
	if err == nil {
		fmt.Println("All closed successfully")
	}

	// Output:
	// Concurrently added: 3 resources
	// All closed successfully
}
