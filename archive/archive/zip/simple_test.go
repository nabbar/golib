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

var _ = Describe("TC-SM-001: Simple Coverage Tests", func() {
	Describe("TC-SM-002: Reader All Methods", func() {
		It("TC-SM-003: should test all Reader methods", func() {
			// Create archive in memory
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Add multiple files
			content1 := []byte("test content 1")
			info1 := createTestFileInfo("test1.txt", int64(len(content1)))
			reader1 := io.NopCloser(bytes.NewReader(content1))
			writer.Add(info1, reader1, "", "")

			content2 := []byte("test content 2")
			info2 := createTestFileInfo("dir/test2.txt", int64(len(content2)))
			reader2 := io.NopCloser(bytes.NewReader(content2))
			writer.Add(info2, reader2, "", "")

			writer.Close()

			// Read archive
			zipReader := newReaderWithSize(buf.Bytes())
			r, err := zip.NewReader(zipReader)
			Expect(err).ToNot(HaveOccurred())
			defer r.Close()

			// Test List
			files, err := r.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(files)).To(Equal(2))

			// Test Has
			Expect(r.Has("test1.txt")).To(BeTrue())
			Expect(r.Has("missing.txt")).To(BeFalse())

			// Test Info
			info, err := r.Info("test1.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Name()).To(Equal("test1.txt"))

			infoMissing, err := r.Info("missing.txt")
			Expect(err).To(HaveOccurred())
			Expect(infoMissing).To(BeNil())

			// Test Get
			rc, err := r.Get("test1.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(rc).ToNot(BeNil())
			content, _ := io.ReadAll(rc)
			rc.Close()
			Expect(string(content)).To(Equal("test content 1"))

			rcMissing, err := r.Get("missing.txt")
			Expect(err).To(HaveOccurred())
			Expect(rcMissing).To(BeNil())

			// Test Walk
			count := 0
			r.Walk(func(info os.FileInfo, rc io.ReadCloser, path, link string) bool {
				count++
				Expect(link).To(BeEmpty())
				if rc != nil {
					rc.Close()
				}
				return true
			})
			Expect(count).To(Equal(2))

			// Test Walk early termination
			countEarly := 0
			r.Walk(func(info os.FileInfo, rc io.ReadCloser, path, link string) bool {
				countEarly++
				if rc != nil {
					rc.Close()
				}
				return countEarly < 1
			})
			Expect(countEarly).To(Equal(1))
		})
	})

	Describe("TC-SM-004: Writer FromPath and addFiltering Coverage", func() {
		It("TC-SM-005: should test FromPath with all filtering paths", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file1.txt": "content1",
				"file2.log": "content2",
				"file3.txt": "content3",
				"data.csv":  "csv data",
			})
			defer os.RemoveAll(testDir)

			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Test 1: Filter with pattern (only .txt files)
			err := writer.FromPath(testDir, "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			// Test 2: Filter with different pattern
			err = writer.FromPath(testDir, "*.log", nil)
			Expect(err).ToNot(HaveOccurred())

			// Test 3: Empty filter (defaults to "*")
			err = writer.FromPath(testDir, "", nil)
			Expect(err).ToNot(HaveOccurred())

			// Test 4: Path replacement function
			replaceFn := func(source string) string {
				return "renamed/" + filepath.Base(source)
			}
			err = writer.FromPath(testDir, "*.csv", replaceFn)
			Expect(err).ToNot(HaveOccurred())

			// Test 5: Single file (not directory)
			singleFile := filepath.Join(testDir, "file1.txt")
			err = writer.FromPath(singleFile, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-SM-006: should handle directories and non-matching patterns", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file.txt": "content",
			})
			defer os.RemoveAll(testDir)

			// Create subdirectory (should be skipped)
			subdir := filepath.Join(testDir, "subdir")
			os.Mkdir(subdir, 0755)

			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Walk will encounter the directory and skip it
			err := writer.FromPath(testDir, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			// Test non-matching filter (should skip files)
			err = writer.FromPath(testDir, "*.xyz", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()
		})

		It("TC-SM-007: should handle filter pattern errors", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"file.txt": "content",
			})
			defer os.RemoveAll(testDir)

			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Invalid glob pattern
			err := writer.FromPath(testDir, "[invalid", nil)
			Expect(err).To(HaveOccurred())

			writer.Close()
		})
	})

	Describe("TC-SM-007: Writer Close and Add Coverage", func() {
		It("TC-SM-008: should handle all Add and Close paths", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Add with nil reader (coverage for early return)
			info0 := createTestFileInfo("test.txt", 0)
			err := writer.Add(info0, nil, "", "")
			Expect(err).ToNot(HaveOccurred())

			// Add with custom path
			content := []byte("test content")
			info1 := createTestFileInfo("original.txt", int64(len(content)))
			reader1 := io.NopCloser(bytes.NewReader(content))
			err = writer.Add(info1, reader1, "custom/path.txt", "")
			Expect(err).ToNot(HaveOccurred())

			// Add with empty custom path
			info2 := createTestFileInfo("test2.txt", int64(len(content)))
			reader2 := io.NopCloser(bytes.NewReader(content))
			err = writer.Add(info2, reader2, "", "")
			Expect(err).ToNot(HaveOccurred())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-SM-009: NewReader All Error Paths", func() {
		It("TC-SM-010: should cover all NewReader validation branches", func() {
			// Branch 1: No Size method
			invalidReader := io.NopCloser(bytes.NewReader([]byte{}))
			_, err := zip.NewReader(invalidReader)
			Expect(err).To(HaveOccurred())

			// Branch 2: Zero size
			zeroReader := &readerWithSize{Reader: bytes.NewReader([]byte{}), size: 0}
			_, err = zip.NewReader(zeroReader)
			Expect(err).To(HaveOccurred())

			// Branch 3: Negative size
			negReader := &readerWithSize{Reader: bytes.NewReader([]byte{}), size: -1}
			_, err = zip.NewReader(negReader)
			Expect(err).To(HaveOccurred())

			// Branch 4: Valid size, successful path (already tested above)
			// Create a valid small archive to test successful NewReader
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)
			content := []byte("test")
			info := createTestFileInfo("test.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))
			writer.Add(info, reader, "", "")
			writer.Close()

			zipReader := newReaderWithSize(buf.Bytes())
			r, err := zip.NewReader(zipReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeNil())
			r.Close()
		})
	})

	Describe("TC-SM-011: Complete Integration Tests", func() {
		It("TC-SM-012: should verify full write-read cycle", func() {
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Write multiple files with various scenarios
			files := map[string]string{
				"file1.txt":     "content 1",
				"dir/file2.txt": "content 2",
				"empty.txt":     "",
			}

			for name, content := range files {
				info := createTestFileInfo(name, int64(len(content)))
				reader := io.NopCloser(bytes.NewReader([]byte(content)))
				writer.Add(info, reader, "", "")
			}

			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Read and verify
			zipReader := newReaderWithSize(buf.Bytes())
			r, err := zip.NewReader(zipReader)
			Expect(err).ToNot(HaveOccurred())
			defer r.Close()

			list, _ := r.List()
			Expect(len(list)).To(Equal(len(files)))

			for name, expectedContent := range files {
				Expect(r.Has(name)).To(BeTrue())

				info, err := r.Info(name)
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())

				rc, err := r.Get(name)
				Expect(err).ToNot(HaveOccurred())
				readContent, _ := io.ReadAll(rc)
				rc.Close()
				Expect(string(readContent)).To(Equal(expectedContent))
			}
		})
	})

	Describe("TC-SM-013: Extensive addFiltering Coverage", func() {
		It("TC-SM-014: should cover all addFiltering code paths", func() {
			testDir, _ := createTestDirectory(map[string]string{
				"match1.txt":  "content1",
				"match2.txt":  "content2",
				"nomatch.log": "logcontent",
			})
			defer os.RemoveAll(testDir)

			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)

			// Path 1: Matching filter with nil fct
			err := writer.FromPath(testDir, "match*.txt", nil)
			Expect(err).ToNot(HaveOccurred())

			// Path 2: Non-matching filter (files will be skipped)
			err = writer.FromPath(testDir, "*.xyz", nil)
			Expect(err).ToNot(HaveOccurred())

			// Path 3: Empty filter (defaults to "*")
			err = writer.FromPath(testDir, "", nil)
			Expect(err).ToNot(HaveOccurred())

			// Path 4: With replacement function
			replaceFn := func(source string) string {
				return "custom/" + filepath.Base(source)
			}
			err = writer.FromPath(testDir, "*.log", replaceFn)
			Expect(err).ToNot(HaveOccurred())

			// Path 5: Single file path (not a directory)
			singleFile := filepath.Join(testDir, "match1.txt")
			err = writer.FromPath(singleFile, "*", nil)
			Expect(err).ToNot(HaveOccurred())

			writer.Close()

			// Verify archive created successfully
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Describe("TC-SM-015: NewReader Seek and Error Paths", func() {
		It("TC-SM-016: should test NewReader with valid archive", func() {
			// Create a valid archive
			buf := newBufferWriteCloser()
			writer, _ := zip.NewWriter(buf)
			content := []byte("test data for reader")
			info := createTestFileInfo("data.txt", int64(len(content)))
			reader := io.NopCloser(bytes.NewReader(content))
			writer.Add(info, reader, "", "")
			writer.Close()

			// Test successful NewReader with valid data
			zipReader := newReaderWithSize(buf.Bytes())
			r, err := zip.NewReader(zipReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeNil())

			// Use the reader to ensure all code paths
			list, err := r.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))

			r.Close()
		})
	})
})
