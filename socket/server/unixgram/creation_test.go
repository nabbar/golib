//go:build linux || darwin

/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
 *
 */

package unixgram_test

import (
	"os"

	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unixgram Creation", func() {
	Describe("New", func() {
		It("should create server", func() {
			srv := scksrv.New(nil, echoHandler)

			Expect(srv).ToNot(BeNil())
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})
	Describe("RegisterSocket", func() {
		It("should accept valid path", func() {
			srv := scksrv.New(nil, echoHandler)

			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			Expect(srv.RegisterSocket(path, 0600, -1)).ToNot(HaveOccurred())
		})
		It("should accept different perms", func() {
			srv := scksrv.New(nil, echoHandler)

			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			Expect(srv.RegisterSocket(path, 0777, -1)).ToNot(HaveOccurred())
		})
		It("should reject invalid GID", func() {
			srv := scksrv.New(nil, echoHandler)

			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			Expect(srv.RegisterSocket(path, 0600, 99999)).To(HaveOccurred())
		})
	})
	Describe("Initial State", func() {
		It("not running", func() {
			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			srv := createAndRegisterServer(path, echoHandler)
			Expect(srv.IsRunning()).To(BeFalse())
		})
		It("is gone initially", func() {
			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			srv := createAndRegisterServer(path, echoHandler)
			Expect(srv.IsGone()).To(BeTrue())
		})
		It("zero connections", func() {
			path := getTempSocketPath()
			defer func() {
				_ = os.Remove(path)
			}()

			srv := createAndRegisterServer(path, echoHandler)
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})
	})
})
