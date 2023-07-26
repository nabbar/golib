/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package httpcli_test

import (
	"io"
	"net/http"

	libptc "github.com/nabbar/golib/network/protocol"

	libtls "github.com/nabbar/golib/certificates"
	"github.com/nabbar/golib/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cli *http.Client
	err error
	opt httpcli.Options
	rsp *http.Response
)

var _ = Describe("HttpCli", func() {
	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()
	Context("Get URL with a Force IP", func() {
		It("Create new client must succeed", func() {
			opt = httpcli.Options{
				Timeout: 0,
				Http2:   true,
				TLS: httpcli.OptionTLS{
					Enable: false,
					Config: libtls.Config{},
				},
				ForceIP: httpcli.OptionForceIP{
					Enable: true,
					Net:    libptc.NetworkTCP,
					IP:     "127.0.0.1:8080",
				},
			}
			cli, err = opt.GetClient(nil, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})
		It("Get URL must succeed", func() {
			rsp, err = cli.Get("http://test.me.example.com/path/any/thing")
			Expect(err).ToNot(HaveOccurred())
			Expect(rsp).ToNot(BeNil())
		})
		It("Result must succeed", func() {
			Expect(rsp.Body).ToNot(BeNil())
			p, e := io.ReadAll(rsp.Body)
			Expect(e).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			Expect(p).ToNot(BeEmpty())
		})
	})
})
