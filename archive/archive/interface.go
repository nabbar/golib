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

import (
	"bufio"
	"io"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// Parse converts a string representation to the corresponding Algorithm.
// The parsing is case-insensitive and returns None for unsupported formats.
//
// This is a convenience function that wraps UnmarshalText for easier string-to-algorithm conversion.
//
// Supported values:
//   - "tar" or "TAR" → Tar
//   - "zip" or "ZIP" → Zip
//   - any other value → None
//
// Parameters:
//   - s: the string to parse (e.g., "tar", "zip", "none")
//
// Returns:
//   - Algorithm: the parsed algorithm, or None if the string is not recognized
//
// Example:
//
//	alg := archive.Parse("tar")
//	if alg == archive.Tar {
//	    // handle tar archive
//	}
//
// Note: This function never returns an error; it defaults to None for invalid input.
func Parse(s string) Algorithm {
	var alg = None
	if e := alg.UnmarshalText([]byte(s)); e != nil {
		return None
	} else {
		return alg
	}
}

// Detect automatically detects the archive format by examining the input stream's header
// and returns the appropriate reader for extracting the archive contents.
//
// The function performs the following steps:
//  1. Peeks at the first 265 bytes of the input stream (without consuming them)
//  2. Checks the magic numbers to identify the archive format (tar or zip)
//  3. Creates and returns a format-specific reader wrapped in the types.Reader interface
//
// Archive format detection:
//   - Tar: checks for "ustar\x00" signature at bytes 257-263 (POSIX tar format)
//   - Zip: checks for 0x504B0304 signature at bytes 0-4 (local file header)
//   - None: if no recognized format is detected
//
// Parameters:
//   - r: the input stream to detect and read. This must be at least 265 bytes long
//     for reliable detection. The caller remains responsible for closing the original stream.
//
// Returns:
//   - Algorithm: the detected archive format (Tar, Zip, or None)
//   - arctps.Reader: the archive-specific reader for extracting files (nil if None or error)
//   - io.ReadCloser: a buffered wrapper around the original stream that preserves the peeked data
//   - error: any error from peeking, detection, or reader creation
//
// Important behavioral differences:
//
// Tar archives:
//   - Work with any io.ReadCloser (sequential access only)
//   - The returned reader can only be used once (sequential tape archive format)
//   - Files are stored sequentially: [metadata][content][metadata][content]...
//   - Best accessed via the Walk() method to iterate through all files
//   - Supports streaming from network, pipes, or non-seekable sources
//
// Zip archives:
//   - Require io.ReaderAt and io.Seeker capabilities for random access
//   - The returned reader can access files multiple times (central directory format)
//   - Files have a central directory storing metadata and file positions
//   - Can use Get() to extract specific files by name without iteration
//   - Requires seekable sources (files, byte buffers, not pipes or streams)
//
// Error cases:
//   - Returns error if the stream has fewer than 265 bytes
//   - Returns error if the stream cannot be peeked
//   - Returns error if zip format is detected but stream lacks ReaderAt/Seeker
//   - Returns None (no error) if format is unrecognized
//
// Usage example:
//
//	file, err := os.Open("archive.tar")
//	if err != nil {
//	    return err
//	}
//	defer file.Close()
//
//	alg, reader, stream, err := archive.Detect(file)
//	if err != nil {
//	    return err
//	}
//	defer stream.Close()
//
//	if reader != nil {
//	    defer reader.Close()
//	    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
//	        // process each file in archive
//	        return true
//	    })
//	}
func Detect(r io.ReadCloser) (Algorithm, arctps.Reader, io.ReadCloser, error) {
	var (
		err error
		buf []byte
		bfr = &rdr{
			r: r,
			b: bufio.NewReader(r),
		}
	)

	if buf, err = bfr.Peek(265); err != nil {
		return None, nil, nil, err
	}

	switch {
	case Tar.DetectHeader(buf):
		if t, e := Tar.Reader(bfr); e != nil {
			return None, nil, nil, e
		} else {
			return Tar, t, bfr, nil
		}

	case Zip.DetectHeader(buf):
		bfr.b = nil
		if z, e := Zip.Reader(bfr); e != nil {
			return None, nil, nil, e
		} else {
			return Zip, z, bfr, nil
		}

	default:
		return None, nil, bfr, nil
	}
}
