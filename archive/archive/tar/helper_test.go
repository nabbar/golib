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

package tar_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nabbar/golib/archive/archive/tar"
)

// nopWriteCloser wraps an io.Writer to implement io.WriteCloser
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

// errorWriteCloser implements io.WriteCloser that always returns an error
type errorWriteCloser struct {
	err error
}

func (e *errorWriteCloser) Write(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorWriteCloser) Close() error {
	return e.err
}

// errorReadCloser implements io.ReadCloser that returns an error after n bytes
type errorReadCloser struct {
	r   io.Reader
	err error
	n   int
	cnt int
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	if e.cnt >= e.n {
		return 0, e.err
	}
	n, err = e.r.Read(p)
	e.cnt += n
	if e.cnt >= e.n {
		return n, e.err
	}
	return n, err
}

func (e *errorReadCloser) Close() error {
	return nil
}

// testFileInfo implements fs.FileInfo for testing
type testFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (t *testFileInfo) Name() string       { return t.name }
func (t *testFileInfo) Size() int64        { return t.size }
func (t *testFileInfo) Mode() fs.FileMode  { return t.mode }
func (t *testFileInfo) ModTime() time.Time { return t.modTime }
func (t *testFileInfo) IsDir() bool        { return t.isDir }
func (t *testFileInfo) Sys() any           { return nil }

// createTestArchive creates a tar archive with the given files
func createTestArchive(files map[string]string) *bytes.Buffer {
	var buf bytes.Buffer
	writer, err := tar.NewWriter(&nopWriteCloser{&buf})
	if err != nil {
		panic(fmt.Sprintf("Failed to create writer: %v", err))
	}

	for name, content := range files {
		rc := io.NopCloser(strings.NewReader(content))
		info := &testFileInfo{
			name:    filepath.Base(name),
			size:    int64(len(content)),
			mode:    0644,
			modTime: time.Now(),
			isDir:   false,
		}
		if err := writer.Add(info, rc, name, ""); err != nil {
			panic(fmt.Sprintf("Failed to add file %s: %v", name, err))
		}
	}

	if err := writer.Close(); err != nil {
		panic(fmt.Sprintf("Failed to close writer: %v", err))
	}

	return &buf
}

// createEmptyArchive creates an empty tar archive
func createEmptyArchive() *bytes.Buffer {
	var buf bytes.Buffer
	writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
	writer.Close()
	return &buf
}

// createTempDir creates a temporary directory with files for testing
func createTempDir(files map[string]string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "tar-test-*")
	if err != nil {
		return "", err
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

// cleanupTempDir removes a temporary directory
func cleanupTempDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}

// countFilesInArchive counts the number of files in an archive
func countFilesInArchive(archiveBuf *bytes.Buffer) int {
	reader, err := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
	if err != nil {
		return 0
	}
	defer reader.Close()

	files, err := reader.List()
	if err != nil {
		return 0
	}

	return len(files)
}

// resetableReader implements a reader that can be reset
type resetableReader struct {
	data []byte
	pos  int
}

func newResetableReader(data []byte) *resetableReader {
	return &resetableReader{data: data, pos: 0}
}

func (r *resetableReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *resetableReader) Close() error {
	return nil
}

func (r *resetableReader) Reset() bool {
	r.pos = 0
	return true
}
