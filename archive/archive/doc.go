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

// Package archive provides unified archive handling for TAR and ZIP formats with automatic detection.
//
// # Overview
//
// The archive package offers a simple, consistent interface for working with TAR and ZIP archive formats.
// It provides automatic format detection, unified Reader/Writer interfaces, and transparent handling of
// the structural differences between sequential (TAR) and random-access (ZIP) archive formats.
//
// Key capabilities:
//   - Automatic format detection from archive headers
//   - Unified Reader interface (types.Reader) for both TAR and ZIP
//   - Unified Writer interface (types.Writer) for both TAR and ZIP
//   - Format-agnostic file extraction via Walk() iterator
//   - JSON and text marshaling support for configuration
//   - File extension and string representation helpers
//
// # Design Philosophy
//
// 1. **Format Abstraction**: Single interface for multiple archive formats (TAR, ZIP)
// 2. **Auto-Detection**: Automatic format identification from magic numbers in headers
// 3. **Standard Compliance**: Implements standard encoding interfaces (TextMarshaler, JSONMarshaler)
// 4. **Zero-Copy Wrapping**: Efficient wrapping of standard library readers/writers
// 5. **Type Safety**: Enum-based algorithm selection prevents invalid format strings
// 6. **Structural Awareness**: Respects and exposes the fundamental differences between formats
//
// # Key Features
//
//   - Algorithm enumeration with 3 values (None, Tar, Zip)
//   - Automatic archive format detection via magic number analysis
//   - Reader factory methods for transparent archive extraction
//   - Writer factory methods for transparent archive creation
//   - JSON and text marshaling support for configuration serialization
//   - File extension and string representation for file naming
//   - Header validation for format verification
//   - Walk-based iteration for format-independent file processing
//
// # Architecture
//
// The package follows a layered architecture with format-specific implementations:
//
//	┌────────────────────────────────────────────────────────────┐
//	│                   archive (main package)                   │
//	├────────────────────────────────────────────────────────────┤
//	│                                                            │
//	│  ┌──────────────────┐      ┌────────────────────────┐      │
//	│  │ Algorithm (enum) │      │  Detection & Parsing   │      │
//	│  │                  │      │                        │      │
//	│  │ • None           │      │ • Parse(string)        │      │
//	│  │ • Tar            │      │ • Detect(io.Reader)    │      │
//	│  │ • Zip            │      │ • DetectHeader([]byte) │      │
//	│  │                  │      │                        │      │
//	│  │ • String()       │      │                        │      │
//	│  │ • Extension()    │      │                        │      │
//	│  │ • IsNone()       │      │                        │      │
//	│  └──────────────────┘      └────────────────────────┘      │
//	│                                                            │
//	│  ┌──────────────────────────────────────────────────┐      │
//	│  │         Reader/Writer Factory                    │      │
//	│  │                                                  │      │
//	│  │ • Algorithm.Reader(io.ReadCloser)                │      │
//	│  │     → types.Reader interface                     │      │
//	│  │                                                  │      │
//	│  │ • Algorithm.Writer(io.WriteCloser)               │      │
//	│  │     → types.Writer interface                     │      │
//	│  └──────────────────────────────────────────────────┘      │
//	│                                                            │
//	│  ┌──────────────────────────────────────────────────┐      │
//	│  │         Encoding/Marshaling                      │      │
//	│  │                                                  │      │
//	│  │ • MarshalText() / UnmarshalText()                │      │
//	│  │ • MarshalJSON() / UnmarshalJSON()                │      │
//	│  └──────────────────────────────────────────────────┘      │
//	│                                                            │
//	└──────────────────────┬─────────────────────────────────────┘
//	                       │
//	         ┌─────────────┴─────────────┐
//	         ▼                           ▼
//	┌─────────────────┐         ┌─────────────────┐
//	│   archive/tar   │         │   archive/zip   │
//	│  (sequential)   │         │ (random access) │
//	└─────────────────┘         └─────────────────┘
//	         │                           │
//	         ▼                           ▼
//	┌─────────────────────────────────────────────┐
//	│     Standard Library (archive/tar, /zip)    │
//	└─────────────────────────────────────────────┘
//
// Component Characteristics:
//
//	Algorithm:         O(1) operations, Stateless, Thread-safe
//	Parse/Detect:      O(1) header scan, Minimal allocation
//	Reader/Writer:     O(1) wrapping, Delegated to stdlib
//	Marshaling:        O(n) string length, Standard encoding
//
// # Supported Archive Formats
//
// The package supports two archive formats with fundamentally different structures:
//
// TAR (Tape ARchive):
//   - Sequential format designed for tape backup systems
//   - Files stored continuously: [header][data][header][data]...
//   - POSIX ustar format (IEEE 1003.1-1988)
//   - Magic number: "ustar\x00" at bytes 257-263
//   - File extension: .tar
//   - Pros: Simple, streaming-friendly, no random access overhead
//   - Cons: Cannot seek to specific files, must read sequentially
//   - Best for: Backups, streaming data, network transfers, piped operations
//
// ZIP:
//   - Random-access format with central directory
//   - Files with central directory: [file1][file2]...[directory]
//   - PKZIP format (APPNOTE.TXT specification)
//   - Magic number: 0x504B0304 at bytes 0-4 (local file header)
//   - File extension: .zip
//   - Pros: Fast random access, can list files without extraction
//   - Cons: Requires seekable storage, higher memory usage
//   - Best for: Distribution packages, random file access, GUI tools
//
// None:
//   - Represents absence of archive format
//   - Used for error handling and default values
//   - Returns empty string for Extension()
//   - Always returns false for DetectHeader()
//
// # Basic Usage
//
// Creating and using algorithms:
//
//	// Direct algorithm selection
//	alg := archive.Tar
//	fmt.Println(alg.String())     // "tar"
//	fmt.Println(alg.Extension())  // ".tar"
//
//	// Parse from string (case-insensitive)
//	alg = archive.Parse("zip")
//	if alg == archive.None {
//	    log.Fatal("Unknown algorithm")
//	}
//
// # Archive Creation (Writing)
//
// Creating a TAR archive:
//
//	outFile, err := os.Create("output.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer outFile.Close()
//
//	writer, err := archive.Tar.Writer(outFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// Add entire directory recursively
//	err = writer.FromPath("/path/to/source", "*", nil)
//
// Creating a ZIP archive with custom naming:
//
//	outFile, err := os.Create("output.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer outFile.Close()
//
//	writer, err := archive.Zip.Writer(outFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// Add files with custom path transformation
//	replaceName := func(sourcePath string) string {
//	    return "archive-root/" + filepath.Base(sourcePath)
//	}
//	err = writer.FromPath("/path/to/files", "*.txt", replaceName)
//
// # Archive Extraction (Reading)
//
// Reading with known format:
//
//	inFile, err := os.Open("input.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer inFile.Close()
//
//	reader, err := archive.Tar.Reader(inFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	// List all files
//	files, err := reader.List()
//	fmt.Println(files)
//
//	// Extract specific file
//	fileReader, err := reader.Get("path/in/archive.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer fileReader.Close()
//	data, _ := io.ReadAll(fileReader)
//
// # Automatic Format Detection
//
// The most powerful feature is automatic detection:
//
//	file, err := os.Open("unknown-format.archive")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Detect format and get appropriate reader
//	alg, reader, stream, err := archive.Detect(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stream.Close()
//
//	if reader == nil {
//	    log.Println("Not a recognized archive format")
//	    return
//	}
//	defer reader.Close()
//
//	fmt.Printf("Detected format: %s\n", alg.String())
//
//	// Walk through all files (works for both TAR and ZIP)
//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
//	    fmt.Printf("File: %s (%d bytes)\n", path, info.Size())
//	    if link != "" {
//	        fmt.Printf("  → Link target: %s\n", link)
//	    }
//	    // Extract file data if needed
//	    // data, _ := io.ReadAll(r)
//	    return true  // continue walking
//	})
//
// # Format-Independent File Processing
//
// Using Walk for unified processing:
//
//	func ExtractArchive(archivePath, destDir string) error {
//	    file, err := os.Open(archivePath)
//	    if err != nil {
//	        return err
//	    }
//	    defer file.Close()
//
//	    _, reader, stream, err := archive.Detect(file)
//	    if err != nil {
//	        return err
//	    }
//	    defer stream.Close()
//	    if reader == nil {
//	        return fmt.Errorf("not an archive")
//	    }
//	    defer reader.Close()
//
//	    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
//	        destPath := filepath.Join(destDir, path)
//
//	        if info.IsDir() {
//	            os.MkdirAll(destPath, info.Mode())
//	            return true
//	        }
//
//	        os.MkdirAll(filepath.Dir(destPath), 0755)
//	        out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, info.Mode())
//	        if err != nil {
//	            return false
//	        }
//	        defer out.Close()
//
//	        io.Copy(out, r)
//	        return true
//	    })
//
//	    return nil
//	}
//
// # Configuration Serialization
//
// JSON marshaling:
//
//	type Config struct {
//	    ArchiveFormat archive.Algorithm `json:"format"`
//	}
//
//	// Marshal
//	cfg := Config{ArchiveFormat: archive.Tar}
//	jsonData, _ := json.Marshal(cfg)
//	// {"format":"tar"}
//
//	// Unmarshal
//	var loaded Config
//	json.Unmarshal([]byte(`{"format":"zip"}`), &loaded)
//	// loaded.ArchiveFormat == archive.Zip
//
//	// None is marshaled as null
//	cfg = Config{ArchiveFormat: archive.None}
//	jsonData, _ = json.Marshal(cfg)
//	// {"format":null}
//
// Text marshaling (for YAML, TOML, etc.):
//
//	alg := archive.Zip
//	text, _ := alg.MarshalText()
//	// []byte("zip")
//
//	var alg2 archive.Algorithm
//	alg2.UnmarshalText([]byte("tar"))
//	// alg2 == archive.Tar
//
// # Use Cases
//
// 1. Backup System with Auto-Detection
//
//	func RestoreBackup(backupFile, targetDir string) error {
//	    file, _ := os.Open(backupFile)
//	    defer file.Close()
//
//	    alg, reader, stream, err := archive.Detect(file)
//	    if err != nil || reader == nil {
//	        return fmt.Errorf("invalid backup archive")
//	    }
//	    defer stream.Close()
//	    defer reader.Close()
//
//	    log.Printf("Restoring %s backup...", alg.String())
//	    // ... extraction logic ...
//	    return nil
//	}
//
// 2. Format Conversion
//
//	func ConvertArchive(src, dst string, dstFormat archive.Algorithm) error {
//	    // Open source
//	    srcFile, _ := os.Open(src)
//	    defer srcFile.Close()
//	    _, srcReader, srcStream, _ := archive.Detect(srcFile)
//	    defer srcStream.Close()
//	    defer srcReader.Close()
//
//	    // Create destination
//	    dstFile, _ := os.Create(dst)
//	    defer dstFile.Close()
//	    dstWriter, _ := dstFormat.Writer(dstFile)
//	    defer dstWriter.Close()
//
//	    // Copy all files
//	    srcReader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
//	        dstWriter.Add(info, r, path, link)
//	        return true
//	    })
//	    return nil
//	}
//
// 3. Archive Validation
//
//	func ValidateArchive(path string) (bool, error) {
//	    file, err := os.Open(path)
//	    if err != nil {
//	        return false, err
//	    }
//	    defer file.Close()
//
//	    alg, reader, stream, err := archive.Detect(file)
//	    defer stream.Close()
//	    if err != nil {
//	        return false, err
//	    }
//	    if reader == nil {
//	        return false, nil  // not an archive
//	    }
//	    defer reader.Close()
//
//	    // Verify all files can be listed
//	    _, err = reader.List()
//	    return err == nil, err
//	}
//
// 4. Selective Extraction
//
//	func ExtractMatching(archivePath, pattern, destDir string) error {
//	    file, _ := os.Open(archivePath)
//	    defer file.Close()
//
//	    _, reader, stream, _ := archive.Detect(file)
//	    defer stream.Close()
//	    defer reader.Close()
//
//	    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
//	        matched, _ := filepath.Match(pattern, filepath.Base(path))
//	        if !matched {
//	            return true  // skip this file
//	        }
//
//	        // Extract matching file
//	        destPath := filepath.Join(destDir, path)
//	        os.MkdirAll(filepath.Dir(destPath), 0755)
//	        out, _ := os.Create(destPath)
//	        defer out.Close()
//	        io.Copy(out, r)
//	        return true
//	    })
//	    return nil
//	}
//
// # Performance Characteristics
//
// Detection Performance:
//   - Header peek: O(1) - reads only 265 bytes
//   - Format matching: O(1) - simple byte comparison
//   - Overhead: ~1-2µs for detection on typical systems
//
// TAR Operations:
//   - List(): O(n) - must scan entire archive sequentially
//   - Get(file): O(n) - must scan until file is found
//   - Has(file): O(n) - must scan until file is found
//   - Walk(): O(n) - single sequential pass through archive
//   - Memory: O(1) - constant memory usage (streaming)
//   - Best for: Sequential processing, streaming, backups
//
// ZIP Operations:
//   - List(): O(1) - reads central directory only
//   - Get(file): O(1) - direct seek to file via directory
//   - Has(file): O(1) - lookup in central directory
//   - Walk(): O(n) - iterates central directory entries
//   - Memory: O(n) - central directory kept in memory
//   - Best for: Random access, selective extraction, GUI tools
//
// # Implementation Details
//
// Algorithm Type:
//   - Implemented as uint8 enum for efficient comparisons
//   - Zero value (None) represents "no archive format"
//   - Iota-based constants ensure sequential values
//
// Detection Mechanism:
//   - Uses bufio.Reader.Peek(265) to examine header without consuming data
//   - TAR detection: checks "ustar\x00" at position 257 (POSIX header)
//   - ZIP detection: checks 0x504B0304 signature at position 0
//   - Returns buffered reader to preserve peeked data
//
// Reader/Writer Wrapping:
//   - Delegates to subpackages (archive/tar, archive/zip)
//   - TAR: wraps standard library archive/tar
//   - ZIP: wraps standard library archive/zip
//   - Both implement unified types.Reader and types.Writer interfaces
//
// Internal Adapter (rdr struct):
//   - Provides io.ReaderAt and io.Seeker capabilities for any io.ReadCloser
//   - Uses bufio.Reader for peeking during detection
//   - Emulates seeking for TAR (sequential access only)
//   - Delegates to native methods for ZIP (true random access)
//
// Error Handling:
//   - Reader() returns ErrInvalidAlgorithm for None algorithm
//   - Writer() returns ErrInvalidAlgorithm for None algorithm
//   - Detect() propagates errors from Peek() or reader creation
//   - UnmarshalJSON() returns errors for malformed JSON
//
// # Limitations and Best Practices
//
// Limitations:
//
//   - Detection requires at least 265 bytes of input data
//   - TAR readers cannot seek or perform random access
//   - ZIP readers require io.ReaderAt and io.Seeker capabilities
//   - DetectHeader() returns false (not error) for truncated data
//   - None algorithm is not a valid reader/writer format
//   - No support for compressed archives (use with archive/compress package)
//
// Best Practices:
//
// DO:
//   - Use Detect() when archive format is unknown
//   - Use Parse() for user-provided format strings
//   - Check IsNone() before attempting operations
//   - Close readers/writers in proper order (reader first, then stream)
//   - Use Walk() for format-independent iteration
//   - Use defer for resource cleanup
//   - Check for nil reader after Detect() (format may be unrecognized)
//
// DON'T:
//   - Don't assume TAR readers support random access (they don't)
//   - Don't use ZIP readers with pipes or non-seekable streams
//   - Don't forget to close both the reader AND the underlying stream
//   - Don't modify Algorithm values outside package constants
//   - Don't use Get() in a loop for TAR (use Walk() instead)
//   - Don't call Writer.Close() without closing the writer first
//
// # Error Handling Patterns
//
// Safe detection with fallback:
//
//	alg, reader, stream, err := archive.Detect(file)
//	if err != nil {
//	    log.Printf("Detection failed: %v", err)
//	    return err
//	}
//	defer stream.Close()
//
//	if reader == nil {
//	    log.Println("Not a recognized archive format")
//	    // Handle non-archive file
//	    return nil
//	}
//	defer reader.Close()
//	// Process archive
//
// Validation before use:
//
//	alg := archive.Parse(userInput)
//	if alg == archive.None && userInput != "none" {
//	    return fmt.Errorf("unsupported format: %s", userInput)
//	}
//
// Proper resource cleanup:
//
//	writer, err := alg.Writer(file)
//	if err != nil {
//	    file.Close()  // Close file if writer creation failed
//	    return err
//	}
//	defer func() {
//	    writer.Close()  // Close writer first (flushes buffers)
//	    file.Close()    // Then close underlying file
//	}()
//
// # Sub-Packages
//
// This package consists of several sub-packages:
//
// archive/archive/types:
//   - Defines Reader and Writer interfaces
//   - Provides FuncExtract callback type for Walk()
//   - Provides ReplaceName callback type for path transformation
//
// archive/archive/tar:
//   - TAR format implementation
//   - Sequential access only
//   - Supports hard links and symbolic links
//
// archive/archive/zip:
//   - ZIP format implementation
//   - Random access via central directory
//   - Does not preserve hard links or symbolic links
//
// # Related Packages
//
// This package is part of github.com/nabbar/golib/archive and works alongside:
//   - github.com/nabbar/golib/archive/compress: Compression wrapper (gzip, bzip2, lz4, xz)
//   - github.com/nabbar/golib/archive/helper: Helper utilities for archive operations
//
// Integration example with compression:
//
//	// Read compressed TAR archive
//	file, _ := os.Open("archive.tar.gz")
//	defer file.Close()
//
//	// Decompress
//	decompressor, _ := compress.Gzip.Reader(file)
//	defer decompressor.Close()
//
//	// Detect and extract archive
//	_, reader, stream, _ := archive.Detect(decompressor)
//	defer stream.Close()
//	defer reader.Close()
//
// # Version Compatibility
//
// Minimum Go version: 1.24
//   - Uses standard library features available since Go 1.24
//   - Relies on io/fs package for file system abstractions
//   - Compatible with all Go 1.24+ releases
//
// Standard library dependencies:
//   - archive/tar: TAR format support
//   - archive/zip: ZIP format support
//   - io: Reader/Writer interfaces
//   - io/fs: File system abstraction
//   - encoding/json: JSON marshaling
//   - bufio: Buffered I/O for detection
//
// No external dependencies required.
package archive
