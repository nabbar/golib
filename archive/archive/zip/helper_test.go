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

package zip_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/nabbar/golib/archive/archive/zip"
)

var testCtx context.Context

func init() {
	testCtx = context.Background()
}

type testFileInfo struct {
	name string
	size int64
	mode os.FileMode
	mdt  time.Time
}

func (t *testFileInfo) Name() string       { return t.name }
func (t *testFileInfo) Size() int64        { return t.size }
func (t *testFileInfo) Mode() os.FileMode  { return t.mode }
func (t *testFileInfo) ModTime() time.Time { return t.mdt }
func (t *testFileInfo) IsDir() bool        { return t.mode.IsDir() }
func (t *testFileInfo) Sys() interface{}   { return nil }

func createTestFileInfo(name string, size int64) *testFileInfo {
	return &testFileInfo{
		name: name,
		size: size,
		mode: 0644,
		mdt:  time.Now(),
	}
}

func createTestZipInMemory(files map[string]string) (*bufferWriteCloser, error) {
	buf := newBufferWriteCloser()

	writer, err := zip.NewWriter(buf)
	if err != nil {
		return nil, err
	}

	for name, content := range files {
		info := createTestFileInfo(name, int64(len(content)))
		reader := io.NopCloser(bytes.NewReader([]byte(content)))
		if err := writer.Add(info, reader, name, ""); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func createTestZipFile(files map[string]string) (string, error) {
	tmpFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		return "", err
	}

	tmpPath := tmpFile.Name()

	writer, err := zip.NewWriter(tmpFile)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", err
	}

	for name, content := range files {
		info := createTestFileInfo(name, int64(len(content)))
		reader := io.NopCloser(bytes.NewReader([]byte(content)))
		if err := writer.Add(info, reader, name, ""); err != nil {
			writer.Close()
			tmpFile.Close()
			os.Remove(tmpPath)
			return "", err
		}
	}

	// Close writer first (this will close tmpFile via its Close method)
	if err := writer.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}

func createTestDirectory(files map[string]string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "test-zip-dir-*")
	if err != nil {
		return "", err
	}

	for name, content := range files {
		fullPath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

type readerWithSize struct {
	*bytes.Reader
	closer io.Closer
	size   int64
}

func newReaderWithSize(data []byte) *readerWithSize {
	return &readerWithSize{
		Reader: bytes.NewReader(data),
		closer: io.NopCloser(nil),
		size:   int64(len(data)),
	}
}

func (r *readerWithSize) Size() int64 { return r.size }

func (r *readerWithSize) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

type bufferWriteCloser struct {
	*bytes.Buffer
}

func (b *bufferWriteCloser) Close() error {
	return nil
}

func newBufferWriteCloser() *bufferWriteCloser {
	return &bufferWriteCloser{Buffer: &bytes.Buffer{}}
}
