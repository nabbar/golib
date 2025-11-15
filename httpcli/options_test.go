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
	"net/url"
	"time"

	. "github.com/nabbar/golib/httpcli"
	libptc "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options", func() {
	Describe("Options Creation", func() {
		It("should create options with default values", func() {
			opts := Options{}

			Expect(opts.Timeout).To(Equal(time.Duration(0)))
			Expect(opts.DisableKeepAlive).To(BeFalse())
			Expect(opts.DisableCompression).To(BeFalse())
			Expect(opts.Http2).To(BeFalse())
		})

		It("should create options with custom timeout", func() {
			opts := Options{
				Timeout: 30 * time.Second,
			}

			Expect(opts.Timeout).To(Equal(30 * time.Second))
		})

		It("should create options with disabled keep-alive", func() {
			opts := Options{
				DisableKeepAlive: true,
			}

			Expect(opts.DisableKeepAlive).To(BeTrue())
		})

		It("should create options with disabled compression", func() {
			opts := Options{
				DisableCompression: true,
			}

			Expect(opts.DisableCompression).To(BeTrue())
		})

		It("should create options with HTTP/2 enabled", func() {
			opts := Options{
				Http2: true,
			}

			Expect(opts.Http2).To(BeTrue())
		})
	})

	Describe("TLS Options", func() {
		It("should create TLS options disabled by default", func() {
			opts := Options{}

			Expect(opts.TLS.Enable).To(BeFalse())
		})

		It("should enable TLS", func() {
			opts := Options{
				TLS: OptionTLS{
					Enable: true,
				},
			}

			Expect(opts.TLS.Enable).To(BeTrue())
		})
	})

	Describe("ForceIP Options", func() {
		It("should create ForceIP options disabled by default", func() {
			opts := Options{}

			Expect(opts.ForceIP.Enable).To(BeFalse())
		})

		It("should enable ForceIP with IPv4", func() {
			opts := Options{
				ForceIP: OptionForceIP{
					Enable: true,
					Net:    libptc.NetworkTCP4,
					IP:     "192.168.1.100",
				},
			}

			Expect(opts.ForceIP.Enable).To(BeTrue())
			Expect(opts.ForceIP.Net).To(Equal(libptc.NetworkTCP4))
			Expect(opts.ForceIP.IP).To(Equal("192.168.1.100"))
		})

		It("should enable ForceIP with IPv6", func() {
			opts := Options{
				ForceIP: OptionForceIP{
					Enable: true,
					Net:    libptc.NetworkTCP6,
					IP:     "::1",
				},
			}

			Expect(opts.ForceIP.Enable).To(BeTrue())
			Expect(opts.ForceIP.Net).To(Equal(libptc.NetworkTCP6))
			Expect(opts.ForceIP.IP).To(Equal("::1"))
		})

		It("should set local address", func() {
			opts := Options{
				ForceIP: OptionForceIP{
					Enable: true,
					Local:  "192.168.1.10",
				},
			}

			Expect(opts.ForceIP.Local).To(Equal("192.168.1.10"))
		})
	})

	Describe("Proxy Options", func() {
		It("should create proxy options disabled by default", func() {
			opts := Options{}

			Expect(opts.Proxy.Enable).To(BeFalse())
		})

		It("should enable proxy with endpoint", func() {
			proxyURL, _ := url.Parse("http://proxy.example.com:8080")
			opts := Options{
				Proxy: OptionProxy{
					Enable:   true,
					Endpoint: proxyURL,
				},
			}

			Expect(opts.Proxy.Enable).To(BeTrue())
			Expect(opts.Proxy.Endpoint).To(Equal(proxyURL))
		})

		It("should enable proxy with credentials", func() {
			proxyURL, _ := url.Parse("http://proxy.example.com:8080")
			opts := Options{
				Proxy: OptionProxy{
					Enable:   true,
					Endpoint: proxyURL,
					Username: "proxyuser",
					Password: "proxypass",
				},
			}

			Expect(opts.Proxy.Username).To(Equal("proxyuser"))
			Expect(opts.Proxy.Password).To(Equal("proxypass"))
		})
	})

	Describe("Options Validation", func() {
		It("should validate empty options", func() {
			opts := Options{}

			err := opts.Validate()
			Expect(err).To(BeNil())
		})

		It("should validate complete options", func() {
			opts := Options{
				Timeout:            30 * time.Second,
				DisableKeepAlive:   false,
				DisableCompression: false,
				Http2:              true,
				TLS: OptionTLS{
					Enable: false,
				},
				ForceIP: OptionForceIP{
					Enable: false,
				},
				Proxy: OptionProxy{
					Enable: false,
				},
			}

			err := opts.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("GetClient", func() {
		It("should get client from options", func() {
			opts := Options{
				Timeout: 10 * time.Second,
			}

			client, err := opts.GetClient(nil, "")
			Expect(err).To(BeNil())
			Expect(client).ToNot(BeNil())
		})

		It("should get client with servername", func() {
			opts := Options{}

			client, err := opts.GetClient(nil, "example.com")
			Expect(err).To(BeNil())
			Expect(client).ToNot(BeNil())
		})
	})

	Describe("DefaultConfig", func() {
		It("should generate default config", func() {
			config := DefaultConfig("")

			Expect(config).ToNot(BeNil())
			Expect(config).ToNot(BeEmpty())
		})

		It("should generate default config with indent", func() {
			config := DefaultConfig("  ")

			Expect(config).ToNot(BeNil())
			Expect(config).ToNot(BeEmpty())
		})
	})

	Describe("Complete Examples", func() {
		It("should create minimal config for development", func() {
			opts := Options{
				Timeout: 5 * time.Second,
			}

			Expect(opts.Timeout).To(Equal(5 * time.Second))
		})

		It("should create config for production with all options", func() {
			proxyURL, _ := url.Parse("http://proxy.corp.com:8080")

			opts := Options{
				Timeout:            30 * time.Second,
				DisableKeepAlive:   false,
				DisableCompression: false,
				Http2:              true,
				TLS: OptionTLS{
					Enable: true,
				},
				ForceIP: OptionForceIP{
					Enable: false,
				},
				Proxy: OptionProxy{
					Enable:   true,
					Endpoint: proxyURL,
					Username: "user",
					Password: "pass",
				},
			}

			Expect(opts.Timeout).To(Equal(30 * time.Second))
			Expect(opts.Http2).To(BeTrue())
			Expect(opts.TLS.Enable).To(BeTrue())
			Expect(opts.Proxy.Enable).To(BeTrue())
		})
	})
})
