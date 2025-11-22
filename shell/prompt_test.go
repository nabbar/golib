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
	"fmt"
	"io"

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shell Prompt Functions", func() {
	Describe("executor behavior", func() {
		var sh shell.Shell
		var outBuf, errBuf *safeBuffer

		BeforeEach(func() {
			sh = shell.New(nil)
			outBuf = newSafeBuffer()
			errBuf = newSafeBuffer()
		})

		Context("with registered commands", func() {
			BeforeEach(func() {
				sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, "Hello!")
				}))
				sh.Add("", command.New("echo", "Echo args", func(out, err io.Writer, args []string) {
					for _, arg := range args {
						fmt.Fprint(out, arg, " ")
					}
				}))
			})

			It("should execute hello command", func() {
				sh.Run(outBuf, errBuf, []string{"hello"})
				Expect(outBuf.String()).To(Equal("Hello!"))
			})

			It("should execute echo with arguments", func() {
				sh.Run(outBuf, errBuf, []string{"echo", "test", "message"})
				Expect(outBuf.String()).To(Equal("test message "))
			})

			It("should handle empty input", func() {
				sh.Run(outBuf, errBuf, []string{})
				Expect(outBuf.String()).To(BeEmpty())
				Expect(errBuf.String()).To(BeEmpty())
			})

			It("should handle invalid command", func() {
				sh.Run(outBuf, errBuf, []string{"invalid"})
				Expect(errBuf.String()).To(ContainSubstring("Invalid command"))
			})
		})

		Context("with prefixed commands", func() {
			BeforeEach(func() {
				sh.Add("sys:", command.New("info", "System info", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, "System Information")
				}))
			})

			It("should execute prefixed command", func() {
				sh.Run(outBuf, errBuf, []string{"sys:info"})
				Expect(outBuf.String()).To(Equal("System Information"))
			})

			It("should fail without prefix", func() {
				sh.Run(outBuf, errBuf, []string{"info"})
				Expect(errBuf.String()).To(ContainSubstring("Invalid command"))
			})
		})
	})

	Describe("command suggestions", func() {
		var sh shell.Shell

		BeforeEach(func() {
			sh = shell.New(nil)
			sh.Add("",
				command.New("hello", "Say hello", nil),
				command.New("help", "Show help", nil),
				command.New("history", "Show history", nil))
			sh.Add("sys:",
				command.New("info", "System info", nil),
				command.New("status", "System status", nil))
		})

		It("should have all commands registered", func() {
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(5))
		})

		It("should retrieve hello command", func() {
			cmd, found := sh.Get("hello")
			Expect(found).To(BeTrue())
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Describe()).To(Equal("Say hello"))
		})

		It("should retrieve prefixed command", func() {
			cmd, found := sh.Get("sys:info")
			Expect(found).To(BeTrue())
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Describe()).To(Equal("System info"))
		})

		It("should return description for valid commands", func() {
			desc := sh.Desc("hello")
			Expect(desc).To(Equal("Say hello"))

			desc = sh.Desc("sys:info")
			Expect(desc).To(Equal("System info"))
		})

		It("should return empty description for invalid commands", func() {
			desc := sh.Desc("invalid")
			Expect(desc).To(BeEmpty())
		})
	})

	Describe("output writer handling", func() {
		var sh shell.Shell
		var counter *callCounter

		BeforeEach(func() {
			sh = shell.New(nil)
			counter = newCallCounter()

			sh.Add("", command.New("test", "Test command", func(out, err io.Writer, args []string) {
				counter.Inc()
				if out != nil {
					fmt.Fprint(out, "output")
				}
				if err != nil {
					fmt.Fprint(err, "error")
				}
			}))
		})

		It("should work with nil output writers", func() {
			sh.Run(nil, nil, []string{"test"})
			Expect(counter.Get()).To(Equal(1))
		})

		It("should work with valid output writers", func() {
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()

			sh.Run(outBuf, errBuf, []string{"test"})
			Expect(counter.Get()).To(Equal(1))
			Expect(outBuf.String()).To(Equal("output"))
			Expect(errBuf.String()).To(Equal("error"))
		})

		It("should work with same writer for out and err", func() {
			buf := newSafeBuffer()

			sh.Run(buf, buf, []string{"test"})
			Expect(counter.Get()).To(Equal(1))
			Expect(buf.String()).To(ContainSubstring("output"))
			Expect(buf.String()).To(ContainSubstring("error"))
		})
	})

	Describe("concurrent command execution", func() {
		It("should handle concurrent Run calls", func() {
			sh := shell.New(nil)
			counter := newCallCounter()

			sh.Add("", command.New("test", "Test", func(out, err io.Writer, args []string) {
				counter.Inc()
			}))

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					sh.Run(nil, nil, []string{"test"})
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			Expect(counter.Get()).To(Equal(10))
		})

		It("should handle mixed concurrent operations", func() {
			sh := shell.New(nil)
			sh.Add("", command.New("cmd1", "Command 1", nil))

			done := make(chan bool, 30)

			// Concurrent Add
			for i := 0; i < 10; i++ {
				go func(idx int) {
					defer GinkgoRecover()
					sh.Add("", command.New(fmt.Sprintf("cmd%d", idx+2), "Test", nil))
					done <- true
				}(i)
			}

			// Concurrent Get
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					_, _ = sh.Get("cmd1")
					done <- true
				}()
			}

			// Concurrent Walk
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					sh.Walk(func(name string, item command.Command) bool {
						return true
					})
					done <- true
				}()
			}

			for i := 0; i < 30; i++ {
				<-done
			}
		})
	})
})
