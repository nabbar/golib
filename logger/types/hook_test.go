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

package types_test

import (
	"context"
	"io"
	"sync"
	"time"

	. "github.com/nabbar/golib/logger/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

// Mock Hook implementation for testing
type mockHook struct {
	mu               sync.RWMutex
	fireCalled       bool
	levelsCalled     bool
	registerCalled   bool
	runCalled        bool
	writeCalled      bool
	closeCalled      bool
	writeData        []byte
	registeredLogger *logrus.Logger
	runContext       context.Context
}

func (m *mockHook) Fire(entry *logrus.Entry) error {
	m.mu.Lock()
	m.fireCalled = true
	m.mu.Unlock()
	return nil
}

func (m *mockHook) Levels() []logrus.Level {
	m.mu.Lock()
	m.levelsCalled = true
	m.mu.Unlock()
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (m *mockHook) RegisterHook(log *logrus.Logger) {
	m.mu.Lock()
	m.registerCalled = true
	m.registeredLogger = log
	m.mu.Unlock()
	log.AddHook(m)
}

func (m *mockHook) Run(ctx context.Context) {
	m.mu.Lock()
	m.runCalled = true
	m.runContext = ctx
	m.mu.Unlock()
	<-ctx.Done()
}

func (m *mockHook) IsRunning() bool {
	return true
}

func (m *mockHook) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeCalled = true
	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func (m *mockHook) Close() error {
	m.mu.Lock()
	m.closeCalled = true
	m.mu.Unlock()
	return nil
}

// Thread-safe getters for testing
func (m *mockHook) wasFireCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fireCalled
}

func (m *mockHook) wasLevelsCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.levelsCalled
}

func (m *mockHook) wasRegisterCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registerCalled
}

func (m *mockHook) wasRunCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runCalled
}

func (m *mockHook) wasWriteCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.writeCalled
}

func (m *mockHook) wasCloseCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closeCalled
}

func (m *mockHook) getWriteData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]byte(nil), m.writeData...)
}

func (m *mockHook) getRegisteredLogger() *logrus.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registeredLogger
}

func getLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

var _ = Describe("Logger Types - Hook Interface", func() {
	var hook *mockHook

	BeforeEach(func() {
		hook = &mockHook{}
	})

	Describe("Hook interface compliance", func() {
		Context("when implementing Hook interface", func() {
			It("should satisfy Hook interface", func() {
				var h Hook = hook
				Expect(h).ToNot(BeNil())
			})

			It("should satisfy logrus.Hook interface", func() {
				var lh logrus.Hook = hook
				Expect(lh).ToNot(BeNil())
			})

			It("should satisfy io.WriteCloser interface", func() {
				var wc io.WriteCloser = hook
				Expect(wc).ToNot(BeNil())
			})
		})
	})

	Describe("Fire method", func() {
		Context("when firing hook", func() {
			It("should be callable", func() {
				entry := &logrus.Entry{
					Logger:  getLogger(),
					Level:   logrus.InfoLevel,
					Message: "test message",
				}

				err := hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook.wasFireCalled()).To(BeTrue())
			})

			It("should handle nil entry gracefully", func() {
				Expect(func() {
					_ = hook.Fire(nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Levels method", func() {
		Context("when getting levels", func() {
			It("should return supported levels", func() {
				levels := hook.Levels()
				Expect(levels).ToNot(BeEmpty())
				Expect(hook.wasLevelsCalled()).To(BeTrue())
			})

			It("should return standard logrus levels", func() {
				levels := hook.Levels()
				Expect(levels).To(ContainElement(logrus.InfoLevel))
				Expect(levels).To(ContainElement(logrus.ErrorLevel))
				Expect(levels).To(ContainElement(logrus.WarnLevel))
			})
		})
	})

	Describe("RegisterHook method", func() {
		Context("when registering hook", func() {
			It("should register with logger", func() {
				logger := getLogger()
				hook.RegisterHook(logger)

				Expect(hook.wasRegisterCalled()).To(BeTrue())
				Expect(hook.getRegisteredLogger()).To(Equal(logger))
			})

			It("should allow logger to use hook", func() {
				logger := getLogger()
				hook.RegisterHook(logger)

				// Trigger a log that will call Fire
				logger.Info("test message")

				Expect(hook.wasFireCalled()).To(BeTrue())
			})
		})
	})

	Describe("Run method", func() {
		Context("when running hook", func() {
			It("should be runnable in goroutine", func(ctx SpecContext) {
				runCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				go hook.Run(runCtx)

				// Give it time to start
				Eventually(func() bool {
					return hook.wasRunCalled()
				}).Should(BeTrue())
			}, NodeTimeout(time.Second))

			It("should respect context cancellation", func(ctx SpecContext) {
				runCtx, cancel := context.WithCancel(context.Background())

				completed := make(chan bool, 1)
				go func() {
					hook.Run(runCtx)
					completed <- true
				}()

				// Cancel context
				cancel()

				// Should complete quickly after cancellation
				Eventually(completed).Should(Receive(BeTrue()))
			}, NodeTimeout(time.Second))
		})
	})

	Describe("Write method", func() {
		Context("when writing data", func() {
			It("should write successfully", func() {
				data := []byte("test log data")
				n, err := hook.Write(data)

				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
				Expect(hook.wasWriteCalled()).To(BeTrue())
				Expect(hook.getWriteData()).To(Equal(data))
			})

			It("should handle empty data", func() {
				data := []byte("")
				n, err := hook.Write(data)

				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle multiple writes", func() {
				data1 := []byte("first write")
				data2 := []byte("second write")

				n1, err1 := hook.Write(data1)
				n2, err2 := hook.Write(data2)

				Expect(err1).ToNot(HaveOccurred())
				Expect(err2).ToNot(HaveOccurred())
				Expect(n1).To(Equal(len(data1)))
				Expect(n2).To(Equal(len(data2)))
				Expect(hook.getWriteData()).To(Equal(append(data1, data2...)))
			})
		})
	})

	Describe("Close method", func() {
		Context("when closing hook", func() {
			It("should close successfully", func() {
				err := hook.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(hook.wasCloseCalled()).To(BeTrue())
			})

			It("should be idempotent", func() {
				err1 := hook.Close()
				err2 := hook.Close()

				Expect(err1).ToNot(HaveOccurred())
				Expect(err2).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Complete hook lifecycle", func() {
		Context("when using hook through its full lifecycle", func() {
			It("should work correctly", func(ctx SpecContext) {
				logger := getLogger()
				runCtx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Register
				hook.RegisterHook(logger)
				Expect(hook.wasRegisterCalled()).To(BeTrue())

				// Run
				go hook.Run(runCtx)
				Eventually(func() bool {
					return hook.wasRunCalled()
				}).Should(BeTrue())

				// Use
				logger.Info("test message")
				Expect(hook.wasFireCalled()).To(BeTrue())

				// Write
				_, err := hook.Write([]byte("direct write"))
				Expect(err).ToNot(HaveOccurred())
				Expect(hook.wasWriteCalled()).To(BeTrue())

				// Close
				err = hook.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(hook.wasCloseCalled()).To(BeTrue())
			}, NodeTimeout(2*time.Second))
		})
	})
})
