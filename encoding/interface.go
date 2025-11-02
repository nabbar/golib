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

// Package encoding provides a unified Coder interface for encoding and decoding operations.
//
// This package defines the Coder interface which is implemented by various sub-packages
// for different encoding/decoding operations including encryption, hashing, hex encoding,
// multiplexing, and more.
//
// Sub-packages:
//   - aes: AES-256-GCM authenticated encryption
//   - hexa: Hexadecimal encoding and decoding
//   - mux: Multiplexing/demultiplexing for multi-channel communication
//   - randRead: Buffered random data reader from remote sources
//   - sha256: SHA-256 cryptographic hashing
//
// The Coder interface provides:
//   - Direct byte slice encoding/decoding
//   - Streaming operations via io.Reader and io.Writer
//   - State management with Reset()
//
// Example usage:
//
//	import (
//	    enchex "github.com/nabbar/golib/encoding/hexa"
//	    encsha "github.com/nabbar/golib/encoding/sha256"
//	)
//
//	// Hex encoding
//	hexCoder := enchex.New()
//	encoded := hexCoder.Encode([]byte("Hello"))
//	decoded, _ := hexCoder.Decode(encoded)
//
//	// SHA-256 hashing
//	hasher := encsha.New()
//	hash := hasher.Encode([]byte("data"))
//
// All implementations follow the same interface pattern, making it easy to swap
// between different encoding schemes.
package encoding

import (
	"io"
)

// Coder is the unified interface for encoding and decoding operations.
//
// This interface is implemented by all encoding sub-packages (aes, hexa, mux, sha256)
// to provide a consistent API for encoding/decoding operations with both direct byte
// slice manipulation and streaming I/O support.
//
// Implementations:
//   - aes.New(): AES-256-GCM encryption/decryption
//   - hexa.New(): Hexadecimal encoding/decoding
//   - sha256.New(): SHA-256 hashing (Decode not applicable)
//   - mux.NewChannel(): Channel writer (multiplexing)
//
// Thread safety depends on the implementation. Refer to specific sub-package
// documentation for concurrency guarantees.
type Coder interface {
	// Encode encodes the given byte slice.
	//
	// Parameter(s): p []byte
	// Return type(s): []byte
	Encode(p []byte) []byte

	// Decode decodes the given byte slice and returns the decoded byte slice and an error if any.
	//
	// Parameters:
	// - p: The byte slice to be decoded.
	//
	// Returns:
	// - []byte: The decoded byte slice.
	// - error: An error if any occurred during decoding.
	Decode(p []byte) ([]byte, error)

	// EncodeReader return a io.Reader that can be used to encode the given byte slice
	//
	// r io.Reader
	// io.Reader
	EncodeReader(r io.Reader) io.ReadCloser

	// DecodeReader return a io.Reader that can be used to decode the given byte slice
	//
	// r io.Reader
	// io.Reader
	DecodeReader(r io.Reader) io.ReadCloser

	// EncodeWriter return a io.writer that can be used to encode the given byte slice
	//
	// w io.Writer parameter.
	// io.Writer return type.
	EncodeWriter(w io.Writer) io.WriteCloser

	// DecodeWriter return a io.writer that can be used to decode the given byte slice
	//
	// w io.Writer parameter.
	// io.Writer return type.
	DecodeWriter(w io.Writer) io.WriteCloser

	// Reset will free memory
	Reset()
}
