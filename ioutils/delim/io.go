/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

import "io"

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
//	bd := delim.New(inputFile, '\n', 0)
//	defer bd.Close()
//	written, err := bd.Copy(outputFile)
//	if err != nil && err != io.EOF {
//	    log.Fatal(err)
//	}
func (o *dlm) Copy(w io.Writer) (n int64, err error) {
	return o.WriteTo(w)
}

// Read reads data up to and including the next delimiter into p.
// It implements the io.Reader interface.
//
// Read returns the number of bytes read into p and any error encountered.
// The data includes the delimiter character if one was found.
//
// If the buffer p is too small to hold the delimited chunk, Read will expand p
// to accommodate the data. Therefore, callers should not rely on p's capacity
// remaining unchanged after the call.
//
// Returns:
//   - n: Number of bytes read (including the delimiter if present)
//   - err: Any error encountered (io.EOF when end of stream is reached, ErrInstance if closed)
//
// Behavior:
//   - If a delimiter is found, returns the data up to and including it
//   - If EOF is reached before a delimiter, returns remaining data with io.EOF
//   - If the instance is closed or invalid, returns ErrInstance
//
// Example:
//
//	buf := make([]byte, 100)
//	n, err := bd.Read(buf)
//	if err != nil && err != io.EOF {
//	    log.Fatal(err)
//	}
//	data := buf[:n]  // data includes the delimiter
func (o *dlm) Read(p []byte) (n int, err error) {
	if o == nil || o.r == nil {
		return 0, ErrInstance
	}

	b, e := o.r.ReadBytes(byte(o.d))

	if len(b) > 0 {
		if cap(p) < len(b) {
			p = append(p, make([]byte, len(b)-len(p))...)
		}
		copy(p, b)
	}

	return len(b), e
}

// UnRead returns the data currently buffered in the internal reader that has not yet been consumed.
//
// This method is useful for peeking at upcoming data without consuming it from the stream.
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
//	// Peek at buffered data without fully reading a delimited chunk
//	buffered, err := bd.UnRead()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(buffered) > 0 {
//	    fmt.Printf("Next %d bytes: %s\n", len(buffered), buffered)
//	}
func (o *dlm) UnRead() ([]byte, error) {
	if o == nil || o.r == nil {
		return nil, ErrInstance
	}

	if s := o.r.Buffered(); s > 0 {
		b := make([]byte, s)
		_, e := o.r.Read(b)
		return b, e
	}

	return nil, nil
}

// ReadBytes reads until the first occurrence of the delimiter in the input,
// returning a slice containing the data up to and including the delimiter.
//
// This is similar to bufio.Reader.ReadBytes but operates on the wrapped reader
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
	if o.r == nil {
		return nil, ErrInstance
	}

	return o.r.ReadBytes(byte(o.d))
}

// Close closes the BufferDelim and releases associated resources.
// It implements the io.Closer interface.
//
// Close performs the following operations:
//  1. Resets the internal bufio.Reader to nil
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
	o.r.Reset(nil)
	o.r = nil

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
//	    log.Fatalf("Failed after writing %d bytes: %v", written, err)
//	}
//	fmt.Printf("Successfully wrote %d bytes\n", written)
func (o *dlm) WriteTo(w io.Writer) (n int64, err error) {
	var (
		e error
		i int
		b []byte
		s int

		d = o.getDelimByte()
	)

	if o.r == nil {
		return 0, ErrInstance
	}

	for err == nil {
		b, err = o.r.ReadBytes(d)
		s = len(b)

		if s > 0 {
			i, e = w.Write(b)
			n += int64(i)
		}

		clear(b)
		b = b[:0] // nolint
		b = nil

		if err == nil && e != nil {
			err = e
		}
	}

	return n, err
}
