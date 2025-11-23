/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package static_test

import (
	"io"
	"os"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PathFile Operations", func() {
	var handler any

	BeforeEach(func() {
		handler = newTestStatic()
	})

	Describe("Has", func() {
		Context("when file exists", func() {
			It("should return true for existing file", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("testdata/test.txt")).To(BeTrue())
			})

			It("should return true for existing nested file", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("testdata/subdir/nested.txt")).To(BeTrue())
			})

			It("should return true for index.html", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("testdata/index.html")).To(BeTrue())
			})
		})

		Context("when file does not exist", func() {
			It("should return false", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("testdata/nonexistent.txt")).To(BeFalse())
			})

			It("should return false for empty path", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("")).To(BeFalse())
			})

			It("should return false for directory that does not exist", func() {
				h := handler.(interface{ Has(string) bool })
				Expect(h.Has("testdata/fakedir/file.txt")).To(BeFalse())
			})
		})
	})

	Describe("List", func() {
		Context("when listing files", func() {
			It("should list all files in root", func() {
				h := handler.(staticList)
				files, err := h.List("testdata")
				Expect(err).ToNot(HaveOccurred())
				Expect(files).ToNot(BeEmpty())
				Expect(files).To(ContainElement("testdata/test.txt"))
				Expect(files).To(ContainElement("testdata/index.html"))
			})

			It("should list files in subdirectory", func() {
				h := handler.(staticList)
				files, err := h.List("testdata/subdir")
				Expect(err).ToNot(HaveOccurred())
				Expect(files).ToNot(BeEmpty())
				Expect(files).To(ContainElement("testdata/subdir/nested.txt"))
			})

			It("should list all files recursively from empty path", func() {
				h := handler.(staticList)
				files, err := h.List("")
				Expect(err).ToNot(HaveOccurred())
				Expect(files).ToNot(BeEmpty())
			})

			It("should not include directories in list", func() {
				h := handler.(staticList)
				files, err := h.List("testdata")
				Expect(err).ToNot(HaveOccurred())
				for _, f := range files {
					Expect(f).ToNot(HaveSuffix("/"))
				}
			})
		})

		Context("when directory does not exist", func() {
			It("should return an error", func() {
				h := handler.(staticList)
				_, err := h.List("testdata/nonexistent")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Find", func() {
		Context("when file exists", func() {
			It("should return file contents", func() {
				h := handler.(staticFind)
				r, err := h.Find("testdata/test.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())

				defer r.Close()

				content, err := io.ReadAll(r)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("This is a test file"))
			})

			It("should return JSON file contents", func() {
				h := handler.(staticFind)
				r, err := h.Find("testdata/test.json")
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())

				defer r.Close()

				content, err := io.ReadAll(r)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("test json file"))
			})

			It("should return nested file contents", func() {
				h := handler.(staticFind)
				r, err := h.Find("testdata/subdir/nested.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())

				defer r.Close()

				content, err := io.ReadAll(r)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("nested test file"))
			})
		})

		Context("when file does not exist", func() {
			It("should return an error", func() {
				h := handler.(staticFind)
				_, err := h.Find("testdata/nonexistent.txt")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when path is empty", func() {
			It("should return an error", func() {
				h := handler.(staticFind)
				_, err := h.Find("")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Info", func() {
		Context("when file exists", func() {
			It("should return file info", func() {
				h := handler.(staticInfo)
				info, err := h.Info("testdata/test.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
				Expect(info.IsDir()).To(BeFalse())
				Expect(info.Size()).To(BeNumerically(">", 0))
			})

			It("should return correct size", func() {
				h := handler.(staticInfo)
				info, err := h.Info("testdata/large.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Size()).To(BeNumerically(">", 100))
			})
		})

		Context("when file does not exist", func() {
			It("should return an error", func() {
				h := handler.(staticInfo)
				_, err := h.Info("testdata/nonexistent.txt")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when path is empty", func() {
			It("should return an error", func() {
				h := handler.(staticInfo)
				_, err := h.Info("")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Temp", func() {
		Context("when creating temp file", func() {
			It("should create temp file from embedded file", func() {
				h := handler.(staticTemp)
				tmp, err := h.Temp("testdata/test.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(tmp).ToNot(BeNil())

				defer func() {
					// Clean up temp file
					if tmp.IsTemp() {
						_ = tmp.CloseDelete()
					} else {
						_ = tmp.Close()
					}
				}()

				content, err := io.ReadAll(tmp)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("This is a test file"))
			})

			It("should create temp file from large file", func() {
				h := handler.(staticTemp)
				tmp, err := h.Temp("testdata/large.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(tmp).ToNot(BeNil())

				defer func() {
					if tmp.IsTemp() {
						_ = tmp.CloseDelete()
					} else {
						_ = tmp.Close()
					}
				}()

				content, err := io.ReadAll(tmp)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(content)).To(BeNumerically(">", 100))
			})
		})

		Context("when file does not exist", func() {
			It("should return an error", func() {
				h := handler.(staticTemp)
				_, err := h.Temp("testdata/nonexistent.txt")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when path is empty", func() {
			It("should return an error", func() {
				h := handler.(staticTemp)
				_, err := h.Temp("")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Map", func() {
		Context("when mapping over files", func() {
			It("should call function for each file", func() {
				h := handler.(staticMap)
				count := 0
				files := []string{}
				err := h.Map(func(pathFile string, inf os.FileInfo) error {
					count++
					files = append(files, pathFile)
					Expect(inf).ToNot(BeNil())
					Expect(inf.IsDir()).To(BeFalse())
					return nil
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(BeNumerically(">", 0))
				Expect(slices.Contains(files, "testdata/test.txt")).To(BeTrue())
			})

			It("should stop on error", func() {
				h := handler.(staticMap)
				count := 0
				err := h.Map(func(pathFile string, inf os.FileInfo) error {
					count++
					if count == 2 {
						return io.EOF
					}
					return nil
				})

				Expect(err).To(HaveOccurred())
				Expect(count).To(Equal(2))
			})

			It("should provide correct file info", func() {
				h := handler.(staticMap)
				err := h.Map(func(pathFile string, inf os.FileInfo) error {
					Expect(inf.Name()).ToNot(BeEmpty())
					Expect(inf.Size()).To(BeNumerically(">=", 0))
					return nil
				})

				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("UseTempForFileSize", func() {
		Context("when setting size threshold", func() {
			It("should use temp files for large files", func() {
				h := handler.(staticFindTempSize)

				// Set threshold to 10 bytes
				h.UseTempForFileSize(10)

				// Small file should be buffered
				r, err := h.Find("testdata/test.json")
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())
				r.Close()

				// Large file should use temp
				r, err = h.Find("testdata/large.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())
				defer r.Close()

				// Clean up temp file if created
				if namer, ok := r.(interface{ Name() string }); ok {
					defer os.Remove(namer.Name())
				}
			})

			It("should accept zero size", func() {
				h := handler.(staticTempSize)
				h.UseTempForFileSize(0)
			})

			It("should accept large size", func() {
				h := handler.(staticTempSize)
				h.UseTempForFileSize(1024 * 1024 * 100) // 100MB
			})
		})
	})
})
