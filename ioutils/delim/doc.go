/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

/*
Package delim provides a buffered reader for reading delimiter-separated data streams.

# Overview

The delim package wraps an io.ReadCloser with a buffered reader that efficiently
processes data separated by a configurable delimiter character. Unlike bufio.Scanner,
which is limited to newline-delimited text, delim allows any rune character as a
delimiter and provides more control over the reading process.

# Key Features

  - Custom delimiter support (newlines, commas, pipes, tabs, null bytes, Unicode characters)
  - Configurable buffer sizes for performance optimization
  - Standard Go interfaces (io.ReadCloser, io.WriterTo)
  - Zero-copy operations where possible
  - Memory-efficient chunk processing
  - Thread-safe when used correctly (one goroutine per instance)

# Basic Usage

The primary type is BufferDelim, created using the New function:

	file, err := os.Open("data.txt")
	if err != nil {
	    log.Fatal(err)
	}
	defer file.Close()

	// Create a BufferDelim with newline delimiter and default buffer
	bd := delim.New(file, '\n', 0)
	defer bd.Close()

	// Read line by line
	for {
	    line, err := bd.ReadBytes()
	    if err == io.EOF {
	        break
	    }
	    if err != nil {
	        log.Fatal(err)
	    }
	    fmt.Printf("Line: %s", line)  // line includes '\n'
	}

# Reading Methods

BufferDelim provides several methods for reading data:

Read(p []byte) - Reads one delimited chunk into the provided buffer

	buf := make([]byte, 1024)
	n, err := bd.Read(buf)
	data := buf[:n]

ReadBytes() - Returns the next delimited chunk as a byte slice

	chunk, err := bd.ReadBytes()
	// chunk includes the delimiter

UnRead() - Peeks at buffered but unread data

	buffered, err := bd.UnRead()
	// Data remains in buffer for next Read

# Writing Methods

BufferDelim also implements io.WriterTo for efficient copying:

	outputFile, _ := os.Create("output.txt")
	defer outputFile.Close()

	written, err := bd.WriteTo(outputFile)
	fmt.Printf("Copied %d bytes\n", written)

# Common Use Cases

CSV/TSV Processing:

	// Read CSV fields separated by commas
	csvFile, _ := os.Open("data.csv")
	bd := delim.New(csvFile, ',', 0)
	defer bd.Close()

	for {
	    field, err := bd.ReadBytes()
	    if err == io.EOF {
	        break
	    }
	    processField(field)
	}

Log File Processing:

	// Process log entries separated by newlines with larger buffer
	logFile, _ := os.Open("app.log")
	bd := delim.New(logFile, '\n', 64*size.KiB)
	defer bd.Close()

	for {
	    entry, err := bd.ReadBytes()
	    if err == io.EOF {
	        break
	    }
	    analyzeLogEntry(entry)
	}

Null-Terminated Strings:

	// Read C-style null-terminated strings
	dataStream, _ := net.Dial("tcp", "server:port")
	bd := delim.New(dataStream, 0, 0)  // 0 is null byte
	defer bd.Close()

	msg, err := bd.ReadBytes()
	// msg is terminated by '\0'

# Performance Considerations

Buffer Size:

The sizeBufferRead parameter in New() controls the internal buffer size.
Larger buffers can improve performance when reading large data chunks:

	// Default buffer (4096 bytes)
	bd := delim.New(reader, '\n', 0)

	// Large buffer for high-throughput scenarios
	bd := delim.New(reader, '\n', 64*size.KiB)

	// Very large buffer for processing huge records
	bd := delim.New(reader, '\n', size.MiB)

Memory Efficiency:

ReadBytes() returns slices backed by the internal buffer, which may be
reused on subsequent reads. Copy the data if you need to retain it:

	data, _ := bd.ReadBytes()
	retained := make([]byte, len(data))
	copy(retained, data)

# Error Handling

The package defines ErrInstance for invalid operations:

	data, err := bd.ReadBytes()
	if err == delim.ErrInstance {
	    // BufferDelim was closed or is invalid
	    log.Fatal("Invalid BufferDelim instance")
	}
	if err == io.EOF {
	    // End of stream reached
	    return
	}

# DiscardCloser

The package also provides DiscardCloser, a no-op io.ReadWriteCloser
useful for testing and benchmarking:

	dc := delim.DiscardCloser{}
	n, _ := dc.Write([]byte("test"))  // n == 4, data discarded
	n, _ = dc.Read(buf)               // n == 0, immediate EOF

# Related Packages

  - github.com/nabbar/golib/size - Convenient size constants (KiB, MiB, etc.)
  - bufio - Standard library buffered I/O (delim builds upon this)
  - io - Standard library I/O interfaces

# Concurrency

BufferDelim instances are not safe for concurrent access. Each instance
should be used by a single goroutine. If multiple goroutines need to process
the same stream, either use a mutex or create separate readers.

# Comparison with bufio.Scanner

Advantages of delim over bufio.Scanner:

  - Custom delimiters (not limited to newlines or custom split functions)
  - Delimiter character included in returned data
  - More predictable buffer behavior
  - Implements standard io.ReadCloser and io.WriterTo interfaces
  - Better control over buffer sizes

When to use bufio.Scanner instead:

  - You only need newline-delimited text
  - You need to remove delimiters from results
  - You want simpler line-by-line iteration

# Examples

See the test files for comprehensive usage examples:
  - constructor_test.go - Construction patterns
  - read_test.go - Reading operations
  - write_test.go - Writing operations
  - edge_cases_test.go - Unicode, binary data, error handling
  - benchmark_test.go - Performance optimization examples
*/
package delim
