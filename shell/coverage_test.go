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

var _ = Describe("Coverage Enhancement", func() {
	var sh shell.Shell

	BeforeEach(func() {
		sh = shell.New(nil)
	})

	Describe("Run edge cases", func() {
		It("should handle nil command gracefully", func() {
			// Add a nil command (shouldn't be added)
			sh.Add("", nil)

			// Try to run a non-existent command
			Expect(func() {
				sh.Run(nil, nil, []string{"test"})
			}).ToNot(Panic())
		})

		It("should handle command that returns nil", func() {
			cmd := command.New("test", "Test", nil)
			sh.Add("", cmd)

			Expect(func() {
				sh.Run(nil, nil, []string{"test"})
			}).ToNot(Panic())
		})

		It("should handle command execution with only stdout", func() {
			outBuf := newSafeBuffer()
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "output")
			})
			sh.Add("", cmd)

			sh.Run(outBuf, nil, []string{"test"})
			Expect(outBuf.String()).To(Equal("output"))
		})

		It("should handle command execution with only stderr", func() {
			errBuf := newSafeBuffer()
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(err, "error")
			})
			sh.Add("", cmd)

			sh.Run(nil, errBuf, []string{"test"})
			Expect(errBuf.String()).To(Equal("error"))
		})

		It("should handle command with many arguments", func() {
			var receivedArgs []string
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				receivedArgs = args
			})
			sh.Add("", cmd)

			args := []string{"test"}
			for i := 0; i < 100; i++ {
				args = append(args, fmt.Sprintf("arg%d", i))
			}

			sh.Run(nil, nil, args)
			Expect(len(receivedArgs)).To(Equal(100))
		})

		It("should handle empty command name in args", func() {
			Expect(func() {
				sh.Run(nil, nil, []string{""})
			}).ToNot(Panic())
		})

		It("should handle single empty string in args", func() {
			Expect(func() {
				sh.Run(nil, nil, []string{""})
			}).ToNot(Panic())
		})
	})

	Describe("Walk edge cases", func() {
		It("should handle Walk that stops early", func() {
			sh.Add("", command.New("cmd1", "Command 1", nil))
			sh.Add("", command.New("cmd2", "Command 2", nil))
			sh.Add("", command.New("cmd3", "Command 3", nil))

			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return false // Stop after first
			})

			Expect(count).To(Equal(1))
		})

		It("should handle Walk with command inspection", func() {
			sh.Add("", command.New("cmd1", "Command 1", nil))
			sh.Add("", command.New("cmd2", "Command 2", nil))

			names := make([]string, 0)
			sh.Walk(func(name string, item command.Command) bool {
				names = append(names, name)
				return true
			})

			Expect(len(names)).To(Equal(2))
			Expect(names).To(ContainElement("cmd1"))
			Expect(names).To(ContainElement("cmd2"))
		})

		It("should handle Walk with command validation", func() {
			sh.Add("", command.New("cmd1", "Command 1", nil))
			sh.Add("", command.New("cmd2", "", nil))

			descriptions := make([]string, 0)
			sh.Walk(func(name string, item command.Command) bool {
				descriptions = append(descriptions, item.Describe())
				return true
			})

			Expect(len(descriptions)).To(Equal(2))
		})
	})

	Describe("Get with various scenarios", func() {
		It("should return false for empty name when shell has commands", func() {
			sh.Add("", command.New("test", "Test", nil))
			_, found := sh.Get("")
			Expect(found).To(BeFalse())
		})

		It("should handle Get after command replacement", func() {
			cmd1 := command.New("test", "Version 1", nil)
			cmd2 := command.New("test", "Version 2", nil)

			sh.Add("", cmd1)
			sh.Add("", cmd2)

			cmd, found := sh.Get("test")
			Expect(found).To(BeTrue())
			Expect(cmd.Describe()).To(Equal("Version 2"))
		})
	})

	Describe("Desc with various scenarios", func() {
		It("should return empty for non-existent command", func() {
			desc := sh.Desc("nonexistent")
			Expect(desc).To(BeEmpty())
		})

		It("should return description after command update", func() {
			sh.Add("", command.New("test", "Description 1", nil))
			sh.Add("", command.New("test", "Description 2", nil))

			desc := sh.Desc("test")
			Expect(desc).To(Equal("Description 2"))
		})

		It("should handle Desc for command with empty description", func() {
			sh.Add("", command.New("test", "", nil))
			desc := sh.Desc("test")
			Expect(desc).To(Equal(""))
		})
	})

	Describe("Add with edge cases", func() {
		It("should handle adding same command multiple times", func() {
			cmd := command.New("test", "Test", nil)

			for i := 0; i < 10; i++ {
				sh.Add("", cmd)
			}

			retrieved, found := sh.Get("test")
			Expect(found).To(BeTrue())
			Expect(retrieved).ToNot(BeNil())
		})

		It("should handle adding commands with various prefixes", func() {
			cmd := command.New("test", "Test", nil)

			sh.Add("", cmd)
			sh.Add("a:", cmd)
			sh.Add("ab:", cmd)
			sh.Add("abc:", cmd)

			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})

			Expect(count).To(Equal(4))
		})

		It("should handle nil in middle of variadic list", func() {
			cmd1 := command.New("cmd1", "Command 1", nil)
			cmd2 := command.New("cmd2", "Command 2", nil)
			cmd3 := command.New("cmd3", "Command 3", nil)

			sh.Add("", cmd1, nil, cmd2, nil, cmd3, nil)

			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})

			Expect(count).To(Equal(3))
		})

		It("should handle all nil commands", func() {
			sh.Add("", nil, nil, nil)

			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})
	})

	Describe("Complex integration scenarios", func() {
		It("should handle operations on shell with many commands", func() {
			// Add many commands
			for i := 0; i < 100; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), fmt.Sprintf("Command %d", i), nil))
			}

			// Get specific commands
			for i := 0; i < 100; i += 10 {
				cmd, found := sh.Get(fmt.Sprintf("cmd%d", i))
				Expect(found).To(BeTrue())
				Expect(cmd).ToNot(BeNil())
			}

			// Walk and count
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(100))

			// Get descriptions
			for i := 0; i < 100; i += 10 {
				desc := sh.Desc(fmt.Sprintf("cmd%d", i))
				Expect(desc).To(Equal(fmt.Sprintf("Command %d", i)))
			}
		})

		It("should handle mixed operations", func() {
			// Add initial commands
			sh.Add("", command.New("cmd1", "Command 1", nil))

			// Get and verify
			_, found := sh.Get("cmd1")
			Expect(found).To(BeTrue())

			// Add more commands
			sh.Add("", command.New("cmd2", "Command 2", nil))
			sh.Add("prefix:", command.New("cmd3", "Command 3", nil))

			// Walk
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(Equal(3))

			// Replace command
			sh.Add("", command.New("cmd1", "Updated Command 1", nil))

			// Verify replacement
			desc := sh.Desc("cmd1")
			Expect(desc).To(Equal("Updated Command 1"))
		})
	})
})
