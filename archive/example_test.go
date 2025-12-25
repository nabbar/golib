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

package archive_test

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
)

// ExampleParseCompression demonstrates parsing compression algorithm names.
func ExampleParseCompression() {
	alg := archive.ParseCompression("gzip")
	fmt.Printf("Algorithm: %s\n", alg.String())
	fmt.Printf("Extension: %s\n", alg.Extension())
	// Output:
	// Algorithm: gzip
	// Extension: .gz
}

// ExampleParseCompression_invalid demonstrates handling invalid algorithm names.
func ExampleParseCompression_invalid() {
	alg := archive.ParseCompression("invalid")
	fmt.Printf("Is None: %v\n", alg.IsNone())
	fmt.Printf("Algorithm: %s\n", alg.String())
	// Output:
	// Is None: true
	// Algorithm: none
}

// ExampleParseArchive demonstrates parsing archive algorithm names.
func ExampleParseArchive() {
	alg := archive.ParseArchive("tar")
	fmt.Printf("Algorithm: %s\n", alg.String())
	fmt.Printf("Extension: %s\n", alg.Extension())
	// Output:
	// Algorithm: tar
	// Extension: .tar
}

// ExampleParseArchive_caseInsensitive demonstrates case-insensitive parsing.
func ExampleParseArchive_caseInsensitive() {
	alg1 := archive.ParseArchive("TAR")
	alg2 := archive.ParseArchive("tar")
	alg3 := archive.ParseArchive("Tar")

	fmt.Printf("All equal: %v\n", alg1 == alg2 && alg2 == alg3)
	// Output:
	// All equal: true
}

