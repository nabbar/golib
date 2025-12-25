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
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive/zip"
)

var _ = Describe("TC-WR-001: ZIP Writer Operations", func() {
	Describe("TC-WR-002: NewWriter", func() {
		It("TC-WR-003: should create valid writer from file", func() {
			tmpFile, err := os.CreateTemp("", "test-*.zip")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, err := zip.NewWriter(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-004: should create writer from buffer", func() {
			buf := newBufferWriteCloser()
			writer, err := zip.NewWriter(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())
			writer.Close()
		})
	})

	Describe("TC-WR-005: Add Operations", func() {
		It("TC-WR-006: should add single file to archive", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			content := []byte("test content")
			info := createTestFileInfo("test.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))

			err := writer.Add(info, reader, "", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("TC-WR-007: should add file with custom path", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			content := []byte("content")
			info := createTestFileInfo("original.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))

			err := writer.Add(info, reader, "custom/path.txt", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-008: should handle nil reader in Add", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			info := createTestFileInfo("test.txt", 0)
			err := writer.Add(info, nil, "", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-009: should add multiple files", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			for i := 0; i < 3; i++ {
				content := []byte("content")
				info := createTestFileInfo("file.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
			}

			writer.Close()
		})

		It("TC-WR-010: should handle empty file content", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			content := []byte("")
			info := createTestFileInfo("empty.txt", 0)
			reader := io.NopCloser(bytes.NewReader(content))

			err := writer.Add(info, reader, "", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-011: should handle large file", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			largeData := make([]byte, 50000)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			info := createTestFileInfo("large.bin", int64(len(largeData)))
			reader := io.NopCloser(bytes.NewReader(largeData))

			err := writer.Add(info, reader, "", "")
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})
	})

	Describe("TC-WR-012: FromPath Operations", func() {
		It("TC-WR-013: should add files from directory", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file1.txt": "content 1",
				"file2.txt": "content 2",
			})
			defer os.RemoveAll(testDir)

			tmpFile, _ := os.CreateTemp("", "test-*.zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, _ := zip.NewWriter(tmpFile)
			err := writer.FromPath(testDir, "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-014: should filter files by pattern", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file.txt": "text",
				"data.csv": "csv",
			})
			defer os.RemoveAll(testDir)

			tmpFile, _ := os.CreateTemp("", "test-*.zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, _ := zip.NewWriter(tmpFile)
			err := writer.FromPath(testDir, "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-015: should transform paths with ReplaceName", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"original.txt": "content",
			})
			defer os.RemoveAll(testDir)

			tmpFile, _ := os.CreateTemp("", "test-*.zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, _ := zip.NewWriter(tmpFile)

			replaceFn := func(source string) string {
				return "renamed/" + filepath.Base(source)
			}

			err := writer.FromPath(testDir, "*", replaceFn)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-016: should add single file by path", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"single.txt": "single content",
			})
			defer os.RemoveAll(testDir)

			tmpFile, _ := os.CreateTemp("", "test-*.zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, _ := zip.NewWriter(tmpFile)
			filePath := filepath.Join(testDir, "single.txt")
			err := writer.FromPath(filePath, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-017: should handle non-existent path", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			err := writer.FromPath("/nonexistent/path", "*", nil)
			Expect(err).To(HaveOccurred())

			writer.Close()
		})

		It("TC-WR-018: should skip non-regular files", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file.txt": "content",
			})
			defer os.RemoveAll(testDir)

			// Create subdirectory (should be skipped)
			os.Mkdir(filepath.Join(testDir, "subdir"), 0755)

			tmpFile, _ := os.CreateTemp("", "test-*.zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, _ := zip.NewWriter(tmpFile)
			err := writer.FromPath(testDir, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})
	})

	Describe("TC-WR-019: Close Operations", func() {
		It("TC-WR-020: should close without error", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-021: should finalize archive on close", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			content := []byte("test")
			info := createTestFileInfo("test.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))
			writer.Add(info, reader, "", "")

			writer.Close()

			// Verify archive is valid
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Describe("TC-WR-022: Integration", func() {
		It("TC-WR-023: should create valid archive readable by Reader", func() {
			// Create and write to buffer first
			buf := newBufferWriteCloser()
			writer, err := zip.NewWriter(buf)
			Expect(err).ToNot(HaveOccurred())

			content := []byte("test content")
			info := createTestFileInfo("test.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))
			err = writer.Add(info, reader, "", "")
			Expect(err).ToNot(HaveOccurred())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Read archive from buffer
			zipReader := newReaderWithSize(buf.Bytes())
			r, err := zip.NewReader(zipReader)
			Expect(err).ToNot(HaveOccurred())
			defer r.Close()

			Expect(r.Has("test.txt")).To(BeTrue())

			// Verify content
			rc, err := r.Get("test.txt")
			Expect(err).ToNot(HaveOccurred())
			defer rc.Close()

			readContent, err := io.ReadAll(rc)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(readContent)).To(Equal(string(content)))
		})

		It("TC-WR-024: should handle errors in FromPath", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Test with non-existent path
			err := writer.FromPath("/nonexistent/path/that/does/not/exist", "*", nil)
			Expect(err).To(HaveOccurred())

			writer.Close()
		})
	})
})
