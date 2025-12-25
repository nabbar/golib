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

package helper_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

// ExampleNew demonstrates creating a Helper using automatic type detection.
// This is the simplest way to create a Helper when the source type determines usage.
func ExampleNew() {
	data := []byte("test data")
	src := bytes.NewReader(data)

	h, err := helper.New(arccmp.Gzip, helper.Compress, src)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fmt.Printf("Helper created: %T\n", h)
	// Output:
	// Helper created: *helper.compressReader
}

// ExampleNewReader_compress demonstrates compressing data while reading.
// Data is read from source, compressed, and returned to the caller.
func ExampleNewReader_compress() {
	input := strings.NewReader("Hello, World!")

	h, err := helper.NewReader(arccmp.Gzip, helper.Compress, input)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	compressed, err := io.ReadAll(h)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: 13 bytes, Compressed: %d bytes\n", len(compressed))
	// Output:
	// Original: 13 bytes, Compressed: 37 bytes
}

// ExampleNewReader_decompress demonstrates decompressing data while reading.
// Compressed data is read from source and returned decompressed.
func ExampleNewReader_decompress() {
	original := "Hello"
	var buf bytes.Buffer
	cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	cw.Write([]byte(original))
	cw.Close()

	h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	data, err := io.ReadAll(h)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", data)
	// Output:
	// Hello
}

// ExampleNewWriter_compress demonstrates compressing data while writing.
// Data written to the Helper is compressed and written to the destination.
func ExampleNewWriter_compress() {
	var buf bytes.Buffer

	h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	if err != nil {
		log.Fatal(err)
	}

	n, err := h.Write([]byte("Hello, World!"))
	if err != nil {
		log.Fatal(err)
	}

	if err := h.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote %d bytes, compressed size: %d bytes\n", n, buf.Len())
	// Output:
	// Wrote 13 bytes, compressed size: 37 bytes
}

// ExampleNewWriter_decompress demonstrates decompressing data while writing.
// Compressed data written to the Helper is decompressed and written to destination.
func ExampleNewWriter_decompress() {
	original := "Hello"
	var compressedBuf bytes.Buffer
	temp, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &compressedBuf)
	temp.Write([]byte(original))
	temp.Close()
	compressed := compressedBuf.Bytes()

	var buf bytes.Buffer

	h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
	if err != nil {
		log.Fatal(err)
	}

	n, err := h.Write(compressed)
	if err != nil {
		log.Fatal(err)
	}

	if err := h.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote %d compressed bytes, decompressed: %s\n", n, buf.String())
	// Output:
	// Wrote 29 compressed bytes, decompressed: Hello
}

// Example_compressStream demonstrates compressing a data stream.
// This is useful for processing large amounts of data efficiently.
func Example_compressStream() {
	input := strings.NewReader("Line 1\nLine 2\nLine 3\n")
	var output bytes.Buffer

	h, err := helper.NewReader(arccmp.Gzip, helper.Compress, input)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	n, err := io.Copy(&output, h)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Compressed %d bytes\n", n)
	// Output:
	// Compressed 37 bytes
}

// Example_decompressStream demonstrates decompressing a data stream.
// Shows how to transparently decompress while reading.
func Example_decompressStream() {
	original := "Hello"
	var buf bytes.Buffer
	cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	cw.Write([]byte(original))
	cw.Close()

	h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	data, err := io.ReadAll(h)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decompressed: %s\n", data)
	// Output:
	// Decompressed: Hello
}

// Example_chainedCompression demonstrates chaining compression operations.
// Shows converting between different compression algorithms.
func Example_chainedCompression() {
	original := "Hello, World!"

	var buf1 bytes.Buffer
	h1, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf1)
	h1.Write([]byte(original))
	h1.Close()

	fmt.Printf("Original: %d bytes\n", len(original))
	fmt.Printf("GZIP compressed: %d bytes\n", buf1.Len())
	// Output:
	// Original: 13 bytes
	// GZIP compressed: 37 bytes
}

// Example_errorHandling demonstrates proper error handling with helpers.
func Example_errorHandling() {
	h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
	if err != nil {
		fmt.Printf("Creation error: %v\n", err)
		return
	}
	defer h.Close()

	buf := make([]byte, 10)
	n, err := h.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	fmt.Printf("Read %d bytes successfully\n", n)
	// Output:
	// Read 10 bytes successfully
}

// Example_invalidOperation demonstrates error handling for invalid operations.
func Example_invalidOperation() {
	_, err := helper.NewReader(arccmp.Gzip, 99, strings.NewReader("test"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	// Output:
	// Error: invalid operation
}

// Example_invalidSource demonstrates error handling for invalid sources.
func Example_invalidSource() {
	_, err := helper.New(arccmp.Gzip, helper.Compress, "invalid source")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	// Output:
	// Error: invalid source
}

// Example_writerToReader demonstrates reading compressed data from a writer operation.
func Example_writerToReader() {
	var buf bytes.Buffer

	w, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte("Data to compress"))
	w.Close()

	r, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	data, _ := io.ReadAll(r)
	fmt.Printf("%s\n", data)
	// Output:
	// Data to compress
}

// Example_multipleWrites demonstrates writing data in multiple chunks.
func Example_multipleWrites() {
	var buf bytes.Buffer

	h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	if err != nil {
		log.Fatal(err)
	}

	chunks := []string{"Hello", " ", "World", "!"}
	for _, chunk := range chunks {
		h.Write([]byte(chunk))
	}
	h.Close()

	r, _ := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
	defer r.Close()

	data, _ := io.ReadAll(r)
	fmt.Printf("%s\n", data)
	// Output:
	// Hello World!
}

// Example_largeData demonstrates handling larger data efficiently.
func Example_largeData() {
	data := strings.Repeat("test data ", 100)

	var buf bytes.Buffer
	h, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
	h.Write([]byte(data))
	h.Close()

	ratio := float64(len(data)) / float64(buf.Len())
	fmt.Printf("Compression ratio: %.2fx\n", ratio)
	// Output:
	// Compression ratio: 23.81x
}
