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

package encoding

import (
	"io"
)

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
