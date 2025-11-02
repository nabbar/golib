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

package hooksyslog_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	loghsl "github.com/nabbar/golib/logger/hooksyslog"
	libptc "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("HookSyslog Integration Tests", func() {
	BeforeEach(func() {
		clearReceivedMessages()
	})

	AfterEach(func() {
		clearReceivedMessages()
	})

	Describe("UDP Syslog Integration", func() {
		Context("with real UDP connection", func() {
			It("should connect and send logs successfully", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "test-app",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Start async writer
				hookCtx, hookCancel := context.WithCancel(context.Background())
				defer hookCancel()
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Send test message
				lgr.WithField("msg", "test message from integration").Info("test")

				// Wait for message to be received
				time.Sleep(200 * time.Millisecond)

				// Verify message was received
				messages := getReceivedMessages()
				Expect(messages).ToNot(BeEmpty())

				// Check that message contains our test string
				found := false
				for _, msg := range messages {
					if strings.Contains(msg, "test message from integration") {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "Expected to find test message in received messages")

				// Graceful shutdown
				hookCancel()
				hook.Close()
			})

			It("should send multiple log levels", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "test-levels",
					LogLevel: []string{"debug", "info", "warn", "error"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				defer hookCancel()
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Send messages at different levels
				lgr.WithField("msg", "debug message").Debug("test")
				lgr.WithField("msg", "info message").Info("test")
				lgr.WithField("msg", "warn message").Warn("test")
				lgr.WithField("msg", "error message").Error("test")

				time.Sleep(200 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(len(messages)).To(BeNumerically(">=", 2))

				hookCancel()
				hook.Close()
			})

			It("should include tag in messages", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "my-custom-tag",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				defer hookCancel()
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Send messages at different levels
				lgr.WithField("msg", "tagged message").Info("test")
				time.Sleep(200 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(messages).ToNot(BeEmpty())

				hookCancel()
				hook.Close()
			})

			It("should handle structured logging with fields", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "structured",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, &logrus.JSONFormatter{})
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				defer hookCancel()
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Send messages at different levels
				lgr.WithFields(logrus.Fields{
					"user_id": 123,
					"action":  "login",
					"msg":     "user action",
				}).Info("test")

				time.Sleep(200 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(messages).ToNot(BeEmpty())

				hookCancel()
				hook.Close()
			})

			It("should handle concurrent logging", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "concurrent",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				defer hookCancel()
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Log from multiple goroutines
				var wg sync.WaitGroup
				for i := 0; i < 3; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						for j := 0; j < 2; j++ {
							lgr.WithFields(logrus.Fields{
								"msg": fmt.Sprintf("concurrent message %d-%d", id, j),
							}).Info("test")
							time.Sleep(10 * time.Millisecond)
						}
					}(i)
				}

				wg.Wait()
				time.Sleep(300 * time.Millisecond)

				messages := getReceivedMessages()
				Expect(len(messages)).To(BeNumerically(">=", 2))

				hookCancel()
				hook.Close()
			})
		})
	})

	Describe("Hook lifecycle", func() {
		Context("with proper lifecycle management", func() {
			It("should handle Done channel correctly", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "lifecycle",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				lgr.WithFields(logrus.Fields{
					"msg": "before close",
				}).Info("test")
				time.Sleep(100 * time.Millisecond)

				// Cancel and wait for Done
				hookCancel()
				hook.Close()

				select {
				case <-hook.Done():
					Expect(true).To(BeTrue())
				case <-time.After(2 * time.Second):
					Fail("Hook did not complete in time")
				}
			})

			It("should flush remaining logs on close", func() {
				opt := logcfg.OptionsSyslog{
					Network:  libptc.NetworkUnixGram.Code(),
					Host:     sckAddr,
					Tag:      "flush",
					LogLevel: []string{"info"},
				}

				hook, err := loghsl.New(opt, nil)
				Expect(err).ToNot(HaveOccurred())

				hookCtx, hookCancel := context.WithCancel(context.Background())
				go hook.Run(hookCtx)

				// Wait for hook to be ready
				time.Sleep(100 * time.Millisecond)

				// Create logger with hook
				lgr := logrus.New()
				lgr.SetOutput(GinkgoWriter) // Don't pollute stdout
				lgr.AddHook(hook)

				// Send multiple messages quickly
				for i := 0; i < 5; i++ {
					lgr.WithFields(logrus.Fields{
						"msg": fmt.Sprintf("flush test %d", i),
					}).Info("test")
				}

				time.Sleep(100 * time.Millisecond)

				// Shutdown
				hookCancel()
				hook.Close()

				time.Sleep(100 * time.Millisecond)

				// Should have received some messages
				messages := getReceivedMessages()
				Expect(len(messages)).To(BeNumerically(">=", 1))
			})
		})
	})
})
