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

// Package helper provides streaming compression and decompression helpers that wrap
// compression algorithms with unified io.ReadWriteCloser interfaces.
//
// # Overview
//
// This package simplifies working with compression algorithms by providing a consistent
// interface for both compression and decompression operations on streaming data. It acts
// as an adapter layer between standard io interfaces and compression algorithms from
// github.com/nabbar/golib/archive/compress.
//
// Key capabilities:
//   - Transparent compression/decompression during read or write operations
//   - Support for any compression algorithm that implements arccmp.Algorithm
//   - Automatic handling of io.Reader/io.Writer to io.ReadCloser/io.WriteCloser conversions
//   - Streaming operations with minimal memory overhead
//   - Thread-safe for single-instance usage
//
// # Design Philosophy
//
// 1. Interface Simplicity: Provide a single Helper interface (io.ReadWriteCloser) for all operations
// 2. Algorithm Agnostic: Work with any compression algorithm without algorithm-specific code
// 3. Streaming First: Process data in chunks without loading entire streams into memory
// 4. Resource Safety: Proper cleanup through Close() and automatic wrapper conversions
// 5. Type Safety: Compile-time guarantees through interface-based design
//
// # Architecture
//
// The package provides four internal implementation types, selected automatically based
// on operation type and I/O direction:
//
//		┌──────────────────────────────────────────────┐
//		│               Helper Interface               │
//		│             (io.ReadWriteCloser)             │
//		└──────────────────────┬───────────────────────┘
//		                       │
//		          ┌────────────┴────────────┐
//		          │                         │
//		   ┌──────▼──────┐           ┌──────▼──────┐
//		   │  Compress   │           │  Decompress │
//		   └──────┬──────┘           └──────┬──────┘
//		     ┌────┴────┐               ┌────┴────┐
//	      ▼         ▼               ▼         ▼
//		 ┌────────┬────────┐       ┌────────┬────────┐
//		 │        │        │       │        │        │
//		 │ Reader │ Writer │       │ Reader │ Writer │
//		 │        │        │       │        │        │
//		 └────────┴────────┘       └────────┴────────┘
//		   compressReader        deCompressReader
//		   compressWriter        deCompressWriter
//
// Component Responsibilities:
//
//   - compressReader: Reads from source, compresses data, provides compressed output
//   - compressWriter: Accepts data, compresses it, writes compressed data to destination
//   - deCompressReader: Reads compressed data from source, provides decompressed output
//   - deCompressWriter: Accepts compressed data, decompresses it, writes to destination
//
// # Data Flow
//
// Compression Read Flow (compressReader):
//  1. Client calls Read(p) on Helper
//  2. Helper reads raw data from source reader
//  3. Data is compressed using algorithm's Writer
//  4. Compressed data is buffered internally
//  5. Compressed chunks are returned to client
//
// Compression Write Flow (compressWriter):
//  1. Client calls Write(p) on Helper
//  2. Data is passed to algorithm's Writer
//  3. Compressed data is written to destination
//  4. Close() finalizes compression stream
//
// Decompression Read Flow (deCompressReader):
//  1. Client calls Read(p) on Helper
//  2. Compressed data is read from source
//  3. Algorithm's Reader decompresses data
//  4. Decompressed data is returned to client
//
// Decompression Write Flow (deCompressWriter):
//  1. Client calls Write(p) on Helper
//  2. Compressed data is buffered
//  3. Background goroutine reads from buffer
//  4. Algorithm's Reader decompresses data
//  5. Decompressed data is written to destination
//
// # Basic Usage
//
// Compress data while reading from a source:
//
//	import (
//	    "io"
//	    "os"
//	    "github.com/nabbar/golib/archive/compress"
//	    "github.com/nabbar/golib/archive/helper"
//	)
//
//	func compressFile(inputPath, outputPath string) error {
//	    input, err := os.Open(inputPath)
//	    if err != nil {
//	        return err
//	    }
//	    defer input.Close()
//
//	    output, err := os.Create(outputPath)
//	    if err != nil {
//	        return err
//	    }
//	    defer output.Close()
//
//	    // Create compression reader
//	    h, err := helper.NewReader(compress.GZIP, helper.Compress, input)
//	    if err != nil {
//	        return err
//	    }
//	    defer h.Close()
//
//	    // Copy compressed data to output
//	    _, err = io.Copy(output, h)
//	    return err
//	}
//
// Compress data while writing to a destination:
//
//	func writeCompressed(data []byte, outputPath string) error {
//	    output, err := os.Create(outputPath)
//	    if err != nil {
//	        return err
//	    }
//	    defer output.Close()
//
//	    // Create compression writer
//	    h, err := helper.NewWriter(compress.GZIP, helper.Compress, output)
//	    if err != nil {
//	        return err
//	    }
//	    defer h.Close()
//
//	    // Write data - it will be compressed automatically
//	    _, err = h.Write(data)
//	    return err
//	}
//
// Decompress data while reading:
//
//	func decompressFile(inputPath string) ([]byte, error) {
//	    input, err := os.Open(inputPath)
//	    if err != nil {
//	        return nil, err
//	    }
//	    defer input.Close()
//
//	    // Create decompression reader
//	    h, err := helper.NewReader(compress.GZIP, helper.Decompress, input)
//	    if err != nil {
//	        return nil, err
//	    }
//	    defer h.Close()
//
//	    // Read decompressed data
//	    return io.ReadAll(h)
//	}
//
// Decompress data while writing:
//
//	func writeDecompressed(compressedData []byte, outputPath string) error {
//	    output, err := os.Create(outputPath)
//	    if err != nil {
//	        return err
//	    }
//	    defer output.Close()
//
//	    // Create decompression writer
//	    h, err := helper.NewWriter(compress.GZIP, helper.Decompress, output)
//	    if err != nil {
//	        return err
//	    }
//	    defer h.Close()
//
//	    // Write compressed data - it will be decompressed automatically
//	    _, err = h.Write(compressedData)
//	    return err
//	}
//
// # Automatic Source Type Detection
//
// The New() function automatically determines the operation type based on source type:
//
//	// Source is io.Reader - creates a reader helper
//	h1, _ := helper.New(compress.GZIP, helper.Compress, input)
//	data, _ := io.ReadAll(h1)
//
//	// Source is io.Writer - creates a writer helper
//	h2, _ := helper.New(compress.GZIP, helper.Compress, output)
//	h2.Write(data)
//
// # Supported Compression Algorithms
//
// The package works with any algorithm implementing arccmp.Algorithm from
// github.com/nabbar/golib/archive/compress. Common algorithms include:
//
//   - compress.GZIP: Standard gzip compression (RFC 1952)
//   - compress.ZLIB: Zlib compression (RFC 1950)
//   - compress.DEFLATE: Raw deflate compression (RFC 1951)
//   - compress.BZIP2: Bzip2 compression
//   - compress.LZ4: LZ4 fast compression
//   - compress.ZSTD: Zstandard compression
//   - compress.SNAPPY: Snappy compression
//
// See the compress package documentation for the complete list and algorithm details.
//
// # Implementation Details
//
// Thread Safety:
//   - Each Helper instance is safe for single-goroutine use
//   - Do not share Helper instances across goroutines
//   - Create separate instances for concurrent operations
//
// Memory Management:
//   - compressReader uses a 512-byte internal buffer (chunkSize)
//   - deCompressWriter spawns a single background goroutine
//   - bufNoEOF implements backpressure through wait loops
//   - Close() must be called to release resources properly
//
// Atomicity:
//   - compressReader uses atomic.Bool for close state
//   - deCompressWriter uses atomic.Bool for close and run states
//   - Prevents race conditions in concurrent read/write scenarios
//
// # Error Handling
//
// The package defines three error types:
//
//   - ErrInvalidSource: Returned when source is neither io.Reader nor io.Writer
//   - ErrClosedResource: Returned when writing to a closed deCompressWriter
//   - ErrInvalidOperation: Returned for unsupported Operation values
//
// Errors from underlying readers/writers and compression algorithms are propagated
// unchanged to the caller.
//
// Read() and Write() methods of incorrect types (e.g., Read() on a writer) return
// ErrInvalidSource with zero bytes to clearly indicate the misuse.
//
// # Performance Considerations
//
// Buffer Sizing:
//   - Default chunkSize is 512 bytes
//   - compressReader allocates buffers based on caller's buffer size
//   - Larger read buffers reduce overhead for large data transfers
//
// Compression Trade-offs:
//   - Compression adds CPU overhead but reduces I/O size
//   - Choose algorithms based on speed vs compression ratio needs
//   - GZIP: Good balance, widely compatible
//   - LZ4/Snappy: Fast compression, lower ratios
//   - ZSTD/BZIP2: Slower, higher compression ratios
//
// Decompression Writer Overhead:
//   - deCompressWriter uses a background goroutine
//   - Adds minimal latency but enables streaming
//   - Wait group ensures proper cleanup on Close()
//
// # Use Cases
//
// 1. Compressed File I/O
//
// Read and write compressed files transparently:
//
//	h, _ := helper.NewWriter(compress.GZIP, helper.Compress, file)
//	defer h.Close()
//	h.Write(data) // Automatically compressed
//
// 2. Network Data Compression
//
// Compress data before sending over network:
//
//	h, _ := helper.NewWriter(compress.LZ4, helper.Compress, conn)
//	defer h.Close()
//	h.Write(payload) // Compressed before transmission
//
// 3. Transparent Decompression
//
// Read compressed data as if uncompressed:
//
//	h, _ := helper.NewReader(compress.GZIP, helper.Decompress, compressedStream)
//	defer h.Close()
//	scanner := bufio.NewScanner(h) // Works with any io.Reader consumer
//
// 4. Format Conversion
//
// Convert between compression formats:
//
//	// Read GZIP, write LZ4
//	src, _ := helper.NewReader(compress.GZIP, helper.Decompress, gzipFile)
//	dst, _ := helper.NewWriter(compress.LZ4, helper.Compress, lz4File)
//	defer src.Close()
//	defer dst.Close()
//	io.Copy(dst, src)
//
// 5. Streaming Processing
//
// Process compressed data in streams without full decompression:
//
//	h, _ := helper.NewReader(compress.ZSTD, helper.Decompress, compressedData)
//	defer h.Close()
//	processInChunks(h) // Read and process incrementally
//
// # Limitations
//
// Read/Write Type Restrictions:
//   - Compression/decompression readers do not support Write()
//   - Compression/decompression writers do not support Read()
//   - Attempting wrong operation returns ErrInvalidSource
//
// Concurrency:
//   - Helper instances are not safe for concurrent Read/Write
//   - Use separate instances per goroutine
//   - Close() should be called from a single goroutine
//
// Algorithm Requirements:
//   - Algorithm must implement arccmp.Algorithm interface
//   - Algorithm must provide working Reader() and Writer() methods
//   - Invalid algorithms cause creation errors
//
// Backpressure:
//   - deCompressWriter uses time.Sleep for backpressure (100µs)
//   - Not suitable for hard real-time systems
//   - May cause latency spikes under memory pressure
//
// # Best Practices
//
// Always Close Resources:
//
//	h, err := helper.NewWriter(algo, helper.Compress, dst)
//	if err != nil {
//	    return err
//	}
//	defer h.Close() // Ensures compression finalization
//
// Check Error Returns:
//
//	n, err := h.Write(data)
//	if err != nil {
//	    return fmt.Errorf("compression failed: %w", err)
//	}
//	if n != len(data) {
//	    return fmt.Errorf("incomplete write: %d of %d bytes", n, len(data))
//	}
//
// Choose Appropriate Algorithms:
//
//	// Fast compression for temporary data
//	helper.NewWriter(compress.LZ4, helper.Compress, dst)
//
//	// Maximum compression for archival
//	helper.NewWriter(compress.ZSTD, helper.Compress, dst)
//
//	// Compatibility with external tools
//	helper.NewWriter(compress.GZIP, helper.Compress, dst)
//
// Handle Large Data Efficiently:
//
//	// Stream large files instead of loading into memory
//	h, _ := helper.NewReader(algo, helper.Compress, largeFile)
//	defer h.Close()
//	io.Copy(destination, h) // Streaming, constant memory
//
// # Related Packages
//
// This package integrates with:
//
//   - github.com/nabbar/golib/archive/compress: Compression algorithm implementations
//   - github.com/nabbar/golib/ioutils/nopwritecloser: io.WriteCloser wrapper utilities
//
// Standard library compatibility:
//
//   - io.Reader, io.Writer, io.Closer: Fully compatible
//   - io.Copy: Works transparently with Helper instances
//   - bufio: Can wrap Helper for buffered operations
//
// # Testing
//
// The package includes comprehensive tests covering:
//   - Constructor functions (New, NewReader, NewWriter)
//   - Compression read/write operations
//   - Decompression read/write operations
//   - Error handling and edge cases
//   - Resource cleanup and close semantics
//   - Concurrent operation safety
//   - Integration with various compression algorithms
//
// See TESTING.md for detailed test documentation and coverage reports.
package helper
