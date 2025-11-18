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
	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get and Desc Methods", func() {
	var sh shell.Shell

	BeforeEach(func() {
		sh = shell.New(nil)
		sh.Add("", command.New("hello", "Say hello", nil))
		sh.Add("", command.New("echo", "Echo text", nil))
		sh.Add("sys:", command.New("info", "System info", nil))
		sh.Add("sys:", command.New("status", "System status", nil))
	})

	Describe("Get method", func() {
		It("should get existing command", func() {
			cmd, found := sh.Get("hello")
			Expect(found).To(BeTrue())
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Name()).To(Equal("hello"))
		})

		It("should get command with prefix", func() {
			cmd, found := sh.Get("sys:info")
			Expect(found).To(BeTrue())
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Name()).To(Equal("info"))
		})

		It("should return false for non-existent command", func() {
			_, found := sh.Get("nonexistent")
			Expect(found).To(BeFalse())
		})

		It("should return false for empty command name on non-empty shell", func() {
			_, found := sh.Get("")
			Expect(found).To(BeFalse())
		})

		It("should handle concurrent Get operations", func() {
			done := make(chan bool, 100)

			for i := 0; i < 100; i++ {
				go func() {
					cmd, found := sh.Get("hello")
					Expect(found).To(BeTrue())
					Expect(cmd).ToNot(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}
		})
	})

	Describe("Desc method", func() {
		It("should get description for existing command", func() {
			desc := sh.Desc("hello")
			Expect(desc).To(Equal("Say hello"))
		})

		It("should get description for command with prefix", func() {
			desc := sh.Desc("sys:info")
			Expect(desc).To(Equal("System info"))
		})

		It("should return empty string for non-existent command", func() {
			desc := sh.Desc("nonexistent")
			Expect(desc).To(BeEmpty())
		})

		It("should handle concurrent Desc operations", func() {
			done := make(chan bool, 100)

			for i := 0; i < 100; i++ {
				go func() {
					desc := sh.Desc("hello")
					Expect(desc).To(Equal("Say hello"))
					done <- true
				}()
			}

			for i := 0; i < 100; i++ {
				<-done
			}
		})
	})

	Describe("empty shell behavior", func() {
		It("should return false on Get for empty shell", func() {
			emptyShell := shell.New(nil)
			_, found := emptyShell.Get("test")
			Expect(found).To(BeFalse())
		})

		It("should return empty string on Desc for empty shell", func() {
			emptyShell := shell.New(nil)
			desc := emptyShell.Desc("test")
			Expect(desc).To(BeEmpty())
		})
	})
})
