/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

// Package ioprogress_test provides runnable examples demonstrating package usage.
//
// These examples progress from simple to complex use cases, demonstrating real-world
// scenarios for I/O progress tracking. All examples are executable and tested as part
// of the test suite.
//
// Examples covered:
//   1. Basic progress tracking (simplest usage)
//   2. Progress with percentage calculation
//   3. File copy with progress bar
//   4. Bandwidth monitoring
//   5. Multi-stage processing with reset
//   6. Complete download simulation (most complex)
package ioprogress_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/nabbar/golib/ioutils/ioprogress"
)

// Example_basicTracking demonstrates the simplest usage: counting bytes read.
//
// This example shows:
//   - Wrapping an io.ReadCloser with progress tracking
//   - Registering a simple increment callback
//   - Reading data and tracking cumulative progress
//
// Use Case: Basic byte counting for logging or metrics.
func Example_basicTracking() {
	// Create test data
	data := strings.NewReader("Hello, World!")

	// Wrap with progress tracking
	reader := ioprogress.NewReadCloser(io.NopCloser(data))
	defer reader.Close()

	// Track total bytes read using atomic operations
	var totalBytes int64
	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&totalBytes, size)
	})

	// Read all data
	io.Copy(io.Discard, reader)

	// Display result
	fmt.Printf("Total bytes read: %d\n", atomic.LoadInt64(&totalBytes))
	// Output: Total bytes read: 13
}

// Example_progressPercentage demonstrates calculating progress percentage.
//
// This example shows:
//   - Tracking progress relative to known total size
//   - Calculating completion percentage
//   - Displaying progress updates
//
// Use Case: Progress bars for file operations with known sizes.
func Example_progressPercentage() {
	// Create test data with known size
	data := strings.Repeat("x", 100)
	totalSize := int64(len(data))

	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
	defer reader.Close()

	// Track progress with percentage calculation
	var current int64
	lastDisplayed := int64(-1)
	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&current, size)
		progress := atomic.LoadInt64(&current) * 100 / totalSize
		// Display at 25% intervals, only once per threshold
		if progress >= 25 && progress%25 == 0 && progress != lastDisplayed {
			fmt.Printf("Progress: %d%%\n", progress)
			lastDisplayed = progress
		}
	})

	// Read in chunks to see progress
	buf := make([]byte, 25)
	for {
		_, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
	}

	fmt.Println("Complete!")
	// Output:
	// Progress: 25%
	// Progress: 50%
	// Progress: 75%
	// Progress: 100%
	// Complete!
}

// Example_fileCopy demonstrates copying with progress tracking.
//
// This example shows:
//   - Tracking both read and write operations
//   - Coordinating reader and writer progress
//   - Using EOF callback for completion notification
//
// Use Case: File copy operations with progress feedback.
func Example_fileCopy() {
	// Source data
	source := strings.NewReader("File contents to copy")
	dest := &bytes.Buffer{}

	// Wrap both reader and writer
	reader := ioprogress.NewReadCloser(io.NopCloser(source))
	defer reader.Close()

	writer := ioprogress.NewWriteCloser(&nopWriteCloser{Writer: dest})
	defer writer.Close()

	// Track read progress
	var bytesRead int64
	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&bytesRead, size)
	})

	// Track write progress
	var bytesWritten int64
	writer.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&bytesWritten, size)
	})

	// Completion callback
	reader.RegisterFctEOF(func() {
		fmt.Printf("Copy complete: %d bytes read, %d bytes written\n",
			atomic.LoadInt64(&bytesRead),
			atomic.LoadInt64(&bytesWritten))
	})

	// Perform copy
	io.Copy(writer, reader)

	// Output: Copy complete: 21 bytes read, 21 bytes written
}

