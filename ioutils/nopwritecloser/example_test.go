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
 *
 */

package nopwritecloser_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nabbar/golib/ioutils/nopwritecloser"
)

// ExampleNew demonstrates the basic usage of nopwritecloser.
// This is the simplest use case: wrapping a bytes.Buffer to satisfy
// an io.WriteCloser interface without actually closing the buffer.
func ExampleNew() {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Wrap it to implement io.WriteCloser
	wc := nopwritecloser.New(&buf)

	// Write some data
	wc.Write([]byte("Hello, World!"))

	// Close is safe to call - it's a no-op
	wc.Close()

	// Buffer is still accessible after close
	fmt.Println(buf.String())
	// Output: Hello, World!
}

// ExampleNew_multipleWrites shows how to use nopwritecloser with multiple writes.
// The wrapper delegates all writes to the underlying buffer without any overhead.
func ExampleNew_multipleWrites() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	// Multiple writes work seamlessly
	wc.Write([]byte("First "))
	wc.Write([]byte("Second "))
	wc.Write([]byte("Third"))

	wc.Close()

	fmt.Println(buf.String())
	// Output: First Second Third
}

// ExampleNew_defer demonstrates the common pattern of using defer with Close().
// This shows that Close() can be safely deferred even though it's a no-op.
func ExampleNew_defer() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	// Safe to defer Close() - it always succeeds
	defer wc.Close()

	wc.Write([]byte("Data written with defer pattern"))

	fmt.Println(buf.String())
	// Output: Data written with defer pattern
}

// ExampleNew_multipleClose shows that Close() can be called multiple times safely.
// This is useful when you're not sure if Close() has already been called.
func ExampleNew_multipleClose() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	wc.Write([]byte("test"))

	// All these close calls are safe
	wc.Close()
	wc.Close()
	wc.Close()

	fmt.Println(buf.String())
	// Output: test
}

// ExampleNew_writeAfterClose demonstrates that writes still work after Close().
// This is a key difference from real io.WriteCloser implementations.
func ExampleNew_writeAfterClose() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	wc.Write([]byte("before"))
	wc.Close()
	wc.Write([]byte("after"))

	fmt.Println(buf.String())
	// Output: beforeafter
}

// ExampleNew_interfaceCompatibility shows type compatibility with standard interfaces.
// The wrapper satisfies io.Writer, io.Closer, and io.WriteCloser interfaces.
func ExampleNew_interfaceCompatibility() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	// Can be used as io.Writer
	var w io.Writer = wc
	w.Write([]byte("as Writer "))

	// Can be used as io.Closer
	var c io.Closer = wc
	c.Close()

	// Can be used as io.WriteCloser
	var wc2 io.WriteCloser = wc
	wc2.Write([]byte("as WriteCloser"))

	fmt.Println(buf.String())
	// Output: as Writer as WriteCloser
}

// ExampleNew_stdoutProtection shows how to protect os.Stdout from being closed.
// This is useful when passing stdout to functions that expect io.WriteCloser.
func ExampleNew_stdoutProtection() {
	// Wrap stdout to prevent it from being closed
	wc := nopwritecloser.New(os.Stdout)

	// This function would close a real WriteCloser
	writeAndClose := func(w io.WriteCloser) {
		w.Write([]byte("Protected output\n"))
		w.Close() // Safe - doesn't close stdout
	}

	writeAndClose(wc)
	// stdout is still usable here
	// Output: Protected output
}

// ExampleNew_jsonEncoder demonstrates using nopwritecloser with json.Encoder.
// This shows integration with standard library encoding packages.
func ExampleNew_jsonEncoder() {
	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	// Create JSON encoder writing to our wrapper
	encoder := json.NewEncoder(wc)

	// Encode some data
	data := map[string]string{
		"name":  "John",
		"email": "john@example.com",
	}
	encoder.Encode(data)

	wc.Close()

	// Output is still available after close
	fmt.Printf("JSON: %s", buf.String())
	// Output: JSON: {"email":"john@example.com","name":"John"}
}

