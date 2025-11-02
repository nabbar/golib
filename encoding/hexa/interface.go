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

// Package hexa provides hexadecimal encoding and decoding with streaming I/O support.
//
// This package implements the encoding.Coder interface for consistent hex operations.
// It uses Go's standard encoding/hex package internally for RFC 4648 compliant encoding.
//
// Features:
//   - Standard hexadecimal encoding (0-9, a-f)
//   - Case-insensitive decoding (accepts both uppercase and lowercase)
//   - Streaming encoding/decoding via io.Reader interfaces
//   - Memory efficient operations
//   - Stateless and thread-safe
//   - Lossless round-trip encoding/decoding
//
// Encoding converts each byte to two hexadecimal characters:
//   - Input size N bytes → Output size 2N bytes
//   - Output format: lowercase hex (e.g., "48656c6c6f")
//   - Character set: 0-9, a-f
//
// Decoding converts hexadecimal strings back to binary:
//   - Input size 2N bytes → Output size N bytes
//   - Accepts uppercase, lowercase, or mixed case
//   - Rejects invalid hex characters or odd length
//
// Example usage:
//
//	import enchex "github.com/nabbar/golib/encoding/hexa"
//
//	// Create coder
//	coder := enchex.New()
//
//	// Encode binary to hex
//	plaintext := []byte("Hello")
//	hex := coder.Encode(plaintext)
//	fmt.Println(string(hex))  // Output: 48656c6c6f
//
//	// Decode hex to binary
//	decoded, err := coder.Decode(hex)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(decoded))  // Output: Hello
//
// Common use cases:
//   - Display binary data in readable format
//   - Store binary data in text files/databases
//   - Debug binary protocols
//   - Encode checksums and hashes
package hexa

import libenc "github.com/nabbar/golib/encoding"

// New creates a new hexadecimal coder instance.
//
// The returned coder implements the encoding.Coder interface and provides
// hexadecimal encoding and decoding functionality. The coder is stateless
// and safe for concurrent use.
//
// Returns:
//   - A new hexadecimal coder instance
//
// Example:
//
//	coder := enchex.New()
//	hex := coder.Encode([]byte("Hello"))
//	fmt.Println(string(hex))  // Output: 48656c6c6f
//
// Encoding format:
//   - Each byte becomes two hex characters
//   - Output is lowercase (0-9, a-f)
//   - No delimiters or spacing
//
// Decoding format:
//   - Accepts uppercase, lowercase, or mixed case
//   - Requires even length input
//   - Rejects invalid hex characters
func New() libenc.Coder {
	return &crt{}
}
