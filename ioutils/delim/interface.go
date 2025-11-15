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

import (
	"bufio"
	"io"

	libsiz "github.com/nabbar/golib/size"
)

// BufferDelim is an interface that extends io.ReadCloser and io.WriterTo with additional
// methods for reading delimited data from an input stream.
//
// It provides functionality to:
//   - Read data until a delimiter is encountered (Read, ReadBytes)
//   - Access buffered but unread data (UnRead)
//   - Copy data to a writer while respecting delimiters (WriteTo, Copy)
//   - Retrieve the current delimiter character (Delim)
//   - Obtain the reader as an io.ReadCloser (Reader)
//
// All read operations will include the delimiter character in the returned data.
// When EOF is reached, the methods return io.EOF error along with any remaining data.
//
// After Close() is called, all subsequent operations will return ErrInstance.
type BufferDelim interface {
	io.ReadCloser
	io.WriterTo

	// Delim returns the delimiter rune used to separate data chunks.
	Delim() rune

	// Reader returns the BufferDelim itself as an io.ReadCloser.
	// This is useful when you need to pass the delimited reader to functions
	// expecting a standard io.ReadCloser interface.
	Reader() io.ReadCloser

	// Copy reads from the BufferDelim and writes to w until EOF or an error occurs.
	// It returns the number of bytes written and any error encountered.
	// This is equivalent to calling WriteTo(w).
	//
	// The data is read in chunks delimited by the delimiter character,
	// and each chunk (including the delimiter) is written to w.
	Copy(w io.Writer) (n int64, err error)

	// ReadBytes reads until the first occurrence of the delimiter in the input,
	// returning a slice containing the data up to and including the delimiter.
	// If ReadBytes encounters an error before finding a delimiter, it returns
	// the data read before the error and the error itself (often io.EOF).
	//
	// Returns ErrInstance if the BufferDelim has been closed.
	ReadBytes() ([]byte, error)

	// UnRead returns the data currently buffered in the internal bufio.Reader
	// that has not yet been read by any Read operation.
	//
	// This is useful for peeking at upcoming data without consuming it.
	// Returns nil if no data is buffered, or ErrInstance if the BufferDelim has been closed.
	UnRead() ([]byte, error)
}

// New creates a new BufferDelim that reads from r, using delim as the delimiter character.
//
// Parameters:
//   - r: The io.ReadCloser to read data from. This will be wrapped with buffering.
//   - delim: The rune character used as delimiter. Common delimiters include:
//     '\n' for newlines, ',' for CSV, '|' for pipes, '\t' for tabs, or any custom character.
//   - sizeBufferRead: The size of the internal buffer. If 0 or negative, the default
//     bufio.Reader buffer size (4096 bytes) is used. For better performance with large
//     data chunks, consider using larger buffer sizes (e.g., 64*libsiz.KiB or libsiz.MiB).
//
// The returned BufferDelim must be closed when done to properly release resources
// and close the underlying reader.
//
// Example:
//
//	// Using default buffer size
//	bd := delim.New(file, '\n', 0)
//	defer bd.Close()
//
//	// Using custom buffer size (64KB)
//	bd := delim.New(file, ',', 64*libsiz.KiB)
//	defer bd.Close()
//
// See also: github.com/nabbar/golib/size package for convenient size constants.
func New(r io.ReadCloser, delim rune, sizeBufferRead libsiz.Size) BufferDelim {
	var b *bufio.Reader

	if sizeBufferRead > 0 {
		b = bufio.NewReaderSize(r, sizeBufferRead.Int())
	} else {
		b = bufio.NewReader(r)
	}

	return &dlm{
		i: r,
		r: b,
		d: delim,
	}
}
