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

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Add Method", func() {
	var sh shell.Shell

	BeforeEach(func() {
		sh = shell.New(nil)
	})

	Describe("single command addition", func() {
		It("should add command without prefix", func() {
			cmd := command.New("test", "Test command", nil)
			sh.Add("", cmd)

			retrieved, found := sh.Get("test")
			Expect(found).To(BeTrue())
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.Name()).To(Equal("test"))
		})

		It("should add command with prefix", func() {
			cmd := command.New("info", "Info command", nil)
			sh.Add("sys:", cmd)

			retrieved, found := sh.Get("sys:info")
			Expect(found).To(BeTrue())
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.Name()).To(Equal("info"))
		})

		It("should allow command with empty name", func() {
			cmd := command.New("", "Empty name", nil)
			sh.Add("", cmd)

			retrieved, found := sh.Get("")
			Expect(found).To(BeTrue())
			Expect(retrieved).ToNot(BeNil())
		})
	})

	Describe("multiple command addition", func() {
		It("should add multiple commands at once", func() {
			cmd1 := command.New("cmd1", "Command 1", nil)
			cmd2 := command.New("cmd2", "Command 2", nil)
			cmd3 := command.New("cmd3", "Command 3", nil)

			sh.Add("", cmd1, cmd2, cmd3)

			c1, found1 := sh.Get("cmd1")
			Expect(found1).To(BeTrue())
			Expect(c1.Name()).To(Equal("cmd1"))

			c2, found2 := sh.Get("cmd2")
			Expect(found2).To(BeTrue())
			Expect(c2.Name()).To(Equal("cmd2"))

			c3, found3 := sh.Get("cmd3")
			Expect(found3).To(BeTrue())
			Expect(c3.Name()).To(Equal("cmd3"))
		})

		It("should add multiple commands with prefix", func() {
			cmd1 := command.New("list", "List command", nil)
			cmd2 := command.New("show", "Show command", nil)

			sh.Add("sys:", cmd1, cmd2)

			c1, found1 := sh.Get("sys:list")
			Expect(found1).To(BeTrue())
			Expect(c1.Name()).To(Equal("list"))

			c2, found2 := sh.Get("sys:show")
			Expect(found2).To(BeTrue())
			Expect(c2.Name()).To(Equal("show"))
		})

		It("should skip nil commands", func() {
			cmd1 := command.New("cmd1", "Command 1", nil)
			cmd2 := command.New("cmd2", "Command 2", nil)

			sh.Add("", cmd1, nil, nil, cmd2, nil)

			c1, found1 := sh.Get("cmd1")
			Expect(found1).To(BeTrue())
			Expect(c1).ToNot(BeNil())

			c2, found2 := sh.Get("cmd2")
			Expect(found2).To(BeTrue())
			Expect(c2).ToNot(BeNil())
		})
	})

	Describe("command replacement", func() {
		It("should replace existing command", func() {
			cmd1 := command.New("test", "First version", nil)
			cmd2 := command.New("test", "Second version", nil)

			sh.Add("", cmd1)
			sh.Add("", cmd2)

			retrieved, found := sh.Get("test")
			Expect(found).To(BeTrue())
			Expect(retrieved.Describe()).To(Equal("Second version"))
		})

		It("should replace command with same prefix", func() {
			cmd1 := command.New("cmd", "Version 1", nil)
			cmd2 := command.New("cmd", "Version 2", nil)

			sh.Add("prefix:", cmd1)
			sh.Add("prefix:", cmd2)

			retrieved, found := sh.Get("prefix:cmd")
			Expect(found).To(BeTrue())
			Expect(retrieved.Describe()).To(Equal("Version 2"))
		})
	})

	Describe("concurrent additions", func() {
		It("should handle concurrent Add operations", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(idx int) {
					cmd := command.New(fmt.Sprintf("cmd%d", idx), fmt.Sprintf("Command %d", idx), nil)
					sh.Add("", cmd)
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			// Verify at least some commands were added
			count := 0
			sh.Walk(func(name string, item command.Command) bool {
				count++
				return true
			})
			Expect(count).To(BeNumerically(">", 0))
		})
	})
})
