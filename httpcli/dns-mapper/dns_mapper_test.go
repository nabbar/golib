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

package dns_mapper_test

import (
	"fmt"
	"os"
	"time"

	libdur "github.com/nabbar/golib/duration"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	idx int = 1
	opt htcdns.Config
	dns htcdns.DNSMapper
)

func init() {
	opt = htcdns.Config{
		DNSMapper:  make(map[string]string),
		TimerClean: libdur.ParseDuration(1000 * time.Hour),
		Transport:  htcdns.TransportConfig{},
	}

	addDns(addIdx("127.0.0."), addIdx("8"), addIdx("127.0.0."), "")
	idx++

	addDns("test.example.com", numIdx("8"), addIdx("127.0.0."), "")
	idx++
	addDns("test.example.com", numIdx("*"), addIdx("127.0.0."), "")
	idx++
	addDns("*.test.example.com", numIdx("8"), addIdx("127.0.0."), "")
	idx++
	addDns("*.test.example.com", numIdx("*"), addIdx("127.0.0."), "")
	idx++

	addDns("test.example.com", numIdx("8"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("test.example.com", numIdx("*"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("*.test.example.com", numIdx("8"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("*.test.example.com", numIdx("*"), addIdx("127.0.0."), numIdx("8"))
	idx++

	addDns("*.test.example.com", numIdx("8"), addIdx("127.0.0."), "")
	idx++
	addDns("*.test.example.com", numIdx("*"), addIdx("127.0.0."), "")
	idx++
	addDns("*.*.test.example.com", numIdx("8"), addIdx("127.0.0."), "")
	idx++
	addDns("*.*.test.example.com", numIdx("*"), addIdx("127.0.0."), "")
	idx++

	addDns("*.test.example.com", numIdx("8"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("*.test.example.com", numIdx("*"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("*.*.test.example.com", numIdx("8"), addIdx("127.0.0."), numIdx("8"))
	idx++
	addDns("*.*.test.example.com", numIdx("*"), addIdx("127.0.0."), numIdx("8"))
	idx++

	dns = htcdns.New(ctx, &opt, nil, func(msg string) {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	})
}

func addIdx(src string) string {
	return fmt.Sprintf("%s%d", src, idx)
}

func numIdx(src string) string {
	if idx < 10 {
		return fmt.Sprintf("%s0%d", src, idx)
	} else {
		return fmt.Sprintf("%s%d", src, idx)
	}
}

func addDns(hostSrc, portSrc, hostDst, portDst string) {
	if portSrc != "" {
		hostSrc = hostSrc + ":" + portSrc
	}

	if portDst != "" {
		hostDst = hostDst + ":" + portDst
	}

	opt.DNSMapper[hostSrc] = hostDst
}

var _ = Describe("DNS Mapper", func() {
	Context("check dns mapper", func() {
		It("be a valid dns mapper", func() {
			Expect(dns).NotTo(BeNil())
		})

		It(fmt.Sprintf("must having '%d' item in dns mapper", idx-2), func() {
			Expect(dns.Len()).To(BeIdenticalTo(idx - 2))
		})

		It("must having one more item if adding a valid item", func() {
			l := dns.Len()
			dns.Add("*.localhost", "127.0.0.1:8080")
			Expect(dns.Len()).To(BeIdenticalTo(l + 1))
		})

		It("must having item just adding and can return it", func() {
			Expect(dns.Get("*.localhost")).To(BeIdenticalTo("127.0.0.1:8080"))
		})

		It("must return a new client to dial to localserver", func() {
			cli := dns.DefaultClient()
			Expect(cli).NotTo(BeNil())

			rsp, err := cli.Get("http://test.localhost")

			Expect(err).NotTo(HaveOccurred())
			Expect(rsp).NotTo(BeNil())
			Expect(rsp.Body).NotTo(BeNil())
			Expect(rsp.Body.Close()).NotTo(HaveOccurred())
		})

		It(fmt.Sprintf("must having '%d' item in dns mapper if delete just added item", idx-2), func() {
			dns.Del("*.localhost")
			Expect(dns.Len()).To(BeIdenticalTo(idx - 2))
		})

		It("must fail for testing source as ip without port '127.0.0.128'", func() {
			var (
				err  error
				src  = "127.0.0.128"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).To(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).To(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(""))

			res2, err = dns.SearchWithCache(src)
			Expect(err).To(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(""))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing source as ip '127.0.0.128:9001' and result must be source ip:port", func() {
			var (
				err  error
				src  = "127.0.0.128:9001"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(src))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(src))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must failt for testing 'test.example.com' without port", func() {
			var (
				err  error
				src  = "test.example.com"
				res1 string
				res2 string
			)

			_, _, err = dns.Clean(src)
			Expect(err).To(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).To(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(""))

			res2, err = dns.SearchWithCache(src)
			Expect(err).To(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(""))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'test.example.com:801' with same source and destination", func() {
			var (
				err  error
				src  = "test.example.com:801"
				res1 string
				res2 string
			)

			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(src))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(src))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'test.example.com:802'", func() {
			var (
				err  error
				src  = "test.example.com:802"
				ctr  = "127.0.0.2:802"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'test.example.com:803'", func() {
			var (
				err  error
				src  = "test.example.com:803"
				ctr  = "127.0.0.3:803"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'any.test.example.com:804'", func() {
			var (
				err  error
				src  = "any.test.example.com:804"
				ctr  = "127.0.0.4:804"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'any.test.example.com:805'", func() {
			var (
				err  error
				src  = "any.test.example.com:805"
				ctr  = "127.0.0.5:805"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'test.example.com:806'", func() {
			var (
				err  error
				src  = "test.example.com:806"
				ctr  = "127.0.0.6:806"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'test.example.com:807'", func() {
			var (
				err  error
				src  = "test.example.com:807"
				ctr  = "127.0.0.7:807"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'any.test.example.com:808'", func() {
			var (
				err  error
				src  = "any.test.example.com:808"
				ctr  = "127.0.0.8:808"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'any.test.example.com:809'", func() {
			var (
				err  error
				src  = "any.test.example.com:809"
				ctr  = "127.0.0.9:809"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.test.example.com:810'", func() {
			var (
				err  error
				src  = "one.test.example.com:810"
				ctr  = "127.0.0.10:810"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.test.example.com:811'", func() {
			var (
				err  error
				src  = "one.test.example.com:811"
				ctr  = "127.0.0.11:811"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.any.test.example.com:812'", func() {
			var (
				err  error
				src  = "one.any.test.example.com:812"
				ctr  = "127.0.0.12:812"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.any.test.example.com:813'", func() {
			var (
				err  error
				src  = "one.any.test.example.com:813"
				ctr  = "127.0.0.13:813"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.test.example.com:814'", func() {
			var (
				err  error
				src  = "one.test.example.com:814"
				ctr  = "127.0.0.14:814"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.test.example.com:815'", func() {
			var (
				err  error
				src  = "one.test.example.com:815"
				ctr  = "127.0.0.15:815"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.any.test.example.com:816'", func() {
			var (
				err  error
				src  = "one.any.test.example.com:816"
				ctr  = "127.0.0.16:816"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})

		It("must success for testing 'one.any.test.example.com:817'", func() {
			var (
				err  error
				src  = "one.any.test.example.com:817"
				ctr  = "127.0.0.17:817"
				res1 string
				res2 string
			)
			_, _, err = dns.Clean(src)
			Expect(err).ToNot(HaveOccurred())

			res1, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1).To(BeIdenticalTo(ctr))

			res2, err = dns.SearchWithCache(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2).To(BeIdenticalTo(ctr))
			Expect(res2).To(BeIdenticalTo(res1))
		})
	})
})
