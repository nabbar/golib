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
	"context"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	logsys "github.com/nabbar/golib/logger/hooksyslog"
	libptc "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("HookSyslog Additional Coverage Tests", func() {
	Describe("SyslogSeverity", func() {
		Context("String method", func() {
			It("should return correct string for all severities", func() {
				tests := map[logsys.SyslogSeverity]string{
					logsys.SyslogSeverityEmerg:   "EMERG",
					logsys.SyslogSeverityAlert:   "ALERT",
					logsys.SyslogSeverityCrit:    "CRIT",
					logsys.SyslogSeverityErr:     "ERR",
					logsys.SyslogSeverityWarning: "WARNING",
					logsys.SyslogSeverityNotice:  "NOTICE",
					logsys.SyslogSeverityInfo:    "INFO",
					logsys.SyslogSeverityDebug:   "DEBUG",
				}

				for sev, expected := range tests {
					Expect(sev.String()).To(Equal(expected))
				}
			})

			It("should return empty string for unknown severity", func() {
				var unknown logsys.SyslogSeverity = 99
				Expect(unknown.String()).To(Equal(""))
			})
		})

		Context("MakeSeverity function", func() {
			It("should parse all valid severity strings", func() {
				tests := map[string]logsys.SyslogSeverity{
					"EMERG":   logsys.SyslogSeverityEmerg,
					"ALERT":   logsys.SyslogSeverityAlert,
					"CRIT":    logsys.SyslogSeverityCrit,
					"ERR":     logsys.SyslogSeverityErr,
					"WARNING": logsys.SyslogSeverityWarning,
					"NOTICE":  logsys.SyslogSeverityNotice,
					"INFO":    logsys.SyslogSeverityInfo,
					"DEBUG":   logsys.SyslogSeverityDebug,
				}

				for str, expected := range tests {
					Expect(logsys.MakeSeverity(str)).To(Equal(expected))
				}
			})

			It("should be case-insensitive", func() {
				Expect(logsys.MakeSeverity("info")).To(Equal(logsys.SyslogSeverityInfo))
				Expect(logsys.MakeSeverity("Info")).To(Equal(logsys.SyslogSeverityInfo))
				Expect(logsys.MakeSeverity("INFO")).To(Equal(logsys.SyslogSeverityInfo))
			})

			It("should return 0 for unknown string", func() {
				Expect(logsys.MakeSeverity("unknown")).To(Equal(logsys.SyslogSeverity(0)))
			})
		})
	})

	Describe("SyslogFacility", func() {
		Context("MakeFacility function", func() {
			It("should parse all valid facility strings", func() {
				tests := map[string]logsys.SyslogFacility{
					"KERN":     logsys.SyslogFacilityKern,
					"USER":     logsys.SyslogFacilityUser,
					"MAIL":     logsys.SyslogFacilityMail,
					"DAEMON":   logsys.SyslogFacilityDaemon,
					"AUTH":     logsys.SyslogFacilityAuth,
					"SYSLOG":   logsys.SyslogFacilitySyslog,
					"LPR":      logsys.SyslogFacilityLpr,
					"NEWS":     logsys.SyslogFacilityNews,
					"UUCP":     logsys.SyslogFacilityUucp,
					"CRON":     logsys.SyslogFacilityCron,
					"AUTHPRIV": logsys.SyslogFacilityAuthPriv,
					"FTP":      logsys.SyslogFacilityFTP,
					"LOCAL0":   logsys.SyslogFacilityLocal0,
					"LOCAL1":   logsys.SyslogFacilityLocal1,
					"LOCAL2":   logsys.SyslogFacilityLocal2,
					"LOCAL3":   logsys.SyslogFacilityLocal3,
					"LOCAL4":   logsys.SyslogFacilityLocal4,
					"LOCAL5":   logsys.SyslogFacilityLocal5,
					"LOCAL6":   logsys.SyslogFacilityLocal6,
					"LOCAL7":   logsys.SyslogFacilityLocal7,
				}

				for str, expected := range tests {
					Expect(logsys.MakeFacility(str)).To(Equal(expected))
				}
			})

			It("should be case-insensitive", func() {
				Expect(logsys.MakeFacility("user")).To(Equal(logsys.SyslogFacilityUser))
				Expect(logsys.MakeFacility("User")).To(Equal(logsys.SyslogFacilityUser))
				Expect(logsys.MakeFacility("USER")).To(Equal(logsys.SyslogFacilityUser))
			})

			It("should return 0 for unknown string", func() {
				Expect(logsys.MakeFacility("unknown")).To(Equal(logsys.SyslogFacility(0)))
			})
		})
	})

	Describe("Hook Methods", func() {
		var (
			hook   logsys.HookSyslog
			ctx    context.Context
			cancel context.CancelFunc
		)

		BeforeEach(func() {
			clearReceivedMessages()

			opts := logcfg.OptionsSyslog{
				Network:  libptc.NetworkUnixGram.Code(),
				Host:     sckAddr,
				Tag:      "coverage-test",
				LogLevel: []string{"info", "debug"},
			}

			var err error
			hook, err = logsys.New(opts, nil)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel = context.WithCancel(context.Background())
			go hook.Run(ctx)

			time.Sleep(100 * time.Millisecond)
		})

		AfterEach(func() {
			if cancel != nil {
				cancel()
			}
			if hook != nil {
				hook.Close()
			}
			clearReceivedMessages()
		})

		Context("RegisterHook method", func() {
			It("should register hook with logger", func() {
				logger := logrus.New()
				hook.RegisterHook(logger)

				logger.WithField("msg", "test via RegisterHook").Info("test")
				time.Sleep(100 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(messages).ToNot(BeEmpty())
			})
		})

		Context("IsRunning method", func() {
			It("should return true when running", func() {
				Eventually(func() bool {
					return hook.IsRunning()
				}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())
			})

			It("should return false after close", func() {
				Eventually(func() bool {
					return hook.IsRunning()
				}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

				cancel()
				hook.Close()

				Eventually(func() bool {
					return hook.IsRunning()
				}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())
			})
		})

		Context("WriteSev method", func() {
			It("should write with custom severity", func() {
				n, err := hook.WriteSev(logsys.SyslogSeverityDebug, []byte("custom severity write"))
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len("custom severity write")))

				time.Sleep(200 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(messages).ToNot(BeEmpty())
			})

			It("should return error when closed", func() {
				hook.Close()
				cancel()

				time.Sleep(100 * time.Millisecond)

				_, err := hook.WriteSev(logsys.SyslogSeverityInfo, []byte("test"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Fire with different levels", func() {
			It("should handle panic level", func() {
				logger := logrus.New()
				logger.AddHook(hook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.PanicLevel,
					Message: "",
					Data:    logrus.Fields{"msg": "panic message"},
				}

				err := hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)
			})

			It("should handle fatal level", func() {
				logger := logrus.New()
				logger.AddHook(hook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.FatalLevel,
					Message: "",
					Data:    logrus.Fields{"msg": "fatal message"},
				}

				err := hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)
			})

			It("should handle debug level", func() {
				logger := logrus.New()
				logger.AddHook(hook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.DebugLevel,
					Message: "",
					Data:    logrus.Fields{"msg": "debug message"},
				}

				err := hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)
			})

			It("should handle trace level (unmapped)", func() {
				logger := logrus.New()
				logger.AddHook(hook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.TraceLevel,
					Message: "",
					Data:    logrus.Fields{"msg": "trace message"},
				}

				err := hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)
			})
		})

		Context("Field filtering edge cases", func() {
			It("should handle empty fields", func() {
				logger := logrus.New()
				logger.AddHook(hook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.InfoLevel,
					Message: "",
					Data:    logrus.Fields{},
				}

				err := hook.Fire(entry)
				Expect(err).To(BeNil()) // Returns nil for empty fields
			})

			It("should handle message-only in access log mode", func() {
				opts := logcfg.OptionsSyslog{
					Network:         libptc.NetworkUnixGram.Code(),
					Host:            sckAddr,
					Tag:             "access-test",
					EnableAccessLog: true,
					LogLevel:        []string{"info"},
				}

				accessHook, err := logsys.New(opts, nil)
				Expect(err).ToNot(HaveOccurred())

				accessCtx, accessCancel := context.WithCancel(context.Background())
				defer accessCancel()
				go accessHook.Run(accessCtx)

				time.Sleep(100 * time.Millisecond)

				logger := logrus.New()
				logger.AddHook(accessHook)

				logger.Info("Access log message without fields")
				time.Sleep(100 * time.Millisecond)

				accessHook.Close()
				accessCancel()

				messages := getReceivedMessages()
				Expect(len(messages)).To(BeNumerically(">=", 1))
			})

			It("should handle empty message in access log mode", func() {
				opts := logcfg.OptionsSyslog{
					Network:         libptc.NetworkUnixGram.Code(),
					Host:            sckAddr,
					Tag:             "empty-msg-test",
					EnableAccessLog: true,
					LogLevel:        []string{"info"},
				}

				emptyHook, err := logsys.New(opts, nil)
				Expect(err).ToNot(HaveOccurred())

				emptyCtx, emptyCancel := context.WithCancel(context.Background())
				defer emptyCancel()
				go emptyHook.Run(emptyCtx)

				time.Sleep(100 * time.Millisecond)

				logger := logrus.New()
				logger.AddHook(emptyHook)

				entry := &logrus.Entry{
					Logger:  logger,
					Level:   logrus.InfoLevel,
					Message: "",
					Data:    logrus.Fields{"user": "test"},
				}

				err = emptyHook.Fire(entry)
				Expect(err).To(BeNil()) // Returns nil for empty message in access mode

				emptyHook.Close()
				emptyCancel()
			})
		})
	})

	Describe("Coverage for platform-specific code", func() {
		Context("Wrapper interface methods", func() {
			It("should test all severity-specific methods", func() {
				opts := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "severity-test",
					LogLevel: []string{},
				}

				hook, err := logsys.New(opts, nil)
				Expect(err).ToNot(HaveOccurred())

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go hook.Run(ctx)

				time.Sleep(100 * time.Millisecond)

				logger := logrus.New()
				logger.AddHook(hook)

				// Test all levels to cover different code paths
				// Note: We don't test Panic() or Fatal() as they terminate execution
				logger.WithField("msg", "error level").Error("test error")
				logger.WithField("msg", "warn level").Warn("test warn")
				logger.WithField("msg", "info level").Info("test info")
				logger.WithField("msg", "debug level").Debug("test debug")
				time.Sleep(200 * time.Millisecond)

				hook.Close()
				cancel()
			})
		})
	})
})
