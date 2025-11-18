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

package command_test

import (
	"fmt"
	"io"

	"github.com/nabbar/golib/shell/command"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command Creation", func() {
	Describe("New function", func() {
		Context("with valid parameters", func() {
			It("should create a command with name and description", func() {
				cmd := command.New("test", "test description", nil)
				Expect(cmd).ToNot(BeNil())
				Expect(cmd.Name()).To(Equal("test"))
				Expect(cmd.Describe()).To(Equal("test description"))
			})

			It("should create a command with a function", func() {
				called := false
				fn := func(out, err io.Writer, args []string) {
					called = true
				}

				cmd := command.New("test", "test description", fn)
				Expect(cmd).ToNot(BeNil())

				cmd.Run(nil, nil, nil)
				Expect(called).To(BeTrue())
			})

			It("should preserve function behavior", func() {
				outBuf := newSafeBuffer()
				fn := func(out, err io.Writer, args []string) {
					fmt.Fprintf(out, "executed with %d args", len(args))
				}

				cmd := command.New("test", "test description", fn)
				cmd.Run(outBuf, nil, []string{"arg1", "arg2"})

				Expect(outBuf.String()).To(Equal("executed with 2 args"))
			})
		})

		Context("with empty strings", func() {
			It("should create a command with empty name", func() {
				cmd := command.New("", "description", nil)
				Expect(cmd).ToNot(BeNil())
				Expect(cmd.Name()).To(Equal(""))
				Expect(cmd.Describe()).To(Equal("description"))
			})

			It("should create a command with empty description", func() {
				cmd := command.New("test", "", nil)
				Expect(cmd).ToNot(BeNil())
				Expect(cmd.Name()).To(Equal("test"))
				Expect(cmd.Describe()).To(Equal(""))
			})

			It("should create a command with both empty", func() {
				cmd := command.New("", "", nil)
				Expect(cmd).ToNot(BeNil())
				Expect(cmd.Name()).To(Equal(""))
				Expect(cmd.Describe()).To(Equal(""))
			})
		})

		Context("with nil function", func() {
			It("should create a valid command", func() {
				cmd := command.New("test", "test description", nil)
				Expect(cmd).ToNot(BeNil())
				Expect(cmd.Name()).To(Equal("test"))
				Expect(cmd.Describe()).To(Equal("test description"))
			})

			It("should not panic when Run is called", func() {
				cmd := command.New("test", "test description", nil)
				Expect(func() {
					cmd.Run(nil, nil, nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Info function", func() {
		Context("with valid parameters", func() {
			It("should create a CommandInfo with name and description", func() {
				info := command.Info("test", "test description")
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test"))
				Expect(info.Describe()).To(Equal("test description"))
			})

			It("should create an info-only object", func() {
				info := command.Info("info", "info description")
				Expect(info).ToNot(BeNil())

				// CommandInfo doesn't have Run method, but if cast to Command it should handle nil function
				if cmd, ok := info.(command.Command); ok {
					Expect(func() {
						cmd.Run(nil, nil, nil)
					}).ToNot(Panic())
				}
			})
		})

		Context("with empty strings", func() {
			It("should create CommandInfo with empty name", func() {
				info := command.Info("", "description")
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal(""))
				Expect(info.Describe()).To(Equal("description"))
			})

			It("should create CommandInfo with empty description", func() {
				info := command.Info("test", "")
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test"))
				Expect(info.Describe()).To(Equal(""))
			})

			It("should create CommandInfo with both empty", func() {
				info := command.Info("", "")
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal(""))
				Expect(info.Describe()).To(Equal(""))
			})
		})
	})

	Describe("Name method", func() {
		It("should return the exact name given", func() {
			cmd := command.New("exact-name", "description", nil)
			Expect(cmd.Name()).To(Equal("exact-name"))
		})

		It("should handle special characters in name", func() {
			cmd := command.New("test:cmd-123_ABC", "description", nil)
			Expect(cmd.Name()).To(Equal("test:cmd-123_ABC"))
		})

		It("should handle unicode in name", func() {
			cmd := command.New("тест-命令", "description", nil)
			Expect(cmd.Name()).To(Equal("тест-命令"))
		})

		It("should be idempotent", func() {
			cmd := command.New("test", "description", nil)
			name1 := cmd.Name()
			name2 := cmd.Name()
			Expect(name1).To(Equal(name2))
		})
	})

	Describe("Describe method", func() {
		It("should return the exact description given", func() {
			cmd := command.New("test", "exact description", nil)
			Expect(cmd.Describe()).To(Equal("exact description"))
		})

		It("should handle long descriptions", func() {
			longDesc := "This is a very long description that contains multiple sentences. " +
				"It describes what the command does in great detail. " +
				"It may even span multiple lines and contain special characters: !@#$%^&*()"
			cmd := command.New("test", longDesc, nil)
			Expect(cmd.Describe()).To(Equal(longDesc))
		})

		It("should handle multiline descriptions", func() {
			multiline := "Line 1\nLine 2\nLine 3"
			cmd := command.New("test", multiline, nil)
			Expect(cmd.Describe()).To(Equal(multiline))
		})

		It("should be idempotent", func() {
			cmd := command.New("test", "description", nil)
			desc1 := cmd.Describe()
			desc2 := cmd.Describe()
			Expect(desc1).To(Equal(desc2))
		})
	})

	Describe("Command interface compliance", func() {
		It("should implement Command interface", func() {
			cmd := command.New("test", "description", nil)
			Expect(cmd).ToNot(BeNil())
			// Verify it implements Command by checking methods exist
			Expect(cmd.Name()).To(Equal("test"))
			Expect(cmd.Describe()).To(Equal("description"))
		})

		It("should implement CommandInfo interface", func() {
			info := command.Info("test", "description")
			Expect(info).ToNot(BeNil())
			// Verify it implements CommandInfo by checking methods exist
			Expect(info.Name()).To(Equal("test"))
			Expect(info.Describe()).To(Equal("description"))
		})

		It("Command should also implement CommandInfo", func() {
			cmd := command.New("test", "description", nil)
			var info command.CommandInfo = cmd
			Expect(info).ToNot(BeNil())
			Expect(info.Name()).To(Equal("test"))
		})
	})
})
