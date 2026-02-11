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

package delim

import (
	"bytes"
	"errors"
	"io"
)

// Reader returns the BufferDelim itself as an io.ReadCloser.
// This allows the BufferDelim to be used wherever an io.ReadCloser is expected.
//
// The returned reader respects the delimiter configuration and will read data
// in delimiter-separated chunks when using the Read method.
func (o *dlm) Reader() io.ReadCloser {
	return o
}

// Copy reads data from the BufferDelim and writes it to w until EOF or an error occurs.
// It returns the total number of bytes written and any write error encountered.
//
// Copy is a convenience method that delegates to WriteTo(w).
// The data is read and written in chunks delimited by the configured delimiter character.
// Each chunk includes the delimiter in the written data.
//
// Returns:
//   - n: Total number of bytes successfully written to w
//   - err: The first error encountered (io.EOF when all data has been read and written)
//
// Example:
//
//	bd := delim.New(inputFile, '\n', 0, false)
//	defer bd.Close()
//	written, err := bd.Copy(outputFile)
//	if err != nil && err != io.EOF {
//	    // Handle error
//	}
func (o *dlm) Copy(w io.Writer) (n int64, err error) {
	return o.WriteTo(w)
}

// Read reads data into p.
// It implements the io.Reader interface.
//
// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p))
// and any error encountered.
//
// Note: Unlike ReadBytes, this method does not stop at the delimiter. It acts as a
// standard buffered reader, filling p with available data.
//
// Returns:
//   - n: Number of bytes read
//   - err: Any error encountered (io.EOF when end of stream is reached, ErrInstance if closed)
//
// Behavior:
//   - Reads available data from the buffer and underlying reader
//   - Does NOT guarantee stopping at a delimiter
//   - If the instance is closed or invalid, returns ErrInstance
//
// Example:
//
//	buf := make([]byte, 100)
//	n, err := bd.Read(buf)
//	if err != nil && err != io.EOF {
//	    log.Fatal(err)
//	}
//	data := buf[:n]  // data contains up to 100 bytes from stream
func (o *dlm) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInstance
	}

	return o.readBuf(p)
}

// UnRead returns the data currently buffered in the internal reader that has not yet been consumed.
//
// Warning: This consumes the data from the buffer. The data returned will
// not be available in subsequent Read calls.
// The returned data represents what has been read into the buffer but not yet returned
// by Read or ReadBytes operations.
//
// Returns:
//   - []byte: The buffered data, or nil if no data is buffered
//   - error: ErrInstance if the BufferDelim is closed or invalid, nil otherwise
//
// Note: Calling UnRead will consume the buffered data, so subsequent UnRead calls
// will return different data (or nil) unless more data has been buffered.
//
// Example:
//
//	// Get buffered data
//	buffered, err := bd.UnRead()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(buffered) > 0 {
//	    fmt.Printf("Buffered %d bytes: %s\n", len(buffered), buffered)
//	}
func (o *dlm) UnRead() ([]byte, error) {
	if o == nil {
		return nil, ErrInstance
	}

	return o.unReadBuf()
}

// ReadBytes reads until the first occurrence of the delimiter in the input,
// returning a slice containing the data up to and including the delimiter.
//
// This operates on the internal buffer to find and return the next chunk of data ending with the delimiter.
// with the configured delimiter.
//
// Returns:
//   - []byte: A slice containing the data read, including the delimiter if found
//   - error: io.EOF if end of stream reached, ErrInstance if closed, or any read error
//
// Behavior:
//   - If the delimiter is found, returns all data up to and including it
//   - If EOF is reached before finding a delimiter, returns the remaining data with io.EOF
//   - Returns ErrInstance if the BufferDelim has been closed
//
// The returned slice points to internal buffer data that may be overwritten
// by subsequent reads. If you need to retain the data, make a copy.
//
// Example:
//
//	// Read lines from a file
//	for {
//	    line, err := bd.ReadBytes()
//	    if err == io.EOF {
//	        if len(line) > 0 {
//	            processLine(line)  // Process last line without delimiter
//	        }
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    processLine(line)  // line includes '\n'
//	}
func (o *dlm) ReadBytes() ([]byte, error) {
	if o == nil {
		return nil, ErrInstance
	}

	o.m.Lock()
	defer o.m.Unlock()

	var (
		err error
		idx int         // index
		mkr = o.r       // marker
		res []byte      // result
		szm = o.s.Int() // size max
	)

	for {
		if len(o.b) > 0 {
			idx = bytes.IndexByte(o.b, mkr)

			if idx >= 0 {
				// Delimiter found in buffer
				needed := idx + 1

				if len(res)+needed > szm {
					// Overflow handling when delimiter is found but total size exceeds limit
					take := szm - len(res)
					if take > 0 {
						res = append(res, o.b[:take]...)
						o.b = o.b[take:]
					}

					if o.d {
						if err = o.discard(mkr); err == nil {
							res[len(res)-1] = o.r
							return res, nil
						}
						return res, err
					}
					return res, ErrBufferFull
				}

				// Zero-copy path: if we haven't accumulated anything yet,
				// return slice directly from internal buffer
				if len(res) == 0 {
					res := o.b[:needed]
					o.b = o.b[needed:]
					return res, nil
				}

				// Accumulation path: append to existing buffer
				res = append(res, o.b[:needed]...)
				o.b = o.b[needed:]
				return res, nil
			}

			// Delimiter not found in buffer
			if len(res)+len(o.b) > szm {
				// Overflow handling when delimiter is NOT found
				take := szm - len(res)
				if take > 0 {
					res = append(res, o.b[:take]...)
					o.b = o.b[take:]
				}

				if o.d {
					if err = o.discard(mkr); err == nil {
						res[len(res)-1] = o.r
						return res, nil
					}
					return res, err
				}
				return res, ErrBufferFull
			}

			res = append(res, o.b...)
			o.b = o.b[:0]
		}

		err = o.fill()
		if err != nil {
			if len(o.b) > 0 {
				continue
			}
			if len(res) > 0 && errors.Is(err, io.EOF) {
				return res, io.EOF
			}
			return res, err
		}
	}
}

