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
	"strings"

	"github.com/nabbar/golib/shell/command"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command Run Method", func() {
	Describe("basic execution", func() {
		It("should execute the function", func() {
			called := false
			fn := func(out, err io.Writer, args []string) {
				called = true
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, nil)
			Expect(called).To(BeTrue())
		})

		It("should write to output writer", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "test output")
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(outBuf, nil, nil)
			Expect(outBuf.String()).To(Equal("test output"))
		})

		It("should write to error writer", func() {
			errBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(err, "error output")
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, errBuf, nil)
			Expect(errBuf.String()).To(Equal("error output"))
		})

		It("should write to both writers", func() {
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "stdout")
				fmt.Fprint(err, "stderr")
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(outBuf, errBuf, nil)
			Expect(outBuf.String()).To(Equal("stdout"))
			Expect(errBuf.String()).To(Equal("stderr"))
		})
	})

	Describe("argument handling", func() {
		It("should pass empty args slice", func() {
			var receivedArgs []string
			fn := func(out, err io.Writer, args []string) {
				receivedArgs = args
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, []string{})
			Expect(receivedArgs).To(BeEmpty())
		})

		It("should pass single argument", func() {
			var receivedArgs []string
			fn := func(out, err io.Writer, args []string) {
				receivedArgs = args
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, []string{"arg1"})
			Expect(receivedArgs).To(Equal([]string{"arg1"}))
		})

		It("should pass multiple arguments", func() {
			var receivedArgs []string
			fn := func(out, err io.Writer, args []string) {
				receivedArgs = args
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, []string{"arg1", "arg2", "arg3"})
			Expect(receivedArgs).To(Equal([]string{"arg1", "arg2", "arg3"}))
		})

		It("should handle arguments with spaces", func() {
			var receivedArgs []string
			fn := func(out, err io.Writer, args []string) {
				receivedArgs = args
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, []string{"arg with spaces", "another arg"})
			Expect(receivedArgs).To(Equal([]string{"arg with spaces", "another arg"}))
		})

		It("should handle special characters in arguments", func() {
			var receivedArgs []string
			fn := func(out, err io.Writer, args []string) {
				receivedArgs = args
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, []string{"!@#$%", "^&*()", "привет", "こんにちは"})
			Expect(receivedArgs).To(HaveLen(4))
		})

		It("should use args to determine output", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				for i, arg := range args {
					if i > 0 {
						fmt.Fprint(out, " ")
					}
					fmt.Fprint(out, arg)
				}
			}

			cmd := command.New("echo", "echo command", fn)
			cmd.Run(outBuf, nil, []string{"hello", "world"})
			Expect(outBuf.String()).To(Equal("hello world"))
		})
	})

	Describe("nil handling", func() {
		It("should handle nil output writer", func() {
			fn := func(out, err io.Writer, args []string) {
				// This should not panic even though out is nil
				if out != nil {
					fmt.Fprint(out, "output")
				}
			}

			cmd := command.New("test", "description", fn)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).ToNot(Panic())
		})

		It("should handle nil error writer", func() {
			fn := func(out, err io.Writer, args []string) {
				// This should not panic even though err is nil
				if err != nil {
					fmt.Fprint(err, "error")
				}
			}

			cmd := command.New("test", "description", fn)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).ToNot(Panic())
		})

		It("should handle nil args", func() {
			called := false
			fn := func(out, err io.Writer, args []string) {
				called = true
				// Should handle nil args gracefully
				Expect(args).To(BeNil())
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, nil)
			Expect(called).To(BeTrue())
		})

		It("should not panic with all nil parameters", func() {
			fn := func(out, err io.Writer, args []string) {
				// Do nothing
			}

			cmd := command.New("test", "description", fn)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).ToNot(Panic())
		})

		It("should not execute when function is nil", func() {
			cmd := command.New("test", "description", nil)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).ToNot(Panic())
		})
	})

	Describe("complex command implementations", func() {
		It("should implement a help command", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				if len(args) == 0 {
					fmt.Fprintln(out, "Available commands:")
					fmt.Fprintln(out, "  help - Show this help")
					fmt.Fprintln(out, "  exit - Exit the shell")
				} else {
					fmt.Fprintf(out, "Help for command: %s\n", args[0])
				}
			}

			cmd := command.New("help", "Show help information", fn)
			cmd.Run(outBuf, nil, nil)
			Expect(outBuf.String()).To(ContainSubstring("Available commands:"))

			outBuf.Reset()
			cmd.Run(outBuf, nil, []string{"exit"})
			Expect(outBuf.String()).To(ContainSubstring("Help for command: exit"))
		})

		It("should implement an echo command", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprintln(out, strings.Join(args, " "))
			}

			cmd := command.New("echo", "Echo arguments", fn)
			cmd.Run(outBuf, nil, []string{"hello", "world", "!"})
			Expect(outBuf.String()).To(Equal("hello world !\n"))
		})

		It("should implement a command with validation", func() {
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				if len(args) == 0 {
					fmt.Fprintln(err, "Error: missing argument")
					return
				}
				fmt.Fprintf(out, "Processing: %s\n", args[0])
			}

			cmd := command.New("process", "Process a file", fn)
			cmd.Run(outBuf, errBuf, nil)
			Expect(errBuf.String()).To(ContainSubstring("Error: missing argument"))
			Expect(outBuf.String()).To(BeEmpty())

			outBuf.Reset()
			errBuf.Reset()
			cmd.Run(outBuf, errBuf, []string{"file.txt"})
			Expect(outBuf.String()).To(ContainSubstring("Processing: file.txt"))
			Expect(errBuf.String()).To(BeEmpty())
		})

		It("should implement a command with flags parsing", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				verbose := false
				var files []string

				for _, arg := range args {
					if arg == "-v" || arg == "--verbose" {
						verbose = true
					} else {
						files = append(files, arg)
					}
				}

				if verbose {
					fmt.Fprintf(out, "Verbose mode enabled\n")
				}
				fmt.Fprintf(out, "Files: %v\n", files)
			}

			cmd := command.New("list", "List files", fn)
			cmd.Run(outBuf, nil, []string{"-v", "file1.txt", "file2.txt"})
			Expect(outBuf.String()).To(ContainSubstring("Verbose mode enabled"))
			Expect(outBuf.String()).To(ContainSubstring("Files: [file1.txt file2.txt]"))
		})
	})

	Describe("error handling within functions", func() {
		It("should write errors to error writer", func() {
			errBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprintf(err, "error: operation failed\n")
			}

			cmd := command.New("fail", "Failing command", fn)
			cmd.Run(nil, errBuf, nil)
			Expect(errBuf.String()).To(Equal("error: operation failed\n"))
		})

		It("should handle panics in function gracefully", func() {
			// Note: The Run method itself doesn't catch panics, but we test that
			// the setup works correctly. In real usage, the shell should handle panics.
			fn := func(out, err io.Writer, args []string) {
				panic("intentional panic")
			}

			cmd := command.New("panic", "Panicking command", fn)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).To(Panic())
		})
	})

	Describe("writer errors", func() {
		It("should propagate write errors", func() {
			tw := newTestWriter()
			tw.SetShouldFail(true)

			fn := func(out, err io.Writer, args []string) {
				n, writeErr := fmt.Fprint(out, "test")
				// The function can check for write errors
				Expect(writeErr).To(HaveOccurred())
				Expect(n).To(Equal(0))
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(tw, nil, nil)
		})
	})

	Describe("multiple invocations", func() {
		It("should be callable multiple times", func() {
			callCount := 0
			fn := func(out, err io.Writer, args []string) {
				callCount++
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(nil, nil, nil)
			cmd.Run(nil, nil, nil)
			cmd.Run(nil, nil, nil)
			Expect(callCount).To(Equal(3))
		})

		It("should maintain independence between calls", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprintf(out, "%v", args)
			}

			cmd := command.New("test", "description", fn)
			cmd.Run(outBuf, nil, []string{"first"})
			first := outBuf.String()

			outBuf.Reset()
			cmd.Run(outBuf, nil, []string{"second"})
			second := outBuf.String()

			Expect(first).To(ContainSubstring("first"))
			Expect(second).To(ContainSubstring("second"))
			Expect(first).ToNot(Equal(second))
		})
	})
})
