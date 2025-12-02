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
	"sync"
	"sync/atomic"
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
	i *atomic.Value // Input reader (stored as *readerWrapper)
	d *atomic.Value // Output writer (stored as io.Writer from io.MultiWriter)
	c *atomic.Int64 // Counter for writer keys
	w sync.Map      // Map of registered writers
}

// readerWrapper wraps an io.ReadCloser to maintain type consistency in atomic.Value.
//
// atomic.Value requires that all stored values have the same concrete type.
// By wrapping all readers in readerWrapper, we ensure this constraint is met
// even when different io.ReadCloser implementations are used.
type readerWrapper struct {
	io.ReadCloser
}

// AddWriter adds one or more writers to the multiplexer.
// All subsequent write operations will be broadcast to these writers.
//
// Implementation notes:
//   - Nil writers are skipped
//   - Writers are stored with unique keys in a sync.Map
//   - A new io.MultiWriter is created encompassing all registered writers
//   - The MultiWriter is stored atomically to ensure thread-safe access
//
// Thread-safety: This method is safe for concurrent use.
func (o *mlt) AddWriter(w ...io.Writer) {
	for _, wrt := range w {
		if wrt != nil {
			o.w.Store(o.c.Add(1), wrt)
		}
	}

	var l = make([]io.Writer, 0)

	o.w.Range(func(key, value any) bool {
		if value != nil {
			if v, k := value.(io.Writer); k {
				l = append(l, v)
			}
		}
		return true
	})

	// Always use MultiWriter to maintain consistent type in atomic.Value.
	// This ensures atomic.Value never stores different concrete types,
	// which would cause a panic.
	if len(l) < 1 {
		o.d.Store(io.MultiWriter(io.Discard))
	} else {
		o.d.Store(io.MultiWriter(l...))
	}
}

// Clean removes all registered writers and resets to io.Discard.
// After calling Clean, all write operations will discard data until
// new writers are added via AddWriter.
//
// Implementation notes:
//   - Atomically sets the writer to io.MultiWriter(io.Discard)
//   - Collects all keys from sync.Map and deletes them
//   - Resets the writer counter to 0
//
// Thread-safety: This method is safe for concurrent use.
func (o *mlt) Clean() {
	o.d.Store(io.MultiWriter(io.Discard))

	var keys = make([]any, 0)

	o.w.Range(func(key, value any) bool {
		keys = append(keys, key)
		return true
	})

	for _, k := range keys {
		o.w.Delete(k)
	}

	o.c.Store(0)
}

// SetInput sets the input source for read operations.
// If i is nil, it defaults to DiscardCloser which returns 0 bytes on reads.
//
// Implementation notes:
//   - The input is wrapped in readerWrapper for type consistency
//   - The wrapped input is stored atomically
//   - Previous input is not automatically closed; caller must manage lifecycle
//
// Thread-safety: This method is safe for concurrent use, though the underlying
// io.ReadCloser itself may not support concurrent reads.
func (o *mlt) SetInput(i io.ReadCloser) {
	if o == nil {
		return
	} else if i == nil {
		i = DiscardCloser{}
	}

	// Wrap in readerWrapper to maintain consistent type in atomic.Value
	o.i.Store(&readerWrapper{ReadCloser: i})
}

// Writer returns the current output writer.
// This is typically an io.MultiWriter combining all registered writers,
// or io.MultiWriter(io.Discard) if no writers have been added.
//
// The returned Writer can be used directly for io.Copy or other operations.
// Changes to registered writers via AddWriter or Clean will not affect
// the returned Writer - call Writer() again to get the updated writer.
func (o *mlt) Writer() io.Writer {
	return o.d.Load().(io.Writer)
}

// Reader returns the current input source.
// Returns the io.ReadCloser set via SetInput, or DiscardCloser if none was set.
//
// The returned ReadCloser can be used directly for io.Copy or other operations.
func (o *mlt) Reader() io.ReadCloser {
	w := o.i.Load().(*readerWrapper)
	return w.ReadCloser
}

// Copy copies data from the input source to all registered writers.
// It's a convenience wrapper around io.Copy(o.Writer(), o.Reader()).
//
// Returns the number of bytes copied and any error encountered.
// The copy operation stops at EOF or the first write error.
//
// Note: If the underlying reader supports WriteTo or the writer supports
// ReadFrom, io.Copy may use those for optimization.
func (o *mlt) Copy() (n int64, err error) {
	return io.Copy(o.Writer(), o.Reader())
}

// Read implements io.Reader by reading from the current input source.
//
// Returns:
//   - The number of bytes read and any error from the underlying reader
//   - ErrInstance if the internal state is invalid (should not occur normally)
//
// The behavior matches the semantics of the underlying io.ReadCloser's Read method.
func (o *mlt) Read(p []byte) (n int, err error) {
	if i := o.i.Load(); i == nil {
		return 0, ErrInstance
	} else if w, ok := i.(*readerWrapper); !ok {
		return 0, ErrInstance
	} else if w.ReadCloser == nil {
		return 0, ErrInstance
	} else {
		return w.Read(p)
	}
}

// Write implements io.Writer by broadcasting data to all registered writers.
//
// Returns:
//   - The number of bytes written (matches len(p) on success)
//   - Any error from the underlying io.MultiWriter
//   - ErrInstance if the internal state is invalid (should not occur normally)
//
// All writes are performed atomically to the current set of writers.
// If any writer returns an error, the error is returned immediately.
func (o *mlt) Write(p []byte) (n int, err error) {
	if i := o.d.Load(); i == nil {
		return 0, ErrInstance
	} else if v, k := i.(io.Writer); !k {
		return 0, ErrInstance
	} else {
		return v.Write(p)
	}
}

// WriteString implements io.StringWriter by broadcasting the string to all registered writers.
//
// This method uses io.WriteString which may be more efficient than Write
// if the underlying writers implement io.StringWriter.
//
// Returns:
//   - The number of bytes written (matches len(s) on success)
//   - Any error from the underlying writers
//   - ErrInstance if the internal state is invalid (should not occur normally)
func (o *mlt) WriteString(s string) (n int, err error) {
	if i := o.d.Load(); i == nil {
		return 0, ErrInstance
	} else if v, k := i.(io.Writer); !k {
		return 0, ErrInstance
	} else {
		return io.WriteString(v, s)
	}
}

// Close implements io.Closer by closing the current input source.
//
// This method closes the reader set via SetInput. It does NOT close
// any of the registered writers - those remain open and usable.
//
// Returns:
//   - Any error from closing the underlying reader
//   - ErrInstance if the internal state is invalid (should not occur normally)
//   - nil if the close succeeds or if using DiscardCloser
func (o *mlt) Close() error {
	if i := o.i.Load(); i == nil {
		return ErrInstance
	} else if w, ok := i.(*readerWrapper); !ok {
		return ErrInstance
	} else if w.ReadCloser == nil {
		return ErrInstance
	} else {
		return w.Close()
	}
}
