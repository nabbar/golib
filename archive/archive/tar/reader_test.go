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
	"io"
	"io/fs"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive/tar"
)

var _ = Describe("TC-RD-001: Tar Reader", func() {
	var (
		testFiles  map[string]string
		archiveBuf *bytes.Buffer
	)

	BeforeEach(func() {
		testFiles = map[string]string{
			"file1.txt":       "content of file 1",
			"file2.txt":       "content of file 2",
			"dir/file3.txt":   "content of file 3",
			"dir/sub/file.go": "package main",
		}
		archiveBuf = createTestArchive(testFiles)
	})

	Describe("TC-RD-002: NewReader", func() {
		It("TC-RD-003: should create a valid reader", func() {
			reader, err := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
		})

		It("TC-RD-004: should create reader from empty archive", func() {
			emptyBuf := createEmptyArchive()
			reader, err := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
		})
	})

	Describe("TC-RD-005: List", func() {
		It("TC-RD-006: should list all files in archive", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(HaveLen(len(testFiles)))
			Expect(files).To(ConsistOf("file1.txt", "file2.txt", "dir/file3.txt", "dir/sub/file.go"))
		})

		It("TC-RD-007: should return empty list for empty archive", func() {
			emptyBuf := createEmptyArchive()
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		It("TC-RD-008: should handle multiple List calls with resetable reader", func() {
			resetReader := newResetableReader(archiveBuf.Bytes())
			reader, _ := tar.NewReader(resetReader)
			defer reader.Close()

			files1, err1 := reader.List()
			Expect(err1).ToNot(HaveOccurred())
			Expect(files1).To(HaveLen(len(testFiles)))

			files2, err2 := reader.List()
			Expect(err2).ToNot(HaveOccurred())
			Expect(files2).To(Equal(files1))
		})
	})

	Describe("TC-RD-009: Info", func() {
		It("TC-RD-010: should get file info for existing file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			info, err := reader.Info("file1.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(info).ToNot(BeNil())
			Expect(info.Name()).To(Equal("file1.txt"))
			Expect(info.Size()).To(Equal(int64(len(testFiles["file1.txt"]))))
			Expect(info.IsDir()).To(BeFalse())
		})

		It("TC-RD-011: should return error for non-existing file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			info, err := reader.Info("nonexistent.txt")
			Expect(err).To(Equal(fs.ErrNotExist))
			Expect(info).To(BeNil())
		})

		It("TC-RD-012: should get info for nested file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			info, err := reader.Info("dir/sub/file.go")
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Name()).To(Equal("file.go"))
			Expect(info.Size()).To(Equal(int64(len(testFiles["dir/sub/file.go"]))))
		})
	})

	Describe("TC-RD-013: Get", func() {
		It("TC-RD-014: should extract file content", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			rc, err := reader.Get("file1.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(rc).ToNot(BeNil())
			defer rc.Close()

			content, err := io.ReadAll(rc)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(testFiles["file1.txt"]))
		})

		It("TC-RD-015: should return error for non-existing file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			rc, err := reader.Get("missing.txt")
			Expect(err).To(Equal(fs.ErrNotExist))
			Expect(rc).To(BeNil())
		})

		It("TC-RD-016: should extract nested file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			rc, err := reader.Get("dir/file3.txt")
			Expect(err).ToNot(HaveOccurred())
			defer rc.Close()

			content, _ := io.ReadAll(rc)
			Expect(string(content)).To(Equal(testFiles["dir/file3.txt"]))
		})

		It("TC-RD-017: should extract multiple files with reset", func() {
			resetReader := newResetableReader(archiveBuf.Bytes())
			reader, _ := tar.NewReader(resetReader)
			defer reader.Close()

			rc1, err1 := reader.Get("file1.txt")
			Expect(err1).ToNot(HaveOccurred())
			content1, _ := io.ReadAll(rc1)
			rc1.Close()

			rc2, err2 := reader.Get("file2.txt")
			Expect(err2).ToNot(HaveOccurred())
			content2, _ := io.ReadAll(rc2)
			rc2.Close()

			Expect(string(content1)).To(Equal(testFiles["file1.txt"]))
			Expect(string(content2)).To(Equal(testFiles["file2.txt"]))
		})
	})

	Describe("TC-RD-018: Has", func() {
		It("TC-RD-019: should return true for existing file", func() {
			resetReader := newResetableReader(archiveBuf.Bytes())
			reader, _ := tar.NewReader(resetReader)
			defer reader.Close()

			Expect(reader.Has("file1.txt")).To(BeTrue())
			Expect(reader.Has("dir/file3.txt")).To(BeTrue())
		})

		It("TC-RD-020: should return false for non-existing file", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			Expect(reader.Has("missing.txt")).To(BeFalse())
			Expect(reader.Has("dir/nonexistent.go")).To(BeFalse())
		})

		It("TC-RD-021: should work with empty archive", func() {
			emptyBuf := createEmptyArchive()
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			defer reader.Close()

			Expect(reader.Has("any.txt")).To(BeFalse())
		})
	})

	Describe("TC-RD-022: Walk", func() {
		It("TC-RD-023: should iterate all files", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			visited := make([]string, 0)
			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				visited = append(visited, path)
				return true
			})

			Expect(visited).To(HaveLen(len(testFiles)))
			Expect(visited).To(ConsistOf("file1.txt", "file2.txt", "dir/file3.txt", "dir/sub/file.go"))
		})

		It("TC-RD-024: should provide correct file info in callback", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				if path == "file1.txt" {
					Expect(info.Size()).To(Equal(int64(len(testFiles["file1.txt"]))))
					Expect(info.IsDir()).To(BeFalse())
					content, _ := io.ReadAll(rc)
					Expect(string(content)).To(Equal(testFiles["file1.txt"]))
				}
				return true
			})
		})

		It("TC-RD-025: should stop walking when callback returns false", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			count := 0
			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				count++
				return count < 2
			})

			Expect(count).To(Equal(2))
		})

		It("TC-RD-026: should handle filter during walk", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			txtFiles := make([]string, 0)
			reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
				if strings.HasSuffix(path, ".txt") {
					txtFiles = append(txtFiles, path)
				}
				return true
			})

			Expect(txtFiles).To(HaveLen(3))
			Expect(txtFiles).To(ConsistOf("file1.txt", "file2.txt", "dir/file3.txt"))
		})
	})

	Describe("TC-RD-027: Close", func() {
		It("TC-RD-028: should close without error", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			err := reader.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-RD-029: should be safe to call multiple times", func() {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			err1 := reader.Close()
			err2 := reader.Close()
			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})
})
