/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

// Package randRead provides a buffered io.ReadCloser for reading data from remote sources
// with automatic reconnection and error handling.
//
// This package enables reading from any remote source (HTTP, gRPC, etc.) with transparent
// connection management, buffering, and automatic reconnection on failures.
//
// Features:
//   - Buffered reading from remote sources
//   - Automatic reconnection on errors
//   - Thread-safe atomic operations
//   - Standard io.ReadCloser interface
//   - Flexible remote source support
//
// The remote source is provided as a function that returns an io.ReadCloser. This function
// is called on initialization and whenever reconnection is needed.
//
// Example usage:
//
//	import (
//	    "net/http"
//	    "github.com/nabbar/golib/encoding/randRead"
//	)
//
//	// Define remote source function
//	source := func() (io.ReadCloser, error) {
//	    resp, err := http.Get("https://random-service.example.com/bytes")
//	    if err != nil {
//	        return nil, err
//	    }
//	    return resp.Body, nil
//	}
//
//	// Create reader
//	reader := randRead.New(source)
//	defer reader.Close()
//
//	// Read random data
//	buffer := make([]byte, 100)
//	n, err := reader.Read(buffer)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Automatic reconnection:
//
//	When a read error occurs, the reader automatically calls the source function
//	to establish a new connection and continues reading. This is transparent to
//	the caller.
//
// Thread safety:
//
//	Uses atomic.Value for buffer and connection management, making it safe for
//	concurrent read operations.
//
// Use cases:
//   - Reading random data from remote services
//   - Streaming data from HTTP endpoints with reconnection
//   - gRPC stream consumers with automatic recovery
//   - Resilient data fetching from unreliable sources
package randRead

import (
	"io"
	"sync/atomic"
)

// FuncRemote is a function type that provides a remote data source.
//
// It should return an io.ReadCloser that provides the data stream and an error
// if the connection cannot be established.
//
// This function is called:
//   - On initial reader creation
//   - When reconnection is needed after errors
//
// Example implementations:
//   - HTTP GET request returning response body
//   - gRPC stream wrapper
//   - Network socket connection
//   - File handle for testing
type FuncRemote func() (io.ReadCloser, error)

// New creates a new buffered reader from a remote source.
//
// The reader will buffer data from the remote source and automatically reconnect
// if errors occur during reading. The remote source is accessed via the provided
// function, which is called on initialization and for reconnection.
//
// Parameters:
//   - fct: Function that returns io.ReadCloser and error
//
// Returns:
//   - io.ReadCloser: Buffered reader with automatic reconnection
//   - Returns nil if fct is nil
//
// The returned reader implements io.ReadCloser and should be closed when done
// to release resources.
//
// Example:
//
//	source := func() (io.ReadCloser, error) {
//	    resp, err := http.Get("https://example.com/data")
//	    if err != nil {
//	        return nil, err
//	    }
//	    return resp.Body, nil
//	}
//
//	reader := randRead.New(source)
//	if reader == nil {
//	    log.Fatal("Failed to create reader")
//	}
//	defer reader.Close()
//
//	data := make([]byte, 1024)
//	n, err := reader.Read(data)
func New(fct FuncRemote) io.ReadCloser {
	if fct == nil {
		return nil
	}

	f := new(atomic.Value)
	f.Store(fct)

	r := &remote{
		r: new(atomic.Value),
		f: f,
	}

	return &prnd{
		b: new(atomic.Value),
		r: r,
	}
}
