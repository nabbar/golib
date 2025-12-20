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

// Package hookfile provides file-based logging hooks for logrus.
// This file handles log file aggregation and rotation functionality.
// It manages multiple writers to the same log file efficiently.
package hookfile

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	iotagg "github.com/nabbar/golib/ioutils/aggregator"
)

// fileAgg represents an aggregated file writer with reference counting.
// It manages a single log file that can be shared by multiple loggers.
type fileAgg struct {
	i *atomic.Int64
	r *os.Root
	f *os.File
	a iotagg.Aggregator
}

// Global map to manage file aggregators by file path
// Uses atomic operations for thread-safe access
var (
	// agg is a thread-safe map that maintains a collection of file aggregators
	// The key is the file path, and the value is the file aggregator instance
	agg = libatm.NewMapTyped[string, *fileAgg]()
)

// init initializes the package and sets up a finalizer to clean up resources
// when the program exits. This ensures all log files are properly closed.
func init() {
	runtime.SetFinalizer(agg, func(a libatm.MapTyped[string, *fileAgg]) {
		a.Range(func(k string, v *fileAgg) bool {
			if v != nil {
				_ = v.a.Close()
				_ = v.f.Close()
				_ = v.r.Close()
			}
			return true
		})
	})
}

// setAgg retrieves or creates a file aggregator for the given file path.
// If an aggregator already exists for the path, its reference count is incremented.
//
// Parameters:
//   - k: The file path to aggregate writes to
//   - m: The file mode to use when creating new files
//   - cre: Whether to create the file if it doesn't exist (enables O_CREATE flag)
//
// Returns:
//   - io.Writer: A writer that writes to the aggregated file
//   - error: Any error that occurred while creating or accessing the file
//
// The function is thread-safe and handles concurrent access to the same file.
func setAgg(k string, m os.FileMode, cre bool) (io.Writer, error) {
	i, l := agg.Load(k)

	if l && i != nil {
		i.i.Add(1)
		agg.Store(k, i)
		return i.a, nil
	}

	var e error
	i, e = newAgg(k, m, cre)

	if e != nil {
		return nil, e
	}

	agg.Store(k, i)
	return i.a, nil
}

// delAgg decreases the reference count for the file aggregator at the given path.
// If the reference count reaches zero, the file and its resources are closed and removed.
//
// Parameters:
//   - k: The file path whose aggregator's reference count should be decremented
//
// This function is thread-safe and ensures proper resource cleanup.
func delAgg(k string) {
	i, _ := agg.Load(k)
	if i == nil {
		return
	}

	if i.i.Add(-1) > 0 {
		agg.Store(k, i)
	} else {
		agg.Delete(k)
		_ = i.a.Close()
		_ = i.f.Close()
		_ = i.r.Close()
	}
}

// newAgg creates a new file aggregator for the specified file path.
// It opens the file in append mode and sets up the necessary writers.
//
// Parameters:
//   - p: The file path to create the aggregator for
//   - m: The file mode to use when creating the file
//   - cre: Whether to create the file if it doesn't exist (enables O_CREATE flag)
//
// Returns:
//   - *fileAgg: The newly created file aggregator
//   - error: Any error that occurred during file operations
//
// The function ensures proper error handling and resource cleanup in case of failures.
// It also sets up a sync function that detects external log rotation and automatically
// reopens the file when rotation is detected.
func newAgg(p string, m os.FileMode, cre bool) (*fileAgg, error) {
	i := &fileAgg{
		i: new(atomic.Int64),
		r: nil,
		f: nil,
		a: nil,
	}

	fl := os.O_WRONLY | os.O_APPEND
	if cre {
		fl = fl | os.O_CREATE
	}

	if r, e := os.OpenRoot(filepath.Dir(p)); e != nil {
		return nil, e
	} else if f, e := r.OpenFile(filepath.Base(p), fl, m); e != nil {
		_ = r.Close()
		return nil, e
	} else if _, e = f.Seek(0, io.SeekEnd); e != nil {
		_ = f.Close()
		_ = r.Close()
		return nil, e
	} else {
		i.r = r
		i.f = f
	}

	a, e := iotagg.New(context.Background(), iotagg.Config{
		AsyncTimer: 0,
		AsyncMax:   0,
		AsyncFct:   nil,
		SyncTimer:  time.Second,
		SyncFct: func(ctx context.Context) {
			// Sync function is called periodically to flush buffered data and detect log rotation

			// First sync the current file to ensure all buffered data is written to disk
			syncErr := i.f.Sync()

			// Check if file has been rotated (only if CreatePath is enabled)
			// Rotation can occur when external tools (like logrotate) rename/move the log file
			needReopen := syncErr != nil
			if !needReopen && cre {
				// Get stats of current file descriptor (the file we're writing to)
				currentStat, err1 := i.f.Stat()
				// Get stats of file on disk at the configured path
				diskStat, err2 := os.Stat(p)

				// If disk file doesn't exist or is different from our FD, rotation occurred
				// os.SameFile compares inode numbers to detect if files are the same
				if err2 != nil || (err1 == nil && !os.SameFile(currentStat, diskStat)) {
					needReopen = true
				}
			}

			if !needReopen {
				return
			}

			// Close current file descriptor (it may point to a rotated file)
			_ = i.f.Close()

			// Reopen the file at the configured path (creates new file if rotated)
			if f, e := i.r.OpenFile(filepath.Base(p), fl, m); e != nil {
				_ = i.r.Close()
				// Log error to stderr since we can't write to the log file
				_, _ = fmt.Fprintf(os.Stderr, "error opening file %s: %v\n", p, e)
			} else {
				// Seek to end to continue appending
				_, _ = f.Seek(0, io.SeekEnd)
				i.f = f
			}
		},
		BufWriter: 250,
		FctWriter: func(p []byte) (n int, err error) {
			return i.f.Write(p)
		},
	})

	if e != nil {
		_ = i.f.Close()
		_ = i.r.Close()
		return nil, e
	}

	a.SetLoggerError(func(msg string, err ...error) {
		for _, er := range err {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", msg, er)
		}
	})

	if e = a.Start(context.Background()); e != nil {
		_ = a.Close()
		_ = i.f.Close()
		_ = i.r.Close()
		return nil, e
	}

	i.a = a

	return i, nil
}

// ResetOpenFiles closes all open file aggregators and clears the aggregator map.
// This function is primarily used for testing and cleanup purposes.
//
// It iterates through all registered file aggregators and:
//   - Closes the aggregator writer
//   - Closes the underlying file descriptor
//   - Closes the root file handle
//   - Removes the aggregator from the global map
//
// This function is thread-safe but should be used with caution in production
// as it will close all active log file handles.
func ResetOpenFiles() {
	agg.Range(func(k string, v *fileAgg) bool {
		_ = v.a.Close()
		_ = v.f.Close()
		_ = v.r.Close()
		agg.Delete(k)
		return true
	})
}
