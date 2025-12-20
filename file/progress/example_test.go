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

package progress_test

import (
	"fmt"
	"io"
	"os"

	"github.com/nabbar/golib/file/progress"
)

// ExampleTemp demonstrates creating a temporary file with automatic cleanup.
// This is the simplest use case - a temporary file that is auto-deleted on close.
func ExampleTemp() {
	// Create temporary file
	p, err := progress.Temp("example-*.tmp")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close() // Automatically deleted because IsTemp() == true

	// Write some data
	data := []byte("temporary data")
	n, err := p.Write(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes to temporary file\n", n)
	// Output: Wrote 14 bytes to temporary file
}

// ExampleOpen demonstrates opening an existing file with basic usage.
func ExampleOpen() {
	// Create a test file first
	testFile := "/tmp/progress-example.txt"
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file with progress tracking
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Read data
	data, err := io.ReadAll(p)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Read: %s\n", string(data))
	// Output: Read: Hello, World!
}

// ExampleCreate demonstrates creating a new file.
func ExampleCreate() {
	testFile := "/tmp/progress-created.txt"
	defer os.Remove(testFile)

	// Create new file
	p, err := progress.Create(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Write data
	data := []byte("New file content")
	n, err := p.Write(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created file and wrote %d bytes\n", n)
	// Output: Created file and wrote 16 bytes
}

// ExampleProgress_RegisterFctIncrement demonstrates tracking read progress with callbacks.
// The callback receives the number of bytes for each operation, not cumulative total.
func ExampleProgress_RegisterFctIncrement() {
	// Create test file
	testFile := "/tmp/progress-increment.txt"
	testData := []byte("0123456789") // 10 bytes
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open with progress tracking
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Register increment callback - receives bytes per operation
	var totalBytes int64
	p.RegisterFctIncrement(func(bytes int64) {
		totalBytes += bytes
		fmt.Printf("Read %d bytes (total: %d)\n", bytes, totalBytes)
	})

	// Read file - callback triggered on each read
	buf := make([]byte, 5)
	p.Read(buf)
	p.Read(buf)

	// Output:
	// Read 5 bytes (total: 5)
	// Read 5 bytes (total: 10)
}

// ExampleProgress_RegisterFctEOF demonstrates EOF detection callback.
func ExampleProgress_RegisterFctEOF() {
	// Create test file
	testFile := "/tmp/progress-eof.txt"
	testData := []byte("EOF Test")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Register EOF callback
	p.RegisterFctEOF(func() {
		fmt.Println("End of file reached!")
	})

	// Read entire file - EOF callback triggered at end
	io.Copy(io.Discard, p)

	// Output: End of file reached!
}

// ExampleProgress_RegisterFctReset demonstrates position reset tracking.
func ExampleProgress_RegisterFctReset() {
	// Create test file
	testFile := "/tmp/progress-reset.txt"
	testData := []byte("0123456789")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file with write mode to enable truncate
	p, err := progress.New(testFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Register reset callback
	p.RegisterFctReset(func(maxSize, currentPos int64) {
		fmt.Printf("Reset callback: max=%d, current=%d\n", maxSize, currentPos)
	})

	// Read some bytes to advance position
	buf := make([]byte, 5)
	p.Read(buf)

	// Truncate triggers reset callback
	p.Truncate(10)

	// Output: Reset callback: max=10, current=5
}

// ExampleProgress_SetBufferSize demonstrates custom buffer sizing.
func ExampleProgress_SetBufferSize() {
	// Create test file
	testFile := "/tmp/progress-buffer.txt"
	testData := make([]byte, 1024) // 1KB
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Set custom buffer size (64KB)
	p.SetBufferSize(64 * 1024)

	// Use file with custom buffer
	data, _ := io.ReadAll(p)
	fmt.Printf("Read %d bytes with custom buffer\n", len(data))

	// Output: Read 1024 bytes with custom buffer
}

// ExampleUnique demonstrates creating unique files with patterns.
func ExampleUnique() {
	// Create unique file in /tmp
	p, err := progress.Unique("/tmp", "myapp-*.dat")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.CloseDelete() // Clean up

	// Write data
	p.Write([]byte("unique file content"))

	// Get file path
	path := p.Path()
	fmt.Printf("Created unique file: %v\n", path != "")

	// Output: Created unique file: true
}

// ExampleProgress_SizeBOF demonstrates tracking bytes read from beginning.
func ExampleProgress_SizeBOF() {
	// Create test file
	testFile := "/tmp/progress-bof.txt"
	testData := []byte("0123456789") // 10 bytes
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Read 5 bytes
	buf := make([]byte, 5)
	p.Read(buf)

	// Check position from beginning
	bof, _ := p.SizeBOF()
	fmt.Printf("Bytes from start: %d\n", bof)

	// Output: Bytes from start: 5
}

// ExampleProgress_SizeEOF demonstrates calculating remaining bytes.
func ExampleProgress_SizeEOF() {
	// Create test file
	testFile := "/tmp/progress-eof-size.txt"
	testData := []byte("0123456789") // 10 bytes
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Read 5 bytes
	buf := make([]byte, 5)
	p.Read(buf)

	// Check remaining bytes
	eof, _ := p.SizeEOF()
	fmt.Printf("Remaining bytes: %d\n", eof)

	// Output: Remaining bytes: 5
}

// ExampleProgress_Stat demonstrates getting file information.
func ExampleProgress_Stat() {
	// Create test file
	testFile := "/tmp/progress-stat.txt"
	testData := []byte("File info test")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Get file info
	info, err := p.Stat()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("File size: %d bytes\n", info.Size())
	// Output: File size: 14 bytes
}

// ExampleProgress_Truncate demonstrates file truncation with reset callback.
func ExampleProgress_Truncate() {
	// Create test file
	testFile := "/tmp/progress-truncate.txt"
	testData := []byte("0123456789") // 10 bytes
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file with write permissions for truncate
	p, err := progress.New(testFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Truncate to 5 bytes
	err = p.Truncate(5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Verify new size
	info, _ := p.Stat()
	fmt.Printf("After truncate: %d bytes\n", info.Size())

	// Output: After truncate: 5 bytes
}

// ExampleProgress_IsTemp demonstrates checking if file is temporary.
func ExampleProgress_IsTemp() {
	// Create temporary file
	tmp, err := progress.Temp("test-*.tmp")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer tmp.Close()

	// Create regular file
	testFile := "/tmp/regular.txt"
	reg, err := progress.Create(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer reg.CloseDelete()

	// Check if temporary
	fmt.Printf("Temp file is temp: %v\n", tmp.IsTemp())
	fmt.Printf("Regular file is temp: %v\n", reg.IsTemp())

	// Output:
	// Temp file is temp: true
	// Regular file is temp: false
}

// ExampleProgress_CloseDelete demonstrates that temporary files are auto-deleted on close.
func ExampleProgress_CloseDelete() {
	// Create temporary file
	p, err := progress.Temp("progress-delete-*.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Write data
	p.Write([]byte("will be deleted"))

	// Check if it's a temp file
	fmt.Printf("Is temp file: %v\n", p.IsTemp())

	// Close - temp files are auto-deleted
	err = p.Close()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("File closed successfully")

	// Output:
	// Is temp file: true
	// File closed successfully
}

// Example_fileCopy demonstrates a real-world file copy operation with progress tracking.
func Example_fileCopy() {
	// Create source file
	srcFile := "/tmp/progress-source.txt"
	srcData := []byte("File copy example with progress tracking")
	if err := os.WriteFile(srcFile, srcData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(srcFile)

	// Open source with progress
	src, err := progress.Open(srcFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer src.Close()

	// Create destination
	dstFile := "/tmp/progress-dest.txt"
	dst, err := progress.Create(dstFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.Remove(dstFile)
	defer dst.Close()

	// Track copy progress with manual copy loop
	var totalBytes int64
	src.RegisterFctIncrement(func(bytes int64) {
		totalBytes += bytes
	})

	src.RegisterFctEOF(func() {
		fmt.Printf("Copy complete: %d bytes\n", totalBytes)
	})

	// Manual copy to trigger progress callbacks
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	// Output: Copy complete: 40 bytes
}

// Example_uploadSimulation demonstrates simulating file upload with progress.
func Example_uploadSimulation() {
	// Create test file
	testFile := "/tmp/progress-upload.dat"
	testData := make([]byte, 100) // 100 bytes
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file for "upload"
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Get file size for percentage calculation
	info, _ := p.Stat()
	fileSize := info.Size()

	// Track upload progress - accumulate bytes
	var totalBytes int64
	var shown bool
	p.RegisterFctIncrement(func(bytes int64) {
		totalBytes += bytes
		percentage := float64(totalBytes) / float64(fileSize) * 100
		if percentage >= 100 && !shown {
			fmt.Printf("Upload: 100%%\n")
			shown = true
		}
	})

	p.RegisterFctEOF(func() {
		fmt.Println("Upload complete!")
	})

	// Simulate upload with manual read loop
	buf := make([]byte, 32)
	for {
		_, err := p.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	// Output:
	// Upload: 100%
	// Upload complete!
}

// Example_batchProcessing demonstrates processing a file in chunks with progress.
func Example_batchProcessing() {
	// Create test file
	testFile := "/tmp/progress-batch.txt"
	testData := []byte("Line1\nLine2\nLine3\nLine4\nLine5\n")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		fmt.Printf("Setup error: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	// Open file
	p, err := progress.Open(testFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer p.Close()

	// Track processing
	lineCount := 0
	p.RegisterFctIncrement(func(bytes int64) {
		// Called on each read
	})

	p.RegisterFctEOF(func() {
		fmt.Printf("Processed %d lines\n", lineCount)
	})

	// Process in chunks
	buf := make([]byte, 10)
	for {
		n, err := p.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					lineCount++
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	// Output: Processed 5 lines
}
