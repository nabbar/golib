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
	"os"

	. "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Progress Creation", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("New", func() {
		It("should create a new progress file", func() {
			path := tempDir + "/test-new.txt"
			p, err := New(path, os.O_CREATE|os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			defer p.Close()

			// Verify the file was created
			Expect(p.Path()).To(Equal(path))
		})

		It("should return error for invalid path", func() {
			_, err := New("/invalid/path/that/doesnt/exist/file.txt", os.O_RDONLY, 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should handle various flags", func() {
			path := tempDir + "/test-flags.txt"
			p, err := New(path, os.O_CREATE|os.O_WRONLY, 0644)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			defer p.Close()
		})
	})

	Describe("Open", func() {
		It("should open an existing file", func() {
			// Create a file first
			path := tempDir + "/test-open.txt"
			err := os.WriteFile(path, []byte("test content"), 0644)
			Expect(err).ToNot(HaveOccurred())

			// Open it
			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			defer p.Close()

			Expect(p.Path()).To(Equal(path))
		})

		It("should return error for non-existent file", func() {
			_, err := Open(tempDir + "/non-existent.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Create", func() {
		It("should create a new file", func() {
			path := tempDir + "/test-create.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			defer p.Close()

			// Verify file exists
			_, err = os.Stat(path)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should truncate existing file", func() {
			path := tempDir + "/test-truncate.txt"

			// Create file with content
			err := os.WriteFile(path, []byte("original content"), 0644)
			Expect(err).ToNot(HaveOccurred())

			// Create (which truncates)
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Verify file is empty
			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(0)))
		})
	})

	Describe("Temp", func() {
		It("should create a temporary file", func() {
			p, err := Temp("progress-test-*.tmp")
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())

			path := p.Path()
			p.Close()

			// Verify file was created in temp directory
			Expect(path).ToNot(BeEmpty())

			// Clean up
			os.Remove(path)
		})

		It("should create unique temporary files", func() {
			p1, err := Temp("test-*.tmp")
			Expect(err).ToNot(HaveOccurred())
			defer p1.Close()
			defer os.Remove(p1.Path())

			p2, err := Temp("test-*.tmp")
			Expect(err).ToNot(HaveOccurred())
			defer p2.Close()
			defer os.Remove(p2.Path())

			// Paths should be different
			Expect(p1.Path()).ToNot(Equal(p2.Path()))
		})
	})

	Describe("Unique", func() {
		It("should create a unique file in specified directory", func() {
			p, err := Unique(tempDir, "unique-*.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			defer p.Close()

			// Verify path is in the specified directory
			path := p.Path()
			Expect(path).To(ContainSubstring(tempDir))
		})

		It("should create unique files with same pattern", func() {
			p1, err := Unique(tempDir, "pattern-*.dat")
			Expect(err).ToNot(HaveOccurred())
			defer p1.Close()

			p2, err := Unique(tempDir, "pattern-*.dat")
			Expect(err).ToNot(HaveOccurred())
			defer p2.Close()

			// Should have different names
			Expect(p1.Path()).ToNot(Equal(p2.Path()))
		})

		It("should return error for invalid directory", func() {
			_, err := Unique("/invalid/dir/path", "test-*.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Interface Compliance", func() {
		It("should implement Progress interface", func() {
			path := tempDir + "/interface-test.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Verify it implements Progress
			var _ Progress = p
		})
	})
})
