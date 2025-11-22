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

package shell_test

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock TTYSaver for testing
type mockTTYSaver struct {
	mu            sync.Mutex
	restoreCalled bool
	signalCalled  bool
	isTerminal    bool
	restoreError  error
	signalError   error
}

func newMockTTYSaver(isTerminal bool) *mockTTYSaver {
	return &mockTTYSaver{
		isTerminal: isTerminal,
	}
}

func (m *mockTTYSaver) IsTerminal() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isTerminal
}

func (m *mockTTYSaver) Restore() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restoreCalled = true
	return m.restoreError
}

func (m *mockTTYSaver) Signal() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalCalled = true
	return m.signalError
}

func (m *mockTTYSaver) WasRestoreCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.restoreCalled
}

func (m *mockTTYSaver) WasSignalCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signalCalled
}

func (m *mockTTYSaver) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restoreCalled = false
	m.signalCalled = false
}

func (m *mockTTYSaver) SetRestoreError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restoreError = err
}

func (m *mockTTYSaver) SetSignalError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalError = err
}

var _ = Describe("TTY Integration", func() {
	Describe("Shell with TTYSaver", func() {
		Context("with nil TTYSaver", func() {
			It("should work with all operations", func() {
				sh := shell.New(nil)

				sh.Add("", command.New("test", "Test", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, "ok")
				}))

				buf := newSafeBuffer()
				sh.Run(buf, nil, []string{"test"})
				Expect(buf.String()).To(Equal("ok"))
			})
		})

		Context("with mock TTYSaver", func() {
			var mock *mockTTYSaver

			BeforeEach(func() {
				mock = newMockTTYSaver(true)
			})

			It("should accept mock TTYSaver", func() {
				sh := shell.New(mock)
				Expect(sh).ToNot(BeNil())
			})

			It("should not call Restore during normal operations", func() {
				sh := shell.New(mock)
				sh.Add("", command.New("test", "Test", nil))
				sh.Run(nil, nil, []string{"test"})

				Expect(mock.WasRestoreCalled()).To(BeFalse())
			})

			It("should support IsTerminal from TTYSaver", func() {
				sh := shell.New(mock)
				Expect(sh).ToNot(BeNil())
				Expect(mock.IsTerminal()).To(BeTrue())
			})
		})

		Context("with real TTYSaver from buffers", func() {
			It("should work with non-terminal input", func() {
				buf := bytes.NewBufferString("test")
				ttySaver, err := tty.New(buf, false)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())

				sh.Add("", command.New("cmd", "Command", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, "executed")
				}))

				outBuf := newSafeBuffer()
				sh.Run(outBuf, nil, []string{"cmd"})
				Expect(outBuf.String()).To(Equal("executed"))
			})
		})
	})

	Describe("TTYSaver lifecycle", func() {
		Context("creation patterns", func() {
			It("should handle nil TTYSaver creation", func() {
				ttySaver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(ttySaver).ToNot(BeNil())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})

			It("should handle TTYSaver with signal handling disabled", func() {
				ttySaver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})

			It("should handle TTYSaver with signal handling enabled", func() {
				ttySaver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})
		})

		Context("with different input sources", func() {
			It("should work with bytes.Buffer input", func() {
				buf := bytes.NewBufferString("test input")
				ttySaver, err := tty.New(buf, false)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})

			It("should work with nil input (defaults to stdin)", func() {
				ttySaver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})
		})
	})

	Describe("concurrent TTYSaver usage", func() {
		It("should handle concurrent shell creation with TTYSavers", func() {
			done := make(chan shell.Shell, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					ttySaver, err := tty.New(nil, false)
					Expect(err).ToNot(HaveOccurred())

					sh := shell.New(ttySaver)
					sh.Add("", command.New("test", "Test", nil))
					done <- sh
				}()
			}

			for i := 0; i < 10; i++ {
				sh := <-done
				Expect(sh).ToNot(BeNil())

				_, found := sh.Get("test")
				Expect(found).To(BeTrue())
			}
		})

		It("should handle shared TTYSaver across operations", func() {
			ttySaver, err := tty.New(nil, false)
			Expect(err).ToNot(HaveOccurred())

			sh := shell.New(ttySaver)
			sh.Add("", command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "ok")
			}))

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					buf := newSafeBuffer()
					sh.Run(buf, nil, []string{"test"})
					Expect(buf.String()).To(Equal("ok"))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("error handling with TTYSaver", func() {
		It("should handle TTYSaver with restore errors", func() {
			mock := newMockTTYSaver(true)
			mock.SetRestoreError(fmt.Errorf("restore failed"))

			sh := shell.New(mock)
			Expect(sh).ToNot(BeNil())

			// Normal operations should still work
			sh.Add("", command.New("test", "Test", nil))
			_, found := sh.Get("test")
			Expect(found).To(BeTrue())
		})

		It("should handle non-terminal TTYSaver", func() {
			mock := newMockTTYSaver(false)

			sh := shell.New(mock)
			Expect(sh).ToNot(BeNil())
			Expect(mock.IsTerminal()).To(BeFalse())

			// Operations should work normally
			sh.Add("", command.New("cmd", "Command", nil))
			_, found := sh.Get("cmd")
			Expect(found).To(BeTrue())
		})
	})

	Describe("TTYSaver state consistency", func() {
		It("should maintain TTYSaver state across operations", func() {
			mock := newMockTTYSaver(true)
			sh := shell.New(mock)

			// Add multiple commands
			for i := 0; i < 10; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			// Execute multiple commands
			for i := 0; i < 10; i++ {
				sh.Run(nil, nil, []string{fmt.Sprintf("cmd%d", i)})
			}

			// TTYSaver should remain consistent
			Expect(mock.IsTerminal()).To(BeTrue())
		})

		It("should handle TTYSaver with mixed operations", func() {
			ttySaver, err := tty.New(nil, false)
			Expect(err).ToNot(HaveOccurred())

			sh := shell.New(ttySaver)

			// Mixed operations
			sh.Add("", command.New("test1", "Test 1", nil))
			_, _ = sh.Get("test1")
			_ = sh.Desc("test1")

			sh.Walk(func(name string, item command.Command) bool {
				return true
			})

			sh.Add("sys:", command.New("info", "Info", nil))
			sh.Run(nil, nil, []string{"sys:info"})

			// Shell should remain functional
			Expect(sh).ToNot(BeNil())
		})
	})
})
