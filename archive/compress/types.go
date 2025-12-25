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

package compress

import "bytes"

// Algorithm represents a compression algorithm supported by this package.
// It is implemented as a uint8 enum for efficient storage and comparison.
// The zero value (None) represents no compression.
type Algorithm uint8

// Compression algorithm constants.
// These represent the supported compression formats.
const (
	None  Algorithm = iota // No compression (pass-through)
	Bzip2                  // Burrows-Wheeler compression (.bz2)
	Gzip                   // GNU zip compression (.gz)
	LZ4                    // LZ4 fast compression (.lz4)
	XZ                     // LZMA2 compression (.xz)
)

// List returns all supported compression algorithms in declaration order.
// This is useful for enumerating available formats in UI or CLI tools.
func List() []Algorithm {
	return []Algorithm{
		None,
		Bzip2,
		Gzip,
		LZ4,
		XZ,
	}
}

// ListString returns the string representation of all supported algorithms.
// This is a convenience function equivalent to calling String() on each algorithm from List().
func ListString() []string {
	var (
		lst = List()
		res = make([]string, len(lst))
	)
	for i := range lst {
		res[i] = lst[i].String()
	}
	return res
}

// IsNone returns true if the algorithm is None (no compression).
// This is useful for checking if compression should be applied.
func (a Algorithm) IsNone() bool {
	return a == None
}

// String returns the lowercase string representation of the algorithm.
// For None, it returns "none". This method is used for text marshaling
// and human-readable output.
func (a Algorithm) String() string {
	switch a {
	case Gzip:
		return "gzip"
	case Bzip2:
		return "bzip2"
	case LZ4:
		return "lz4"
	case XZ:
		return "xz"
	default:
		return "none"
	}
}

// Extension returns the standard file extension for the algorithm,
// including the leading dot (e.g., ".gz" for Gzip).
// For None, it returns an empty string.
// This is useful for constructing compressed filenames.
func (a Algorithm) Extension() string {
	switch a {
	case Gzip:
		return ".gz"
	case Bzip2:
		return ".bz2"
	case LZ4:
		return ".lz4"
	case XZ:
		return ".xz"
	default:
		return ""
	}
}

// DetectHeader examines the provided byte slice to determine if it matches
// the magic number (file signature) of this algorithm.
// It requires at least 6 bytes for accurate detection (XZ header size).
// Returns false if the header doesn't match or if h is too short.
//
// Magic numbers:
//   - Gzip:  0x1F 0x8B
//   - Bzip2: 'B' 'Z' 'h' [0-9]
//   - LZ4:   0x04 0x22 0x4D 0x18
//   - XZ:    0xFD 0x37 0x7A 0x58 0x5A 0x00
func (a Algorithm) DetectHeader(h []byte) bool {
	if len(h) < 6 {
		return false
	}

	switch a {
	case Gzip:
		exp := []byte{31, 139}
		return bytes.Equal(h[0:2], exp)
	case Bzip2:
		exp := []byte{'B', 'Z', 'h'}
		return bytes.Equal(h[0:3], exp) && h[3] >= '0' && h[3] <= '9'
	case LZ4:
		exp := []byte{0x04, 0x22, 0x4D, 0x18}
		return bytes.Equal(h[0:4], exp)
	case XZ:
		exp := []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}
		alt := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		return bytes.Equal(h[0:6], exp) || bytes.Equal(h[0:6], alt)
	default:
		return false
	}
}
