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
 */

package tty_test

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TTY Input Types", func() {
	Describe("Different input sources", func() {
		Context("with nil input", func() {
			It("should default to os.Stdin", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should handle nil input with signals", func() {
				saver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with standard file descriptors", func() {
			It("should accept os.Stdin", func() {
				saver, err := tty.New(os.Stdin, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should accept os.Stdout", func() {
				saver, err := tty.New(os.Stdout, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should accept os.Stderr", func() {
				saver, err := tty.New(os.Stderr, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with file inputs", func() {
			var tempFile *os.File

			BeforeEach(func() {
				var err error
				tempFile, err = os.CreateTemp("", "tty-test-*.txt")
				Expect(err).ToNot(HaveOccurred())
				_, err = tempFile.WriteString("test content\n")
				Expect(err).ToNot(HaveOccurred())
				_, err = tempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				if tempFile != nil {
					name := tempFile.Name()
					_ = tempFile.Close()
					_ = os.Remove(name)
				}
			})

			It("should handle regular file input", func() {
				saver, err := tty.New(tempFile, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				// Regular files are not terminals
				Expect(saver.IsTerminal()).To(BeFalse())
			})

			It("should not error on restore with file input", func() {
				saver, err := tty.New(tempFile, false)
				Expect(err).ToNot(HaveOccurred())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with reader without Fd() method", func() {
			It("should handle bytes.Buffer", func() {
				buf := bytes.NewBufferString("test data")
				saver, err := tty.New(buf, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				// Buffer has no Fd(), so not a terminal
				Expect(saver.IsTerminal()).To(BeFalse())
			})

			It("should handle strings.Reader", func() {
				reader := strings.NewReader("test data")
				saver, err := tty.New(reader, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				Expect(saver.IsTerminal()).To(BeFalse())
			})

			It("should handle io.Reader interface", func() {
				var reader io.Reader = strings.NewReader("test")
				saver, err := tty.New(reader, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with pipes", func() {
			It("should handle pipe input", func() {
				r, w, err := os.Pipe()
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					_ = r.Close()
					_ = w.Close()
				}()

				// Write some data to pipe
				go func() {
					_, _ = w.WriteString("test data\n")
					_ = w.Close()
				}()

				saver, err := tty.New(r, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				// Pipes are not terminals
				Expect(saver.IsTerminal()).To(BeFalse())
			})
		})
	})

	Describe("Signal handling with different inputs", func() {
		Context("signal flag variations", func() {
			It("should create without signals (false)", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should create with signals (true)", func() {
				saver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should handle signals with file input", func() {
				tempFile, err := os.CreateTemp("", "tty-sig-*.txt")
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					name := tempFile.Name()
					_ = tempFile.Close()
					_ = os.Remove(name)
				}()

				saver, err := tty.New(tempFile, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should handle signals with buffer input", func() {
				buf := bytes.NewBufferString("data")
				saver, err := tty.New(buf, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})
	})

	Describe("Restore behavior with different inputs", func() {
		Context("non-terminal inputs", func() {
			It("should restore gracefully with file", func() {
				tempFile, err := os.CreateTemp("", "tty-restore-*.txt")
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					name := tempFile.Name()
					_ = tempFile.Close()
					_ = os.Remove(name)
				}()

				saver, err := tty.New(tempFile, false)
				Expect(err).ToNot(HaveOccurred())

				// Should not error even though it's not a terminal
				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should restore gracefully with buffer", func() {
				buf := bytes.NewBufferString("test")
				saver, err := tty.New(buf, false)
				Expect(err).ToNot(HaveOccurred())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should restore multiple times with non-terminal", func() {
				reader := strings.NewReader("test")
				saver, err := tty.New(reader, false)
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < 10; i++ {
					err = saver.Restore()
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})
	})

	Describe("Edge cases with input types", func() {
		Context("closed file descriptors", func() {
			It("should handle closed file gracefully", func() {
				tempFile, err := os.CreateTemp("", "tty-closed-*.txt")
				Expect(err).ToNot(HaveOccurred())
				name := tempFile.Name()
				_ = tempFile.Close()
				defer func() {
					_ = os.Remove(name)
				}()

				// Try to open and immediately close
				f, err := os.Open(name)
				Expect(err).ToNot(HaveOccurred())
				_ = f.Close()

				// Create saver with closed file
				saver, err := tty.New(f, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("concurrent access to same input", func() {
			It("should handle concurrent New() calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						saver, err := tty.New(nil, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(saver).ToNot(BeNil())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					<-done
				}
			})
		})
	})
})