// Close closes the BufferDelim and releases associated resources.
// It implements the io.Closer interface.
//
// Close performs the following operations:
//  1. Resets the internal buffer
//  2. Closes the underlying io.ReadCloser
//
// After Close is called, all subsequent operations on the BufferDelim will return ErrInstance.
// It is safe to call Close multiple times, though subsequent calls after the first will
// return the error from closing the already-closed underlying reader.
//
// Returns:
//   - error: Any error from closing the underlying reader, or nil on success
//
// Example:
//
//	bd := delim.New(file, '\n', 0)
//	defer bd.Close()  // Ensure resources are released
//
//	// Use bd...
func (o *dlm) Close() error {
	o.m.Lock()
	defer o.m.Unlock()

	o.b = o.b[:0]
	return o.i.Close()
}

// WriteTo reads data from the BufferDelim and writes it to w until EOF or an error occurs.
// It implements the io.WriterTo interface.
//
// WriteTo reads the input in delimiter-separated chunks and writes each chunk (including
// the delimiter) to w. This continues until the end of the input stream is reached.
//
// The method handles both read and write errors appropriately:
//   - If a write error occurs, it stops immediately and returns the write error
//   - If a read error occurs (including io.EOF), it returns that error after writing any buffered data
//
// Returns:
//   - n: Total number of bytes written to w
//   - err: io.EOF when all data has been successfully written, or the first error encountered
//
// Memory management:
//   - Each chunk is cleared after writing to help the garbage collector
//   - The method allocates minimal additional memory beyond the internal buffer
//
// Example:
//
//	// Copy all data from input to output, respecting delimiters
//	bd := delim.New(inputFile, '\n', 64*size.KiB)
//	defer bd.Close()
//
//	written, err := bd.WriteTo(outputFile)
//	if err != nil && err != io.EOF {
//	    log.Fatalf("Failed after writing %r bytes: %v", written, err)
//	}
//	fmt.Printf("Successfully wrote %r bytes\n", written)
func (o *dlm) WriteTo(w io.Writer) (n int64, err error) {
	var (
		e error
		i int
		b []byte
	)

	if o == nil {
		return 0, ErrInstance
	}

	for err == nil {
		b, err = o.ReadBytes()

		if len(b) > 0 {
			i, e = w.Write(b)
			n += int64(i)
		}

		b = b[:0] // nolint

		if err == nil && e != nil {
			err = e
		}
	}

	return n, err
}

func (o *dlm) unReadBuf() ([]byte, error) {
	o.m.Lock()
	defer o.m.Unlock()

	if len(o.b) > 0 {
		res := make([]byte, len(o.b))
		copy(res, o.b)

		o.b = o.b[:0]

		return res, nil
	}

	return nil, nil
}

func (o *dlm) readBuf(p []byte) (n int, err error) {
	o.m.Lock()
	defer o.m.Unlock()

	var (
		mxp = len(p)
		rst = mxp
		pos int
	)

	if mxp == 0 {
		return 0, nil
	}

	for {
		if len(o.b) > 0 {
			if len(o.b) < rst {
				copy(p[pos:], o.b)

				pos += len(o.b)
				rst -= len(o.b)

				o.b = o.b[:0]
			} else {
				copy(p[pos:], o.b[:rst])
				o.b = o.b[rst:]

				pos += rst
				rst = 0
			}
		}

		if rst < 1 {
			return pos, err
		} else if err != nil {
			return pos, err
		}

		err = o.fill()
	}
}

func (o *dlm) fill() error {
	var (
		nbr int
		err error
		req = o.s.Int()
	)

	if cap(o.b) < req {
		newBuf := make([]byte, len(o.b), req)
		copy(newBuf, o.b)
		o.b = newBuf
	}

	start := len(o.b)
	if cap(o.b) >= start+req {
		o.b = o.b[:start+req]
	} else {
		// Should not happen with check above, but safe fallback
		o.b = o.b[:cap(o.b)]
	}

	nbr, err = o.i.Read(o.b[start:])

	if nbr > 0 {
		o.b = o.b[:start+nbr]
	} else {
		o.b = o.b[:start]
		if err == nil {
			return io.EOF
		}
	}

	return err
}

func (o *dlm) discard(dlm byte) error {
	var (
		err error
		idx int
	)

	for {
		if len(o.b) > 0 {
			idx = bytes.IndexByte(o.b, dlm)
			if idx >= 0 {
				o.b = o.b[idx+1:]
				return nil
			}
			o.b = o.b[:0]
		}

		err = o.fill()
		if err != nil {
			if len(o.b) > 0 {
				continue
			}
			return err
		}
	}
}
