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
 *
 */

package httpcli_test

import (
	"fmt"
	"os"
	"time"

	libdur "github.com/nabbar/golib/duration"
	. "github.com/nabbar/golib/httpcli"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Management", func() {
	Describe("GetClient", func() {
		It("should get default client", func() {
			client := GetClient()

			Expect(client).ToNot(BeNil())
		})

		It("should return same client on multiple calls", func() {
			client1 := GetClient()
			client2 := GetClient()

			// Both should be non-nil
			Expect(client1).ToNot(BeNil())
			Expect(client2).ToNot(BeNil())
		})
	})

	Describe("DNS Mapper Management", func() {
		var originalMapper htcdns.DNSMapper

		BeforeEach(func() {
			// Save original mapper
			originalMapper = DefaultDNSMapper()
		})

		AfterEach(func() {
			// Restore original mapper
			if originalMapper != nil {
				SetDefaultDNSMapper(originalMapper)
			}
		})

		It("should get default DNS mapper", func() {
			mapper := DefaultDNSMapper()

			Expect(mapper).ToNot(BeNil())
		})

		It("should set custom DNS mapper", func() {
			cfg := &htcdns.Config{
				DNSMapper:  make(map[string]string),
				TimerClean: libdur.ParseDuration(5 * time.Minute),
				Transport:  htcdns.TransportConfig{},
			}

			customMapper := htcdns.New(ctx, cfg, nil, func(msg string) {
				fmt.Fprintln(os.Stdout, msg)
			})

			SetDefaultDNSMapper(customMapper)

			mapper := DefaultDNSMapper()
			Expect(mapper).ToNot(BeNil())
		})

		It("should not set nil DNS mapper", func() {
			original := DefaultDNSMapper()

			SetDefaultDNSMapper(nil)

			// Should still have the original
			current := DefaultDNSMapper()
			Expect(current).ToNot(BeNil())
			Expect(current).To(Equal(original))
		})

		It("should replace DNS mapper on set", func() {
			cfg1 := &htcdns.Config{
				DNSMapper: map[string]string{
					"test1.com:80": "127.0.0.1:8080",
				},
				TimerClean: libdur.ParseDuration(5 * time.Minute),
			}
			mapper1 := htcdns.New(ctx, cfg1, nil, nil)

			SetDefaultDNSMapper(mapper1)

			cfg2 := &htcdns.Config{
				DNSMapper: map[string]string{
					"test2.com:80": "127.0.0.1:9090",
				},
				TimerClean: libdur.ParseDuration(5 * time.Minute),
			}
			mapper2 := htcdns.New(ctx, cfg2, nil, nil)

			SetDefaultDNSMapper(mapper2)

			current := DefaultDNSMapper()
			Expect(current).ToNot(BeNil())
		})
	})

	Describe("Client Timeout", func() {
		It("should have default timeout constant", func() {
			Expect(ClientTimeout5Sec).To(Equal(5 * time.Second))
		})
	})

	Describe("Integration with DNS Mapper", func() {
		var testMapper htcdns.DNSMapper

		BeforeEach(func() {
			cfg := &htcdns.Config{
				DNSMapper: map[string]string{
					"integration.test:80": "127.0.0.1:8080",
				},
				TimerClean: libdur.ParseDuration(10 * time.Minute),
				Transport:  htcdns.TransportConfig{},
			}

			testMapper = htcdns.New(ctx, cfg, nil, nil)
			SetDefaultDNSMapper(testMapper)
		})

		AfterEach(func() {
			if testMapper != nil {
				testMapper.Close()
			}
		})

		It("should use DNS mapper for client creation", func() {
			client := GetClient()

			Expect(client).ToNot(BeNil())
			Expect(client.Transport).ToNot(BeNil())
		})
	})
})