// ExampleNew_gzipCompression shows using nopwritecloser with gzip compression.
// This demonstrates chaining writers while protecting the final destination.
func ExampleNew_gzipCompression() {
	var buf bytes.Buffer

	// Wrap buffer to protect it from gzip.Writer.Close()
	wc := nopwritecloser.New(&buf)

	// Create gzip writer on top
	gzWriter := gzip.NewWriter(wc)

	// Write compressed data
	gzWriter.Write([]byte("This will be compressed"))
	gzWriter.Close() // Closes gzip writer

	wc.Close() // Safe - doesn't affect buffer

	// Buffer contains compressed data
	fmt.Printf("Data written: %t\n", buf.Len() > 0)
	// Output: Data written: true
}

// ExampleNew_multiWriter demonstrates using io.MultiWriter with nopwritecloser.
// This shows how to tee output to multiple destinations.
func ExampleNew_multiWriter() {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	// Write to both buffers
	wc1 := nopwritecloser.New(&buf1)
	wc2 := nopwritecloser.New(&buf2)
	multi := io.MultiWriter(wc1, wc2)

	multi.Write([]byte("duplicated data"))

	wc1.Close()
	wc2.Close()

	fmt.Printf("Buffer 1: %s\n", buf1.String())
	fmt.Printf("Buffer 2: %s\n", buf2.String())
	// Output:
	// Buffer 1: duplicated data
	// Buffer 2: duplicated data
}

// ExampleNew_functionParameter shows passing nopwritecloser to a function.
// This is the most common use case: satisfying an io.WriteCloser parameter.
func ExampleNew_functionParameter() {
	// Function that requires io.WriteCloser
	processData := func(wc io.WriteCloser, data string) {
		defer wc.Close()
		fmt.Fprintf(wc, "Processed: %s", data)
	}

	var buf bytes.Buffer
	wc := nopwritecloser.New(&buf)

	processData(wc, "my data")

	// Buffer is still accessible
	fmt.Println(buf.String())
	// Output: Processed: my data
}

// ExampleNew_logSink demonstrates using nopwritecloser as a log sink.
// This shows a real-world pattern for collecting log output.
func ExampleNew_logSink() {
	var logBuf bytes.Buffer
	wc := nopwritecloser.New(&logBuf)

	// Simulate logging function
	logMessage := func(w io.WriteCloser, level, msg string) {
		fmt.Fprintf(w, "[%s] %s\n", level, msg)
	}

	logMessage(wc, "INFO", "Application started")
	logMessage(wc, "DEBUG", "Loading config")
	logMessage(wc, "ERROR", "Connection failed")

	wc.Close()

	// Logs are captured in buffer
	fmt.Print(logBuf.String())
	// Output:
	// [INFO] Application started
	// [DEBUG] Loading config
	// [ERROR] Connection failed
}

// ExampleNew_copy demonstrates using io.Copy with nopwritecloser.
// This shows compatibility with standard library I/O functions.
func ExampleNew_copy() {
	source := bytes.NewBufferString("data to copy")
	var dest bytes.Buffer

	wc := nopwritecloser.New(&dest)

	// Use standard io.Copy
	n, _ := io.Copy(wc, source)

	wc.Close()

	fmt.Printf("Copied %d bytes: %s\n", n, dest.String())
	// Output: Copied 12 bytes: data to copy
}

// ExampleNew_chained demonstrates chaining multiple nopwritecloser instances.
// While not common, this shows that the wrapper can be stacked.
func ExampleNew_chained() {
	var buf bytes.Buffer

	// Chain multiple wrappers (unusual but valid)
	wc1 := nopwritecloser.New(&buf)
	wc2 := nopwritecloser.New(wc1)
	wc3 := nopwritecloser.New(wc2)

	wc3.Write([]byte("nested"))
	wc3.Close()
	wc2.Close()
	wc1.Close()

	fmt.Println(buf.String())
	// Output: nested
}
