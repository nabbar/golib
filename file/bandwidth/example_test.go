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

package bandwidth_test

import (
	"fmt"
	"io"
	"os"

	"github.com/nabbar/golib/file/bandwidth"
	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

// Example_basic demonstrates the simplest usage of the bandwidth package
// with unlimited bandwidth (no throttling).
func Example_basic() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-basic-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("Hello, World!"))
	tmpFile.Close()

	// Create bandwidth limiter with 0 (unlimited)
	bw := bandwidth.New(0)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register bandwidth limiting (no-op with 0 limit)
	bw.RegisterIncrement(fpg, nil)

	// Read file content
	data, _ := io.ReadAll(fpg)
	fmt.Printf("Read %d bytes\n", len(data))

	// Output:
	// Read 13 bytes
}

// Example_withLimit demonstrates bandwidth limiting with a specific bytes-per-second rate.
func Example_withLimit() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-limit-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 1024)) // 1KB
	tmpFile.Close()

	// Create bandwidth limiter: 10 MB/s (high enough to avoid throttling in test)
	bw := bandwidth.New(10 * libsiz.SizeMega)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register bandwidth limiting
	bw.RegisterIncrement(fpg, nil)

	// Read file content (limited but fast for testing)
	data, _ := io.ReadAll(fpg)
	fmt.Printf("Read %d bytes with bandwidth limit\n", len(data))

	// Output:
	// Read 1024 bytes with bandwidth limit
}

// Example_withCallback demonstrates bandwidth limiting with a progress callback.
func Example_withCallback() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-callback-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 2048)) // 2KB
	tmpFile.Close()

	// Create bandwidth limiter: 10 MB/s (high enough for testing)
	bw := bandwidth.New(10 * libsiz.SizeMega)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Track progress with callback
	var totalBytes int64
	bw.RegisterIncrement(fpg, func(size int64) {
		totalBytes += size // Accumulate total
	})

	// Read file content
	io.ReadAll(fpg)
	fmt.Printf("Transferred %d bytes total\n", totalBytes)

	// Output:
	// Transferred 2048 bytes total
}

// Example_withResetCallback demonstrates handling reset events during file operations.
func Example_withResetCallback() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-reset-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 1024))
	tmpFile.Close()

	// Create bandwidth limiter
	bw := bandwidth.New(0)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register reset callback
	var resetCalled bool
	bw.RegisterReset(fpg, func(size, current int64) {
		resetCalled = true
		fmt.Printf("Reset called with size=%d, current=%d\n", size, current)
	})

	// Read part of file and reset
	buffer := make([]byte, 512)
	fpg.Read(buffer)
	fpg.Reset(1024)

	if resetCalled {
		fmt.Println("Reset was triggered")
	}

	// Output:
	// Reset called with size=1024, current=512
	// Reset was triggered
}

// Example_multipleCallbacks demonstrates using both increment and reset callbacks together.
func Example_multipleCallbacks() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-multi-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 1024))
	tmpFile.Close()

	// Create bandwidth limiter: unlimited for testing
	bw := bandwidth.New(0)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Track increments
	var incrementCount int
	bw.RegisterIncrement(fpg, func(size int64) {
		incrementCount++
	})

	// Track resets
	var resetCount int
	bw.RegisterReset(fpg, func(size, current int64) {
		resetCount++
	})

	// Perform operations
	buffer := make([]byte, 256)
	fpg.Read(buffer)
	fpg.Reset(1024)

	fmt.Printf("Increments: %d, Resets: %d\n", incrementCount, resetCount)

	// Output:
	// Increments: 1, Resets: 1
}

// Example_customLimit demonstrates using a custom bandwidth limit value.
func Example_customLimit() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-custom-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 512))
	tmpFile.Close()

	// Create bandwidth limiter with custom limit: 10 MB/s
	customLimit := libsiz.Size(10 * libsiz.SizeMega)
	bw := bandwidth.New(customLimit)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register with no callback
	bw.RegisterIncrement(fpg, nil)

	// Read file
	data, _ := io.ReadAll(fpg)
	fmt.Printf("Read %d bytes with custom limit\n", len(data))

	// Output:
	// Read 512 bytes with custom limit
}

