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

// Package zip provides a simple interface for reading and writing ZIP archives.
//
// # Overview
//
// This package wraps Go's standard archive/zip library with a unified interface
// (github.com/nabbar/golib/archive/archive/types) that allows for consistent
// archive handling across different formats (ZIP, TAR, etc.).
//
// Key features:
//   - Simple Reader interface for extracting files from ZIP archives
//   - Simple Writer interface for creating ZIP archives
//   - File filtering and path transformation during archive creation
//   - Walk functionality to iterate through archive contents
//   - Consistent error handling using fs.ErrInvalid and fs.ErrNotExist
//
// # Architecture
//
// The package provides two main types:
//
//  1. rdr: Wraps zip.Reader to provide the types.Reader interface
//  2. wrt: Wraps zip.Writer to provide the types.Writer interface
//
// Both types manage the underlying io.ReadCloser/io.WriteCloser and ensure
// proper resource cleanup.
//
//	┌──────────────────┐
//	│  Client Code     │
//	└────────┬─────────┘
//	         │
//	         ├──────────────────┐
//	         │                  │
//	    ┌────▼──────┐    ┌──────▼────┐
//	    │  Reader   │    │  Writer   │
//	    │ (types)   │    │ (types)   │
//	    └────┬──────┘    └──────┬────┘
//	         │                  │
//	    ┌────▼──────┐    ┌──────▼────┐
//	    │    rdr    │    │    wrt    │
//	    │ (zip pkg) │    │ (zip pkg) │
//	    └────┬──────┘    └──────┬────┘
//	         │                  │
//	    ┌────▼──────┐    ┌──────▼────┐
//	    │zip.Reader │    │zip.Writer │
//	    │ (stdlib)  │    │ (stdlib)  │
//	    └───────────┘    └───────────┘
//
// # Data Flow
//
// Reading from ZIP:
//  1. Client calls NewReader with io.ReadCloser (file, buffer, etc.)
//  2. NewReader validates the reader implements required interfaces:
//     - readerSize: provides Size() method for archive size
//     - readerAt: provides ReadAt for random access
//     - io.Seeker: provides Seek for positioning
//  3. Creates zip.Reader using the validated reader
//  4. Returns rdr instance implementing types.Reader
//  5. Client uses List(), Get(), Info(), Has(), or Walk() to access contents
//  6. Client calls Close() to release resources
//
// Writing to ZIP:
//  1. Client calls NewWriter with io.WriteCloser (file, buffer, etc.)
//  2. NewWriter creates zip.Writer wrapping the WriteCloser
//  3. Returns wrt instance implementing types.Writer
//  4. Client uses Add() or FromPath() to add files
//  5. Client calls Close() to finalize archive (flushes and closes)
//
// # Basic Usage
//
// Reading a ZIP archive:
//
//	// Open ZIP file
//	f, err := os.Open("archive.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer f.Close()
//
//	// Create reader
//	reader, err := zip.NewReader(f)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	// List all files
//	files, err := reader.List()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Files in archive:", files)
//
//	// Extract specific file
//	if reader.Has("data.txt") {
//	    rc, err := reader.Get("data.txt")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer rc.Close()
//
//	    data, _ := io.ReadAll(rc)
//	    fmt.Println("Content:", string(data))
//	}
//
// Writing a ZIP archive:
//
//	// Create ZIP file
//	f, err := os.Create("output.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer f.Close()
//
//	// Create writer
//	writer, err := zip.NewWriter(f)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// Add files from directory
//	err = writer.FromPath("/path/to/source", "*.txt", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Advanced Features
//
// Walking through archive contents:
//
//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, name string, link string) bool {
//	    fmt.Printf("File: %s, Size: %d\n", name, info.Size())
//	    if r != nil {
//	        defer r.Close()
//	        // Process file content
//	    }
//	    return true // Continue walking
//	})
//
// Adding individual files with custom names:
//
//	// Open source file
//	srcFile, _ := os.Open("source.txt")
//	defer srcFile.Close()
//
//	srcInfo, _ := srcFile.Stat()
//
//	// Add with custom name in archive
//	err := writer.Add(srcInfo, srcFile, "custom/path/file.txt", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Filtering and path transformation:
//
//	// Only add .go files and transform paths
//	replaceFn := func(source string) string {
//	    return "backup/" + filepath.Base(source)
//	}
//
//	err := writer.FromPath("/project/src", "*.go", replaceFn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Reader Requirements
//
// The NewReader function requires the io.ReadCloser to implement specific interfaces
// for proper ZIP file handling:
//
// Required interfaces:
//
//  1. readerSize: Must provide Size() int64 method
//     - Returns the total size of the ZIP archive
//     - Required by zip.NewReader for random access
//
//  2. readerAt: Must implement io.ReaderAt interface
//     - Provides ReadAt(p []byte, off int64) (n int, err error)
//     - Allows random access to archive contents
//
//  3. io.Seeker: Must implement Seek method
//     - Provides Seek(offset int64, whence int) (int64, error)
//     - Allows repositioning within the archive
//
// Common types that satisfy these requirements:
//   - *os.File: Fully implements all required interfaces
//   - bytes.Reader: Implements all required interfaces
//   - Custom types wrapping file-like resources
//
// If any interface is not implemented, NewReader returns fs.ErrInvalid.
//
// # Error Handling
//
// The package uses standard fs package errors for consistency:
//
//   - fs.ErrInvalid: Returned when reader doesn't implement required interfaces,
//     archive size is invalid (<=0), seek fails, or file is not a regular file
//
//   - fs.ErrNotExist: Returned when requested file is not found in archive
//
// Additional errors from archive/zip standard library may be returned:
//   - zip.ErrFormat: Malformed ZIP archive
//   - zip.ErrAlgorithm: Unsupported compression algorithm
//   - zip.ErrChecksum: Checksum verification failed
//
// # Writer Behavior
//
// The Writer implementation handles files as follows:
//
// Add method:
//   - Accepts nil io.ReadCloser (treated as no-op, returns nil)
//   - Automatically closes the provided io.ReadCloser via defer
//   - Respects forcePath parameter to override file name in archive
//   - Copies file content using io.Copy for efficiency
//
// FromPath method:
//   - Walks directory tree recursively
//   - Applies glob filter pattern (default "*" if empty)
//   - Skips directories (only adds regular files)
//   - Uses os.OpenRoot for secure file access
//   - Applies ReplaceName function for path transformation
//   - Returns fs.ErrInvalid for non-regular files
//
// Close method:
//   - Flushes zip.Writer to ensure all data is written
//   - Closes zip.Writer to finalize archive structure
//   - Closes underlying io.WriteCloser
//   - Returns first error encountered during the sequence
//
// # Implementation Details
//
// Reader (rdr type):
//   - Stores underlying io.ReadCloser for resource management
//   - Stores *zip.Reader for archive operations
//   - List: Pre-allocates slice with archive file count capacity
//   - Info/Get/Has: Performs linear search through zip.Reader.File slice
//   - Walk: Iterates through all files, opens each, calls callback
//   - Close: Delegates to underlying io.ReadCloser
//
// Writer (wrt type):
//   - Stores underlying io.WriteCloser for resource management
//   - Stores *zip.Writer for archive operations
//   - Uses zip.FileInfoHeader to preserve file metadata
//   - Uses zip.Writer.CreateHeader for file creation
//   - Uses io.Copy for efficient data transfer
//   - Implements proper error propagation and resource cleanup
//
// # Performance Considerations
//
// Reading:
//   - List(): O(n) where n is number of files, pre-allocates result slice
//   - Info/Get/Has(): O(n) linear search through file list
//   - Walk(): O(n) iteration, opens each file once
//   - For frequent lookups, consider caching file list in application
//
// Writing:
//   - Add(): Single file write, O(size) where size is file content length
//   - FromPath(): O(n) where n is number of matching files in tree
//   - Filter pattern matching: O(pattern_length) per file
//   - Uses os.OpenRoot for safe file access (may have overhead)
//
// Memory:
//   - Reader keeps entire file list in memory (*zip.Reader requirement)
//   - Writer buffers according to zip.Writer's internal buffering
//   - Large file transfers use io.Copy (efficient streaming)
//
// # Thread Safety
//
// This package is NOT thread-safe by design:
//   - Reader and Writer instances should not be shared between goroutines
//   - Multiple goroutines can create separate Reader/Writer instances
//   - Underlying zip.Reader and zip.Writer are not thread-safe
//   - Application must provide external synchronization if needed
//
// # Limitations
//
// Archive Format:
//   - Only supports ZIP64 format (inherits from archive/zip)
//   - Compression methods depend on archive/zip support
//   - No direct support for encrypted archives
//   - No support for multi-volume ZIP archives
//
// Reader Constraints:
//   - Requires random access (ReaderAt, Seeker interfaces)
//   - Cannot read from streaming/pipe sources
//   - Archive size must be known upfront
//   - Linear search for file lookups (no indexing)
//
// Writer Constraints:
//   - FromPath only adds regular files (skips symlinks, devices)
//   - Filter pattern uses filepath.Match (simple glob, not regex)
//   - No support for adding directories as entries
//   - No support for file permissions beyond fs.FileInfo
//
// Error Handling:
//   - Walk callback errors are not captured or returned
//   - Add silently returns nil for nil io.ReadCloser
//   - Close errors are returned but may lose subsequent close errors
//
// # Use Cases
//
// 1. Configuration Backup:
//
// Create a ZIP backup of configuration files:
//
//	writer, _ := zip.NewWriter(backupFile)
//	defer writer.Close()
//
//	// Backup all .conf files
//	writer.FromPath("/etc/myapp", "*.conf", func(path string) string {
//	    return "configs/" + filepath.Base(path)
//	})
//
// 2. Application Bundle:
//
// Read application resources from embedded ZIP:
//
//	reader, _ := zip.NewReader(embeddedZip)
//	defer reader.Close()
//
//	// Extract template
//	if reader.Has("template.html") {
//	    rc, _ := reader.Get("template.html")
//	    defer rc.Close()
//	    template, _ := io.ReadAll(rc)
//	    // Use template...
//	}
//
// 3. Log Archival:
//
// Archive log files by date:
//
//	replaceFn := func(path string) string {
//	    base := filepath.Base(path)
//	    return fmt.Sprintf("logs/%s/%s", date, base)
//	}
//
//	writer.FromPath("/var/log/app", "*.log", replaceFn)
//
// 4. Data Export:
//
// Export database query results to ZIP:
//
//	writer, _ := zip.NewWriter(exportFile)
//	defer writer.Close()
//
//	for _, record := range records {
//	    content := marshalRecord(record)
//	    buf := io.NopCloser(bytes.NewReader(content))
//	    info := createFileInfo(record)
//	    writer.Add(info, buf, record.Filename, "")
//	}
//
// # Best Practices
//
// DO:
//   - Always defer Close() on both Reader and Writer
//   - Check errors from NewReader and NewWriter
//   - Use FromPath for directory archival (simpler than manual Add)
//   - Close io.ReadCloser returned by Get() after use
//   - Pre-allocate when building file lists
//
// DON'T:
//   - Don't share Reader/Writer instances between goroutines
//   - Don't ignore errors from Close() operations
//   - Don't assume Walk will stop on callback errors
//   - Don't use with streaming sources (pipes, network streams)
//   - Don't modify archive while reading
//
// # Integration
//
// This package integrates with:
//
// Standard Library:
//   - archive/zip: Core ZIP format support
//   - io/fs: File system abstraction and error types
//   - path/filepath: Path manipulation and pattern matching
//   - os: File operations and metadata
//
// Related golib Packages:
//   - github.com/nabbar/golib/archive/archive/types: Common archive interfaces
//   - github.com/nabbar/golib/archive/archive/tar: TAR archive support
//   - github.com/nabbar/golib/archive/compress: Compression algorithms
//
// External References:
//   - ZIP File Format Specification: PKWARE APPNOTE
//   - Go archive/zip documentation: https://pkg.go.dev/archive/zip
//
// # Testing
//
// The package includes comprehensive testing:
//   - Unit tests for Reader and Writer operations
//   - Integration tests with temporary files
//   - Edge case tests (empty archives, large files, invalid inputs)
//   - Concurrent access tests (when applicable)
//   - Benchmark tests for performance validation
//
// Target code coverage: >80% without striving for 100%.
//
// For detailed test documentation, see TESTING.md (if available).
package zip
