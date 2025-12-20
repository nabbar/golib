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

package multi

import (
	"io"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
)

// Multi is a thread-safe I/O multi-writer that broadcasts writes to multiple
// destinations and manages a single input source.
//
// The Multi interface extends io.ReadWriteCloser and io.StringWriter with
// methods for managing multiple write destinations and an input source.
// All methods are safe for concurrent use.
//
// Write operations (Write, WriteString) are broadcast to all registered
// writers via io.MultiWriter. Read operations are performed on the currently
// set input source.
//
// Example usage:
//
//	m := multi.New()
//	var buf1, buf2 bytes.Buffer
//	m.AddWriter(&buf1, &buf2)
//	m.Write([]byte("data")) // written to both buf1 and buf2
type Multi interface {
	io.ReadWriteCloser
	io.StringWriter

	// AddWriter adds one or more io.Writer destinations to the multi-writer.
	// All subsequent Write and WriteString operations will be broadcast to
	// these writers in addition to any previously added writers.
	// Nil writers are silently skipped.
	//
	// AddWriter is safe to call concurrently.
	AddWriter(w ...io.Writer)

	// Clean removes all registered writers from the multi-writer.
	// After Clean(), all writes will be discarded until new writers are added.
	//
	// Clean is safe to call concurrently.
	Clean()

	// SetInput sets or replaces the input source for read operations.
	// If i is nil, a DiscardCloser is used as the default.
	//
	// SetInput is safe to call concurrently, though note that the underlying
	// io.ReadCloser itself may not be safe for concurrent reads.
	SetInput(i io.Reader)

	// Stats returns current performance statistics for adaptive mode.
	// Returns metrics including write count, mean latency, and current mode.
	Stats() Stats

	// IsParallel reports whether the Multi is currently in parallel write mode.
	// Returns true if writes are executed concurrently to multiple destinations.
	IsParallel() bool

	// IsSequential reports whether the Multi is currently in sequential write mode.
	// Returns true if writes are executed sequentially to destinations.
	IsSequential() bool

	// IsAdaptive reports whether the Multi is in adaptive mode.
	// Returns true if the Multi automatically switches between sequential and
	// parallel modes based on observed write latency.
	IsAdaptive() bool

	// Reader returns the current input source.
	// Returns the io.ReadCloser set via SetInput, or DiscardCloser if none was set.
	Reader() io.ReadCloser

	// Writer returns the current write destination.
	// This is typically an io.MultiWriter wrapping all registered writers,
	// or io.Discard if no writers have been added.
	Writer() io.Writer

	// Copy copies data from the input source (Reader) to all registered
	// write destinations (Writer). It's a convenience wrapper around io.Copy.
	//
	// Returns the number of bytes copied and any error encountered.
	// The copy stops at EOF or the first error.
	Copy() (n int64, err error)
}

// New creates and initializes a new Multi instance with specified mode and configuration.
//
// Parameters:
//   - adaptive: if true, enables adaptive mode that switches between sequential and parallel
//     based on measured write latency. If false, uses sticky mode (sequential or parallel).
//   - parallel: initial write mode. If adaptive is false, this mode is permanent.
//     If adaptive is true, this is the starting mode.
//   - cfg: configuration for adaptive behavior (thresholds, sample sizes, etc.)
//
// The returned Multi is ready for use and safe for concurrent operations.
// It is initialized with:
//   - DiscardCloser as the default input source (reads return 0 bytes)
//   - io.Discard as the default output destination (writes are discarded)
//
// You should add writers via AddWriter and optionally set an input via
// SetInput before performing I/O operations.
//
// Example:
//
//	m := multi.New(false, false, multi.DefaultConfig())
//	var output bytes.Buffer
//	m.AddWriter(&output)
//	m.Write([]byte("hello"))
func New(adaptive, parallel bool, cfg Config) Multi {
	if cfg.SampleWrite <= 0 {
		cfg.SampleWrite = 100
	}
	if cfg.ThresholdLatency <= 0 {
		cfg.ThresholdLatency = 5000 // 5µs
	}
	if cfg.MinimalWriter <= 0 {
		cfg.MinimalWriter = 3
	}
	if cfg.MinimalSize <= 0 {
		cfg.MinimalSize = 512
	}

	m := &mlt{
		i:   libatm.NewValue[*readerWrapper](),
		d:   libatm.NewValue[*writeWrapper](),
		c:   new(atomic.Int64),
		w:   libatm.NewMapTyped[int64, io.Writer](),
		g:   cfg,
		adp: new(atomic.Bool),
		par: new(atomic.Bool),
		lst: new(atomic.Int64),
	}

	// Initialize with safe defaults to prevent panics
	m.i.Store(newReadWrapper(nil))
	m.d.Store(newWriteSeq(0, nil))

	m.adp.Store(adaptive)
	m.par.Store(parallel)

	return m
}
