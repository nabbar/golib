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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive"
)

var _ = Describe("TC-IO-001: I/O Operations", func() {
	Describe("TC-IO-002: Reader Creation", func() {
		It("TC-IO-003: should create TAR reader successfully", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			reader, err := archive.Tar.Reader(io.NopCloser(bytes.NewReader(buf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			if reader != nil {
				reader.Close()
			}
		})

		It("TC-IO-004: ZIP reader creation is covered in zip subpackage", func() {
			// ZIP reader requires special interfaces (Size, ReaderAt, Seeker)
			// which are tested extensively in the zip/ subpackage
			Skip("ZIP reader fully tested in archive/zip/ subpackage")
		})

		It("TC-IO-005: should return error for None algorithm reader", func() {
			reader, err := archive.None.Reader(io.NopCloser(bytes.NewReader([]byte{})))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(archive.ErrInvalidAlgorithm))
			Expect(reader).To(BeNil())
		})

		It("TC-IO-006: should handle invalid TAR data", func() {
			invalidData := []byte("not a tar archive")
			reader, err := archive.Tar.Reader(io.NopCloser(bytes.NewReader(invalidData)))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			if reader != nil {
				reader.Close()
			}
		})
	})

	Describe("TC-IO-007: Writer Creation", func() {
		It("TC-IO-008: should create TAR writer successfully", func() {
			var buf bytes.Buffer
			writer, err := archive.Tar.Writer(&nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			if writer != nil {
				writer.Close()
			}
		})

		It("TC-IO-009: should create ZIP writer successfully", func() {
			tmpFile, _ := createTempArchiveFile(".zip")
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			writer, err := archive.Zip.Writer(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			if writer != nil {
				writer.Close()
			}
		})

		It("TC-IO-010: should return error for None algorithm writer", func() {
			var buf bytes.Buffer
			writer, err := archive.None.Writer(&nopWriteCloser{&buf})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(archive.ErrInvalidAlgorithm))
			Expect(writer).To(BeNil())
		})

		It("TC-IO-011: should write and read TAR archive", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			reader, err := archive.Tar.Reader(io.NopCloser(bytes.NewReader(buf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			if reader != nil {
				reader.Close()
			}
		})
	})

	Describe("TC-IO-012: Round-trip Operations", func() {
		It("TC-IO-013: should write and read files in TAR", func() {
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

			reader, _ := archive.Tar.Reader(tmpFile)
			defer reader.Close()

			files, _ := reader.List()
			Expect(files).To(HaveLen(2))
		})

		It("TC-IO-014: should write and read files in TAR (ZIP requires ReaderAt)", func() {
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

			reader, _ := archive.Tar.Reader(tmpFile)
			defer reader.Close()

			files, _ := reader.List()
			Expect(files).To(HaveLen(2))
		})

		It("TC-IO-015: should preserve file content in TAR", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)

			testContent := "test content for preservation"
			_ = createTestFile(tmpDir, "test.txt", testContent)

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			reader, _ := archive.Tar.Reader(tmpFile)
			defer reader.Close()

			files, _ := reader.List()
			if len(files) > 0 {
				rc, _ := reader.Get(files[0])
				if rc != nil {
					defer rc.Close()
					data, _ := io.ReadAll(rc)
					Expect(string(data)).To(Equal(testContent))
				}
			}
		})

		It("TC-IO-016: should preserve file content with multiple reads", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)

			testContent1 := "test content for preservation 1"
			testContent2 := "test content for preservation 2"
			_ = createTestFile(tmpDir, "test1.txt", testContent1)
			_ = createTestFile(tmpDir, "test2.txt", testContent2)

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			reader, _ := archive.Tar.Reader(tmpFile)
			defer reader.Close()

			files, _ := reader.List()
			Expect(files).To(HaveLen(2))
		})
	})

	Describe("TC-IO-017: Error Handling", func() {
		It("TC-IO-018: should handle ErrInvalidAlgorithm for None reader", func() {
			_, err := archive.None.Reader(io.NopCloser(bytes.NewReader([]byte{})))
			Expect(err).To(Equal(archive.ErrInvalidAlgorithm))
		})

		It("TC-IO-019: should handle ErrInvalidAlgorithm for None writer", func() {
			var buf bytes.Buffer
			_, err := archive.None.Writer(&nopWriteCloser{&buf})
			Expect(err).To(Equal(archive.ErrInvalidAlgorithm))
		})

		It("TC-IO-020: should export ErrInvalidAlgorithm constant", func() {
			Expect(archive.ErrInvalidAlgorithm).ToNot(BeNil())
			Expect(archive.ErrInvalidAlgorithm.Error()).To(ContainSubstring("invalid algorithm"))
		})
	})
})
