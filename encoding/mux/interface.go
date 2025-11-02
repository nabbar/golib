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

// Package mux provides thread-safe multiplexing and demultiplexing for routing multiple
// logical channels over a single I/O stream.
//
// This package enables multiple writers to share a single io.Writer (multiplexing) and
// a single reader to route data to multiple io.Writers (demultiplexing). It uses CBOR
// encoding for message structure and hexadecimal encoding for payload safety.
//
// Features:
//   - Multiple logical channels over single physical stream
//   - Rune-based channel identification (Unicode keys)
//   - Delimiter-based message framing
//   - Thread-safe concurrent access
//   - Lock-free demultiplexing with sync.Map
//   - CBOR encoding for compact message structure
//   - Hexadecimal payload encoding for safety
//
// Message format:
//
//	CBOR{K: rune, D: hex_string} + delimiter
//
// Example usage (Multiplexer):
//
//	import encmux "github.com/nabbar/golib/encoding/mux"
//
//	// Create multiplexer
//	mux := encmux.NewMultiplexer(outputStream, '\n')
//
//	// Create channels
//	channelA := mux.NewChannel('a')
//	channelB := mux.NewChannel('b')
//
//	// Write to channels (thread-safe)
//	channelA.Write([]byte("Message A"))
//	channelB.Write([]byte("Message B"))
//
// Example usage (DeMultiplexer):
//
//	import encmux "github.com/nabbar/golib/encoding/mux"
//
//	// Create demultiplexer
//	demux := encmux.NewDeMultiplexer(inputStream, '\n', 4096)
//
//	// Register output channels
//	demux.NewChannel('a', outputA)
//	demux.NewChannel('b', outputB)
//
//	// Start routing (blocks until EOF)
//	err := demux.Copy()
//
// Thread safety:
//   - Multiplexer uses mutex for atomic message writes
//   - DeMultiplexer uses sync.Map for lock-free channel lookup
//   - Safe for concurrent goroutine access
//
// Use cases:
//   - Network protocol multiplexing
//   - Log aggregation from multiple sources
//   - Test output routing by category
//   - Multi-channel communication over single connection
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

	// NewChannel returns a new io.Writer that writes to the multiplexer.
	// The returned io.Writer is a channel associated with the given key.
	//
	// Parameters:
	//   key: rune/byte parameter to identify the channel.
	//
	// Returns:
	//   io.Writer: the io.Writer associated with the channel.
	NewChannel(key rune) io.Writer
}

// NewMultiplexer creates a new Multiplexer using the given io.Writer and delimiter byte.
//
// Parameters:
//
//	w io.Writer - the output writer
//	delim byte - the delimiter byte used to identify each block of data
//
// Returns:
//
//	Multiplexer: the newly created Multiplexer
func NewMultiplexer(w io.Writer, delim byte) Multiplexer {
	return &mux{
		w: w,
		d: delim,
	}
}

type DeMultiplexer interface {
	io.Reader

	// Copy continuously reads data from the demultiplexer and writes it to the
	// corresponding channels until an error occurs, which is then returned.
	//
	// If the error is io.EOF, it is ignored and the function will return nil.
	//
	// Parameters: None
	//
	// Returns:
	//   error: an error if any occurred during the copy process.
	Copy() error

	// NewChannel adds a new channel to the demultiplexer by mapping the given key to the provided writer.
	//
	// Parameters:
	//   key: rune/byte parameter to identify the channel
	//   w: io.Writer parameter to write the channel data to
	NewChannel(key rune, w io.Writer)
}

// NewDeMultiplexer creates a new DeMultiplexer using the given io.Reader and delimiter byte.
//
// Parameters:
//
//	r io.Reader - the input reader
//	delim byte - the delimiter byte used to identify each block of data
//	size int - the size of the buffered reader, 0 for no buffering
//
// Returns:
//
//	DeMultiplexer: the newly created DeMultiplexer
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
		// m is a sync.Map, no initialization needed
	}
}