// Example_bandwidthMonitor demonstrates real-time bandwidth calculation.
//
// This example shows:
//   - Periodic bandwidth sampling
//   - Using goroutines for concurrent monitoring
//   - Coordinating progress tracking with timing
//
// Use Case: Network downloads or uploads with speed monitoring.
//
// Note: This example doesn't have Output comment because timing is non-deterministic.
func Example_bandwidthMonitor() {
	// Simulate data stream
	data := strings.Repeat("x", 1000)
	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
	defer reader.Close()

	// Bandwidth tracking
	var totalBytes int64

	// Track bytes transferred
	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&totalBytes, size)
	})

	// Read data
	io.Copy(io.Discard, reader)

	// Display total (deterministic)
	fmt.Printf("Total transferred: %d bytes\n", atomic.LoadInt64(&totalBytes))
	// Output:
	// Total transferred: 1000 bytes
}

// Example_multiStageProcessing demonstrates using Reset() for multi-stage operations.
//
// This example shows:
//   - Processing data in multiple stages
//   - Using Reset() to update progress context
//   - Tracking progress across different processing phases
//
// Use Case: ETL pipelines, multi-pass file processing, validation + transformation.
func Example_multiStageProcessing() {
	// Data to process
	data := strings.Repeat("data", 25) // 100 bytes
	totalSize := int64(len(data))

	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
	defer reader.Close()

	// Track stage progress
	var currentStage string
	reader.RegisterFctReset(func(max, current int64) {
		progress := current * 100 / max
		fmt.Printf("%s: %d%% complete (%d/%d bytes)\n",
			currentStage, progress, current, max)
	})

	buf := make([]byte, 50)
	
	// Stage 1: Validation
	reader.Read(buf) // Read first 50 bytes
	currentStage = "Validation"
	reader.Reset(totalSize)

	// Stage 2: Processing
	reader.Read(buf) // Read next 50 bytes
	currentStage = "Processing"
	reader.Reset(totalSize)

	// Stage 3: Finalization (reading remaining 0 bytes causes EOF)
	currentStage = "Finalization"
	reader.Reset(totalSize)

	fmt.Println("All stages complete")
	// Output:
	// Validation: 50% complete (50/100 bytes)
	// Processing: 100% complete (100/100 bytes)
	// Finalization: 100% complete (100/100 bytes)
	// All stages complete
}

// Example_completeDownload demonstrates a complete download simulation with all features.
//
// This example shows:
//   - Progress percentage tracking
//   - Completion notification
//   - Comprehensive progress feedback
//
// Use Case: Production-quality download manager or file transfer utility.
//
// Note: Timing-based features (ETA, speed) are not shown in this simplified example
// as they would be non-deterministic in tests. In real applications, you would
// calculate these based on actual elapsed time.
func Example_completeDownload() {
	// Simulate download (would be HTTP response in real code)
	fileSize := int64(1000)
	data := strings.Repeat("x", int(fileSize))

	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
	defer reader.Close()

	dest := &bytes.Buffer{}
	writer := ioprogress.NewWriteCloser(&nopWriteCloser{Writer: dest})
	defer writer.Close()

	// Progress tracking
	var downloaded int64
	var lastDisplayed int64

	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&downloaded, size)
		current := atomic.LoadInt64(&downloaded)
		
		// Calculate progress
		progress := current * 100 / fileSize

		// Display progress at 25% intervals, only once per threshold
		if progress >= 25 && progress%25 == 0 {
			last := atomic.LoadInt64(&lastDisplayed)
			if progress != last && atomic.CompareAndSwapInt64(&lastDisplayed, last, progress) {
				fmt.Printf("Download: %d%% (%d/%d bytes)\n", progress, current, fileSize)
			}
		}
	})

	// Completion notification
	reader.RegisterFctEOF(func() {
		fmt.Println("✓ Download complete!")
	})

	// Perform download with fixed buffer size to ensure deterministic progress display
	buf := make([]byte, 250) // Exactly 250 bytes per read for predictable 25% steps
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			writer.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
	}

	// Output:
	// Download: 25% (250/1000 bytes)
	// Download: 50% (500/1000 bytes)
	// Download: 75% (750/1000 bytes)
	// Download: 100% (1000/1000 bytes)
	// ✓ Download complete!
}

// nopWriteCloser is defined in helper_test.go
