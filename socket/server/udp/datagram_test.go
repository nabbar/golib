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

package udp_test

import (
	"time"

	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Datagram Handling", func() {
	var (
		srv     libsck.Server
		address string
	)
	BeforeEach(func() {
		address = getTestAddress()
		srv = createAndRegisterServer(address, echoHandler, nil)
		startServer(x, srv)
		waitForServerRunning(srv, 2*time.Second)
	})
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
	})
	It("should handle single datagram", func() {
		data := []byte("test")
		Expect(sendDatagram(address, data)).ToNot(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
	})
	It("should handle multiple datagrams", func() {
		for i := 0; i < 10; i++ {
			Expect(sendDatagram(address, []byte("test"))).ToNot(HaveOccurred())
		}
		time.Sleep(200 * time.Millisecond)
	})
})
