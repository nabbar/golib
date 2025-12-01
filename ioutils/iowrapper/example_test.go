/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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
 *
 */

package iowrapper_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/nabbar/golib/ioutils/iowrapper"
)

// Example_basic demonstrates the simplest usage: wrapping an io.Reader
// and using it as normal without any customization.
func Example_basic() {
	// Create a simple buffer
	buf := bytes.NewBufferString("hello world")

	// Wrap it
	wrapper := iowrapper.New(buf)

	// Use it like any io.Reader
	data := make([]byte, 5)
	n, _ := wrapper.Read(data)

	fmt.Printf("%s\n", data[:n])
	// Output: hello
}

// Example_customRead demonstrates how to intercept and transform data
// during read operations using a custom function.
func Example_customRead() {
	// Original data source
	buf := bytes.NewBufferString("hello")
	wrapper := iowrapper.New(buf)

	// Set custom read function that converts to uppercase
	wrapper.SetRead(func(p []byte) []byte {
		// Read from the underlying buffer
		n, _ := buf.Read(p)
		data := p[:n]

		// Transform to uppercase
		for i := range data {
			if data[i] >= 'a' && data[i] <= 'z' {
				data[i] -= 32
			}
		}

		return data
	})

	// Read through the wrapper
	data := make([]byte, 10)
	n, _ := wrapper.Read(data)

	fmt.Printf("%s\n", data[:n])
	// Output: HELLO
}

// Example_logging demonstrates using a wrapper to log I/O operations
// without modifying the underlying code.
func Example_logging() {
	buf := bytes.NewBufferString("data")
	wrapper := iowrapper.New(buf)

	// Track bytes read
	var cnt atomic.Int64

	wrapper.SetRead(func(p []byte) []byte {
		n, _ := buf.Read(p)
		data := p[:n]
		cnt.Add(int64(len(data)))
		return data
	})

	// Perform reads
	tmp := make([]byte, 2)
	wrapper.Read(tmp)
	wrapper.Read(tmp)

	fmt.Printf("Total bytes read: %d\n", cnt.Load())
	// Output: Total bytes read: 4
}

// Example_validation demonstrates using a wrapper to validate data
// before writing, rejecting invalid input.
func Example_validation() {
	var buf bytes.Buffer
	wrapper := iowrapper.New(&buf)

	// Maximum allowed write size
	const maxSize = 10

	wrapper.SetWrite(func(p []byte) []byte {
		// Validate size
		if len(p) > maxSize {
			// Return nil to signal error
			return nil
		}

		// Write to underlying buffer
		n, _ := buf.Write(p)
		return p[:n]
	})

	// Try to write small data (succeeds)
	n1, err1 := wrapper.Write([]byte("small"))
	fmt.Printf("Write 1: %d bytes, error: %v\n", n1, err1)

	// Try to write large data (fails)
	n2, err2 := wrapper.Write([]byte("this is a very long string"))
	fmt.Printf("Write 2: %d bytes, error: %v\n", n2, err2 != nil)

	// Output:
	// Write 1: 5 bytes, error: <nil>
	// Write 2: 0 bytes, error: true
}

// Example_checksum demonstrates calculating a checksum while reading data
// transparently using a custom read function.
func Example_checksum() {
	// Source data
	buf := bytes.NewBufferString("hello world")
	wrapper := iowrapper.New(buf)

	// Hash calculator
	h := md5.New()

	// Custom read function that updates hash
	wrapper.SetRead(func(p []byte) []byte {
		n, err := buf.Read(p)
		if err != nil || n == 0 {
			// Return nil to signal EOF
			return nil
		}
		data := p[:n]

		// Update hash with data read
		h.Write(data)

		return data
	})

	// Read all data (hash is calculated transparently)
	io.Copy(io.Discard, wrapper)

	// Get final checksum
	fmt.Printf("%x\n", h.Sum(nil))
	// Output: 5eb63bbbe01eeed093cb22bb8f5acdc3
}

// Example_resetToDefault demonstrates how to reset custom functions
// back to default behavior using nil.
func Example_resetToDefault() {
	buf := bytes.NewBufferString("test")
	wrapper := iowrapper.New(buf)

	// Set custom function
	wrapper.SetRead(func(p []byte) []byte {
		return []byte("custom")
	})

	// Read with custom function
	data1 := make([]byte, 10)
	n1, _ := wrapper.Read(data1)
	fmt.Printf("Custom: %s\n", data1[:n1])

	// Reset to default by passing nil
	wrapper.SetRead(nil)

	// Now reads from underlying buffer
	data2 := make([]byte, 10)
	n2, _ := wrapper.Read(data2)
	fmt.Printf("Default: %s\n", data2[:n2])

	// Output:
	// Custom: custom
	// Default: test
}

