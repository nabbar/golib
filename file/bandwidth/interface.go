/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

// Package bandwidth provides bandwidth throttling and rate limiting for file I/O operations.
//
// This package integrates seamlessly with the file/progress package to control data transfer
// rates in bytes per second. It implements time-based throttling using atomic operations for
// thread-safe concurrent usage.
//
// Key features:
//   - Configurable bytes-per-second limits
//   - Zero-cost when set to unlimited (0 bytes/second)
//   - Thread-safe atomic operations
//   - Seamless integration with progress tracking
//
// Example usage:
//
//	import (
//	    "github.com/nabbar/golib/file/bandwidth"
//	    "github.com/nabbar/golib/file/progress"
//	    "github.com/nabbar/golib/size"
//	)
//
//	// Create bandwidth limiter (1 MB/s)
//	bw := bandwidth.New(size.SizeMiB)
//
//	// Open file with progress tracking
//	fpg, _ := progress.Open("largefile.dat")
//	defer fpg.Close()
//
//	// Register bandwidth limiting
//	bw.RegisterIncrement(fpg, nil)
//
//	// All I/O operations will be throttled to 1 MB/s
package bandwidth

import (
	"sync/atomic"

	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

// BandWidth defines the interface for bandwidth control and rate limiting.
//
// This interface provides methods to register bandwidth limiting callbacks
// with progress-enabled file operations. It integrates seamlessly with the
// progress package to enforce bytes-per-second transfer limits.
type BandWidth interface {

	// RegisterIncrement registers a function to be called when the progress of
	// a file being read or written reaches a certain number of bytes. The
	// function will be called with the number of bytes that have been read or
	// written from the start of the file. The function is called even if the
	// registered progress is not reached (i.e. if the file is smaller than
	// the registered progress). The function is called with the current
	// progress when the file is closed (i.e. when io.Copy returns io.EOF).
	//
	// The function is called with the following signature:
	//
	// func(size int64)
	//
	// If the function is nil, it is simply ignored.
	//
	RegisterIncrement(fpg libfpg.Progress, fi libfpg.FctIncrement)

	// RegisterReset registers a function to be called when the progress of a
	// file being read or written is reset. The function will be called with the
	// maximum progress that has been reached and the current progress when
	// the file is closed (i.e. when io.Copy returns io.EOF).
	//
	// The function is called with the following signature:
	//
	// func(size, current int64)
	//
	// If the function is nil, it is simply ignored.
	RegisterReset(fpg libfpg.Progress, fr libfpg.FctReset)
}

// New returns a new BandWidth instance with the given bytes by second limit.
// The instance returned by New implements the BandWidth interface.
//
// The bytesBySecond argument specifies the maximum number of bytes that
// can be read or written to the underlying file per second. If the
// underlying file is smaller than the maximum number of bytes, the
// registered functions will be called with the size of the underlying
// file. The registered functions will be called with the current progress
// when the file is closed (i.e. when io.Copy returns io.EOF).
//
// The returned instance is safe for concurrent use.
//
// The returned instance is not safe for concurrent writes. If the
// returned instance is used concurrently, the caller must ensure that
// the instance is not modified concurrently.
//
// The returned instance is not safe for concurrent reads. If the
// returned instance is used concurrently, the caller must ensure that
// the instance is not modified concurrently.
//
// The returned instance is not safe for concurrent seeks. If the
// returned instance is used concurrently, the caller must ensure that
// the instance is not modified concurrently.
func New(bytesBySecond libsiz.Size) BandWidth {
	return &bw{
		t: new(atomic.Value),
		l: bytesBySecond,
	}
}
