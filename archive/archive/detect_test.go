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

package archive_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive"
)

var _ = Describe("TC-DT-001: Detection Operations", func() {
	Describe("TC-DT-002: TAR Detection", func() {
		It("TC-DT-003: should detect TAR archive from file", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "test.txt", "content")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			alg, reader, stream, err := archive.Detect(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
			Expect(reader).ToNot(BeNil())
			Expect(stream).ToNot(BeNil())

			if reader != nil {
				reader.Close()
			}
			if stream != nil {
				stream.Close()
			}
		})

		It("TC-DT-004: should detect TAR archive from file with content", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "data.txt", "test data")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			alg, reader, stream, err := archive.Detect(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
			Expect(reader).ToNot(BeNil())

			if reader != nil {
				reader.Close()
			}
			if stream != nil {
				stream.Close()
			}
		})
	})

	Describe("TC-DT-005: ZIP Detection", func() {
		It("TC-DT-006: ZIP detection is covered in zip subpackage tests", func() {
			// ZIP reader requires special interfaces (Size, ReaderAt, Seeker)
			// which are tested extensively in the zip/ subpackage
			// This test is intentionally skipped as it duplicates those tests
			Skip("ZIP detection fully tested in archive/zip/ subpackage")
		})
	})

	Describe("TC-DT-007: Unknown Format Detection", func() {
		It("TC-DT-008: should return None for non-archive data", func() {
			data := make([]byte, 300)
			for i := range data {
				data[i] = byte(i % 256)
			}
			alg, reader, stream, err := archive.Detect(io.NopCloser(bytes.NewReader(data)))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.None))
			Expect(reader).To(BeNil())

			if stream != nil {
				stream.Close()
			}
		})

		It("TC-DT-009: should return error for truncated data", func() {
			data := []byte("short")
			_, _, _, err := archive.Detect(io.NopCloser(bytes.NewReader(data)))
			Expect(err).To(HaveOccurred())
		})

		It("TC-DT-010: should handle empty data", func() {
			data := []byte{}
			_, _, _, err := archive.Detect(io.NopCloser(bytes.NewReader(data)))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TC-DT-011: Reader Functionality After Detection", func() {
		It("TC-DT-012: should list files from detected TAR archive", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			testFiles := map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			}
			_ = createTestFiles(tmpDir, testFiles)

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, reader, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()
			defer reader.Close()

			files, err := reader.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(HaveLen(2))
		})

		It("TC-DT-013: should walk files from detected TAR archive", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "test.txt", "content")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, reader, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()
			if reader != nil {
				defer reader.Close()
			}

			count := 0
			if reader != nil {
				reader.Walk(func(info os.FileInfo, r io.ReadCloser, path, link string) bool {
					if !info.IsDir() {
						count++
					}
					return true
				})
			}
			Expect(count).To(Equal(1))
		})
	})

	Describe("TC-DT-014: Edge Cases", func() {
		It("TC-DT-015: should handle archive with single file", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "single.txt", "single content")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			alg, reader, stream, err := archive.Detect(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))

			if reader != nil {
				defer reader.Close()
			}
			if stream != nil {
				defer stream.Close()
			}
		})

		It("TC-DT-016: should handle archive with nested directories", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)

			subDir := filepath.Join(tmpDir, "sub", "nested")
			_ = os.MkdirAll(subDir, 0755)
			_ = createTestFile(subDir, "deep.txt", "deep content")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			alg, reader, stream, err := archive.Detect(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))

			if reader != nil {
				defer reader.Close()
			}
			if stream != nil {
				defer stream.Close()
			}
		})
	})
})

type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}