// ExampleDetectCompression demonstrates detecting compression format from data.
func ExampleDetectCompression() {
	// Create some gzip compressed data
	buf := &wcBuffer{Buffer: &bytes.Buffer{}}
	writer, _ := arccmp.Gzip.Writer(buf)
	writer.Write([]byte("Hello, World!"))
	writer.Close()

	// Detect the compression format
	alg, reader, err := archive.DetectCompression(bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Detected: %s\n", alg.String())

	// Read decompressed data
	data, _ := io.ReadAll(reader)
	fmt.Printf("Data: %s\n", string(data))
	// Output:
	// Detected: gzip
	// Data: Hello, World!
}

// ExampleDetectCompression_uncompressed demonstrates detection of uncompressed data.
func ExampleDetectCompression_uncompressed() {
	data := []byte("plain text data")

	alg, reader, err := archive.DetectCompression(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Detected: %s\n", alg.String())
	fmt.Printf("Is None: %v\n", alg.IsNone())
	// Output:
	// Detected: none
	// Is None: true
}

// ExampleDetectArchive demonstrates detecting archive format from data.
func ExampleDetectArchive() {
	// This example requires an actual archive file
	// For demonstration, we'll show the expected usage pattern

	// Open an archive file
	// file, err := os.Open("example.tar")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer file.Close()

	// Detect the archive format
	// alg, reader, closer, err := archive.DetectArchive(file)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer closer.Close()

	// if alg.IsNone() {
	//     fmt.Println("Not an archive")
	//     return
	// }

	// List files in the archive
	// files, _ := reader.List()
	// for _, f := range files {
	//     fmt.Println(f)
	// }

	fmt.Println("See DetectArchive usage in package documentation")
	// Output:
	// See DetectArchive usage in package documentation
}

// ExampleExtractAll demonstrates extracting an archive with auto-detection.
func ExampleExtractAll() {
	// This example requires an actual archive file
	// For demonstration, we'll show the expected usage pattern

	// Open archive file
	// file, err := os.Open("archive.tar.gz")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer file.Close()

	// Extract to destination directory
	// err = archive.ExtractAll(file, "archive.tar.gz", "/output/dir")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// The function automatically:
	// 1. Detects compression (gz)
	// 2. Decompresses the stream
	// 3. Detects archive format (tar)
	// 4. Extracts all files with metadata

	fmt.Println("See ExtractAll usage in package documentation")
	// Output:
	// See ExtractAll usage in package documentation
}

// Example_compressionWorkflow demonstrates a complete compression workflow.
func Example_compressionWorkflow() {
	// Original data
	original := []byte("This is the data to compress and decompress")

	// Compress
	compressed := &wcBuffer{Buffer: &bytes.Buffer{}}
	compressor, err := arccmp.Gzip.Writer(compressed)
	if err != nil {
		log.Fatal(err)
	}
	compressor.Write(original)
	compressor.Close()

	fmt.Printf("Original size: %d bytes\n", len(original))
	fmt.Printf("Compressed size: %d bytes\n", compressed.Len())

	// Decompress
	decompressor, err := arccmp.Gzip.Reader(io.NopCloser(compressed.Buffer))
	if err != nil {
		log.Fatal(err)
	}
	defer decompressor.Close()

	decompressed, err := io.ReadAll(decompressor)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Match: %v\n", bytes.Equal(original, decompressed))
	// Output:
	// Original size: 43 bytes
	// Compressed size: 61 bytes
	// Match: true
}

// Example_algorithmComparison demonstrates comparing different compression algorithms.
func Example_algorithmComparison() {
	algorithms := []arccmp.Algorithm{
		arccmp.Gzip,
		arccmp.Bzip2,
		arccmp.LZ4,
		arccmp.XZ,
	}

	for _, alg := range algorithms {
		fmt.Printf("%-6s: %s\n", alg.String(), alg.Extension())
	}
	// Output:
	// gzip  : .gz
	// bzip2 : .bz2
	// lz4   : .lz4
	// xz    : .xz
}

// Example_archiveFormats demonstrates listing available archive formats.
func Example_archiveFormats() {
	formats := []arcarc.Algorithm{
		arcarc.Tar,
		arcarc.Zip,
	}

	for _, format := range formats {
		fmt.Printf("%-4s: %s\n", format.String(), format.Extension())
	}
	// Output:
	// tar : .tar
	// zip : .zip
}

// Example_errorHandling demonstrates proper error handling patterns.
func Example_errorHandling() {
	// Try to detect compression on empty data
	_, _, err := archive.DetectCompression(bytes.NewReader([]byte{}))
	if err != nil {
		fmt.Printf("Error detected: %v\n", err != nil)
	}

	// Try to parse invalid algorithm
	alg := archive.ParseCompression("invalid_algorithm")
	if alg.IsNone() {
		fmt.Println("Invalid algorithm returns None")
	}

	// Output:
	// Error detected: true
	// Invalid algorithm returns None
}

// Example_streamingExtraction demonstrates memory-efficient streaming extraction.
func Example_streamingExtraction() {
	// This pattern ensures constant memory usage regardless of archive size
	//
	// file, err := os.Open("large_archive.tar.gz")
	// if err != nil {
	//     return err
	// }
	// defer file.Close()
	//
	// // ExtractAll streams data, using constant memory
	// err = archive.ExtractAll(file, "large_archive.tar.gz", "/output")
	// // Even a 10GB archive uses only ~10MB RAM

	fmt.Println("See streaming extraction pattern in documentation")
	// Output:
	// See streaming extraction pattern in documentation
}

// Example_multipleFormats demonstrates handling multiple compression formats.
func Example_multipleFormats() {
	testData := []byte("test data")

	// Try different compression algorithms
	algorithms := []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2, arccmp.LZ4}

	for _, alg := range algorithms {
		buf := &wcBuffer{Buffer: &bytes.Buffer{}}

		// Compress
		w, _ := alg.Writer(buf)
		w.Write(testData)
		w.Close()

		// Detect
		detected, r, _ := archive.DetectCompression(bytes.NewReader(buf.Bytes()))
		r.Close()

		fmt.Printf("%s detected: %v\n", alg.String(), detected == alg)
	}
	// Output:
	// gzip detected: true
	// bzip2 detected: true
	// lz4 detected: true
}

// Example_fileExtraction demonstrates extracting specific files from an archive.
func Example_fileExtraction() {
	// This pattern shows how to work with archive readers directly
	// for selective extraction rather than extracting everything
	//
	// file, _ := os.Open("archive.tar")
	// defer file.Close()
	//
	// alg, reader, closer, _ := archive.DetectArchive(file)
	// defer closer.Close()
	//
	// if alg == arcarc.Tar {
	//     // Check if specific file exists
	//     if reader.Has("important.txt") {
	//         // Get file info
	//         info, _ := reader.Info("important.txt")
	//         fmt.Printf("Size: %d\n", info.Size())
	//
	//         // Extract just this file
	//         data, _ := reader.Get("important.txt")
	//         defer data.Close()
	//         // Process data...
	//     }
	// }

	fmt.Println("See selective extraction pattern in documentation")
	// Output:
	// See selective extraction pattern in documentation
}

// Example_createArchive demonstrates creating a TAR archive programmatically.
func Example_createArchive() {
	// This pattern shows creating archives from scratch
	//
	// output, _ := os.Create("output.tar")
	// defer output.Close()
	//
	// writer, _ := arcarc.Tar.Writer(output)
	// defer writer.Close()
	//
	// // Add files to archive
	// for _, filePath := range filesToAdd {
	//     info, _ := os.Stat(filePath)
	//     file, _ := os.Open(filePath)
	//     defer file.Close()
	//
	//     writer.Add(info, file, filePath, "")
	// }

	fmt.Println("See archive creation pattern in documentation")
	// Output:
	// See archive creation pattern in documentation
}

// Example_nestedCompression demonstrates handling nested compression (tar.gz).
func Example_nestedCompression() {
	// ExtractAll automatically handles nested compression:
	//
	// file, _ := os.Open("archive.tar.gz")
	// defer file.Close()
	//
	// // Automatically:
	// // 1. Detects GZIP compression
	// // 2. Decompresses to get TAR stream
	// // 3. Detects TAR format
	// // 4. Extracts all files
	// archive.ExtractAll(file, "archive.tar.gz", "/output")

	fmt.Println("ExtractAll handles nested formats automatically")
	// Output:
	// ExtractAll handles nested formats automatically
}
