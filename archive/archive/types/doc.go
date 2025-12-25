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

// Package types defines common interfaces for archive reading and writing operations.
//
// # Overview
//
// This package provides a unified interface abstraction for working with various archive
// formats (ZIP, TAR, BZIP2, GZIP, etc.). It defines two main interfaces:
//   - Reader: for reading and extracting files from archives
//   - Writer: for creating archives and adding files to them
//
// These interfaces allow implementing format-specific archive handlers while maintaining
// a consistent API across different archive types. This approach enables:
//   - Format-agnostic archive manipulation
//   - Easy switching between archive formats
//   - Consistent error handling and behavior
//   - Simplified testing through interface mocking
//
// # Architecture
//
// The package follows a simple contract-based architecture:
//
//	┌──────────────────────────────────────────────┐
//	│          types.Reader Interface              │
//	│  ┌────────────────────────────────────────┐  │
//	│  │ Close() error                          │  │
//	│  │ List() ([]string, error)               │  │
//	│  │ Info(path) (fs.FileInfo, error)        │  │
//	│  │ Get(path) (io.ReadCloser, error)       │  │
//	│  │ Has(path) bool                         │  │
//	│  │ Walk(FuncExtract)                      │  │
//	│  └────────────────────────────────────────┘  │
//	└──────────────────────────────────────────────┘
//	           ▲          ▲          ▲
//	           │          │          │
//	    ┌──────┴───┬──────┴───┬──────┴──────┐
//	    │          │          │             │
//	  ZIP.rdr   TAR.rdr   BZIP.rdr        Other
//	 Readers    Readers   Readers        Readers
//
//	┌──────────────────────────────────────────────┐
//	│          types.Writer Interface              │
//	│  ┌────────────────────────────────────────┐  │
//	│  │ Close() error                          │  │
//	│  │ Add(info, reader, path, link) error    │  │
//	│  │ FromPath(src, filter, fn) error        │  │
//	│  └────────────────────────────────────────┘  │
//	└──────────────────────────────────────────────┘
//	           ▲          ▲          ▲
//	           │          │          │
//	    ┌──────┴───┬──────┴───┬──────┴──────┐
//	    │          │          │             │
//	  ZIP.wrt   TAR.wrt   BZIP.wrt        Other
//	 Writers    Writers   Writers        Writers
//
// # Reader Interface
//
// The Reader interface provides methods for reading archives:
//
//   - Close(): Releases resources associated with the reader
//   - List(): Returns all file paths in the archive
//   - Info(path): Gets file metadata for a specific path
//   - Get(path): Opens a file for reading from the archive
//   - Has(path): Checks if a file exists in the archive
//   - Walk(fn): Iterates through all files with a callback
//
// # Writer Interface
//
// The Writer interface provides methods for creating archives:
//
//   - Close(): Finalizes and closes the archive
//   - Add(info, reader, path, link): Adds a single file to the archive
//   - FromPath(src, filter, fn): Recursively adds files from a directory
//
// # Function Types
//
// FuncExtract: Callback function used by Walk() method
//
//		type FuncExtract func(info fs.FileInfo, reader io.ReadCloser,
//		                      path string, link string) bool
//
//	  - info: File metadata (size, permissions, timestamps)
//	  - reader: Stream to read file content (may be nil)
//	  - path: File path within the archive
//	  - link: Symlink target (empty for regular files)
//	  - Return: true to continue walking, false to stop
//
// ReplaceName: Callback function for path transformation in FromPath()
//
//		type ReplaceName func(sourcePath string) string
//
//	  - sourcePath: Original file path
//	  - Return: Transformed path to use in archive
//
// # Usage Patterns
//
// Reading from an Archive:
//
//	// Open an archive (implementation-specific)
//	reader, err := someformat.NewReader(file)
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
//
//	// List all files
//	files, err := reader.List()
//	if err != nil {
//	    return err
//	}
//
//	// Check if file exists
//	if reader.Has("config.json") {
//	    // Get file info
//	    info, err := reader.Info("config.json")
//	    if err != nil {
//	        return err
//	    }
//
//	    // Extract file
//	    rc, err := reader.Get("config.json")
//	    if err != nil {
//	        return err
//	    }
//	    defer rc.Close()
//
//	    // Process file content...
//	}
//
// Writing to an Archive:
//
//	// Create an archive (implementation-specific)
//	writer, err := someformat.NewWriter(file)
//	if err != nil {
//	    return err
//	}
//	defer writer.Close()
//
//	// Add single file
//	info, _ := os.Stat("myfile.txt")
//	file, _ := os.Open("myfile.txt")
//	defer file.Close()
//
//	err = writer.Add(info, file, "", "")
//	if err != nil {
//	    return err
//	}
//
//	// Add directory recursively with filter
//	err = writer.FromPath("/path/to/dir", "*.txt", nil)
//	if err != nil {
//	    return err
//	}
//
// Walking through an Archive:
//
//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
//	    if r != nil {
//	        defer r.Close()
//	    }
//
//	    fmt.Printf("File: %s, Size: %d\n", path, info.Size())
//
//	    // Process file...
//
//	    return true // Continue walking
//	})
//
// # Error Handling
//
// Implementations should follow Go's standard error handling conventions:
//   - Use fs.ErrNotExist for missing files
//   - Use fs.ErrInvalid for invalid operations or parameters
//   - Return descriptive errors for format-specific issues
//   - Ensure Close() is idempotent and safe to call multiple times
//
// # Implementation Guidelines
//
// When implementing these interfaces:
//
// 1. Resource Management:
//   - Always release resources in Close()
//   - Make Close() idempotent
//   - Document whether concurrent access is safe
//
// 2. Reader Implementations:
//   - Return fs.ErrNotExist for missing files in Info/Get
//   - Close() should not invalidate List() results
//   - Walk() should handle errors gracefully
//   - Has() should be fast (use caching if needed)
//
// 3. Writer Implementations:
//   - Add() should handle nil readers (directory entries)
//   - FromPath() should respect the filter pattern
//   - Apply ReplaceName before adding files
//   - Flush buffers in Close()
//
// 4. Thread Safety:
//   - Document whether implementation is thread-safe
//   - Consider adding mutex protection for concurrent access
//   - Walk() callback may receive nil readers on errors
//
// # Performance Considerations
//
// For Reader Implementations:
//   - Cache file listings to optimize List() and Has()
//   - Use lazy loading for large archives
//   - Consider streaming for large file extraction
//
// For Writer Implementations:
//   - Buffer writes to improve throughput
//   - Avoid excessive memory allocation
//   - Handle large files with streaming
//
// # Limitations
//
//   - No support for archive modification (append/delete)
//   - No built-in encryption/decryption
//   - No automatic format detection
//   - No progress reporting (implement externally)
//   - Symlink handling varies by format
//
// # Use Cases
//
// 1. Format-Agnostic Archive Processing:
//
//	func ProcessArchive(r types.Reader) error {
//	    files, err := r.List()
//	    if err != nil {
//	        return err
//	    }
//
//	    for _, file := range files {
//	        // Process each file...
//	    }
//	    return nil
//	}
//
// 2. Archive Format Conversion:
//
//	func ConvertArchive(src types.Reader, dst types.Writer) error {
//	    return src.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
//	        if r != nil {
//	            defer r.Close()
//	            dst.Add(info, r, path, link)
//	        }
//	        return true
//	    })
//	}
//
// 3. Selective Extraction:
//
//	func ExtractMatching(r types.Reader, pattern string) error {
//	    return r.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
//	        if matched, _ := filepath.Match(pattern, path); matched {
//	            // Extract file...
//	        }
//	        if rc != nil {
//	            rc.Close()
//	        }
//	        return true
//	    })
//	}
//
// # Best Practices
//
// 1. Always close readers and writers:
//
//	reader, err := format.NewReader(file)
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
//
// 2. Close extracted files:
//
//	rc, err := reader.Get("file.txt")
//	if err != nil {
//	    return err
//	}
//	defer rc.Close()
//
// 3. Check file existence before extraction:
//
//	if reader.Has("config.json") {
//	    rc, _ := reader.Get("config.json")
//	    defer rc.Close()
//	    // Process...
//	}
//
// 4. Handle Walk() errors:
//
//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
//	    if r == nil {
//	        log.Printf("Error opening %s", path)
//	        return true // Continue despite error
//	    }
//	    defer r.Close()
//	    // Process...
//	    return true
//	})
//
// 5. Use path transformation wisely:
//
//	writer.FromPath(srcDir, "*.txt", func(src string) string {
//	    return "backup/" + filepath.Base(src)
//	})
//
// # Integration
//
// This package is designed to work with:
//   - github.com/nabbar/golib/archive/archive/zip
//   - github.com/nabbar/golib/archive/archive/tar
//   - Other archive format implementations
//
// # Testing
//
// Testing implementations:
//   - Use temporary files for integration tests
//   - Test edge cases (empty archives, large files, special characters)
//   - Verify error handling
//   - Test concurrent access if supported
//   - Validate symlink handling
//
// See TESTING.md for detailed testing guidelines.
package types