// Example_multipleOperations demonstrates wrapping an object that supports
// multiple I/O operations (Read, Write, Seek).
func Example_multipleOperations() {
	// bytes.Buffer supports Read and Write
	buf := bytes.NewBuffer([]byte("hello"))
	wrapper := iowrapper.New(buf)

	// Read some data
	data := make([]byte, 5)
	n, _ := wrapper.Read(data)
	fmt.Printf("Read: %s\n", data[:n])

	// Write new data
	wrapper.Write([]byte(" world"))

	// Read again
	n, _ = wrapper.Read(data)
	fmt.Printf("Read again: %s\n", data[:n])

	// Output:
	// Read: hello
	// Read again:  worl
}

// Example_chainedWrappers demonstrates composing multiple wrappers
// for layered transformations.
func Example_chainedWrappers() {
	// Original data
	buf := bytes.NewBufferString("hello")

	// Layer 1: Logging wrapper
	logged := iowrapper.New(buf)
	var logMsg string
	logged.SetRead(func(p []byte) []byte {
		n, _ := buf.Read(p)
		logMsg = fmt.Sprintf("Read %d bytes", n)
		return p[:n]
	})

	// Layer 2: Uppercase transformation wrapper
	transformed := iowrapper.New(logged)
	transformed.SetRead(func(p []byte) []byte {
		n, _ := logged.Read(p)
		data := p[:n]
		for i := range data {
			if data[i] >= 'a' && data[i] <= 'z' {
				data[i] -= 32
			}
		}
		return data
	})

	// Read through the chain: buf → logged → transformed
	data := make([]byte, 10)
	n, _ := transformed.Read(data)

	fmt.Printf("%s\n", logMsg)
	fmt.Printf("%s\n", data[:n])
	// Output:
	// Read 5 bytes
	// HELLO
}

// Example_writeWithTransform demonstrates transforming data during write
// operations, such as adding a prefix to all written data.
func Example_writeWithTransform() {
	var buf bytes.Buffer
	wrapper := iowrapper.New(&buf)

	// Add timestamp prefix to all writes
	prefix := "[LOG] "
	wrapper.SetWrite(func(p []byte) []byte {
		// Create combined data with prefix
		combined := append([]byte(prefix), p...)

		// Write to underlying buffer
		buf.Write(combined)

		// Return original data length (not including prefix)
		return p
	})

	// Write data
	wrapper.Write([]byte("message"))

	// Check what was actually written
	fmt.Printf("%s\n", buf.String())
	// Output: [LOG] message
}

// Example_seekOperation demonstrates using Seek with a custom function
// to track position changes.
func Example_seekOperation() {
	// Use strings.Reader which supports Seek
	reader := strings.NewReader("hello world")
	wrapper := iowrapper.New(reader)

	var position int64

	// Custom seek function that tracks position
	wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
		pos, err := reader.Seek(offset, whence)
		position = pos
		return pos, err
	})

	// Seek to position 6
	wrapper.Seek(6, io.SeekStart)

	// Read from new position
	data := make([]byte, 5)
	wrapper.Read(data)

	fmt.Printf("Position: %d, Data: %s\n", position, data)
	// Output: Position: 6, Data: world
}

// Example_errorHandling demonstrates how custom functions signal errors
// by returning nil, which causes operations to return io.ErrUnexpectedEOF.
func Example_errorHandling() {
	buf := bytes.NewBufferString("test")
	wrapper := iowrapper.New(buf)

	// Custom function that rejects reads by returning nil
	wrapper.SetRead(func(p []byte) []byte {
		// Signal error by returning nil
		return nil
	})

	// Attempt to read
	data := make([]byte, 10)
	n, err := wrapper.Read(data)

	fmt.Printf("Bytes read: %d, Error: %v\n", n, err == io.ErrUnexpectedEOF)
	// Output: Bytes read: 0, Error: true
}

// Example_closeWithCleanup demonstrates using a custom close function
// to perform additional cleanup operations.
func Example_closeWithCleanup() {
	// Simulate a resource that needs cleanup
	type resource struct {
		closed bool
	}

	res := &resource{}
	wrapper := iowrapper.New(res)

	// Custom close function
	var cleanupDone bool
	wrapper.SetClose(func() error {
		res.closed = true
		cleanupDone = true
		// Perform cleanup (logging omitted for test output)
		return nil
	})

	// Close the wrapper
	wrapper.Close()

	fmt.Printf("Cleanup done: %v\n", cleanupDone)
	// Output: Cleanup done: true
}

// Example_concurrentAccess demonstrates that the wrapper is safe for
// concurrent use from multiple goroutines.
func Example_concurrentAccess() {
	// Create wrapper without underlying reader (to avoid data races on bytes.Buffer)
	wrapper := iowrapper.New(nil)

	var cnt atomic.Int64

	// Custom read function with counter (thread-safe with atomic)
	wrapper.SetRead(func(p []byte) []byte {
		cnt.Add(1)
		// Return fixed data without accessing shared state
		copy(p, []byte("data"))
		return p[:4]
	})

	// Simulate concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			data := make([]byte, 10)
			wrapper.Read(data)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	fmt.Printf("Total reads: %d\n", cnt.Load())
	// Output: Total reads: 10
}
