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

package hooksyslog_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	logsys "github.com/nabbar/golib/logger/hooksyslog"
	libptc "github.com/nabbar/golib/network/protocol"
)

var _ = Describe("HookSyslog Configuration and Options", func() {
	AfterEach(func() {
		logsys.ResetOpenSyslog()
	})
	Describe("New", func() {
		Context("with minimal valid configuration", func() {
			It("should validate configuration structure", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					Host:     "", // will use localhost system syslog
					LogLevel: []string{"info"},
				}

				// Note: This will fail without actual syslog connection
				// We're testing the configuration structure
				_, err := logsys.New(opt, nil)
				// Error is expected without real syslog
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with log levels", func() {
			It("should accept custom levels", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					LogLevel: []string{"error", "fatal"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				// Error expected without syslog connection
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with formatter", func() {
			It("should accept JSON formatter", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					LogLevel: []string{"info"},
				}
				//no hostname => using localhost syslog system

				formatter := &logrus.JSONFormatter{}

				_, err := logsys.New(opt, formatter)
				// Error expected without syslog connection
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with DisableStack option", func() {
			It("should accept DisableStack", func() {
				opt := logcfg.OptionsSyslog{
					Network:      libptc.NetworkUnix.Code(),
					DisableStack: true,
					LogLevel:     []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with DisableTimestamp option", func() {
			It("should accept DisableTimestamp", func() {
				opt := logcfg.OptionsSyslog{
					Network:          libptc.NetworkUnix.Code(),
					DisableTimestamp: true,
					LogLevel:         []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with EnableTrace option", func() {
			It("should accept EnableTrace", func() {
				opt := logcfg.OptionsSyslog{
					Network:     libptc.NetworkUnix.Code(),
					EnableTrace: true,
					LogLevel:    []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with EnableAccessLog option", func() {
			It("should accept EnableAccessLog", func() {
				opt := logcfg.OptionsSyslog{
					Network:         libptc.NetworkUnix.Code(),
					EnableAccessLog: true,
					LogLevel:        []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with Tag option", func() {
			It("should accept Tag", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					Tag:      "myapp",
					LogLevel: []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Configuration validation", func() {
		Context("with various network types", func() {
			It("should accept Unix network", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					LogLevel: []string{"info"},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred()) // Expected without connection
			})

			It("should accept TCP network", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkTCP.Code(),
					Host:     "localhost:514",
					LogLevel: []string{"info"},
				}

				_, err := logsys.New(opt, nil)
				Expect(err).To(HaveOccurred()) // Expected without connection
			})

			It("should accept UDP network", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUDP.Code(),
					Host:     "localhost:514",
					LogLevel: []string{"info"},
				}

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred()) // UDP succeeds without actual connection
			})
		})

		Context("with all log levels", func() {
			It("should accept all standard levels", func() {
				opt := logcfg.OptionsSyslog{
					Network: libptc.NetworkUnix.Code(),
					LogLevel: []string{
						"panic", "fatal", "error",
						"warn", "info", "debug",
					},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with empty log levels", func() {
			It("should handle empty levels", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnix.Code(),
					LogLevel: []string{},
				}
				//no hostname => using localhost syslog system

				_, err := logsys.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
