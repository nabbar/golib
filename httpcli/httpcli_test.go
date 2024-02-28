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
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	libdur "github.com/nabbar/golib/duration"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cli *http.Client
	err error
	opt htcdns.Config
	dns htcdns.DNSMapper
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
			opt = htcdns.Config{
				DNSMapper:  make(map[string]string),
				TimerClean: libdur.ParseDuration(30 * time.Second),
				Transport:  htcdns.TransportConfig{},
			}

			dns = htcdns.New(ctx, &opt, nil, func(msg string) {
				_, _ = fmt.Fprintln(os.Stdout, msg)
			})

			cli = dns.DefaultClient()
			Expect(cli).ToNot(BeNil())
		})
		It("Get URL must succeed for DNS Mapper with host", func() {
			dns.Del("test.me.example.com:80")
			dns.Add("test.me.example.com", "127.0.0.1:8080")

			rsp, err = cli.Get("http://test.me.example.com/path/any/thing")
			Expect(err).ToNot(HaveOccurred())
			Expect(rsp).ToNot(BeNil())
		})
		It("Get URL must succeed for DNS Mapper with host & port", func() {
			dns.Del("test.me.example.com")
			dns.Add("test.me.example.com:80", "127.0.0.1:8080")

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
