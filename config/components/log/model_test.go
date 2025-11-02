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
 *
 */

package log_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Model tests verify the Log, Level, Field and Options management methods.
var _ = Describe("Model Methods", func() {
	var (
		cpt CptLog
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, DefaultLevel)
	})

	Describe("Log method", func() {
		Context("logger instance", func() {
			It("should return nil for unstarted component", func() {
				logger := cpt.Log()
				Expect(logger).To(BeNil())
			})

			It("should be callable without panic", func() {
				Expect(func() {
					_ = cpt.Log()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Level management", func() {
		Context("SetLevel method", func() {
			It("should set log level", func() {
				cpt.SetLevel(loglvl.DebugLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.DebugLevel))
			})

			It("should accept all log levels", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
					loglvl.NilLevel,
				}

				for _, level := range levels {
					cpt.SetLevel(level)
					Expect(cpt.GetLevel()).To(Equal(level))
				}
			})

			It("should not panic when setting level", func() {
				Expect(func() {
					cpt.SetLevel(loglvl.ErrorLevel)
				}).NotTo(Panic())
			})

			It("should allow changing level multiple times", func() {
				cpt.SetLevel(loglvl.DebugLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.DebugLevel))

				cpt.SetLevel(loglvl.ErrorLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.ErrorLevel))

				cpt.SetLevel(loglvl.InfoLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("GetLevel method", func() {
			It("should return default level initially", func() {
				Expect(cpt.GetLevel()).To(Equal(DefaultLevel))
			})

			It("should return level set by SetLevel", func() {
				cpt.SetLevel(loglvl.WarnLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.WarnLevel))
			})

			It("should be consistent across multiple calls", func() {
				cpt.SetLevel(loglvl.DebugLevel)
				level1 := cpt.GetLevel()
				level2 := cpt.GetLevel()
				Expect(level1).To(Equal(level2))
			})
		})
	})

	Describe("Field management", func() {
		Context("SetField method", func() {
			It("should not panic when setting fields on unstarted component", func() {
				fields := logfld.New(ctx)
				Expect(func() {
					cpt.SetField(fields)
				}).NotTo(Panic())
			})

			It("should accept nil context fields", func() {
				Expect(func() {
					cpt.SetField(nil)
				}).NotTo(Panic())
			})
		})

		Context("GetField method", func() {
			It("should return fields instance", func() {
				fields := cpt.GetField()
				Expect(fields).NotTo(BeNil())
			})

			It("should not panic when getting fields", func() {
				Expect(func() {
					_ = cpt.GetField()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Options management", func() {
		Context("SetOptions method", func() {
			It("should return error for unstarted component", func() {
				err := cpt.SetOptions(nil)
				Expect(err).To(HaveOccurred())
			})

			It("should not panic when setting options", func() {
				Expect(func() {
					_ = cpt.SetOptions(nil)
				}).NotTo(Panic())
			})
		})

		Context("GetOptions method", func() {
			It("should return nil for unstarted component", func() {
				opts := cpt.GetOptions()
				Expect(opts).To(BeNil())
			})

			It("should not panic when getting options", func() {
				Expect(func() {
					_ = cpt.GetOptions()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("thread-safety of level operations", func() {
			It("should handle concurrent GetLevel calls", func() {
				done := make(chan bool, 10)

				cpt.SetLevel(loglvl.InfoLevel)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						level := cpt.GetLevel()
						Expect(level).To(Equal(loglvl.InfoLevel))
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent SetLevel calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						if idx%2 == 0 {
							cpt.SetLevel(loglvl.DebugLevel)
						} else {
							cpt.SetLevel(loglvl.InfoLevel)
						}
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}

				// Level should be one of the set values
				level := cpt.GetLevel()
				Expect(level).To(SatisfyAny(
					Equal(loglvl.DebugLevel),
					Equal(loglvl.InfoLevel),
				))
			})

			It("should handle concurrent mixed operations", func() {
				done := make(chan bool, 20)

				for i := 0; i < 20; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						switch idx % 4 {
						case 0:
							cpt.SetLevel(loglvl.DebugLevel)
						case 1:
							_ = cpt.GetLevel()
						case 2:
							_ = cpt.Log()
						case 3:
							_ = cpt.GetField()
						}
						done <- true
					}(i)
				}

				for i := 0; i < 20; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Level initialization", func() {
		Context("creating component with different levels", func() {
			It("should initialize with PanicLevel", func() {
				c := New(ctx, loglvl.PanicLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.PanicLevel))
			})

			It("should initialize with FatalLevel", func() {
				c := New(ctx, loglvl.FatalLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.FatalLevel))
			})

			It("should initialize with ErrorLevel", func() {
				c := New(ctx, loglvl.ErrorLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.ErrorLevel))
			})

			It("should initialize with WarnLevel", func() {
				c := New(ctx, loglvl.WarnLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.WarnLevel))
			})

			It("should initialize with InfoLevel", func() {
				c := New(ctx, loglvl.InfoLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.InfoLevel))
			})

			It("should initialize with DebugLevel", func() {
				c := New(ctx, loglvl.DebugLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.DebugLevel))
			})

			It("should initialize with NilLevel", func() {
				c := New(ctx, loglvl.NilLevel)
				Expect(c.GetLevel()).To(Equal(loglvl.NilLevel))
			})
		})
	})
})
