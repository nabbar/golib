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

// Package tar provides a high-level interface for reading and writing tar archives.
//
// # Overview
//
// This package wraps Go's standard archive/tar library with a more convenient
// interface that implements the github.com/nabbar/golib/archive/archive/types
// Reader and Writer interfaces. It provides simplified methods for common archive
// operations while maintaining compatibility with standard Go interfaces.
//
// # Key Features
//
//   - Simple reader/writer constructors (NewReader, NewWriter)
//   - High-level operations (List, Info, Get, Has, Walk)
//   - Recursive directory archiving with filtering
//   - Symbolic and hard link preservation
//   - Path transformation during archiving
//   - Reset support for re-reading archives (when underlying reader supports it)
//
// # Design Philosophy
//
// The package follows these design principles:
//
//  1. Interface Compliance: Implements standard archive types interfaces for
//     consistency across different archive formats (tar, zip, etc.)
//
//  2. Simplicity: Provides high-level methods that handle common use cases
//     without requiring detailed knowledge of tar format internals.
//
//  3. Flexibility: Supports advanced scenarios like path filtering, renaming,
//     and link preservation while keeping simple cases simple.
//
//  4. Safety: Uses defer for resource cleanup and properly handles errors
//     during archive operations.
//
// # Architecture
//
// The package consists of two main components:
//
//	┌────────────────────────────────────────────────┐
//	│              tar Package                       │
//	├────────────────────────────────────────────────┤
//	│                                                │
//	│  ┌──────────────┐         ┌──────────────┐     │
//	│  │   NewReader  │         │   NewWriter  │     │
//	│  └──────┬───────┘         └──────┬───────┘     │
//	│         │                        │             │
//	│         ▼                        ▼             │
//	│  ┌──────────────┐         ┌──────────────┐     │
//	│  │     rdr      │         │     wrt      │     │
//	│  │  (Reader)    │         │  (Writer)    │     │
//	│  └──────┬───────┘         └──────┬───────┘     │
//	│         │                        │             │
//	│         ▼                        ▼             │
//	│  ┌──────────────────────────────────────┐      │
//	│  │    archive/tar (std library)         │      │
//	│  └──────────────────────────────────────┘      │
//	│                                                │
//	└────────────────────────────────────────────────┘

