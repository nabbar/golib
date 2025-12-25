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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive/tar"
)

var _ = Describe("TC-WR-001: Tar Writer", func() {
	Describe("TC-WR-002: NewWriter", func() {
		It("TC-WR-003: should create a valid writer", func() {
			var buf bytes.Buffer
			writer, err := tar.NewWriter(&nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())
		})
	})

	Describe("TC-WR-004: Add", func() {
		It("TC-WR-005: should add a single file", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			content := strings.NewReader("test content")
			info := &testFileInfo{
				name:    "test.txt",
				size:    12,
				mode:    0644,
				modTime: time.Now(),
			}

			err := writer.Add(info, io.NopCloser(content), "test.txt", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			Expect(countFilesInArchive(&buf)).To(Equal(1))
		})

		It("TC-WR-006: should add multiple files", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			files := map[string]string{
				"file1.txt": "content 1",
				"file2.txt": "content 2",
				"file3.txt": "content 3",
			}

			for name, content := range files {
				rc := io.NopCloser(strings.NewReader(content))
				info := &testFileInfo{
					name:    name,
					size:    int64(len(content)),
					mode:    0644,
					modTime: time.Now(),
				}
				err := writer.Add(info, rc, name, "")
				Expect(err).ToNot(HaveOccurred())
			}

			writer.Close()
			Expect(countFilesInArchive(&buf)).To(Equal(len(files)))
		})

		It("TC-WR-007: should add file with custom path", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			content := strings.NewReader("content")
			info := &testFileInfo{name: "original.txt", size: 7, mode: 0644, modTime: time.Now()}

			err := writer.Add(info, io.NopCloser(content), "custom/path/file.txt", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			Expect(reader.Has("custom/path/file.txt")).To(BeTrue())
		})

		It("TC-WR-008: should add file with link target", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			info := &testFileInfo{
				name:    "link.txt",
				size:    0,
				mode:    0777 | os.ModeSymlink,
				modTime: time.Now(),
			}

			err := writer.Add(info, nil, "link.txt", "/target/path")
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-009: should handle empty file", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			content := strings.NewReader("")
			info := &testFileInfo{name: "empty.txt", size: 0, mode: 0644, modTime: time.Now()}

			err := writer.Add(info, io.NopCloser(content), "empty.txt", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			rc, _ := reader.Get("empty.txt")
			defer rc.Close()
			content2, _ := io.ReadAll(rc)
			Expect(content2).To(BeEmpty())
		})

		It("TC-WR-010: should handle large file", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			largeContent := strings.Repeat("x", 10000)
			content := strings.NewReader(largeContent)
			info := &testFileInfo{name: "large.txt", size: int64(len(largeContent)), mode: 0644, modTime: time.Now()}

			err := writer.Add(info, io.NopCloser(content), "large.txt", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			rc, _ := reader.Get("large.txt")
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			Expect(len(data)).To(Equal(len(largeContent)))
		})
	})

	Describe("TC-WR-011: FromPath", func() {
		var tmpDir string

		AfterEach(func() {
			cleanupTempDir(tmpDir)
		})

		It("TC-WR-012: should add single file", func() {
			var err error
			tmpDir, err = createTempDir(map[string]string{
				"test.txt": "test content",
			})
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			filePath := filepath.Join(tmpDir, "test.txt")
			err = writer.FromPath(filePath, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			Expect(countFilesInArchive(&buf)).To(Equal(1))
		})

		It("TC-WR-013: should add directory recursively", func() {
			var err error
			tmpDir, err = createTempDir(map[string]string{
				"file1.txt":        "content 1",
				"file2.txt":        "content 2",
				"sub/file3.txt":    "content 3",
				"sub/deep/file.go": "package main",
			})
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			err = writer.FromPath(tmpDir, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			Expect(countFilesInArchive(&buf)).To(Equal(4))
		})

		It("TC-WR-014: should filter files by pattern", func() {
			var err error
			tmpDir, err = createTempDir(map[string]string{
				"file1.txt": "text file 1",
				"file2.txt": "text file 2",
				"file.go":   "go file",
				"file.md":   "markdown file",
			})
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			err = writer.FromPath(tmpDir, "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			files, _ := reader.List()
			for _, f := range files {
				Expect(f).To(HaveSuffix(".txt"))
			}
		})

		It("TC-WR-015: should use path replacement function", func() {
			var err error
			tmpDir, err = createTempDir(map[string]string{
				"data.txt": "content",
			})
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			replaceFn := func(path string) string {
				return "renamed/" + filepath.Base(path)
			}

			err = writer.FromPath(tmpDir, "*", replaceFn)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			Expect(reader.Has("renamed/data.txt")).To(BeTrue())
		})

		It("TC-WR-016: should skip directories", func() {
			var err error
			tmpDir, err = createTempDir(map[string]string{
				"file.txt":     "file content",
				"sub/file.txt": "sub file content",
			})
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			err = writer.FromPath(tmpDir, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			// Walk through files and verify none are directories
			dirFound := false
			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				if info.IsDir() {
					dirFound = true
				}
				return true
			})
			Expect(dirFound).To(BeFalse())
		})

		It("TC-WR-017: should handle empty directory", func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "tar-empty-*")
			Expect(err).ToNot(HaveOccurred())

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			err = writer.FromPath(tmpDir, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			Expect(countFilesInArchive(&buf)).To(Equal(0))
		})

		It("TC-WR-018: should handle non-existent path", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			defer writer.Close()

			err := writer.FromPath("/nonexistent/path", "*", nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TC-WR-019: Close", func() {
		It("TC-WR-020: should close without error", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-021: should finalize archive on close", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			content := strings.NewReader("test")
			info := &testFileInfo{name: "test.txt", size: 4, mode: 0644, modTime: time.Now()}
			writer.Add(info, io.NopCloser(content), "test.txt", "")

			writer.Close()

			// Archive should be valid after close
			reader, err := tar.NewReader(io.NopCloser(&buf))
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			files, _ := reader.List()
			Expect(files).To(HaveLen(1))
		})

		It("TC-WR-022: should be safe to call multiple times", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			err1 := writer.Close()
			err2 := writer.Close()
			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).To(HaveOccurred()) // Second close may fail
		})

		It("TC-WR-023: should propagate close errors", func() {
			errWriter := &errorWriteCloser{err: fmt.Errorf("close error")}
			writer, _ := tar.NewWriter(errWriter)

			err := writer.Close()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TC-WR-024: Integration", func() {
		It("TC-WR-025: should create valid archive that can be read", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			files := map[string]string{
				"doc1.txt":     "Document 1",
				"doc2.txt":     "Document 2",
				"dir/doc3.txt": "Document 3",
			}

			for name, content := range files {
				rc := io.NopCloser(strings.NewReader(content))
				info := &testFileInfo{
					name:    filepath.Base(name),
					size:    int64(len(content)),
					mode:    0644,
					modTime: time.Now(),
				}
				writer.Add(info, rc, name, "")
			}

			writer.Close()

			// Verify by reading with Walk
			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			found := make(map[string]string)
			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				content, _ := io.ReadAll(rc)
				found[path] = string(content)
				return true
			})

			Expect(found).To(HaveLen(len(files)))
			for name, expectedContent := range files {
				Expect(found[name]).To(Equal(expectedContent))
			}
		})

		It("TC-WR-026: should handle mixed operations", func() {
			tmpDir, err := createTempDir(map[string]string{
				"temp1.txt": "temp file 1",
				"temp2.txt": "temp file 2",
			})
			Expect(err).ToNot(HaveOccurred())
			defer cleanupTempDir(tmpDir)

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			// Add from path
			err = writer.FromPath(tmpDir, "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			// Add manually
			rc := io.NopCloser(strings.NewReader("manual"))
			info := &testFileInfo{name: "manual.txt", size: 6, mode: 0644, modTime: time.Now()}
			writer.Add(info, rc, "manual.txt", "")

			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			files, _ := reader.List()
			Expect(len(files)).To(BeNumerically(">=", 1)) // At least manual.txt
			Expect(files).To(ContainElement("manual.txt"))
		})
	})
})
