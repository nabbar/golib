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
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive/tar"
)

var _ = Describe("TC-EC-001: Edge Cases", func() {
	Describe("TC-EC-002: Empty Archive", func() {
		It("TC-EC-003: should handle empty archive in List", func() {
			emptyBuf := createEmptyArchive()
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		It("TC-EC-004: should handle empty archive in Walk", func() {
			emptyBuf := createEmptyArchive()
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			defer reader.Close()

			called := false
			reader.Walk(func(_ fs.FileInfo, _ io.ReadCloser, _ string, _ string) bool {
				called = true
				return true
			})

			Expect(called).To(BeFalse())
		})

		It("TC-EC-005: should handle empty archive in Has", func() {
			emptyBuf := createEmptyArchive()
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(emptyBuf.Bytes())))
			defer reader.Close()

			Expect(reader.Has("any.txt")).To(BeFalse())
		})
	})

	Describe("TC-EC-006: Large Data", func() {
		It("TC-EC-007: should handle large file content", func() {
			largeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 10000)
			archiveBuf := createTestArchive(map[string]string{
				"large.txt": largeContent,
			})

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			rc, err := reader.Get("large.txt")
			Expect(err).ToNot(HaveOccurred())
			defer rc.Close()

			content, err := io.ReadAll(rc)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(content)).To(Equal(len(largeContent)))
			Expect(string(content)).To(Equal(largeContent))
		})

		It("TC-EC-008: should handle many files", func() {
			manyFiles := make(map[string]string)
			for i := 0; i < 100; i++ {
				manyFiles[strings.Repeat("a", i)+".txt"] = strings.Repeat("x", i)
			}

			archiveBuf := createTestArchive(manyFiles)
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(HaveLen(100))
		})
	})

	Describe("TC-EC-009: Special Characters", func() {
		It("TC-EC-010: should handle filenames with spaces", func() {
			archiveBuf := createTestArchive(map[string]string{
				"file with spaces.txt": "content",
			})

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			Expect(reader.Has("file with spaces.txt")).To(BeTrue())
		})

		It("TC-EC-011: should handle filenames with special characters", func() {
			archiveBuf := createTestArchive(map[string]string{
				"file-name_123.txt": "content",
			})

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			rc, err := reader.Get("file-name_123.txt")
			Expect(err).ToNot(HaveOccurred())
			defer rc.Close()
		})

		It("TC-EC-012: should handle deep directory paths", func() {
			archiveBuf := createTestArchive(map[string]string{
				"a/b/c/d/e/f/g/h/file.txt": "deep content",
			})

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			Expect(reader.Has("a/b/c/d/e/f/g/h/file.txt")).To(BeTrue())
		})
	})

	Describe("TC-EC-013: Binary Content", func() {
		It("TC-EC-014: should handle binary data", func() {
			binaryData := make([]byte, 256)
			for i := 0; i < 256; i++ {
				binaryData[i] = byte(i)
			}

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "binary.bin", size: int64(len(binaryData)), mode: 0644}
			writer.Add(info, io.NopCloser(bytes.NewReader(binaryData)), "binary.bin", "")
			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			rc, _ := reader.Get("binary.bin")
			defer rc.Close()
			readData, _ := io.ReadAll(rc)

			Expect(readData).To(Equal(binaryData))
		})

		It("TC-EC-015: should handle zero bytes in content", func() {
			dataWithZeros := []byte("hello\x00world\x00test")

			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "zeros.bin", size: int64(len(dataWithZeros)), mode: 0644}
			writer.Add(info, io.NopCloser(bytes.NewReader(dataWithZeros)), "zeros.bin", "")
			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			rc, _ := reader.Get("zeros.bin")
			defer rc.Close()
			readData, _ := io.ReadAll(rc)

			Expect(readData).To(Equal(dataWithZeros))
		})
	})

	Describe("TC-EC-016: Concurrent Access", func() {
		It("TC-EC-017: should handle sequential reads safely", func() {
			archiveBuf := createTestArchive(map[string]string{
				"file1.txt": "content 1",
				"file2.txt": "content 2",
			})

			// Use resetable reader for multiple operations
			resetReader := newResetableReader(archiveBuf.Bytes())
			reader, _ := tar.NewReader(resetReader)
			defer reader.Close()

			// Sequential access should work fine
			Expect(reader.Has("file1.txt")).To(BeTrue())
			Expect(reader.Has("file2.txt")).To(BeTrue())

			files, _ := reader.List()
			Expect(files).To(HaveLen(2))
		})
	})

	Describe("TC-EC-018: Reset Behavior", func() {
		It("TC-EC-019: should support reset with resetable reader", func() {
			archiveBuf := createTestArchive(map[string]string{
				"test.txt": "test content",
			})

			resetReader := newResetableReader(archiveBuf.Bytes())
			reader, _ := tar.NewReader(resetReader)
			defer reader.Close()

			// First read
			files1, _ := reader.List()
			Expect(files1).To(HaveLen(1))

			// Second read after reset
			files2, _ := reader.List()
			Expect(files2).To(Equal(files1))
		})

		It("TC-EC-020: should handle non-resetable reader", func() {
			archiveBuf := createTestArchive(map[string]string{
				"test.txt": "test content",
			})

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()

			// First read
			files1, _ := reader.List()
			Expect(files1).To(HaveLen(1))

			// Second read without reset - may return empty
			files2, _ := reader.List()
			Expect(files2).To(BeEmpty())
		})
	})

	Describe("TC-EC-021: Malformed Input", func() {
		It("TC-EC-022: should handle corrupted archive gracefully", func() {
			corruptedData := []byte("This is not a tar archive")
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(corruptedData)))
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred()) // List returns empty, not error
			Expect(files).To(BeEmpty())
		})

		It("TC-EC-023: should handle truncated archive", func() {
			archiveBuf := createTestArchive(map[string]string{
				"test.txt": "content",
			})

			// Truncate the archive
			truncated := archiveBuf.Bytes()[:len(archiveBuf.Bytes())/2]

			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(truncated)))
			defer reader.Close()

			// Should handle gracefully
			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			_ = files // May be empty or partial
		})
	})

	Describe("TC-EC-024: Permissions and Modes", func() {
		It("TC-EC-025: should preserve file permissions", func() {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

			info := &testFileInfo{
				name: "exec.sh",
				size: 10,
				mode: 0755,
			}

			writer.Add(info, io.NopCloser(strings.NewReader("#!/bin/sh\n")), "exec.sh", "")
			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			defer reader.Close()

			fileInfo, err := reader.Info("exec.sh")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileInfo.Mode().Perm()).To(Equal(os.FileMode(0755)))
		})
	})
})
