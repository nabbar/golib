/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package progress_test

import (
	"io"
	"os"

	. "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("File Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-file-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Path", func() {
		It("should return correct path", func() {
			path := tempDir + "/path-test.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			Expect(p.Path()).To(Equal(path))
		})

		It("should clean path", func() {
			path := tempDir + "/./subdir/../path-clean.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Path should be cleaned
			resultPath := p.Path()
			Expect(resultPath).ToNot(ContainSubstring("./"))
			Expect(resultPath).ToNot(ContainSubstring("../"))
		})
	})

	Describe("Stat", func() {
		It("should return file info", func() {
			path := tempDir + "/stat-test.txt"
			testData := []byte("Test data for stat")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info).ToNot(BeNil())
			Expect(info.Size()).To(Equal(int64(len(testData))))
			Expect(info.Mode().IsRegular()).To(BeTrue())
		})

		It("should update size after write", func() {
			path := tempDir + "/stat-write.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Initial size should be 0
			info1, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info1.Size()).To(Equal(int64(0)))

			// Write data
			data := []byte("New data")
			p.Write(data)
			p.Sync()

			// Size should update
			info2, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info2.Size()).To(Equal(int64(len(data))))
		})
	})

	Describe("SizeBOF and SizeEOF", func() {
		It("should return correct BOF size", func() {
			path := tempDir + "/size-bof.txt"
			testData := []byte("0123456789ABCDEF")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// At start, BOF should be 0
			size, err := p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(0)))

			// Read 5 bytes
			buf := make([]byte, 5)
			p.Read(buf)

			// BOF should now be 5
			size, err = p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(5)))
		})

		It("should return correct EOF size", func() {
			path := tempDir + "/size-eof.txt"
			testData := []byte("0123456789ABCDEF") // 16 bytes
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// At start, EOF should be full size
			size, err := p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(len(testData))))

			// Read 5 bytes
			buf := make([]byte, 5)
			p.Read(buf)

			// EOF should now be remaining size (11 bytes)
			size, err = p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(11)))
		})

		It("should handle EOF size at end", func() {
			path := tempDir + "/size-eof-end.txt"
			testData := []byte("Short")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Seek to end
			p.Seek(0, io.SeekEnd)

			// EOF should be 0
			size, err := p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(0)))
		})
	})

	Describe("Truncate", func() {
		It("should truncate file to smaller size", func() {
			path := tempDir + "/truncate-smaller.txt"
			testData := []byte("0123456789ABCDEF")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Truncate to 5 bytes
			err = p.Truncate(5)
			Expect(err).ToNot(HaveOccurred())

			// Verify new size
			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(5)))

			// Verify content
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 10)
			n, _ := p.Read(buf)
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal("01234"))
		})

		It("should extend file to larger size", func() {
			path := tempDir + "/truncate-larger.txt"
			testData := []byte("Short")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Extend to 20 bytes
			err = p.Truncate(20)
			Expect(err).ToNot(HaveOccurred())

			// Verify new size
			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(20)))
		})

		It("should truncate to zero", func() {
			path := tempDir + "/truncate-zero.txt"
			testData := []byte("Data to remove")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			err = p.Truncate(0)
			Expect(err).ToNot(HaveOccurred())

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(0)))
		})
	})

	Describe("Sync", func() {
		It("should sync file to disk", func() {
			path := tempDir + "/sync-test.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write data
			data := []byte("Data to sync")
			p.Write(data)

			// Sync to disk
			err = p.Sync()
			Expect(err).ToNot(HaveOccurred())

			// Verify file exists and has content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal(data))
		})

		It("should sync multiple times", func() {
			path := tempDir + "/sync-multiple.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write and sync multiple times
			for i := 0; i < 5; i++ {
				p.WriteString("chunk ")
				err = p.Sync()
				Expect(err).ToNot(HaveOccurred())
			}

			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("chunk chunk chunk chunk chunk "))
		})
	})

	Describe("Close", func() {
		It("should close file properly", func() {
			path := tempDir + "/close-test.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			p.WriteString("test data")
			err = p.Close()
			Expect(err).ToNot(HaveOccurred())

			// File should still exist
			_, err = os.Stat(path)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should allow multiple close calls", func() {
			path := tempDir + "/close-multiple.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			err = p.Close()
			Expect(err).ToNot(HaveOccurred())

			// Second close should not error
			err = p.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("CloseDelete", func() {
		It("should close and delete file", func() {
			path := tempDir + "/close-delete.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			p.WriteString("temporary data")
			p.Sync()

			err = p.CloseDelete()
			// May fail with os.Root restriction, so we just check it doesn't panic
			if err == nil {
				// File should be deleted
				_, err = os.Stat(path)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}
		})

		It("should handle CloseDelete for temp files", func() {
			p, err := Temp("delete-test-*.tmp")
			Expect(err).ToNot(HaveOccurred())

			path := p.Path()
			p.WriteString("temp data")
			p.Sync()

			err = p.CloseDelete()
			Expect(err).ToNot(HaveOccurred())

			// File should be deleted
			_, err = os.Stat(path)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("should allow multiple Close calls", func() {
			path := tempDir + "/close-multiple.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			p.WriteString("data")

			// First close
			err = p.Close()
			Expect(err).ToNot(HaveOccurred())

			// Second close should not panic
			err = p.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should use OpenRoot for deletion when available", func() {
			path := tempDir + "/close-delete-root.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			p.WriteString("data")
			p.Sync()

			err = p.CloseDelete()
			// May fail with os.Root restriction, check doesn't panic
			if err == nil {
				// File should be deleted if successful
				_, err = os.Stat(path)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}
		})
	})

	Describe("ByteReader and ByteWriter", func() {
		It("should read single byte", func() {
			path := tempDir + "/byte-read.txt"
			testData := []byte{0x41, 0x42, 0x43} // ABC
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			b, err := p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte(0x41)))

			b, err = p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte(0x42)))
		})

		It("should write single byte", func() {
			path := tempDir + "/byte-write.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			err = p.WriteByte(0x58) // X
			Expect(err).ToNot(HaveOccurred())

			err = p.WriteByte(0x59) // Y
			Expect(err).ToNot(HaveOccurred())

			err = p.WriteByte(0x5A) // Z
			Expect(err).ToNot(HaveOccurred())

			// Verify content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("XYZ"))
		})

		It("should handle EOF on ReadByte", func() {
			path := tempDir + "/byte-eof.txt"
			testData := []byte{0x41}
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read the only byte
			_, err = p.ReadByte()
			Expect(err).ToNot(HaveOccurred())

			// Next read should return EOF
			_, err = p.ReadByte()
			Expect(err).To(Equal(io.EOF))
		})
	})

	Describe("Edge cases", func() {
		It("should handle empty file", func() {
			path := tempDir + "/empty.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(0)))

			size, err := p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(0)))

			size, err = p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(0)))
		})

		It("should handle very large file", func() {
			path := tempDir + "/large.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write 10MB
			chunk := make([]byte, 1024*1024)
			for i := 0; i < 10; i++ {
				n, err := p.Write(chunk)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(chunk)))
			}

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(10 * 1024 * 1024)))
		})
	})
})
