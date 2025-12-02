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

/*
Package progress provides file I/O operations with integrated real-time progress tracking and event callbacks.

# Overview

The progress package wraps standard Go file operations with comprehensive progress monitoring capabilities.
It implements all standard I/O interfaces (io.Reader, io.Writer, io.Seeker, etc.) while adding transparent
progress tracking through configurable callbacks. This enables applications to monitor file operations in
real-time without modifying existing I/O code patterns.

# Design Philosophy

1. Standard Interface Compliance: Fully implements all Go I/O interfaces for drop-in replacement
2. Non-Intrusive Monitoring: Progress tracking doesn't alter I/O semantics or performance characteristics
3. Callback-Based Events: Flexible notification system for progress updates, resets, and EOF detection
4. Thread-Safe Operations: Uses atomic operations for safe concurrent access to progress state
5. Resource Management: Proper cleanup with support for temporary file auto-deletion

# Key Features

  - Full standard I/O interface implementation (io.Reader, io.Writer, io.Seeker, io.Closer, etc.)
  - Real-time progress tracking with increment callbacks
  - Reset callbacks for seek operations and position changes
  - EOF detection callbacks for completion notifications
  - Configurable buffer sizes for optimized performance (default: 32KB)
  - Temporary file creation with automatic cleanup
  - Unique file generation with custom patterns
  - File position tracking (beginning-of-file and end-of-file calculations)
  - Thread-safe atomic operations for progress state
  - Support for both regular and temporary files

# Basic Usage

Opening an existing file with progress tracking:

	import "github.com/nabbar/golib/file/progress"

	// Open file
	p, err := progress.Open("largefile.dat")
	if err != nil {
	    log.Fatal(err)
	}
	defer p.Close()

	// Register progress callback (called on every read/write operation)
	p.RegisterFctIncrement(func(bytes int64) {
	    fmt.Printf("Processed: %d bytes\n", bytes)
	})

	// Register EOF callback (called when file reaches end)
	p.RegisterFctEOF(func() {
	    fmt.Println("File processing complete!")
	})

	// Use like any io.Reader - callbacks are invoked transparently
	io.Copy(io.Discard, p)

# File Creation

Creating new files:

	// Create new file
	p, err := progress.Create("/path/to/newfile.dat")
	if err != nil {
	    log.Fatal(err)
	}
	defer p.Close()

	// Write data - progress callbacks are triggered
	data := []byte("Hello, World!")
	n, err := p.Write(data)

Creating with custom flags and permissions:

	// Custom file creation with flags
	import "os"

	p, err := progress.New("/path/to/file.dat",
	    os.O_RDWR|os.O_CREATE|os.O_TRUNC,
	    0644)

# Temporary Files

Creating temporary files with automatic cleanup:

	// Create temporary file (auto-deleted on close if IsTemp() == true)
	p, err := progress.Temp("myapp-*.tmp")
	if err != nil {
	    log.Fatal(err)
	}
	defer p.Close() // Automatically deleted

	// Write temporary data
	p.Write([]byte("temporary content"))

Creating unique files in specific directory:

	// Create unique file in custom location
	p, err := progress.Unique("/tmp/myapp", "data-*.bin")
	if err != nil {
	    log.Fatal(err)
	}
	defer p.Close()

	// Check if file is temporary
	if p.IsTemp() {
	    fmt.Println("This is a temporary file")
	}

# Progress Callbacks

The package provides three types of callbacks for monitoring file operations:

Increment Callback:

Called after every successful read or write operation with the cumulative byte count.

	p.RegisterFctIncrement(func(bytes int64) {
	    fmt.Printf("Total bytes processed: %d\n", bytes)
	})

Reset Callback:

Called when file position is reset (e.g., via Seek operations) with the maximum
position reached and current position.

	p.RegisterFctReset(func(maxSize, currentPos int64) {
	    fmt.Printf("Position reset: was at %d, now at %d\n", maxSize, currentPos)
	})

EOF Callback:

Called when end-of-file is reached during read operations.

	p.RegisterFctEOF(func() {
	    fmt.Println("Reached end of file")
	})

# Buffer Configuration

Customizing buffer size for performance optimization:

	p, err := progress.Open("file.dat")
	if err != nil {
	    log.Fatal(err)
	}
	defer p.Close()

	// Set custom buffer size (64KB)
	p.SetBufferSize(64 * 1024)

	// Larger buffers reduce callback frequency but use more memory
	// Default buffer size is 32KB (DefaultBuffSize constant)

# Advanced Features

File Position Tracking:

	// Get bytes from beginning to current position
	bof, err := p.SizeBOF()
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("Read %d bytes so far\n", bof)

	// Get remaining bytes from current position to EOF
	eof, err := p.SizeEOF()
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("Remaining bytes: %d\n", eof)

File Truncation:

	// Truncate file to specific size
	err := p.Truncate(1024) // Truncate to 1KB
	if err != nil {
	    log.Fatal(err)
	}
	// Reset callback is automatically triggered

Syncing to Disk:

	// Force write of buffered data to disk
	err := p.Sync()
	if err != nil {
	    log.Fatal(err)
	}

# Callback Propagation

Progress callbacks can be propagated to other Progress instances:

	// Source file with progress tracking
	src, _ := progress.Open("source.dat")
	src.RegisterFctIncrement(func(bytes int64) {
	    fmt.Printf("Source: %d bytes\n", bytes)
	})

	// Destination file - inherits source's callbacks
	dst, _ := progress.Create("dest.dat")
	src.SetRegisterProgress(dst)

	// Both files now share the same progress callbacks
	io.Copy(dst, src)

# File Information

Accessing file metadata:

	// Get file statistics
	info, err := p.Stat()
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("File size: %d bytes\n", info.Size())
	fmt.Printf("Modified: %v\n", info.ModTime())

	// Get file path
	path := p.Path()
	fmt.Printf("File path: %s\n", path)

# Cleanup Operations

Standard close:

	// Close file (keeps file on disk)
	err := p.Close()
	if err != nil {
	    log.Fatal(err)
	}

Close and delete:

	// Close and delete file
	err := p.CloseDelete()
	if err != nil {
	    log.Fatal(err)
	}
	// File is removed from filesystem

# Use Cases

File Upload/Download Progress:

Monitor file transfer operations in web applications or CLI tools.

	p, _ := progress.Open("upload.bin")
	defer p.Close()

	var totalBytes int64
	p.RegisterFctIncrement(func(bytes int64) {
	    totalBytes = bytes
	    percentage := float64(bytes) / float64(fileSize) * 100
	    fmt.Printf("\rUploading: %.2f%%", percentage)
	})

	// Upload to server
	http.Post(url, "application/octet-stream", p)

Large File Processing:

Track progress when processing large data files.

	p, _ := progress.Open("largefile.csv")
	defer p.Close()

	p.RegisterFctIncrement(func(bytes int64) {
	    fmt.Printf("Processed %d MB\n", bytes/(1024*1024))
	})

	scanner := bufio.NewScanner(p)
	for scanner.Scan() {
	    processLine(scanner.Text())
	}

Temporary Work Files:

Use temporary files for intermediate processing stages.

	// Create temp file for processing
	tmp, _ := progress.Temp("processing-*.dat")
	defer tmp.Close() // Auto-deleted

	// Process data through temporary file
	io.Copy(tmp, dataSource)
	tmp.Seek(0, io.SeekStart)
	processData(tmp)

# Error Handling

The package defines error codes in the errors.go file:

	var (
	    ErrorParamEmpty          // Empty parameters
	    ErrorNilPointer          // Nil pointer dereference
	    ErrorIOFileStat          // File stat error
	    ErrorIOFileSeek          // File seek error
	    ErrorIOFileTruncate      // File truncate error
	    ErrorIOFileSync          // File sync error
	    ErrorIOFileOpen          // File open error
	    ErrorIOFileTempNew       // Temporary file creation error
	    ErrorIOFileTempClose     // Temporary file close error
	    ErrorIOFileTempRemove    // Temporary file removal error
	)

All errors are wrapped using github.com/nabbar/golib/errors for enhanced error handling.

# Performance Considerations

Buffer Sizing:

The default buffer size (32KB) is optimized for general use. Adjust based on your workload:

  - Small files (<1MB): Default buffer is sufficient
  - Large files (>100MB): Increase to 64KB or 128KB for better performance
  - High-frequency operations: Larger buffers reduce callback overhead
  - Memory-constrained: Use smaller buffers (16KB) to reduce memory footprint

Memory Usage:

  - Base overhead: ~200 bytes per Progress instance
  - Buffer allocation: Configurable (default 32KB)
  - No additional allocations during normal I/O operations
  - Atomic operations ensure minimal lock contention

Callback Overhead:

  - Increment callback: Called on every Read/Write (can be frequent)
  - Reset callback: Called on Seek/Truncate operations (infrequent)
  - EOF callback: Called once per file read completion
  - Nil callbacks have minimal overhead (simple atomic load check)

# Thread Safety

The Progress interface uses atomic operations for callback storage and retrieval,
making callback registration thread-safe. However, file I/O operations themselves
follow standard Go file semantics:

  - Multiple goroutines can read from the same file concurrently using ReadAt
  - Concurrent Read/Write/Seek operations require external synchronization
  - Callback invocations are sequential (not concurrent)

# Dependencies

The package depends on:

  - Standard library: os, io, path/filepath, sync/atomic
  - github.com/nabbar/golib/errors: Enhanced error handling

# Interface Compliance

The Progress interface implements:

  - io.Reader, io.ReaderAt, io.ReaderFrom
  - io.Writer, io.WriterAt, io.WriterTo, io.StringWriter
  - io.Seeker
  - io.Closer
  - io.ByteReader, io.ByteWriter
  - All combined interfaces (io.ReadCloser, io.ReadWriteCloser, etc.)

This ensures drop-in compatibility with any code expecting standard I/O interfaces.

# Examples

See example_test.go for comprehensive usage examples including:

  - Basic file operations with progress tracking
  - Temporary file creation and management
  - Progress callback registration and usage
  - Buffer size configuration
  - File position tracking
  - Error handling patterns
  - Real-world use cases (file copy, upload simulation, batch processing)
*/
package progress
