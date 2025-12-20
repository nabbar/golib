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

package multi_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/nabbar/golib/ioutils/multi"
)

// ExampleNew demonstrates the basic creation of a Multi instance.
// This is the simplest usage pattern - creating the multi-writer.
func ExampleNew() {
	m := multi.New(false, false, multi.DefaultConfig())
	fmt.Printf("Multi instance created: %T\n", m)
	// Output:
	// Multi instance created: *multi.mlt
}

// Example_simpleWrite demonstrates writing to a single destination.
// This is the most basic use case for broadcasting writes.
func Example_simpleWrite() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf bytes.Buffer
	m.AddWriter(&buf)

	m.Write([]byte("Hello, World!"))

	fmt.Println(buf.String())
	// Output:
	// Hello, World!
}

// Example_multipleWriters demonstrates broadcasting to multiple destinations.
// Data is automatically replicated to all registered writers.
func Example_multipleWriters() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf1, buf2, buf3 bytes.Buffer
	m.AddWriter(&buf1, &buf2, &buf3)

	m.Write([]byte("broadcast message"))

	fmt.Printf("Buffer 1: %s\n", buf1.String())
	fmt.Printf("Buffer 2: %s\n", buf2.String())
	fmt.Printf("Buffer 3: %s\n", buf3.String())
	// Output:
	// Buffer 1: broadcast message
	// Buffer 2: broadcast message
	// Buffer 3: broadcast message
}

// Example_writeString demonstrates using WriteString for efficient string writes.
// This can be more efficient than Write for string data.
func Example_writeString() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf bytes.Buffer
	m.AddWriter(&buf)

	m.WriteString("String data")

	fmt.Println(buf.String())
	// Output:
	// String data
}

// Example_setInputAndRead demonstrates setting an input source and reading from it.
// The input can be any io.ReadCloser.
func Example_setInputAndRead() {
	m := multi.New(false, false, multi.DefaultConfig())

	input := io.NopCloser(strings.NewReader("input data"))
	m.SetInput(input)

	buf := make([]byte, 10)
	n, _ := m.Read(buf)

	fmt.Printf("%s\n", buf[:n])
	// Output:
	// input data
}

// Example_copy demonstrates copying from input to all outputs.
// This is useful for replicating streams to multiple destinations.
func Example_copy() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf1, buf2 bytes.Buffer
	m.AddWriter(&buf1, &buf2)

	input := io.NopCloser(strings.NewReader("data to copy"))
	m.SetInput(input)

	n, err := m.Copy()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Copied %d bytes\n", n)
	fmt.Printf("Buffer 1: %s\n", buf1.String())
	fmt.Printf("Buffer 2: %s\n", buf2.String())
	// Output:
	// Copied 12 bytes
	// Buffer 1: data to copy
	// Buffer 2: data to copy
}

// Example_clean demonstrates removing all writers.
// Useful for resetting the multi-writer or changing output destinations.
func Example_clean() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf1 bytes.Buffer
	m.AddWriter(&buf1)
	m.Write([]byte("before clean"))

	m.Clean()

	var buf2 bytes.Buffer
	m.AddWriter(&buf2)
	m.Write([]byte("after clean"))

	fmt.Printf("Buffer 1: %s\n", buf1.String())
	fmt.Printf("Buffer 2: %s\n", buf2.String())
	// Output:
	// Buffer 1: before clean
	// Buffer 2: after clean
}

// Example_replaceInput demonstrates changing the input source.
// The input can be replaced at any time.
func Example_replaceInput() {
	m := multi.New(false, false, multi.DefaultConfig())

	input1 := io.NopCloser(strings.NewReader("first"))
	m.SetInput(input1)

	buf := make([]byte, 5)
	n, _ := m.Read(buf)
	fmt.Printf("First read: %s\n", buf[:n])

	input2 := io.NopCloser(strings.NewReader("second"))
	m.SetInput(input2)

	buf2 := make([]byte, 6)
	n, _ = m.Read(buf2)
	fmt.Printf("Second read: %s\n", buf2[:n])
	// Output:
	// First read: first
	// Second read: second
}

// Example_dynamicWriters demonstrates adding writers dynamically.
// Writers can be added at any time, even after data has been written.
func Example_dynamicWriters() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf1 bytes.Buffer
	m.AddWriter(&buf1)
	m.Write([]byte("message 1"))

	var buf2 bytes.Buffer
	m.AddWriter(&buf2)
	m.Write([]byte(" message 2"))

	fmt.Printf("Buffer 1: %s\n", buf1.String())
	fmt.Printf("Buffer 2: %s\n", buf2.String())
	// Output:
	// Buffer 1: message 1 message 2
	// Buffer 2:  message 2
}

