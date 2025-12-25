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

package compress_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/nabbar/golib/archive/compress"
)

// Example demonstrates basic algorithm usage
func Example() {
	// Get algorithm by name
	alg := compress.Gzip
	fmt.Printf("Algorithm: %s\n", alg.String())
	fmt.Printf("Extension: %s\n", alg.Extension())
	fmt.Printf("Is None: %v\n", alg.IsNone())
	// Output:
	// Algorithm: gzip
	// Extension: .gz
	// Is None: false
}

// ExampleParse demonstrates parsing algorithm from string
func ExampleParse() {
	alg := compress.Parse("gzip")
	fmt.Printf("Parsed: %s\n", alg.String())

	unknown := compress.Parse("unknown")
	fmt.Printf("Unknown: %s\n", unknown.String())
	// Output:
	// Parsed: gzip
	// Unknown: none
}

// ExampleList demonstrates listing all algorithms
func ExampleList() {
	algorithms := compress.List()
	fmt.Printf("Count: %d\n", len(algorithms))

	names := compress.ListString()
	fmt.Printf("Names: %v\n", names)
	// Output:
	// Count: 5
	// Names: [none bzip2 gzip lz4 xz]
}

// ExampleAlgorithm_String demonstrates string representation
func ExampleAlgorithm_String() {
	fmt.Println(compress.Gzip.String())
	fmt.Println(compress.Bzip2.String())
	fmt.Println(compress.None.String())
	// Output:
	// gzip
	// bzip2
	// none
}

// ExampleAlgorithm_Extension demonstrates file extension
func ExampleAlgorithm_Extension() {
	fmt.Println(compress.Gzip.Extension())
	fmt.Println(compress.LZ4.Extension())
	fmt.Println(compress.None.Extension())
	// Output:
	// .gz
	// .lz4
	//
}

// ExampleAlgorithm_IsNone demonstrates checking for None algorithm
func ExampleAlgorithm_IsNone() {
	fmt.Println(compress.None.IsNone())
	fmt.Println(compress.Gzip.IsNone())
	// Output:
	// true
	// false
}

// ExampleAlgorithm_Writer demonstrates basic compression
func ExampleAlgorithm_Writer() {
	data := []byte("Hello, compression!")
	var buf bytes.Buffer

	// Create gzip writer
	writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
	if err != nil {
		log.Fatal(err)
	}

	// Write data
	_, err = writer.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	// Must close to flush
	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %d bytes\n", len(data))
	fmt.Printf("Compressed: %d bytes\n", buf.Len())
	fmt.Printf("Ratio: %.1f%%\n", float64(buf.Len())/float64(len(data))*100)
	// Output:
	// Original: 19 bytes
	// Compressed: 43 bytes
	// Ratio: 226.3%
}

// ExampleAlgorithm_Reader demonstrates basic decompression
func ExampleAlgorithm_Reader() {
	// Sample gzip data
	compressed := []byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0xff, 0xf2, 0x48, 0xcd, 0xc9, 0xc9, 0x07,
		0x04, 0x00, 0x00, 0xff, 0xff, 0x82, 0x89, 0xd1,
		0xf7, 0x05, 0x00, 0x00, 0x00,
	}

	// Create gzip reader
	reader, err := compress.Gzip.Reader(bytes.NewReader(compressed))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// Read decompressed data
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decompressed: %s\n", string(data))
	// Output:
	// Decompressed: Hello
}

// ExampleDetect demonstrates automatic format detection
func ExampleDetect() {
	// Sample gzip data
	compressed := []byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0xff, 0xf2, 0x48, 0xcd, 0xc9, 0xc9, 0x07,
		0x04, 0x00, 0x00, 0xff, 0xff, 0x82, 0x89, 0xd1,
		0xf7, 0x05, 0x00, 0x00, 0x00,
	}

	// Detect and decompress
	alg, reader, err := compress.Detect(bytes.NewReader(compressed))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Detected: %s\n", alg.String())

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Data: %s\n", string(data))
	// Output:
	// Detected: gzip
	// Data: Hello
}

// ExampleDetectOnly demonstrates format detection without decompression
func ExampleDetectOnly() {
	// Sample gzip header
	data := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}

	alg, reader, err := compress.DetectOnly(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Format: %s\n", alg.String())
	// Output:
	// Format: gzip
}

// ExampleAlgorithm_DetectHeader demonstrates header validation
func ExampleAlgorithm_DetectHeader() {
	gzipHeader := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}
	bzip2Header := []byte{'B', 'Z', 'h', '9', 0x00, 0x00}

	fmt.Printf("Gzip matches gzip header: %v\n", compress.Gzip.DetectHeader(gzipHeader))
	fmt.Printf("Gzip matches bzip2 header: %v\n", compress.Gzip.DetectHeader(bzip2Header))
	fmt.Printf("Bzip2 matches bzip2 header: %v\n", compress.Bzip2.DetectHeader(bzip2Header))
	// Output:
	// Gzip matches gzip header: true
	// Gzip matches bzip2 header: false
	// Bzip2 matches bzip2 header: true
}

