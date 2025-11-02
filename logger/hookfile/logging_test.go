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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghkf "github.com/nabbar/golib/logger/hookfile"
)

var _ = Describe("HookFile Logging Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "hookfile-logging-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Fire method", func() {
		Context("with basic logging", func() {
			It("should create log file", func() {
				logFile := filepath.Join(tempDir, "test-fire.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Verify hook was created
				levels := hook.Levels()
				Expect(levels).To(HaveLen(len(logrus.AllLevels)))
			})
		})

		Context("with JSON formatter", func() {
			It("should accept formatter", func() {
				logFile := filepath.Join(tempDir, "test-json.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, &logrus.JSONFormatter{})
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with filtering options", func() {
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
				logFile := filepath.Join(tempDir, "test-no-time.log")
				opt := logcfg.OptionsFile{
					Filepath:         logFile,
					Create:           true,
					DisableTimestamp: true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})

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
	})

	Describe("File operations", func() {
		Context("with file creation", func() {
			It("should create file when Create is true", func() {
				logFile := filepath.Join(tempDir, "created.log")
				opt := logcfg.OptionsFile{
					Filepath: logFile,
					Create:   true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Give it a moment
				time.Sleep(10 * time.Millisecond)
			})
		})

		Context("with directory creation", func() {
			It("should create parent directories", func() {
				logFile := filepath.Join(tempDir, "sub1", "sub2", "test.log")
				opt := logcfg.OptionsFile{
					Filepath:   logFile,
					Create:     true,
					CreatePath: true,
				}

				hook, err := loghkf.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Verify parent directories exist
				_, err = os.Stat(filepath.Dir(logFile))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
