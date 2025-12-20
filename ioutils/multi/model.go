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

// mlt is the concrete implementation of the Multi interface.
// It uses atomic operations and sync.Map for thread-safe concurrent access.
//
// Fields:
//   - i: atomic.Value storing a *readerWrapper (input source)
//   - d: atomic.Value storing an io.Writer (output destinations via io.MultiWriter)
//   - c: atomic counter for generating unique writer keys
//   - w: sync.Map storing registered writers (key: int64, value: io.Writer)
//
// The use of atomic.Value requires consistent types. To achieve this:
//   - Input readers are always wrapped in readerWrapper
//   - Output writers always use io.MultiWriter, even for single writers or io.Discard
type mlt struct {
	i libatm.Value[*readerWrapper]      // Input reader (stored as *readerWrapper)
	d libatm.Value[*writeWrapper]       // Output writer (stored as io.Writer from io.MultiWriter)
	c *atomic.Int64                     // Counter for writer keys
	w libatm.MapTyped[int64, io.Writer] // Map of registered writers
	g Config                            // Adaptive configuration parameters

	// Performance tracking for adaptive mode
	adp *atomic.Bool  // False=sequential/parallel sticky, True=adaptive mode
	par *atomic.Bool  // False=sequential, True=parallel
	lst *atomic.Int64 // Last known writer count
}

// AddWriter adds one or more io.Writer destinations to the multi-writer.
// Nil writers are silently skipped. Thread-safe for concurrent use.
func (o *mlt) AddWriter(w ...io.Writer) {
	for _, wrt := range w {
		if wrt != nil {
			o.w.Store(o.c.Add(1), wrt)
		}
	}

	o.update()
}

// Clean removes all registered writers from the multi-writer.
// After Clean, writes will be discarded until new writers are added.
func (o *mlt) Clean() {
	o.w.Range(func(k int64, v io.Writer) bool {
		o.w.Delete(k)
		return true
	})
	o.update()
}

// update rebuilds the internal writer based on current mode and registered writers.
// Creates either a sequential or parallel writer wrapper depending on the current mode.
func (o *mlt) update() {
	var (
		l = make([]io.Writer, 0)
		c = int64(0)
	)
	o.w.Range(func(k int64, v io.Writer) bool {
		if v != nil {
			c++
			l = append(l, v)
		}
		return true
	})

	if o.par.Load() {
		o.d.Store(newWritePar(int64(o.g.SampleWrite), o.check, o.g.MinimalSize, l...))
	} else {
		o.d.Store(newWriteSeq(int64(o.g.SampleWrite), o.check, l...))
	}

	o.lst.Store(c)
}

// check evaluates write latency and switches between sequential and parallel modes
// in adaptive mode. Only active when adaptive mode is enabled.
func (o *mlt) check(lat int64) {
	if !o.adp.Load() {
		return
	}

	var (
		p = o.par.Load()
		c = o.lst.Load()
	)

	if p {
		if lat < o.g.ThresholdLatency {
			o.par.Store(false)
			o.update()
		}
	} else if c >= int64(o.g.MinimalWriter) {
		if lat > o.g.ThresholdLatency {
			o.par.Store(true)
			o.update()
		}
	}
}

// SetInput sets or replaces the input source for read operations.
// If i is nil, a DiscardCloser is used as default. Closes the previous reader if any.
func (o *mlt) SetInput(i io.Reader) {
	// Wrap in readerWrapper to maintain consistent type in atomic.Value
	l := o.i.Swap(newReadWrapper(i))

	if l == nil {
		return
	}

	_ = l.Close()
}

// IsParallel reports whether the Multi is currently in parallel write mode.
func (o *mlt) IsParallel() bool {
	return o.par.Load()
}

// IsSequential reports whether the Multi is currently in sequential write mode.
func (o *mlt) IsSequential() bool {
	return !o.par.Load()
}

// IsAdaptive reports whether the Multi is in adaptive mode.
func (o *mlt) IsAdaptive() bool {
	return o.adp.Load()
}

// Writer returns the current write destination wrapper.
func (o *mlt) Writer() io.Writer {
	return o.d.Load()
}

// Reader returns the current input source reader.
func (o *mlt) Reader() io.ReadCloser {
	return o.i.Load()
}

// Copy copies data from the input source to all registered writers.
// Returns the number of bytes copied and any error encountered.
func (o *mlt) Copy() (n int64, err error) {
	return io.Copy(o.Writer(), o.Reader())
}

// Read reads data from the input source into p.
// Implements io.Reader interface.
func (o *mlt) Read(p []byte) (n int, err error) {
	return o.i.Load().Read(p)
}

// Write writes data to all registered writers.
// Implements io.Writer interface.
func (o *mlt) Write(p []byte) (n int, err error) {
	return o.d.Load().Write(p)
}

// WriteString writes a string to all registered writers.
// Implements io.StringWriter interface.
func (o *mlt) WriteString(s string) (n int, err error) {
	return o.Write([]byte(s))
}

// Close closes the input reader and cleans up all writers.
// Implements io.Closer interface.
func (o *mlt) Close() error {
	e := o.i.Load().Close()
	o.d.Store(newWriteSeq(0, nil))
	o.w.Range(func(k int64, v io.Writer) bool {
		o.w.Delete(k)
		return true
	})
	o.c.Store(0)
	return e
}
