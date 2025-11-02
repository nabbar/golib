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

// Package sha256 provides SHA-256 cryptographic hashing implementing the encoding.Coder interface.
//
// This package wraps Go's standard crypto/sha256 implementation and provides a consistent
// interface for hash operations with hex-encoded output.
//
// Features:
//   - FIPS 180-4 compliant SHA-256 hashing
//   - 256-bit (32 byte) hash output
//   - Hex-encoded output (64 characters)
//   - Streaming I/O support via io.Reader
//   - Thread-safe stateless operations
//   - Standard encoding.Coder interface
//
// SHA-256 properties:
//   - One-way hash function (cannot reverse)
//   - Collision resistant (2^128 operations)
//   - Deterministic (same input = same output)
//   - Avalanche effect (small input change = completely different hash)
//
// Example usage:
//
//	import encsha "github.com/nabbar/golib/encoding/sha256"
//
//	// Create hasher
//	hasher := encsha.New()
//
//	// Hash data
//	data := []byte("Hello, World!")
//	hash := hasher.Encode(data)
//	fmt.Printf("SHA-256: %s\n", hash)
//	// Output: dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
//
//	// Hash file
//	file, _ := os.Open("data.bin")
//	defer file.Close()
//	hashReader := hasher.EncodeReader(file)
//	io.Copy(io.Discard, hashReader)
//
// Use cases:
//   - File integrity verification
//   - Data deduplication
//   - Checksum generation
//   - Content-addressed storage
//   - Digital signatures (as part of signature schemes)
//
// Note: For password hashing, use bcrypt or argon2 instead. SHA-256 is too fast
// for passwords and lacks salt/iteration features needed for password security.
package sha256

import (
	"crypto/sha256"

	libenc "github.com/nabbar/golib/encoding"
)

// New creates a new SHA-256 hasher implementing the encoding.Coder interface.
//
// The returned hasher can be used to compute SHA-256 hashes of data. The hash
// output is hex-encoded (64 characters representing 32 bytes).
//
// Returns:
//   - encoding.Coder: SHA-256 hasher instance
//
// The hasher is stateless and can be reused. Call Reset() between operations
// to ensure clean state.
//
// Example:
//
//	hasher := sha256.New()
//	hash := hasher.Encode([]byte("data"))
//	fmt.Println(string(hash))  // 64 hex characters
//
// Properties:
//   - Output: 256 bits (32 bytes, 64 hex chars)
//   - Algorithm: FIPS 180-4 SHA-256
//   - Security: Cryptographic-strength
//   - Performance: ~400 MB/s on modern CPUs
func New() libenc.Coder {
	return &crt{
		hsh: sha256.New(),
	}
}
