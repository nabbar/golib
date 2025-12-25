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
 */

package helper_test

import (
	"bytes"
	"errors"
	"io"
)

type errReader struct {
	err error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errReader) Close() error {
	return nil
}

type errWriter struct {
	err error
}

func (e *errWriter) Write(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errWriter) Close() error {
	return nil
}

type limitReader struct {
	r     io.Reader
	limit int
	read  int
}

func (l *limitReader) Read(p []byte) (n int, err error) {
	if l.read >= l.limit {
		return 0, io.EOF
	}
	if len(p) > l.limit-l.read {
		p = p[:l.limit-l.read]
	}
	n, err = l.r.Read(p)
	l.read += n
	return n, err
}

func (l *limitReader) Close() error {
	if c, ok := l.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func newLimitReader(r io.Reader, limit int) io.ReadCloser {
	return &limitReader{r: r, limit: limit}
}

type countWriter struct {
	w     io.Writer
	count int
}

func (c *countWriter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.count += n
	return n, err
}

func (c *countWriter) Close() error {
	if cl, ok := c.w.(io.Closer); ok {
		return cl.Close()
	}
	return nil
}

func newCountWriter(w io.Writer) *countWriter {
	return &countWriter{w: w}
}

func compressData(data []byte) []byte {
	var buf bytes.Buffer
	return buf.Bytes()
}

func decompressData(data []byte) ([]byte, error) {
	return data, nil
}

var testErr = errors.New("test error")
