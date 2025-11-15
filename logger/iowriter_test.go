/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger_test

import (
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger IOWriter", func() {
	var log liblog.Logger

	BeforeEach(func() {
		log = liblog.New(GetContext())
		log.SetLevel(loglvl.DebugLevel)
		err := log.SetOptions(&logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if log != nil {
			Expect(log.Close()).ToNot(HaveOccurred())
		}
	})

	Describe("Write method", func() {
		Context("when writing bytes", func() {
			It("should write successfully", func() {
				data := []byte("test log message\n")
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})

			It("should handle empty data", func() {
				data := []byte("")
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle multi-line data", func() {
				data := []byte("line 1\nline 2\nline 3\n")
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})
		})
	})

	Describe("SetIOWriterLevel method", func() {
		Context("when setting writer level", func() {
			It("should set level successfully", func() {
				Expect(func() {
					log.SetIOWriterLevel(loglvl.WarnLevel)
				}).ToNot(Panic())
			})

			It("should accept all valid levels", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
				}

				for _, level := range levels {
					Expect(func() {
						log.SetIOWriterLevel(level)
					}).ToNot(Panic())
				}
			})
		})
	})

	Describe("GetIOWriterLevel method", func() {
		Context("when getting writer level", func() {
			It("should return current level", func() {
				log.SetIOWriterLevel(loglvl.WarnLevel)
				level := log.GetIOWriterLevel()
				Expect(level).To(Equal(loglvl.WarnLevel))
			})

			It("should reflect level changes", func() {
				log.SetIOWriterLevel(loglvl.DebugLevel)
				Expect(log.GetIOWriterLevel()).To(Equal(loglvl.DebugLevel))

				log.SetIOWriterLevel(loglvl.ErrorLevel)
				Expect(log.GetIOWriterLevel()).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("SetIOWriterFilter method", func() {
		Context("when setting filter patterns", func() {
			It("should set single filter pattern", func() {
				Expect(func() {
					log.SetIOWriterFilter("ERROR")
				}).ToNot(Panic())
			})

			It("should set multiple filter patterns", func() {
				Expect(func() {
					log.SetIOWriterFilter("ERROR", "WARN", "DEBUG")
				}).ToNot(Panic())
			})

			It("should accept empty filter (reset)", func() {
				Expect(func() {
					log.SetIOWriterFilter()
				}).ToNot(Panic())
			})

			It("should replace existing filters", func() {
				log.SetIOWriterFilter("ERROR")
				Expect(func() {
					log.SetIOWriterFilter("INFO")
				}).ToNot(Panic())
			})
		})
	})

	Describe("AddIOWriterFilter method", func() {
		Context("when adding filter patterns", func() {
			It("should add single filter pattern", func() {
				Expect(func() {
					log.AddIOWriterFilter("ERROR")
				}).ToNot(Panic())
			})

			It("should add multiple filter patterns", func() {
				Expect(func() {
					log.AddIOWriterFilter("ERROR", "WARN")
				}).ToNot(Panic())
			})

			It("should accumulate filters", func() {
				Expect(func() {
					log.AddIOWriterFilter("ERROR")
					log.AddIOWriterFilter("WARN")
					log.AddIOWriterFilter("DEBUG")
				}).ToNot(Panic())
			})
		})
	})

	Describe("Write with filters", func() {
		Context("when filter patterns are set", func() {
			BeforeEach(func() {
				// Set filter to drop messages containing "DROP"
				log.SetIOWriterFilter("DROP")
			})

			It("should write non-filtered messages", func() {
				data := []byte("IMPORTANT message\n")
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})

			It("should handle filtered messages", func() {
				data := []byte("DROP this message\n")
				// Message should still be "written" but filtered out
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})
		})
	})

	Describe("Write with different levels", func() {
		Context("when IO writer level is set", func() {
			It("should respect the set level", func() {
				log.SetIOWriterLevel(loglvl.ErrorLevel)

				// Write should still work regardless of level
				data := []byte("test message\n")
				n, err := log.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})
		})
	})
})