// Example_highBandwidth demonstrates using high bandwidth limits for fast transfers.
func Example_highBandwidth() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-high-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 1024))
	tmpFile.Close()

	// Create bandwidth limiter: 100 MB/s (very high)
	bw := bandwidth.New(100 * libsiz.SizeMega)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register bandwidth limiting
	bw.RegisterIncrement(fpg, nil)

	// Read file content (essentially no throttling for small files)
	data, _ := io.ReadAll(fpg)
	fmt.Printf("Read %d bytes with high bandwidth limit\n", len(data))

	// Output:
	// Read 1024 bytes with high bandwidth limit
}

// Example_progressTracking demonstrates combining bandwidth limiting with detailed progress tracking.
func Example_progressTracking() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-progress-*.dat")
	defer os.Remove(tmpFile.Name())
	testData := make([]byte, 4096) // 4KB
	tmpFile.Write(testData)
	tmpFile.Close()

	// Create bandwidth limiter: unlimited for testing
	bw := bandwidth.New(0)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Track detailed progress
	var totalBytes int64

	bw.RegisterIncrement(fpg, func(size int64) {
		totalBytes += size
	})

	// Read file content
	io.ReadAll(fpg)

	fmt.Printf("Transfer complete: %d bytes total\n", totalBytes)

	// Output:
	// Transfer complete: 4096 bytes total
}

// Example_nilCallbacks demonstrates that nil callbacks are safely handled.
func Example_nilCallbacks() {
	// Create a temporary test file
	tmpFile, _ := os.CreateTemp("", "example-nil-*.dat")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(make([]byte, 256))
	tmpFile.Close()

	// Create bandwidth limiter: unlimited
	bw := bandwidth.New(0)

	// Open file with progress tracking
	fpg, _ := libfpg.Open(tmpFile.Name())
	defer fpg.Close()

	// Register with nil callbacks (safe to do)
	bw.RegisterIncrement(fpg, nil)
	bw.RegisterReset(fpg, nil)

	// Read file content
	data, _ := io.ReadAll(fpg)
	fmt.Printf("Read %d bytes with nil callbacks\n", len(data))

	// Output:
	// Read 256 bytes with nil callbacks
}

// Example_variousSizes demonstrates bandwidth limiting with different size constants.
func Example_variousSizes() {
	// Create test files
	tmpFile1, _ := os.CreateTemp("", "example-sizes-1-*.dat")
	tmpFile2, _ := os.CreateTemp("", "example-sizes-2-*.dat")
	tmpFile3, _ := os.CreateTemp("", "example-sizes-3-*.dat")
	defer os.Remove(tmpFile1.Name())
	defer os.Remove(tmpFile2.Name())
	defer os.Remove(tmpFile3.Name())

	tmpFile1.Write(make([]byte, 128))
	tmpFile2.Write(make([]byte, 256))
	tmpFile3.Write(make([]byte, 512))
	tmpFile1.Close()
	tmpFile2.Close()
	tmpFile3.Close()

	// Different bandwidth limits (high values for testing)
	limits := []libsiz.Size{
		10 * libsiz.SizeMega,  // 10 MB/s
		100 * libsiz.SizeMega, // 100 MB/s
		libsiz.SizeGiga,       // 1 GB/s
	}

	files := []string{tmpFile1.Name(), tmpFile2.Name(), tmpFile3.Name()}

	for i, limit := range limits {
		bw := bandwidth.New(limit)
		fpg, _ := libfpg.Open(files[i])
		bw.RegisterIncrement(fpg, nil)
		data, _ := io.ReadAll(fpg)
		fpg.Close()

		fmt.Printf("File %d: %d bytes with high bandwidth limit\n", i+1, len(data))
	}

	// Output:
	// File 1: 128 bytes with high bandwidth limit
	// File 2: 256 bytes with high bandwidth limit
	// File 3: 512 bytes with high bandwidth limit
}
