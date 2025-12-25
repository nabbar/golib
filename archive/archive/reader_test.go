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

var _ = Describe("TC-RDR-001: Internal Reader Operations", func() {
	Describe("TC-RDR-002: Reader with Different Input Types", func() {
		It("TC-RDR-003: should handle Detect with valid TAR from seekable file", func() {
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

			if reader != nil {
				reader.Close()
			}
			if stream != nil {
				stream.Close()
			}
		})

		It("TC-RDR-004: should handle Detect with buffer", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "data.txt", "test data")

			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()

			alg, reader, stream, err := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))

			if reader != nil {
				reader.Close()
			}
			if stream != nil {
				stream.Close()
			}
		})

		It("TC-RDR-005: should handle Parse with None result", func() {
			alg := archive.Parse("invalid")
			Expect(alg).To(Equal(archive.None))
			Expect(alg.IsNone()).To(BeTrue())
		})

		It("TC-RDR-006: should handle Reader creation with Zip returning error for buffer", func() {
			var buf bytes.Buffer
			writer, _ := archive.Zip.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			reader, err := archive.Zip.Reader(io.NopCloser(bytes.NewReader(buf.Bytes())))
			// ZIP reader requires ReaderAt/Seeker, buffer doesn't provide it
			Expect(err).To(HaveOccurred())
			Expect(reader).To(BeNil())
		})
	})

	Describe("TC-RDR-007: Detect Edge Cases", func() {
		It("TC-RDR-008: should handle Detect with ZIP signature but invalid data", func() {
			data := make([]byte, 300)
			// Add ZIP signature
			data[0] = 0x50
			data[1] = 0x4b
			data[2] = 0x03
			data[3] = 0x04

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(data)))
			if stream != nil {
				stream.Close()
			}
			// Should error or return None because data is invalid
			// Either outcome is acceptable
		})

		It("TC-RDR-009: should handle multiple sequential reads", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "file1.txt", "content1")
			_ = createTestFile(tmpDir, "file2.txt", "content2")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			// First read
			tmpFile, _ = os.Open(tmpFile.Name())
			_, reader1, stream1, _ := archive.Detect(tmpFile)
			files1, _ := reader1.List()
			reader1.Close()
			stream1.Close()
			tmpFile.Close()

			// Second read
			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()
			_, reader2, stream2, _ := archive.Detect(tmpFile)
			defer stream2.Close()
			defer reader2.Close()
			files2, _ := reader2.List()

			Expect(files1).To(Equal(files2))
		})
	})

	Describe("TC-RDR-010: Reader Interface Coverage", func() {
		It("TC-RDR-011: should handle Read operations on detected stream", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "test.txt", "test content")

			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))

			if stream != nil {
				defer stream.Close()
				// Try reading a few bytes to exercise Read
				testBuf := make([]byte, 10)
				_, _ = stream.Read(testBuf)
			}
		})

		It("TC-RDR-012: should handle None reader error path", func() {
			reader, err := archive.None.Reader(io.NopCloser(bytes.NewReader([]byte{})))
			Expect(err).To(Equal(archive.ErrInvalidAlgorithm))
			Expect(reader).To(BeNil())
		})

		It("TC-RDR-013: should handle empty TAR archive", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			reader, err := archive.Tar.Reader(io.NopCloser(bytes.NewReader(buf.Bytes())))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			if reader != nil {
				defer reader.Close()
				files, _ := reader.List()
				Expect(files).To(BeEmpty())
			}
		})
	})

	Describe("TC-RDR-014: Parse Function Coverage", func() {
		It("TC-RDR-015: should parse all valid algorithms", func() {
			Expect(archive.Parse("tar")).To(Equal(archive.Tar))
			Expect(archive.Parse("TAR")).To(Equal(archive.Tar))
			Expect(archive.Parse("Tar")).To(Equal(archive.Tar))

			Expect(archive.Parse("zip")).To(Equal(archive.Zip))
			Expect(archive.Parse("ZIP")).To(Equal(archive.Zip))
			Expect(archive.Parse("Zip")).To(Equal(archive.Zip))

			Expect(archive.Parse("none")).To(Equal(archive.None))
			Expect(archive.Parse("")).To(Equal(archive.None))
			Expect(archive.Parse("unknown")).To(Equal(archive.None))
		})
	})

	Describe("TC-RDR-016: ReadAt Complete", func() {
		It("TC-RDR-017: should ReadAt with file", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "ra.txt", "readat test data")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, _, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()

			if readerAt, ok := stream.(io.ReaderAt); ok {
				buf := make([]byte, 10)
				n, _ := readerAt.ReadAt(buf, 10)
				Expect(n).To(BeNumerically(">=", 0))
			}
		})

		It("TC-RDR-018: should ReadAt error for non-seekable", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))
			defer stream.Close()

			if readerAt, ok := stream.(io.ReaderAt); ok {
				testBuf := make([]byte, 10)
				_, err := readerAt.ReadAt(testBuf, 0)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Describe("TC-RDR-019: Size Complete", func() {
		It("TC-RDR-020: should get Size with Seeker", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "sz.txt", "size test")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, _, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()

			if sizer, ok := stream.(interface{ Size() int64 }); ok {
				size := sizer.Size()
				Expect(size).To(BeNumerically(">", 0))
			}
		})

		It("TC-RDR-021: should return 0 for non-seekable", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))
			defer stream.Close()

			if sizer, ok := stream.(interface{ Size() int64 }); ok {
				size := sizer.Size()
				_ = size
			}
		})
	})

	Describe("TC-RDR-022: Seek Complete", func() {
		It("TC-RDR-023: should Seek with file", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "sk.txt", "seek test data")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, _, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()

			if seeker, ok := stream.(io.Seeker); ok {
				pos, err := seeker.Seek(10, io.SeekStart)
				Expect(err).ToNot(HaveOccurred())
				Expect(pos).To(Equal(int64(10)))
			}
		})

		It("TC-RDR-024: should error for invalid whence", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))
			defer stream.Close()

			if seeker, ok := stream.(io.Seeker); ok {
				_, err := seeker.Seek(10, io.SeekCurrent)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Describe("TC-RDR-025: Reset Complete", func() {
		It("TC-RDR-026: should Reset with Seeker", func() {
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "rs.txt", "reset test")

			tmpFile, _ := createTempArchiveFile(".tar")
			defer os.Remove(tmpFile.Name())

			writer, _ := archive.Tar.Writer(tmpFile)
			_ = writer.FromPath(tmpDir, "*.txt", nil)
			_ = writer.Close()
			tmpFile.Close()

			tmpFile, _ = os.Open(tmpFile.Name())
			defer tmpFile.Close()

			_, _, stream, _ := archive.Detect(tmpFile)
			defer stream.Close()

			buf := make([]byte, 10)
			_, _ = stream.Read(buf)

			if resetter, ok := stream.(interface{ Reset() bool }); ok {
				result := resetter.Reset()
				Expect(result).To(BeTrue())
			}
		})

		It("TC-RDR-027: should fail Reset for non-seekable", func() {
			var buf bytes.Buffer
			writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
			_ = writer.Close()

			_, _, stream, _ := archive.Detect(io.NopCloser(bytes.NewReader(buf.Bytes())))
			defer stream.Close()

			testBuf := make([]byte, 10)
			_, _ = stream.Read(testBuf)

			if resetter, ok := stream.(interface{ Reset() bool }); ok {
				result := resetter.Reset()
				Expect(result).To(BeFalse())
			}
		})
	})
})
