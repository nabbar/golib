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
	"strings"

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration Tests", func() {
	Describe("complete workflow", func() {
		It("should handle a full command lifecycle", func() {
			sh := shell.New(nil)
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()

			// Add commands
			sh.Add("",
				command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
					if len(args) > 0 {
						fmt.Fprintf(out, "Hello, %s!", args[0])
					} else {
						fmt.Fprint(out, "Hello, World!")
					}
				}),
				command.New("echo", "Echo arguments", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, strings.Join(args, " "))
				}),
			)

			sh.Add("sys:",
				command.New("info", "System info", func(out, err io.Writer, args []string) {
					fmt.Fprint(out, "System Information")
				}),
			)

			// Verify commands were added
			cmd1, found1 := sh.Get("hello")
			Expect(found1).To(BeTrue())
			Expect(cmd1.Name()).To(Equal("hello"))

			cmd2, found2 := sh.Get("sys:info")
			Expect(found2).To(BeTrue())
			Expect(cmd2.Name()).To(Equal("info"))

			// Execute commands
			sh.Run(outBuf, errBuf, []string{"hello"})
			Expect(outBuf.String()).To(Equal("Hello, World!"))

			outBuf.Reset()
			sh.Run(outBuf, errBuf, []string{"hello", "Alice"})
			Expect(outBuf.String()).To(Equal("Hello, Alice!"))

			outBuf.Reset()
			sh.Run(outBuf, errBuf, []string{"echo", "test", "message"})
			Expect(outBuf.String()).To(Equal("test message"))

			outBuf.Reset()
			sh.Run(outBuf, errBuf, []string{"sys:info"})
			Expect(outBuf.String()).To(Equal("System Information"))

			// Walk through commands
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(3))

			// Get descriptions
			Expect(sh.Desc("hello")).To(Equal("Say hello"))
			Expect(sh.Desc("sys:info")).To(Equal("System info"))
		})
	})

	Describe("edge cases", func() {
		var sh shell.Shell

		BeforeEach(func() {
			sh = shell.New(nil)
		})

		It("should handle unicode command names", func() {
			sh.Add("", command.New("命令", "Unicode command", nil))
			sh.Add("", command.New("🚀", "Emoji command", nil))

			cmd1, found1 := sh.Get("命令")
			Expect(found1).To(BeTrue())
			Expect(cmd1).ToNot(BeNil())

			cmd2, found2 := sh.Get("🚀")
			Expect(found2).To(BeTrue())
			Expect(cmd2).ToNot(BeNil())
		})

		It("should handle very long command names", func() {
			longName := strings.Repeat("a", 10000)
			sh.Add("", command.New(longName, "Long name", nil))

			cmd, found := sh.Get(longName)
			Expect(found).To(BeTrue())
			Expect(cmd).ToNot(BeNil())
		})

		It("should handle special characters in names", func() {
			specialNames := []string{
				"test-cmd",
				"test_cmd",
				"test.cmd",
				"test:cmd",
				"test/cmd",
				"test@cmd",
			}

			for _, name := range specialNames {
				sh.Add("", command.New(name, "Special", nil))
			}

			for _, name := range specialNames {
				cmd, found := sh.Get(name)
				Expect(found).To(BeTrue())
				Expect(cmd).ToNot(BeNil())
			}
		})

		It("should handle thousands of commands", func() {
			for i := 0; i < 5000; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), fmt.Sprintf("Command %d", i), nil))
			}

			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(5000))

			// Verify random commands
			cmd, found := sh.Get("cmd2500")
			Expect(found).To(BeTrue())
			Expect(cmd.Name()).To(Equal("cmd2500"))
		})

		It("should handle command with large output", func() {
			outBuf := newSafeBuffer()
			largeData := strings.Repeat("x", 1000000)

			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, largeData)
			})

			sh.Add("", cmd)
			sh.Run(outBuf, nil, []string{"test"})

			Expect(outBuf.Len()).To(Equal(1000000))
		})

		It("should handle alternating stdout/stderr writes", func() {
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()

			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				for i := 0; i < 100; i++ {
					if i%2 == 0 {
						fmt.Fprint(out, "O")
					} else {
						fmt.Fprint(err, "E")
					}
				}
			})

			sh.Add("", cmd)
			sh.Run(outBuf, errBuf, []string{"test"})

			Expect(outBuf.Len()).To(Equal(50))
			Expect(errBuf.Len()).To(Equal(50))
		})

		It("should handle commands with error conditions", func() {
			errBuf := newSafeBuffer()

			cmd := command.New("fail", "Failing command", func(out, err io.Writer, args []string) {
				fmt.Fprint(err, "Error: command failed")
			})

			sh.Add("", cmd)
			sh.Run(nil, errBuf, []string{"fail"})

			Expect(errBuf.String()).To(Equal("Error: command failed"))
		})
	})

	Describe("namespace management", func() {
		var sh shell.Shell

		BeforeEach(func() {
			sh = shell.New(nil)
		})

		It("should handle multiple commands with same name in different namespaces", func() {
			sh.Add("sys:", command.New("list", "System list", nil))
			sh.Add("user:", command.New("list", "User list", nil))
			sh.Add("group:", command.New("list", "Group list", nil))

			sysDesc := sh.Desc("sys:list")
			userDesc := sh.Desc("user:list")
			groupDesc := sh.Desc("group:list")

			Expect(sysDesc).To(Equal("System list"))
			Expect(userDesc).To(Equal("User list"))
			Expect(groupDesc).To(Equal("Group list"))
		})

		It("should distinguish command with and without namespace", func() {
			sh.Add("", command.New("test", "Test without namespace", nil))
			sh.Add("ns:", command.New("test", "Test with namespace", nil))

			cmd1, found1 := sh.Get("test")
			Expect(found1).To(BeTrue())
			Expect(cmd1.Describe()).To(Equal("Test without namespace"))

			cmd2, found2 := sh.Get("ns:test")
			Expect(found2).To(BeTrue())
			Expect(cmd2.Describe()).To(Equal("Test with namespace"))
		})
	})

	Describe("stress tests", func() {
		It("should handle rapid command registration and execution", func() {
			sh := shell.New(nil)
			done := make(chan bool, 200)

			// Rapidly add commands
			for i := 0; i < 100; i++ {
				go func(idx int) {
					cmd := command.New(fmt.Sprintf("cmd%d", idx), "Test", func(out, err io.Writer, args []string) {
						fmt.Fprint(out, "ok")
					})
					sh.Add("", cmd)
					done <- true
				}(i)
			}

			// Rapidly execute commands
			for i := 0; i < 100; i++ {
				go func(idx int) {
					outBuf := newSafeBuffer()
					sh.Run(outBuf, nil, []string{fmt.Sprintf("cmd%d", idx)})
					done <- true
				}(i)
			}

			// Wait for all operations
			for i := 0; i < 200; i++ {
				<-done
			}
		})
	})
})
