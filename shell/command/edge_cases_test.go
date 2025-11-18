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

var _ = Describe("Edge Cases", func() {
	Describe("nil receiver handling", func() {
		It("should handle nil command for Name", func() {
			// This tests the internal model's nil check
			// We can't directly test it without reflection, but we ensure the API is safe
			Expect(func() {
				_ = command.New("test", "desc", nil).Name()
			}).ToNot(Panic())
		})

		It("should handle nil command for Describe", func() {
			Expect(func() {
				_ = command.New("test", "desc", nil).Describe()
			}).ToNot(Panic())
		})

		It("should handle nil command for Run", func() {
			Expect(func() {
				command.New("test", "desc", nil).Run(nil, nil, nil)
			}).ToNot(Panic())
		})
	})

	Describe("extreme string lengths", func() {
		It("should handle very long names", func() {
			longName := strings.Repeat("a", 10000)
			cmd := command.New(longName, "description", nil)
			Expect(cmd.Name()).To(Equal(longName))
			Expect(len(cmd.Name())).To(Equal(10000))
		})

		It("should handle very long descriptions", func() {
			longDesc := strings.Repeat("description ", 1000)
			cmd := command.New("test", longDesc, nil)
			Expect(cmd.Describe()).To(Equal(longDesc))
		})

		It("should handle both very long", func() {
			longName := strings.Repeat("n", 5000)
			longDesc := strings.Repeat("d", 5000)
			cmd := command.New(longName, longDesc, nil)
			Expect(len(cmd.Name())).To(Equal(5000))
			Expect(len(cmd.Describe())).To(Equal(5000))
		})
	})

	Describe("special characters", func() {
		It("should handle null bytes in name", func() {
			name := "test\x00name"
			cmd := command.New(name, "desc", nil)
			Expect(cmd.Name()).To(Equal(name))
		})

		It("should handle null bytes in description", func() {
			desc := "desc\x00ription"
			cmd := command.New("test", desc, nil)
			Expect(cmd.Describe()).To(Equal(desc))
		})

		It("should handle all whitespace characters", func() {
			name := "test\t\n\r\v\f name"
			desc := "desc\t\n\r\v\f ription"
			cmd := command.New(name, desc, nil)
			Expect(cmd.Name()).To(Equal(name))
			Expect(cmd.Describe()).To(Equal(desc))
		})

		It("should handle emoji in name", func() {
			name := "🚀test🎉command🔥"
			cmd := command.New(name, "desc", nil)
			Expect(cmd.Name()).To(Equal(name))
		})

		It("should handle emoji in description", func() {
			desc := "A cool 😎 command that rocks 🎸"
			cmd := command.New("test", desc, nil)
			Expect(cmd.Describe()).To(Equal(desc))
		})

		It("should handle control characters", func() {
			name := "test\x01\x02\x03"
			cmd := command.New(name, "desc", nil)
			Expect(cmd.Name()).To(Equal(name))
		})
	})

	Describe("very large argument lists", func() {
		It("should handle thousands of arguments", func() {
			args := make([]string, 10000)
			for i := range args {
				args[i] = fmt.Sprintf("arg%d", i)
			}

			var receivedCount int
			fn := func(out, err io.Writer, a []string) {
				receivedCount = len(a)
			}

			cmd := command.New("test", "desc", fn)
			cmd.Run(nil, nil, args)
			Expect(receivedCount).To(Equal(10000))
		})

		It("should handle arguments with extreme sizes", func() {
			args := []string{
				strings.Repeat("a", 100000),
				strings.Repeat("b", 100000),
			}

			var totalLen int
			fn := func(out, err io.Writer, a []string) {
				for _, arg := range a {
					totalLen += len(arg)
				}
			}

			cmd := command.New("test", "desc", fn)
			cmd.Run(nil, nil, args)
			Expect(totalLen).To(Equal(200000))
		})
	})

	Describe("concurrent access", func() {
		It("should be safe for concurrent Name calls", func() {
			cmd := command.New("test", "description", nil)
			done := make(chan bool, 100)

			for i := 0; i < 100; i++ {
				go func() {
					name := cmd.Name()
					Expect(name).To(Equal("test"))
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}
		})

		It("should be safe for concurrent Describe calls", func() {
			cmd := command.New("test", "description", nil)
			done := make(chan bool, 100)

			for i := 0; i < 100; i++ {
				go func() {
					desc := cmd.Describe()
					Expect(desc).To(Equal("description"))
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}
		})

		It("should be safe for concurrent Run calls", func() {
			counter := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(out, ".")
			}

			cmd := command.New("test", "description", fn)
			done := make(chan bool, 100)

			for i := 0; i < 100; i++ {
				go func() {
					cmd.Run(counter, nil, nil)
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}

			Expect(counter.Len()).To(Equal(100))
		})

		It("should be safe for mixed concurrent operations", func() {
			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "run")
			}

			cmd := command.New("test", "description", fn)
			done := make(chan bool, 300)

			// 100 Name calls
			for i := 0; i < 100; i++ {
				go func() {
					_ = cmd.Name()
					done <- true
				}()
			}

			// 100 Describe calls
			for i := 0; i < 100; i++ {
				go func() {
					_ = cmd.Describe()
					done <- true
				}()
			}

			// 100 Run calls
			for i := 0; i < 100; i++ {
				go func() {
					cmd.Run(newSafeBuffer(), nil, nil)
					done <- true
				}()
			}

			for i := 0; i < 300; i++ {
				<-done
			}
		})
	})

	Describe("function behavior edge cases", func() {
		It("should handle function that does nothing", func() {
			fn := func(out, err io.Writer, args []string) {
				// Intentionally empty
			}

			cmd := command.New("noop", "does nothing", fn)
			Expect(func() {
				cmd.Run(nil, nil, nil)
			}).ToNot(Panic())
		})

		It("should handle function that only reads args", func() {
			fn := func(out, err io.Writer, args []string) {
				_ = len(args)
				if len(args) > 0 {
					_ = args[0]
				}
			}

			cmd := command.New("reader", "reads args", fn)
			Expect(func() {
				cmd.Run(nil, nil, []string{"test"})
			}).ToNot(Panic())
		})

		It("should handle function with multiple writes", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				for i := 0; i < 1000; i++ {
					fmt.Fprint(out, "x")
				}
			}

			cmd := command.New("writer", "writes a lot", fn)
			cmd.Run(outBuf, nil, nil)
			Expect(outBuf.Len()).To(Equal(1000))
		})

		It("should handle alternating stdout/stderr writes", func() {
			outBuf := newSafeBuffer()
			errBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				for i := 0; i < 10; i++ {
					fmt.Fprintf(out, "out%d ", i)
					fmt.Fprintf(err, "err%d ", i)
				}
			}

			cmd := command.New("alternate", "alternates output", fn)
			cmd.Run(outBuf, errBuf, nil)
			Expect(outBuf.String()).To(ContainSubstring("out0"))
			Expect(outBuf.String()).To(ContainSubstring("out9"))
			Expect(errBuf.String()).To(ContainSubstring("err0"))
			Expect(errBuf.String()).To(ContainSubstring("err9"))
		})
	})

	Describe("CommandInfo edge cases", func() {
		It("should handle CommandInfo created with extreme values", func() {
			info := command.Info(strings.Repeat("x", 1000), strings.Repeat("y", 1000))
			Expect(info.Name()).To(HaveLen(1000))
			Expect(info.Describe()).To(HaveLen(1000))
		})

		It("should handle CommandInfo with special characters", func() {
			info := command.Info("test\n\r\t", "desc\x00\x01")
			Expect(info.Name()).To(ContainSubstring("\n"))
			Expect(info.Describe()).To(ContainSubstring("\x00"))
		})
	})

	Describe("memory and resource handling", func() {
		It("should not leak memory with large outputs", func() {
			outBuf := newSafeBuffer()
			fn := func(out, err io.Writer, args []string) {
				// Write a large amount of data
				data := strings.Repeat("x", 1000000)
				fmt.Fprint(out, data)
			}

			cmd := command.New("large", "large output", fn)
			cmd.Run(outBuf, nil, nil)
			Expect(outBuf.Len()).To(Equal(1000000))

			// Clear buffer
			outBuf.Reset()
			Expect(outBuf.Len()).To(Equal(0))
		})

		It("should handle rapid repeated calls", func() {
			callCount := 0
			fn := func(out, err io.Writer, args []string) {
				callCount++
			}

			cmd := command.New("rapid", "rapid calls", fn)
			for i := 0; i < 10000; i++ {
				cmd.Run(nil, nil, nil)
			}
			Expect(callCount).To(Equal(10000))
		})
	})

	Describe("interface type assertions", func() {
		It("should allow Command to be used as CommandInfo", func() {
			cmd := command.New("test", "desc", nil)
			var info command.CommandInfo = cmd
			Expect(info.Name()).To(Equal("test"))
			Expect(info.Describe()).To(Equal("desc"))
		})

		It("should allow CommandInfo to be cast to Command if applicable", func() {
			info := command.Info("test", "desc")
			if cmd, ok := info.(command.Command); ok {
				Expect(cmd.Name()).To(Equal("test"))
				cmd.Run(nil, nil, nil) // Should not panic
			}
		})
	})
})
