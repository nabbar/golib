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

var _ = Describe("Walk and Run Methods", func() {
	var sh shell.Shell

	BeforeEach(func() {
		sh = shell.New(nil)
	})

	Describe("Walk method", func() {
		BeforeEach(func() {
			sh.Add("", command.New("cmd1", "Command 1", nil))
			sh.Add("", command.New("cmd2", "Command 2", nil))
			sh.Add("sys:", command.New("info", "System info", nil))
		})

		It("should iterate over all commands", func() {
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(3))
		})

		It("should provide correct command names", func() {
			names := make(map[string]bool)
			sh.Walk(func(name string, item command.Command) bool {
				names[name] = true
				return true
			})
			Expect(names).To(HaveKey("cmd1"))
			Expect(names).To(HaveKey("cmd2"))
			Expect(names).To(HaveKey("sys:info"))
		})

		It("should stop iteration when func returns false", func() {
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return count < 2
			})
			Expect(count).To(Equal(2))
		})

		It("should handle nil function gracefully", func() {
			Expect(func() {
				sh.Walk(nil)
			}).ToNot(Panic())
		})

		It("should handle empty shell", func() {
			emptyShell := shell.New(nil)
			count := 0
			emptyShell.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(0))
		})

		It("should handle concurrent Walk operations", func() {
			done := make(chan bool, 50)

			for i := 0; i < 50; i++ {
				go func() {
					count := 0
					sh.Walk(func(name string, item command.Command) bool {
						count++
						return true
					})
					Expect(count).To(Equal(3))
					done <- true
				}()
			}

			for i := 0; i < 50; i++ {
				<-done
			}
		})
	})

	Describe("Run method", func() {
		It("should execute existing command", func() {
			outBuf := newSafeBuffer()
			executed := false

			cmd := command.New("test", "Test command", func(out, err io.Writer, args []string) {
				executed = true
				fmt.Fprint(out, "executed")
			})

			sh.Add("", cmd)
			sh.Run(outBuf, nil, []string{"test"})

			Expect(executed).To(BeTrue())
			Expect(outBuf.String()).To(Equal("executed"))
		})

		It("should pass arguments to command", func() {
			var receivedArgs []string

			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				receivedArgs = args
			})

			sh.Add("", cmd)
			sh.Run(nil, nil, []string{"test", "arg1", "arg2", "arg3"})

			Expect(receivedArgs).To(Equal([]string{"arg1", "arg2", "arg3"}))
		})

		It("should handle command with prefix", func() {
			outBuf := newSafeBuffer()

			cmd := command.New("info", "Info", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "info output")
			})

			sh.Add("sys:", cmd)
			sh.Run(outBuf, nil, []string{"sys:info"})

			Expect(outBuf.String()).To(Equal("info output"))
		})

		It("should handle non-existent command gracefully", func() {
			Expect(func() {
				sh.Run(nil, nil, []string{"nonexistent"})
			}).ToNot(Panic())
		})

		It("should handle empty args", func() {
			Expect(func() {
				sh.Run(nil, nil, []string{})
			}).ToNot(Panic())
		})

		It("should handle nil writers", func() {
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				// Command that doesn't write
			})

			sh.Add("", cmd)

			Expect(func() {
				sh.Run(nil, nil, []string{"test"})
			}).ToNot(Panic())
		})

		It("should write to error writer", func() {
			errBuf := newSafeBuffer()

			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(err, "error message")
			})

			sh.Add("", cmd)
			sh.Run(nil, errBuf, []string{"test"})

			Expect(errBuf.String()).To(Equal("error message"))
		})

		It("should handle concurrent Run operations", func() {
			done := make(chan bool, 100)

			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "output")
			})
			sh.Add("", cmd)

			for i := 0; i < 100; i++ {
				go func() {
					outBuf := newSafeBuffer()
					sh.Run(outBuf, nil, []string{"test"})
					Expect(outBuf.String()).To(Equal("output"))
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}
		})
	})

	Describe("concurrent Add and Run", func() {
		It("should handle concurrent Add and Run operations", func() {
			done := make(chan bool, 200)

			// Add commands concurrently
			for i := 0; i < 100; i++ {
				go func(idx int) {
					cmd := command.New(fmt.Sprintf("cmd%d", idx), "Test", func(out, err io.Writer, args []string) {
						fmt.Fprint(out, "output")
					})
					sh.Add("", cmd)
					done <- true
				}(i)
			}

			// Run commands concurrently
			for i := 0; i < 100; i++ {
				go func(idx int) {
					outBuf := newSafeBuffer()
					sh.Run(outBuf, nil, []string{fmt.Sprintf("cmd%d", idx)})
					done <- true
				}(i)
			}

			for i := 0; i < 200; i++ {
				<-done
			}
		})
	})
})