// ExampleAlgorithm_MarshalText demonstrates text marshaling
func ExampleAlgorithm_MarshalText() {
	alg := compress.Gzip
	data, err := alg.MarshalText()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Marshaled: %s\n", string(data))

	var decoded compress.Algorithm
	err = decoded.UnmarshalText(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decoded: %s\n", decoded.String())
	// Output:
	// Marshaled: gzip
	// Decoded: gzip
}

// ExampleAlgorithm_MarshalJSON demonstrates JSON marshaling
func ExampleAlgorithm_MarshalJSON() {
	type Config struct {
		Compression compress.Algorithm `json:"compression"`
	}

	cfg := Config{Compression: compress.Gzip}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JSON: %s\n", string(data))

	var decoded Config
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Algorithm: %s\n", decoded.Compression.String())
	// Output:
	// JSON: {"compression":"gzip"}
	// Algorithm: gzip
}

// Example_roundTrip demonstrates full compression/decompression cycle
func Example_roundTrip() {
	original := []byte("This is test data for compression")

	// Compress
	var compressed bytes.Buffer
	writer, err := compress.Gzip.Writer(nopWriteCloser{&compressed})
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write(original)
	if err != nil {
		log.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Decompress
	reader, err := compress.Gzip.Reader(&compressed)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Match: %v\n", bytes.Equal(original, decompressed))
	fmt.Printf("Data: %s\n", string(decompressed))
	// Output:
	// Match: true
	// Data: This is test data for compression
}

// Example_compareAlgorithms demonstrates comparing compression algorithms
func Example_compareAlgorithms() {
	data := bytes.Repeat([]byte("test data "), 1024*1024)

	algorithms := []compress.Algorithm{
		compress.Gzip,
		compress.Bzip2,
		compress.LZ4,
		compress.XZ,
	}

	for _, alg := range algorithms {
		var buf bytes.Buffer
		w, err := alg.Writer(nopWriteCloser{&buf})
		if err != nil {
			continue
		}

		w.Write(data)
		w.Close()

		ratio := float64(buf.Len()) / float64(len(data)) * 100
		fmt.Printf("%s: %.2f%%\n", alg.String(), ratio)
	}
	// Output:
	// gzip: 0.19%
	// bzip2: 0.01%
	// lz4: 0.39%
	// xz: 0.02%
}

// Example_autoDetection demonstrates automatic format detection and decompression
func Example_autoDetection() {
	// Create some compressed data
	original := []byte("Auto-detect this!")
	var buf bytes.Buffer
	w, _ := compress.Bzip2.Writer(nopWriteCloser{&buf})
	w.Write(original)
	w.Close()

	// Auto-detect and decompress
	alg, reader, err := compress.Detect(&buf)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Detected: %s\n", alg.String())

	data, _ := io.ReadAll(reader)
	fmt.Printf("Data: %s\n", string(data))
	// Output:
	// Detected: bzip2
	// Data: Auto-detect this!
}

// Example_jsonConfiguration demonstrates using algorithms in JSON configuration
func Example_jsonConfiguration() {
	type AppConfig struct {
		LogCompression  compress.Algorithm `json:"log_compression"`
		DataCompression compress.Algorithm `json:"data_compression"`
	}

	// Create config
	cfg := AppConfig{
		LogCompression:  compress.Gzip,
		DataCompression: compress.LZ4,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JSON: %s\n", string(jsonData))

	// Unmarshal from JSON
	var loaded AppConfig
	err = json.Unmarshal(jsonData, &loaded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Log: %s\n", loaded.LogCompression.String())
	fmt.Printf("Data: %s\n", loaded.DataCompression.String())
	// Output:
	// JSON: {"log_compression":"gzip","data_compression":"lz4"}
	// Log: gzip
	// Data: lz4
}

// Example_noneAlgorithm demonstrates using None (no compression)
func Example_noneAlgorithm() {
	data := []byte("Not compressed")
	var buf bytes.Buffer

	// None algorithm passes through without compression
	writer, err := compress.None.Writer(nopWriteCloser{&buf})
	if err != nil {
		log.Fatal(err)
	}

	writer.Write(data)
	writer.Close()

	fmt.Printf("Original: %d bytes\n", len(data))
	fmt.Printf("Output: %d bytes\n", buf.Len())
	fmt.Printf("Match: %v\n", bytes.Equal(data, buf.Bytes()))
	// Output:
	// Original: 14 bytes
	// Output: 14 bytes
	// Match: true
}

// Example_multipleWrites demonstrates multiple writes to the same compressor
func Example_multipleWrites() {
	var buf bytes.Buffer
	writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
	if err != nil {
		log.Fatal(err)
	}

	// Multiple writes
	writer.Write([]byte("First "))
	writer.Write([]byte("Second "))
	writer.Write([]byte("Third"))
	writer.Close()

	// Decompress
	reader, err := compress.Gzip.Reader(&buf)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	fmt.Printf("Result: %s\n", string(data))
	// Output:
	// Result: First Second Third
}
