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

// Package compress provides unified compression and decompression utilities for multiple algorithms.
//
// # Overview
//
// The compress package offers a simple, consistent interface for working with various compression
// formats including Gzip, Bzip2, LZ4, and XZ. It provides automatic format detection, encoding/decoding
// support (JSON, text marshaling), and transparent Reader/Writer wrapping for seamless integration
// with Go's standard io interfaces.
//
// # Design Philosophy
//
// 1. **Algorithm Agnostic**: Single interface for multiple compression formats (Gzip, Bzip2, LZ4, XZ)
// 2. **Auto-Detection**: Automatic compression format detection from data headers
// 3. **Standard Compliance**: Implements encoding.TextMarshaler/Unmarshaler and json.Marshaler/Unmarshaler
// 4. **Zero-Copy Wrapping**: Efficient Reader/Writer wrapping without data buffering
// 5. **Type Safety**: Enum-based algorithm selection prevents invalid format strings
//
// # Key Features
//
//   - Unified Algorithm enumeration with 5 supported formats (None, Gzip, Bzip2, LZ4, XZ)
//   - Automatic compression detection via magic number analysis
//   - Reader/Writer factory methods for transparent compression/decompression
//   - JSON and text marshaling support for configuration serialization
//   - File extension and string representation helpers
//   - Header validation for format verification
//
// # Architecture
//
// The package follows a functional architecture with stateless operations:
//
//	┌─────────────────────────────────────────────────────┐
//	│              Algorithm (enum type)                  │
//	├─────────────────────────────────────────────────────┤
//	│                                                     │
//	│  ┌──────────────┐    ┌──────────────────────────┐   │
//	│  │   Format     │    │   Detection & Parsing    │   │
//	│  │              │    │                          │   │
//	│  │ • String()   │    │ • Parse(string)          │   │
//	│  │ • Extension()│    │ • Detect(io.Reader)      │   │
//	│  │ • IsNone()   │    │ • DetectOnly(io.Reader)  │   │
//	│  └──────────────┘    │ • DetectHeader([]byte)   │   │
//	│                      └──────────────────────────┘   │
//	│                                                     │
//	│  ┌──────────────────────────────────────────────┐   │
//	│  │         I/O Wrapping                         │   │
//	│  │                                              │   │
//	│  │ • Reader(io.Reader) → io.ReadCloser          │   │
//	│  │ • Writer(io.WriteCloser) → io.WriteCloser    │   │
//	│  └──────────────────────────────────────────────┘   │
//	│                                                     │
//	│  ┌──────────────────────────────────────────────┐   │
//	│  │         Encoding/Marshaling                  │   │
//	│  │                                              │   │
//	│  │ • MarshalText() / UnmarshalText()            │   │
//	│  │ • MarshalJSON() / UnmarshalJSON()            │   │
//	│  └──────────────────────────────────────────────┘   │
//	│                                                     │
//	└─────────────────────────────────────────────────────┘
//	                       │
//	                       ▼
//	┌─────────────────────────────────────────────────────┐
//	│          Standard Library & External                │
//	│                                                     │
//	│  compress/gzip  compress/bzip2  lz4  xz             │
//	└─────────────────────────────────────────────────────┘
//
// Component Characteristics:
//
//	Algorithm:      O(1) operations, Stateless, Thread-safe
//	Parse/Detect:   O(1) header scan, Minimal allocation
//	Reader/Writer:  O(1) wrapping, Delegated to stdlib/external
//	Marshaling:     O(n) string length, Standard encoding
//
// # Supported Algorithms
//
// The package supports the following compression algorithms:
//
//   - None: No compression (pass-through)
//   - Gzip: GNU zip format (RFC 1952) - Good balance of speed and compression
//   - Bzip2: Burrows-Wheeler algorithm - High compression ratio, slower
//   - LZ4: Extremely fast compression/decompression - Low compression ratio
//   - XZ: LZMA2-based compression - Highest compression ratio, slowest
//
// Magic Numbers (Header Detection):
//
//	Gzip:   0x1F 0x8B
//	Bzip2:  'B' 'Z' 'h' [0-9]
//	LZ4:    0x04 0x22 0x4D 0x18
//	XZ:     0xFD 0x37 0x7A 0x58 0x5A 0x00
//
// File Extensions:
//
//	Gzip:   .gz
//	Bzip2:  .bz2
//	LZ4:    .lz4
//	XZ:     .xz
//	None:   (empty string)
//
// # Basic Usage
//
// Creating and using algorithms:
//
//	// Direct algorithm selection
//	alg := compress.Gzip
//	fmt.Println(alg.String())     // "gzip"
//	fmt.Println(alg.Extension())  // ".gz"
//
//	// Parse from string
//	alg = compress.Parse("bzip2")
//	if alg == compress.None {
//	    log.Fatal("Unknown algorithm")
//	}
//
//	// List all algorithms
//	algorithms := compress.List()        // []Algorithm
//	names := compress.ListString()       // []string
//
// # Compression (Writing)
//
// Compress data while writing:
//
//	file, err := os.Create("output.txt.gz")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Wrap the file writer with gzip compression
//	writer, err := compress.Gzip.Writer(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// Write compressed data
//	writer.Write([]byte("This data will be compressed"))
//
// # Decompression (Reading)
//
// Decompress data while reading:
//
//	file, err := os.Open("input.txt.gz")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Wrap the file reader with gzip decompression
//	reader, err := compress.Gzip.Reader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	// Read decompressed data
//	data, err := io.ReadAll(reader)
//	fmt.Println(string(data))
//
// # Automatic Format Detection
//
// Detect compression format from data:
//
//	file, err := os.Open("unknown.dat")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Detect and decompress automatically
//	alg, reader, err := compress.Detect(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	fmt.Printf("Detected: %s\n", alg.String())
//	data, _ := io.ReadAll(reader)
//
// # Detection Without Wrapping
//
// Detect format without creating a decompression reader:
//
//	file, err := os.Open("data.bz2")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Detect only, returns buffered reader
//	alg, bufferedReader, err := compress.DetectOnly(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Format: %s\n", alg.String())
//	// Use bufferedReader as needed (preserves peeked data)
//
// # Header Validation
//
// Verify if data matches a specific format:
//
//	data := []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00}
//	if compress.Gzip.DetectHeader(data) {
//	    fmt.Println("Valid Gzip header")
//	}
//
//	// Check all formats
//	for _, alg := range compress.List() {
//	    if alg.DetectHeader(data) {
//	        fmt.Printf("Detected: %s\n", alg.String())
//	        break
//	    }
//	}
//
// # JSON Marshaling
//
// Serialize and deserialize algorithms in JSON:
//
//	type Config struct {
//	    Compression compress.Algorithm `json:"compression"`
//	}
//
//	// Marshal
//	cfg := Config{Compression: compress.Gzip}
//	json, _ := json.Marshal(cfg)
//	// {"compression":"gzip"}
//
//	// Unmarshal
//	var cfg2 Config
//	json.Unmarshal([]byte(`{"compression":"lz4"}`), &cfg2)
//	// cfg2.Compression == compress.LZ4
//
//	// None is marshaled as null
//	cfg3 := Config{Compression: compress.None}
//	json, _ := json.Marshal(cfg3)
//	// {"compression":null}
//
// # Text Marshaling
//
// Serialize algorithms as text:
//
//	alg := compress.Bzip2
//	text, _ := alg.MarshalText()
//	// []byte("bzip2")
//
//	var alg2 compress.Algorithm
//	alg2.UnmarshalText([]byte("xz"))
//	// alg2 == compress.XZ
//
// # Use Cases
//
// 1. File Archiving with Auto-Detection
//
// Extract files regardless of compression format:
//
//	func ExtractFile(src, dst string) error {
//	    in, err := os.Open(src)
//	    if err != nil {
//	        return err
//	    }
//	    defer in.Close()
//
//	    alg, reader, err := compress.Detect(in)
//	    if err != nil {
//	        return err
//	    }
//	    defer reader.Close()
//
//	    log.Printf("Detected compression: %s", alg.String())
//
//	    out, err := os.Create(dst)
//	    if err != nil {
//	        return err
//	    }
//	    defer out.Close()
//
//	    _, err = io.Copy(out, reader)
//	    return err
//	}
//
// 2. HTTP Response Compression
//
// Compress HTTP responses based on client capabilities:
//
//	func CompressResponse(w http.ResponseWriter, data []byte, format string) error {
//	    alg := compress.Parse(format)
//	    if alg == compress.None {
//	        w.Write(data)
//	        return nil
//	    }
//
//	    w.Header().Set("Content-Encoding", alg.String())
//
//	    writer, err := alg.Writer(struct {
//	        io.Writer
//	        io.Closer
//	    }{w, io.NopCloser(nil)})
//	    if err != nil {
//	        return err
//	    }
//	    defer writer.Close()
//
//	    _, err = writer.Write(data)
//	    return err
//	}
//
// 3. Log File Rotation with Compression
//
// Compress rotated log files:
//
//	func RotateLog(path string, compression compress.Algorithm) error {
//	    src, err := os.Open(path)
//	    if err != nil {
//	        return err
//	    }
//	    defer src.Close()
//
//	    dstPath := path + compression.Extension()
//	    dst, err := os.Create(dstPath)
//	    if err != nil {
//	        return err
//	    }
//	    defer dst.Close()
//
//	    writer, err := compression.Writer(dst)
//	    if err != nil {
//	        return err
//	    }
//	    defer writer.Close()
//
//	    _, err = io.Copy(writer, src)
//	    return err
//	}
//
// 4. Configuration with Compression Settings
//
// Store compression preferences in config files:
//
//	type AppConfig struct {
//	    DataCompression compress.Algorithm `json:"data_compression"`
//	    LogCompression  compress.Algorithm `json:"log_compression"`
//	}
//
//	// Save config
//	cfg := AppConfig{
//	    DataCompression: compress.LZ4,
//	    LogCompression:  compress.Gzip,
//	}
//	data, _ := json.Marshal(cfg)
//	os.WriteFile("config.json", data, 0644)
//
//	// Load config
//	data, _ = os.ReadFile("config.json")
//	var loaded AppConfig
//	json.Unmarshal(data, &loaded)
//
// # Performance Characteristics
//
// Compression Speed (fastest to slowest):
//  1. None (no compression)
//  2. LZ4 (extremely fast, ~500 MB/s)
//  3. Gzip (fast, ~100 MB/s)
//  4. Bzip2 (medium, ~10 MB/s)
//  5. XZ (slow, ~5 MB/s)
//
// Compression Ratio (best to worst):
//  1. XZ (highest compression, ~50-70% reduction)
//  2. Bzip2 (high compression, ~40-60% reduction)
//  3. Gzip (good compression, ~30-50% reduction)
//  4. LZ4 (low compression, ~20-30% reduction)
//  5. None (no compression)
//
// Memory Usage:
//   - Detection: Minimal (6-byte peek buffer)
//   - Reader wrapping: Depends on algorithm (Gzip: ~256KB, LZ4: ~64KB)
//   - Writer wrapping: Depends on algorithm and buffer settings
//
// Trade-offs:
//   - LZ4: Best for real-time data, logging, network protocols
//   - Gzip: Good balance for web content, moderate-sized files
//   - Bzip2: Best for archival, cold storage (rarely accessed data)
//   - XZ: Best for distribution packages, maximum compression needed
//
// # Implementation Details
//
// Algorithm Type:
//   - Implemented as uint8 enum for efficient comparisons and storage
//   - Iota-based constants ensure sequential values (0-4)
//   - Zero value (None) represents "no compression" for safe defaults
//
// Detection Mechanism:
//   - Uses bufio.Reader.Peek(6) to examine first 6 bytes without consuming data
//   - Returns buffered reader to preserve peeked data for subsequent reads
//   - Header detection uses bytes.Equal for exact magic number matching
//
// Reader/Writer Wrapping:
//   - Leverages standard library (compress/gzip, compress/bzip2)
//   - Uses external packages for LZ4 (pierrec/lz4) and XZ (ulikunitz/xz)
//   - io.NopCloser wrapping for algorithms that don't provide Close() method
//   - Returns io.ReadCloser/io.WriteCloser for consistent interface
//
// Error Handling:
//   - Reader() returns error for algorithm-specific initialization failures
//   - Writer() returns error for writer creation or configuration issues
//   - Detect() propagates errors from Peek() or Reader() operations
//   - UnmarshalJSON() returns errors for invalid JSON structure
//
// # Limitations and Best Practices
//
// Limitations:
//
//   - Detection requires at least 6 bytes of data (XZ header size)
//   - DetectHeader() returns false for truncated or corrupted headers
//   - None algorithm passes data through without validation
//   - No support for compression levels (uses default settings)
//   - No support for custom compression parameters
//   - XZ alternative header (all 0xFF) is detected but may be false positive
//
// Best Practices:
//
// DO:
//   - Always check error returns from Reader() and Writer()
//   - Use defer closer.Close() to ensure resource cleanup
//   - Use Detect() when format is unknown or variable
//   - Use Parse() for user-provided format strings
//   - Check IsNone() before attempting compression/decompression
//   - Use List() for enumerating supported formats in UI/CLI
//
// DON'T:
//   - Don't assume all readers have at least 6 bytes (check DetectOnly errors)
//   - Don't forget to Close() writers (buffered data may not be flushed)
//   - Don't use DetectHeader() with truncated data (returns false, not error)
//   - Don't modify Algorithm values outside package constants
//   - Don't parse untrusted format strings without validating the result
//
// # Error Handling Patterns
//
// Safe detection with fallback:
//
//	alg, reader, err := compress.Detect(input)
//	if err != nil {
//	    // Fallback to uncompressed read
//	    reader = io.NopCloser(input)
//	    alg = compress.None
//	}
//	defer reader.Close()
//
// Validation before use:
//
//	alg := compress.Parse(userInput)
//	if alg == compress.None && userInput != "none" {
//	    return fmt.Errorf("unsupported compression: %s", userInput)
//	}
//
// Proper resource cleanup:
//
//	writer, err := alg.Writer(file)
//	if err != nil {
//	    file.Close()  // Close file if writer creation failed
//	    return err
//	}
//	defer writer.Close()  // Writer.Close() also flushes buffers
//
// # Dependencies
//
// Standard Library:
//   - bytes: Header comparison
//   - bufio: Buffered reading for detection
//   - compress/bzip2: Bzip2 decompression (read-only)
//   - compress/gzip: Gzip compression and decompression
//   - encoding/json: JSON marshaling
//   - io: Reader/Writer interfaces
//   - strings: Text parsing
//
// External Packages:
//   - github.com/dsnet/compress/bzip2: Bzip2 compression (write support)
//   - github.com/pierrec/lz4/v4: LZ4 compression and decompression
//   - github.com/ulikunitz/xz: XZ compression and decompression
//
// # Testing
//
// The package includes comprehensive testing with BDD methodology (Ginkgo v2 + Gomega):
//   - Algorithm operations (String, Extension, IsNone)
//   - Format detection (DetectHeader, Detect, DetectOnly)
//   - Parsing and marshaling (Parse, JSON, Text)
//   - Reader/Writer wrapping for all algorithms
//   - Round-trip compression/decompression tests
//   - Edge cases (empty data, invalid formats, corrupted headers)
//   - Concurrent access tests (thread-safety validation)
//   - Performance benchmarks (compression speed, detection latency)
//
// Target: 80% code coverage with race detector validation.
//
// For detailed test documentation, see TESTING.md in the package directory.
//
// # Related Packages
//
// This package is part of github.com/nabbar/golib/archive and works alongside:
//   - github.com/nabbar/golib/archive/tar: Tar archive format support
//   - github.com/nabbar/golib/archive/zip: Zip archive format support
//
// # Version Compatibility
//
// Minimum Go version: 1.24
//   - Relies on standard library features available since Go 1.24
//
// External package versions:
//   - github.com/dsnet/compress: v0.0.1 or later
//   - github.com/pierrec/lz4/v4: v4.0.0 or later
//   - github.com/ulikunitz/xz: v0.5.10 or later
package compress
