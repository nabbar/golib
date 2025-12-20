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

package delim_test

import (
	"fmt"
	"io"
	"strings"

	"github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"
)

// Example_basic demonstrates the most basic usage of BufferDelim.
// This example reads lines from a string using newline as delimiter.
func Example_basic() {
	// Create a simple reader with newline-delimited data
	data := "first line\nsecond line\nthird line\n"
	r := io.NopCloser(strings.NewReader(data))

	// Create BufferDelim with newline delimiter and default buffer
	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Read first line
	line, err := bd.ReadBytes()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Line: %s", string(line))
	// Output: Line: first line
}

// Example_readAllLines demonstrates reading all lines from a stream.
func Example_readAllLines() {
	// Prepare data with multiple lines
	data := "line 1\nline 2\nline 3\n"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Read all lines until EOF
	count := 0
	for {
		line, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		count++
		fmt.Printf("Line %d: %s", count, string(line))
	}
	// Output:
	// Line 1: line 1
	// Line 2: line 2
	// Line 3: line 3
}

// Example_csvProcessing demonstrates processing CSV data with comma delimiter.
func Example_csvProcessing() {
	// CSV-like data with comma delimiter
	data := "Alice,Bob,Charlie,David"
	r := io.NopCloser(strings.NewReader(data))

	// Use comma as delimiter
	bd := delim.New(r, ',', 0, false)
	defer bd.Close()

	// Process each field
	fields := []string{}
	for {
		field, err := bd.ReadBytes()
		if err == io.EOF {
			// Handle last field without delimiter
			if len(field) > 0 {
				fields = append(fields, string(field))
			}
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		// Remove delimiter from field
		fieldStr := strings.TrimSuffix(string(field), ",")
		fields = append(fields, fieldStr)
	}

	fmt.Printf("Fields: %v\n", fields)
	// Output: Fields: [Alice Bob Charlie David]
}

// Example_customDelimiter demonstrates using a custom delimiter character.
func Example_customDelimiter() {
	// Data separated by pipe character
	data := "section1|section2|section3|"
	r := io.NopCloser(strings.NewReader(data))

	// Use pipe as delimiter
	bd := delim.New(r, '|', 0, false)
	defer bd.Close()

	// Read sections
	for i := 1; ; i++ {
		section, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Printf("Section %d: %s\n", i, strings.TrimSuffix(string(section), "|"))
	}
	// Output:
	// Section 1: section1
	// Section 2: section2
	// Section 3: section3
}

// Example_readMethod demonstrates using the Read method for more control.
func Example_readMethod() {
	data := "hello\nworld\n"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Using Read method with buffer
	buf := make([]byte, 100)
	n, err := bd.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Read %d bytes: %s", n, string(buf[:n]))
	// Output: Read 12 bytes: hello
	//world

}

// Example_writeTo demonstrates efficient copying using WriteTo.
func Example_writeTo() {
	data := "line1\nline2\nline3\n"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Copy all data to a strings.Builder
	var output strings.Builder
	written, err := bd.WriteTo(&output)
	if err != nil && err != io.EOF {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes:\n%s", written, output.String())
	// Output:
	// Wrote 18 bytes:
	// line1
	// line2
	// line3
}

// Example_customBufferSize demonstrates using a custom buffer size for performance.
func Example_customBufferSize() {
	// Large data stream
	data := strings.Repeat("This is a line with some content\n", 10)
	r := io.NopCloser(strings.NewReader(data))

	// Use 64KB buffer for better performance with large data
	bd := delim.New(r, '\n', 64*libsiz.SizeKilo, false)
	defer bd.Close()

	// Count lines
	count := 0
	for {
		_, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		count++
	}

	fmt.Printf("Processed %d lines with 64KB buffer\n", count)
	// Output: Processed 10 lines with 64KB buffer
}

// Example_nullTerminatedStrings demonstrates reading null-terminated strings.
func Example_nullTerminatedStrings() {
	// Null-terminated data (like C strings)
	data := "first\x00second\x00third\x00"
	r := io.NopCloser(strings.NewReader(data))

	// Use null byte as delimiter
	bd := delim.New(r, 0, 0, false)
	defer bd.Close()

	// Read null-terminated strings
	for {
		str, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		// Remove null terminator for display
		cleanStr := strings.TrimSuffix(string(str), "\x00")
		fmt.Printf("String: %s\n", cleanStr)
	}
	// Output:
	// String: first
	// String: second
	// String: third
}

// Example_tabDelimitedData demonstrates processing tab-separated values.
func Example_tabDelimitedData() {
	// TSV data
	data := "Name\tAge\tCity\tCountry\t"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\t', 0, false)
	defer bd.Close()

	// Read fields
	for {
		field, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Printf("[%s] ", strings.TrimSuffix(string(field), "\t"))
	}
	fmt.Println()
	// Output: [Name] [Age] [City] [Country]
}

// Example_errorHandling demonstrates proper error handling patterns.
func Example_errorHandling() {
	data := "line1\nline2\nline3"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	for {
		line, err := bd.ReadBytes()

		// Handle EOF - end of data is normal
		if err == io.EOF {
			// Process last line if present
			if len(line) > 0 {
				fmt.Printf("Last line: %s\n", string(line))
			}
			break
		}

		// Handle other errors
		if err != nil {
			fmt.Printf("error reading: %v\n", err)
			return
		}

		fmt.Printf("Line: %s", string(line))
	}
	// Output:
	// Line: line1
	// Line: line2
	// Last line: line3
}

// Example_copyMethod demonstrates using the Copy convenience method.
func Example_copyMethod() {
	data := "data1\ndata2\ndata3\n"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Copy is an alias for WriteTo
	var output strings.Builder
	n, err := bd.Copy(&output)
	if err != nil && err != io.EOF {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Copied %d bytes\n", n)
	// Output: Copied 18 bytes
}

// Example_multipleDelimiters demonstrates handling data with different delimiters.
func Example_multipleDelimiters() {
	// Process colon-separated key-value pairs
	data := "name:John:age:30:city:NYC:"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, ':', 0, false)
	defer bd.Close()

	// Read pairs
	values := []string{}
	for {
		val, err := bd.ReadBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		values = append(values, strings.TrimSuffix(string(val), ":"))
	}

	// Process key-value pairs
	for i := 0; i < len(values); i += 2 {
		if i+1 < len(values) {
			fmt.Printf("%s = %s\n", values[i], values[i+1])
		}
	}
	// Output:
	// name = John
	// age = 30
	// city = NYC
}

// Example_readerInterface demonstrates using BufferDelim as io.ReadCloser.
func Example_readerInterface() {
	data := "test\ndata\n"
	r := io.NopCloser(strings.NewReader(data))

	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Get as io.ReadCloser interface
	reader := bd.Reader()

	// Use as standard reader
	buf := make([]byte, 100)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Read: %s", string(buf[:n]))
	// Output: Read: test
	//data
}

// Example_discardCloser demonstrates using DiscardCloser for testing.
func Example_discardCloser() {
	// Create a DiscardCloser - useful for testing
	dc := delim.DiscardCloser{}

	// Write operation - data is discarded
	n, err := dc.Write([]byte("test data"))
	fmt.Printf("Written: %d bytes, error: %v\n", n, err)

	// Read operation - returns immediately
	buf := make([]byte, 100)
	n, err = dc.Read(buf)
	fmt.Printf("Read: %d bytes, error: %v\n", n, err)

	// Close operation - no-op
	err = dc.Close()
	fmt.Printf("Close error: %v\n", err)

	// Output:
	// Written: 9 bytes, error: <nil>
	// Read: 0 bytes, error: <nil>
	// Close error: <nil>
}

// Example_unRead demonstrates how UnRead retrieves and consumes buffered data.
func Example_unRead() {
	data := "line1\nline2\n"
	r := io.NopCloser(strings.NewReader(data))

	// Use a small buffer to ensure some data is buffered but not all
	// Here 0 means default (32KB), which is plenty for our small data
	bd := delim.New(r, '\n', 0, false)
	defer bd.Close()

	// Read the first line
	// This will read "line1\n" and buffer "line2\n" (and potentially more if available)
	line, err := bd.ReadBytes()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Read: %s", string(line))

	// Check what's left in the buffer
	// This consumes the buffer content ("line2\n")
	buffered, err := bd.UnRead()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Buffered: %s", string(buffered))

	// Output:
	// Read: line1
	// Buffered: line2
}
