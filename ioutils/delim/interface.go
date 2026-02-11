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
	"encoding/binary"
	"io"
	"math"
	"unicode/utf8"

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

	// UnRead returns the data currently buffered in the internal buffer
	// that has not yet been read by any Read operation.
	//
	// Warning: This consumes the data from the buffer. The data returned will
	// not be available in subsequent Read calls.
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
//     buffer size (32KB) is used. For better performance with large data chunks,
//     consider using larger buffer sizes (e.g., 64*libsiz.SizeKilo or libsiz.SizeMega).
//   - discardOverflow: If true, when the buffer is full and no delimiter is found,
//     the buffer content is discarded until a delimiter is found or EOF. If false,
//     ErrBufferFull is returned when the buffer is full and no delimiter is found.
//
// The returned BufferDelim must be closed when done to properly release resources
// and close the underlying reader.
//
// Supported delimiters include all ASCII characters (0-127) and extended ASCII (128-255):
//   - '\n' (newline), '\r' (carriage return), '\t' (tab)
//   - ',', '|', ';', ':', ' ' (common separators)
//   - '\x00' (null byte for C-style strings)
//   - Any single-byte character in range 0-255
//
// Example:
//
//	// Using default buffer size
//	bd := delim.New(file, '\n', 0, false)
//	defer bd.Close()
//
//	// Using custom buffer size (64KB) with overflow discard enabled
//	bd := delim.New(file, ',', 64*libsiz.SizeKilo, true)
//	defer bd.Close()
//
// See also: github.com/nabbar/golib/size package for convenient size constants.
func New(r io.ReadCloser, delim rune, sizeBuffer libsiz.Size, discardOverflow bool) BufferDelim {
	if sizeBuffer < 1 {
		sizeBuffer = 32 * libsiz.SizeKilo
	}

	var b = sizeBuffer.Uint64()

	if i := uint64(math.MaxInt/2) - 1; i < b {
		sizeBuffer = libsiz.ParseUint64(i)
	}

	if int32(delim) < 0 {
		return nil
	}

	v := make([]byte, utf8.UTFMax)
	binary.BigEndian.PutUint32(v, uint32(delim))

	return &dlm{
		i: r,
		r: v[len(v)-1],
		b: make([]byte, 0, sizeBuffer.Int()*2),
		s: sizeBuffer,
		d: discardOverflow,
	}
}