// Reader Component:
//   - Wraps io.ReadCloser with tar.Reader
//   - Provides query methods (List, Info, Has)
//   - Supports extraction (Get, Walk)
//   - Optional reset capability for re-reading
//
// Writer Component:
//   - Wraps io.WriteCloser with tar.Writer
//   - Adds individual files (Add method)
//   - Recursively archives directories (FromPath method)
//   - Handles links, filtering, and path transformation
//
// # Basic Usage
//
// Reading a tar archive:
//
//	file, err := os.Open("archive.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	reader, err := tar.NewReader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	// List all files
//	files, err := reader.List()
//	for _, path := range files {
//	    fmt.Println(path)
//	}
//
// Creating a tar archive:
//
//	file, err := os.Create("archive.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	writer, err := tar.NewWriter(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// Add files from directory
//	err = writer.FromPath("/path/to/files", "*", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Reader Operations
//
// The Reader interface provides several methods for working with archives:
//
// List() - Enumerate all files in the archive:
//
//	files, err := reader.List()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, path := range files {
//	    fmt.Println(path)
//	}
//
// Info() - Get file metadata:
//
//	info, err := reader.Info("path/to/file.txt")
//	if err == fs.ErrNotExist {
//	    fmt.Println("File not found")
//	} else if err == nil {
//	    fmt.Printf("Size: %d bytes\n", info.Size())
//	}
//
// Get() - Extract a specific file:
//
//	rc, err := reader.Get("path/to/file.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rc.Close()
//	data, _ := io.ReadAll(rc)
//
// Has() - Check if a file exists:
//
//	if reader.Has("config.json") {
//	    fmt.Println("Config file found")
//	}
//
// Walk() - Process all files with a callback:
//
//	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
//	    fmt.Printf("Processing: %s (%d bytes)\n", path, info.Size())
//	    if strings.HasSuffix(path, ".txt") {
//	        data, _ := io.ReadAll(rc)
//	        processTextFile(data)
//	    }
//	    return true // Continue to next file
//	})
//
// # Writer Operations
//
// The Writer interface provides methods for creating archives:
//
// Add() - Add a single file:
//
//	fileInfo, _ := os.Stat("/path/to/file.txt")
//	fileData, _ := os.Open("/path/to/file.txt")
//	defer fileData.Close()
//
//	err := writer.Add(fileInfo, fileData, "custom/path.txt", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// FromPath() - Add files recursively:
//
//	// Add all files
//	err := writer.FromPath("/source/directory", "*", nil)
//
//	// Add only Go files
//	err := writer.FromPath("/source/directory", "*.go", nil)
//
//	// Add files with path transformation
//	err := writer.FromPath("/home/user/project", "*", func(path string) string {
//	    return strings.TrimPrefix(path, "/home/user/project/")
//	})
//
// # Reset Capability
//
// The reader supports resetting to the beginning of the archive if the
// underlying io.ReadCloser implements a Reset() bool method. This is useful
// for re-reading the archive without reopening the file:
//
//	// First pass - list files
//	files, _ := reader.List()
//
//	// Reset to beginning (if supported)
//	if reader.(*tar.rdr).Reset() {
//	    // Second pass - extract files
//	    reader.Walk(extractFunc)
//	}
//
// Note: File-based readers typically support reset via seeking, while
// network streams do not.
//
// # Path Filtering and Transformation
//
// When adding files to an archive, you can filter which files to include
// and transform their paths:
//
// Filtering by pattern:
//
//	// Include only text files
//	writer.FromPath("/docs", "*.txt", nil)
//
//	// Include only markdown files
//	writer.FromPath("/docs", "*.md", nil)
//
// Path transformation:
//
//	// Strip base directory from archive paths
//	writer.FromPath("/var/log/app", "*", func(path string) string {
//	    return strings.TrimPrefix(path, "/var/log/app/")
//	})
//
//	// Flatten directory structure
//	writer.FromPath("/etc/config", "*", func(path string) string {
//	    return filepath.Base(path)
//	})
//
// # Link Handling
//
// The package preserves symbolic and hard links during archiving:
//
//   - Symbolic links: Stored with their target path (relative or absolute)
//   - Hard links: Currently treated as symbolic links
//   - Regular files: Copied with full contents
//
// Example of link preservation:
//
//	// Archive a directory with symlinks
//	writer.FromPath("/path/with/symlinks", "*", nil)
//	// Symlinks are preserved in the archive
//
//	// Extract and check link targets
//	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
//	    if info.Mode()&os.ModeSymlink != 0 {
//	        fmt.Printf("%s -> %s\n", path, link)
//	    }
//	    return true
//	})
//
// # Error Handling
//
// The package returns standard errors from the archive/tar package and
// filesystem operations:
//
//   - fs.ErrNotExist: File not found in archive (Info, Get methods)
//   - fs.ErrInvalid: Invalid file type or operation
//   - io.EOF: End of archive reached (not returned by high-level methods)
//   - Other errors: I/O errors, permission errors, etc.
//
// Example error handling:
//
//	info, err := reader.Info("config.json")
//	if err == fs.ErrNotExist {
//	    fmt.Println("Config not found, using defaults")
//	} else if err != nil {
//	    log.Fatalf("Error reading archive: %v", err)
//	} else {
//	    fmt.Printf("Config size: %d\n", info.Size())
//	}
//
// # Performance Considerations
//
// Reading Performance:
//   - Sequential reads are efficient (tar format is sequential)
//   - Random access requires scanning from the beginning
//   - Reset capability avoids reopening files for multiple passes
//   - Walk is more efficient than multiple Get calls
//
// Writing Performance:
//   - Buffering is handled by tar.Writer internally
//   - Large directories are walked incrementally
//   - File contents are streamed (no full buffering)
//
// Memory Usage:
//   - Reader: O(1) memory per operation (streams data)
//   - Writer: O(1) memory per file (streams data)
//   - List: O(n) memory for path list (n = number of files)
//
// # Best Practices
//
// DO:
//   - Always close readers and writers to release resources
//   - Call writer.Close() before closing the underlying io.WriteCloser
//   - Use Walk for processing all files (more efficient than List + Get)
//   - Check for fs.ErrNotExist when looking up specific files
//   - Use defer for cleanup in error scenarios
//
// DON'T:
//   - Don't forget to close the writer (results in corrupted archive)
//   - Don't call multiple operations simultaneously (not thread-safe)
//   - Don't assume Reset will work (check return value)
//   - Don't archive the same directory twice in one archive
//
// Example of proper resource management:
//
//	func createArchive(archivePath, sourcePath string) error {
//	    file, err := os.Create(archivePath)
//	    if err != nil {
//	        return err
//	    }
//	    defer file.Close()
//
//	    writer, err := tar.NewWriter(file)
//	    if err != nil {
//	        return err
//	    }
//	    defer writer.Close() // Must close before file.Close()
//
//	    return writer.FromPath(sourcePath, "*", nil)
//	}
//
// # Thread Safety
//
// The package is NOT thread-safe. Each reader or writer instance should
// be used by only one goroutine at a time. If you need concurrent access:
//
//   - Use separate reader instances for each goroutine
//   - Use external synchronization (mutex) if sharing instances
//   - Consider using a queue pattern for concurrent writes
//
// # Limitations
//
//   - Sequential access only (tar format limitation)
//   - Reset requires underlying reader support (not always available)
//   - Hard links are treated as symbolic links
//   - Special files (devices, pipes) are not supported
//   - Directory entries are implicit (not stored separately)
//
// # Related Packages
//
//   - archive/tar: Standard library tar implementation (used internally)
//   - github.com/nabbar/golib/archive/archive/types: Common archive interfaces
//   - github.com/nabbar/golib/archive/archive/zip: ZIP archive implementation
//   - io/fs: Standard filesystem interfaces used for file info
//
// # Examples
//
// For complete usage examples, see example_test.go in this package.
//
// # Testing
//
// The package includes comprehensive tests covering:
//   - Reader operations (list, info, get, has, walk)
//   - Writer operations (add, fromPath, close)
//   - Edge cases (empty archives, missing files, invalid inputs)
//   - Concurrency safety (no races, atomic operations where applicable)
//   - Performance benchmarks
//
// Run tests with:
//
//	go test -v
//	go test -race -v  # With race detector
//	go test -cover    # With coverage analysis
package tar