// Example_readerWriter demonstrates accessing the underlying reader and writer.
// Direct access to reader/writer allows integration with standard io functions.
func Example_readerWriter() {
	m := multi.New(false, false, multi.DefaultConfig())

	var buf bytes.Buffer
	m.AddWriter(&buf)

	input := io.NopCloser(strings.NewReader("direct access"))
	m.SetInput(input)

	// Use io.Copy with direct access
	n, err := io.Copy(m.Writer(), m.Reader())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Copied %d bytes: %s\n", n, buf.String())
	// Output:
	// Copied 13 bytes: direct access
}

// Example_loggingMultiWriter demonstrates a practical use case:
// logging to both console and file simultaneously.
func Example_loggingMultiWriter() {
	m := multi.New(false, false, multi.DefaultConfig())

	var fileBuffer bytes.Buffer // Simulating a file

	// Add both stdout and "file" as outputs
	m.AddWriter(os.Stdout, &fileBuffer)

	m.Write([]byte("Log entry\n"))

	fmt.Printf("Also captured in file: %d bytes\n", fileBuffer.Len())
	// Output:
	// Log entry
	// Also captured in file: 10 bytes
}

// Example_streamReplication demonstrates replicating an input stream
// to multiple outputs, useful for data pipelines.
func Example_streamReplication() {
	m := multi.New(false, false, multi.DefaultConfig())

	var processed, archived, backup bytes.Buffer
	m.AddWriter(&processed, &archived, &backup)

	input := io.NopCloser(strings.NewReader("stream data"))
	m.SetInput(input)

	n, _ := m.Copy()

	fmt.Printf("Replicated %d bytes to 3 destinations\n", n)
	fmt.Printf("Processed: %s\n", processed.String())
	fmt.Printf("Archived: %s\n", archived.String())
	fmt.Printf("Backup: %s\n", backup.String())
	// Output:
	// Replicated 11 bytes to 3 destinations
	// Processed: stream data
	// Archived: stream data
	// Backup: stream data
}

// Example_conditionalWriters demonstrates dynamically managing writers
// based on conditions, useful for selective logging or output routing.
func Example_conditionalWriters() {
	m := multi.New(false, false, multi.DefaultConfig())

	var normalLog, errorLog bytes.Buffer

	// Normal operation
	m.AddWriter(&normalLog)
	m.Write([]byte("info: operation started\n"))

	// Error occurred - add error log
	m.AddWriter(&errorLog)
	m.Write([]byte("error: something went wrong\n"))

	fmt.Printf("Normal log:\n%s", normalLog.String())
	fmt.Printf("Error log:\n%s", errorLog.String())
	// Output:
	// Normal log:
	// info: operation started
	// error: something went wrong
	// Error log:
	// error: something went wrong
}

// Example_mixedOperations demonstrates a complex scenario combining
// multiple features: reading, writing, copying, and managing destinations.
func Example_mixedOperations() {
	m := multi.New(false, false, multi.DefaultConfig())

	var primary, secondary bytes.Buffer

	// Phase 1: Direct write
	m.AddWriter(&primary)
	m.Write([]byte("Phase 1: "))

	// Phase 2: Copy from input
	input := io.NopCloser(strings.NewReader("Phase 2: "))
	m.SetInput(input)
	m.Copy()

	// Phase 3: Add secondary and continue
	m.AddWriter(&secondary)
	m.Write([]byte("Phase 3"))

	fmt.Printf("Primary: %s\n", primary.String())
	fmt.Printf("Secondary: %s\n", secondary.String())
	// Output:
	// Primary: Phase 1: Phase 2: Phase 3
	// Secondary: Phase 3
}

// Example_closeInput demonstrates proper resource management by closing
// the input source when done.
func Example_closeInput() {
	m := multi.New(false, false, multi.DefaultConfig())

	input := io.NopCloser(strings.NewReader("data"))
	m.SetInput(input)

	buf := make([]byte, 4)
	m.Read(buf)

	err := m.Close() // Closes the input
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Input closed successfully")
	// Output:
	// Input closed successfully
}

// Example_discardCloser demonstrates the use of DiscardCloser,
// a no-op ReadWriteCloser used as default.
func Example_discardCloser() {
	d := multi.DiscardCloser{}

	// Write is discarded
	n, err := d.Write([]byte("discarded"))
	fmt.Printf("Write: %d bytes, error: %v\n", n, err)

	// Read returns nothing
	buf := make([]byte, 10)
	n, err = d.Read(buf)
	fmt.Printf("Read: %d bytes, error: %v\n", n, err)

	// Close is a no-op
	err = d.Close()
	fmt.Printf("Close error: %v\n", err)
	// Output:
	// Write: 9 bytes, error: <nil>
	// Read: 0 bytes, error: <nil>
	// Close error: <nil>
}
