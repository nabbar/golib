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

package archive

import "bytes"

// Algorithm represents an archive format algorithm (Tar, Zip, or None).
// It is implemented as a uint8 enum for efficient comparisons and storage.
// The zero value (None) represents "no archive format".
type Algorithm uint8

const (
	// None represents no archive algorithm. This is the zero value.
	None Algorithm = iota

	// Tar represents the TAR (Tape ARchive) format. Tar archives store files
	// sequentially, with each file preceded by its metadata header.
	Tar

	// Zip represents the ZIP archive format. Zip archives use a central directory
	// to store file metadata and positions, allowing random access to files.
	Zip
)

// IsNone returns true if the algorithm is None (no archive format).
// This is useful for checking if an algorithm was successfully parsed or detected.
func (a Algorithm) IsNone() bool {
	return a == None
}

// String returns the string representation of the algorithm.
// Returns "tar" for Tar, "zip" for Zip, and "none" for None.
// This method is used for logging, configuration serialization, and user display.
func (a Algorithm) String() string {
	switch a {
	case Tar:
		return "tar"
	case Zip:
		return "zip"
	default:
		return "none"
	}
}

// Extension returns the file extension for the algorithm (including the dot).
// Returns ".tar" for Tar, ".zip" for Zip, and "" (empty string) for None.
// This is useful for automatic file naming and format detection.
func (a Algorithm) Extension() string {
	switch a {
	case Tar:
		return ".tar"
	case Zip:
		return ".zip"
	default:
		return ""
	}
}

// DetectHeader validates if the provided header bytes match the algorithm's format.
// It examines magic numbers at specific positions in the header:
//   - Tar: checks for "ustar\x00" at bytes 257-263 (POSIX tar format)
//   - Zip: checks for 0x504B0304 at bytes 0-4 (local file header signature)
//
// Parameters:
//   - h: the header bytes to validate (must be at least 263 bytes for reliable detection)
//
// Returns:
//   - true if the header matches the expected format for this algorithm
//   - false if the header doesn't match or is too short
//
// Note: This method returns false for None algorithm and for truncated headers.
func (a Algorithm) DetectHeader(h []byte) bool {
	if len(h) < 263 {
		return false
	}

	switch a {
	case Tar:
		exp := append([]byte("ustar"), 0x00)
		val := h[257:263]
		return bytes.Equal(val, exp)
	case Zip:
		exp := []byte{0x50, 0x4b, 0x03, 0x04}
		return bytes.Equal(h[0:4], exp)
	default:
		return false
	}
}
