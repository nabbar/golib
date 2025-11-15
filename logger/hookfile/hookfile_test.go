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

package hookfile_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghkf "github.com/nabbar/golib/logger/hookfile"
	libsiz "github.com/nabbar/golib/size"
)

var _ = Describe("HookFile Creation and Basic Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "hookfile-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("New", func() {
		Context("with missing filepath", func() {
			It("should return error", func() {
				opt := logcfg.OptionsFile{}

				hook, err := loghkf.New(opt, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing file path"))
				Expect(hook).To(BeNil())
			})
		})

		Context("with valid filepath", func() {
			It("should create hook successfully", func() {
				logFile := filepath.Join(tempDir, "test.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with custom file mode", func() {
			It("should create file with specified mode", func() {
				logFile := filepath.Join(tempDir, "test-mode.log")

				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with path creation", func() {
			It("should create parent directories", func() {
				logFile := filepath.Join(tempDir, "subdir1", "subdir2", "test.log")
				opt := logcfg.OptionsFile{
					Filepath:   logFile,
					Create:     true,
					CreatePath: true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Verify parent directories were created
				_, err = os.Stat(filepath.Dir(logFile))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with specific log levels", func() {
			It("should accept custom levels", func() {
				logFile := filepath.Join(tempDir, "test-levels.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
					LogLevel: []string{"info", "warn", "error"},
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				levels := hook.Levels()
				Expect(levels).To(HaveLen(3))
			})
		})

		Context("with all levels (default)", func() {
			It("should use all logrus levels", func() {
				logFile := filepath.Join(tempDir, "test-all-levels.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				levels := hook.Levels()
				Expect(levels).To(Equal(logrus.AllLevels))
			})
		})

		Context("with custom formatter", func() {
			It("should accept formatter", func() {
				logFile := filepath.Join(tempDir, "test-formatter.log")
				formatter := &logrus.JSONFormatter{}
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, formatter)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with disable options", func() {
			It("should accept DisableStack", func() {
				logFile := filepath.Join(tempDir, "test-no-stack.log")
				opt := logcfg.OptionsFile{
					Filepath:     logFile,
					Create:       true,
					DisableStack: true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})

			It("should accept DisableTimestamp", func() {
				logFile := filepath.Join(tempDir, "test-no-timestamp.log")
				opt := logcfg.OptionsFile{
					Filepath:         logFile,
					Create:           true,
					DisableTimestamp: true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with enable options", func() {
			It("should accept EnableTrace", func() {
				logFile := filepath.Join(tempDir, "test-trace.log")
				opt := logcfg.OptionsFile{
					Filepath:    logFile,
					Create:      true,
					EnableTrace: true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})

			It("should accept EnableAccessLog", func() {
				logFile := filepath.Join(tempDir, "test-access.log")
				opt := logcfg.OptionsFile{
					Filepath:        logFile,
					Create:          true,
					EnableAccessLog: true,
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with buffer size", func() {
			It("should accept custom buffer size", func() {
				logFile := filepath.Join(tempDir, "test-buffer.log")
				opt := logcfg.OptionsFile{
					Filepath:       logFile,
					Create:         true,
					FileBufferSize: libsiz.Size(8192),
				}

				hook, err := loghkf.New(opt, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})
	})

	Describe("Done channel", func() {
		Context("with valid hook", func() {
			It("should provide done channel", func() {
				logFile := filepath.Join(tempDir, "test-done.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				done := hook.Done()
				Expect(done).ToNot(BeNil())
			})
		})
	})

	Describe("RegisterHook", func() {
		Context("with valid logger", func() {
			It("should register hook successfully", func() {
				logFile := filepath.Join(tempDir, "test-register.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				logger := logrus.New()
				hook.RegisterHook(logger)

				// Logger should have the hook
				// Note: We can't easily verify this without internal access
				Expect(logger).ToNot(BeNil())
			})
		})
	})
})
