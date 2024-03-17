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

package mux

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrInvalidInstance = errors.New("invalid de-multiplexer")
	ErrInvalidChannel  = errors.New("invalid channel key")
)

type Multiplexer interface {
	// NewChannel returns an io.Writer that can be used to mux the given byte slice with other channel into the main writer
	//
	// key is a rune/byte parameter to identify the channel.
	// io.Writer is the return type.
	NewChannel(key rune) io.Writer
}

// NewMultiplexer creates a new Multiplexer with the given io.Writer and delimiter.
//
// w: io.Writer to write to.
// delim: byte delimiter to use.
// Returns a Multiplexer.
func NewMultiplexer(w io.Writer, delim byte) Multiplexer {
	return &mux{
		w: w,
		d: delim,
	}
}

type DeMultiplexer interface {
	io.Reader

	// Copy launch a continuously reading from the main io.Reader and writing to the correct channel io.Writer
	// A good use case is to use it in a goroutine.
	//
	// returns any error.
	Copy() error

	// NewChannel registers a new channel to the given io.Writer for the given key as discriminant
	//
	// Parameters:
	//   key is a rune/byte parameter to identify the channel.
	//   w io.Writer the io.writer to write to.
	NewChannel(key rune, w io.Writer)
}

// NewDeMultiplexer creates a new DeMultiplexer using the given io.Reader, delimiter byte, and optional buffer size.
// If the buffer size is 0, a default buffer size of 4096 will be used.
//
// Parameters:
//
//	r io.Reader - the input reader
//	delim byte - the delimiter byte used to identify each block of data
//	size int - the buffer size
//
// Return type: DeMultiplexer
func NewDeMultiplexer(r io.Reader, delim byte, size int) DeMultiplexer {
	var rb *bufio.Reader

	if size > 0 {
		rb = bufio.NewReaderSize(r, size)
	} else {
		rb = bufio.NewReader(r)
	}

	return &dmux{
		r: rb,
		d: delim,
		m: make(map[rune]io.Writer),
	}
}
