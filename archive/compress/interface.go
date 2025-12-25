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

import (
	"bufio"
	"io"
)

// Parse is a convenience function to parse a string and return the corresponding Algorithm.
// The parsing is case-insensitive and trims whitespace, quotes, and apostrophes.
// If the string doesn't match any known algorithm, it returns None.
//
// Supported strings: "gzip", "bzip2", "lz4", "xz", "none" (case-insensitive).
//
// Example:
//
//	alg := Parse("gzip")        // Returns Gzip
//	alg := Parse("BZIP2")       // Returns Bzip2 (case-insensitive)
//	alg := Parse("unknown")     // Returns None
//	alg := Parse("  lz4  ")     // Returns LZ4 (whitespace trimmed)
func Parse(s string) Algorithm {
	var alg = None
	if e := alg.UnmarshalText([]byte(s)); e != nil {
		return None
	} else {
		return alg
	}
}

// Detect automatically detects the compression algorithm from the input stream
// and returns both the detected algorithm and a decompression reader ready for use.
//
// It combines DetectOnly (format detection) and Reader (decompression wrapping)
// into a single convenience function.
//
// The function reads the first 6 bytes to detect the format without consuming them
// from the stream (using bufio.Reader.Peek). The returned reader is properly wrapped
// for transparent decompression.
//
// Returns:
//   - Algorithm: The detected compression format (or None if uncompressed)
//   - io.ReadCloser: A reader that transparently decompresses the data
//   - error: Any error during detection or reader creation
//
// Example:
//
//	file, _ := os.Open("data.gz")
//	defer file.Close()
//	alg, reader, err := compress.Detect(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//	fmt.Printf("Detected: %s\n", alg.String())
//	data, _ := io.ReadAll(reader)  // Automatically decompressed
func Detect(r io.Reader) (Algorithm, io.ReadCloser, error) {
	var (
		err error
		alg Algorithm
		rdr io.ReadCloser
	)

	if alg, rdr, err = DetectOnly(r); err != nil {
		return None, nil, err
	} else if rdr, err = alg.Reader(rdr); err != nil {
		return None, nil, err
	} else {
		return alg, rdr, nil
	}
}

// DetectOnly detects the compression algorithm by examining the input stream's header
// without wrapping it in a decompression reader. This is useful when you need to know
// the format but want to handle the decompression yourself.
//
// Unlike Detect(), this function only identifies the format and returns a buffered reader
// that preserves the peeked header bytes. The returned reader can be passed to Reader()
// to create the appropriate decompression wrapper.
//
// Returns:
//   - Algorithm: The detected compression format (or None if uncompressed/unknown)
//   - io.ReadCloser: A buffered reader with preserved header data (via io.NopCloser)
//   - error: Any error during the peek operation
//
// Example:
//
//	file, _ := os.Open("unknown.dat")
//	defer file.Close()
//	alg, bufferedReader, err := compress.DetectOnly(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Format: %s\n", alg.String())
//	// Use bufferedReader for further processing
func DetectOnly(r io.Reader) (Algorithm, io.ReadCloser, error) {
	var (
		err error
		alg Algorithm
		bfr = bufio.NewReader(r)
		buf []byte
	)

	if buf, err = bfr.Peek(6); err != nil {
		return None, nil, err
	}

	switch {
	case Gzip.DetectHeader(buf): // gzip
		alg = Gzip
	case Bzip2.DetectHeader(buf): // bzip2
		alg = Bzip2
	case LZ4.DetectHeader(buf): // lz4
		alg = LZ4
	case XZ.DetectHeader(buf): // xz
		alg = XZ
	default:
		alg = None
	}

	return alg, io.NopCloser(bfr), err
}
