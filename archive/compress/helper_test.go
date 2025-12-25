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

package compress_test

import (
	"bytes"
	"io"

	"github.com/nabbar/golib/archive/compress"
)

type bnc struct {
	alg compress.Algorithm
	nbr int
	txt string
	buf []byte
}

func newTestBenchDataOpe(alg compress.Algorithm, size int, msg string) *bnc {
	data := bytes.Repeat([]byte("test data for compression "), size)

	return &bnc{
		alg: alg,
		nbr: size,
		txt: msg,
		buf: data[:size],
	}
}

// tst holds test data used across multiple test cases
type tst struct {
	dat []byte
	str string
}

// newTestData creates a test data instance with various sizes
func newTestData(size int) tst {
	data := bytes.Repeat([]byte("test data for compression "), size/26+1)
	return tst{
		dat: data[:size],
		str: string(data[:size]),
	}
}

// errReader is an io.Reader that always returns an error
type errReader struct {
	err error
}

func (e errReader) Read(p []byte) (int, error) {
	return 0, e.err
}

func (e errReader) Close() error {
	return e.err
}

// errWriter is an io.WriteCloser that always returns an error
type errWriter struct {
	err error
}

func (e errWriter) Write(p []byte) (int, error) {
	return 0, e.err
}

func (e errWriter) Close() error {
	return e.err
}

// limitedWriter writes up to N bytes then returns an error
type limitedWriter struct {
	w io.Writer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return 0, io.ErrShortWrite
	}
	if len(p) > l.n {
		p = p[:l.n]
	}
	n, err := l.w.Write(p)
	l.n -= n
	if l.n <= 0 && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func (l *limitedWriter) Close() error {
	if c, ok := l.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// nopWriteCloser wraps an io.Writer to provide Close method
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

// compressTestData compresses data using the provided algorithm for testing
func compressTestData(alg interface {
	Writer(io.WriteCloser) (io.WriteCloser, error)
}, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := alg.Writer(nopWriteCloser{&buf})
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// roundTripTest performs a compression and decompression round-trip test
func roundTripTest(
	writer func(io.WriteCloser) (io.WriteCloser, error),
	reader func(io.Reader) (io.ReadCloser, error),
	data []byte,
) ([]byte, error) {
	// Compress
	var compressed bytes.Buffer
	w, err := writer(nopWriteCloser{&compressed})
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	// Decompress
	r, err := reader(&compressed)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return io.ReadAll(r)
}
