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

package httpcli_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tlscas "github.com/nabbar/golib/certificates/ca"
	libcfg "github.com/nabbar/golib/config"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

// Helper and edge case tests verify internal functions, root CA handling,
// registration functions, and various edge cases.
var _ = Describe("Helper Functions and Edge Cases", func() {
	var (
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
	})

	Describe("GetRootCaCert function", func() {
		Context("with nil function", func() {
			It("should handle nil root CA function", func() {
				fct := func() []string { return nil }
				cert := GetRootCaCert(fct)
				Expect(cert).To(BeNil())
			})
		})

		Context("with empty list", func() {
			It("should handle empty certificate list", func() {
				fct := func() []string { return []string{} }
				cert := GetRootCaCert(fct)
				Expect(cert).To(BeNil())
			})
		})

		Context("with valid certificates", func() {
			It("should parse single certificate", func() {
				// Simple PEM certificate for testing
				pemCert := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU3eGMA0GCSqGSIb3DQEBBQUAMA0xCzAJBgNVBAYTAlVT
MB4XDTA5MDUwNjE1NDUwNloXDTEwMDUwNjE1NDUwNlowDTELMAkGA1UEBhMCVVMw
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALMiPHmLATCYg6vfb1eGJqHTW4dL
j8x2xjkD4I5SHxOPbBCE4jQLO23JTmVuKwqZGmJe5eKgHxBxGi2wLcJvJ3q9jDjX
lxS3V9YqJzVYq6CJfg5V7Kj3KKj3FYBIVEHblPwKBp5V1BUjr0KJQagCpUJBo4GE
9DqKRLLH2rXQVk3xAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAg4VkLBQjVdDpZVCM
wk6jVqMl3k7YgNfQb5Ar7r5oByvhLmTEckE4Q7bVD9DQHK3L9LQXMR4K6D7MiBUo
2lPG7Xx2vFvBnvQ1KKJxc7PAE7WS5xPV0HPVLzQCwPjJZKXKJgLHdqVmDqWLHPJQ
kGJZQp3MJ7RpJvVcRjJZCPvSbLM=
-----END CERTIFICATE-----`

				fct := func() []string { return []string{pemCert} }
				cert := GetRootCaCert(fct)
				Expect(cert).NotTo(BeNil())
			})

			It("should combine multiple certificates", func() {
				pemCert := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU3eGMA0GCSqGSIb3DQEBBQUAMA0xCzAJBgNVBAYTAlVT
MB4XDTA5MDUwNjE1NDUwNloXDTEwMDUwNjE1NDUwNlowDTELMAkGA1UEBhMCVVMw
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALMiPHmLATCYg6vfb1eGJqHTW4dL
j8x2xjkD4I5SHxOPbBCE4jQLO23JTmVuKwqZGmJe5eKgHxBxGi2wLcJvJ3q9jDjX
lxS3V9YqJzVYq6CJfg5V7Kj3KKj3FYBIVEHblPwKBp5V1BUjr0KJQagCpUJBo4GE
9DqKRLLH2rXQVk3xAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAg4VkLBQjVdDpZVCM
wk6jVqMl3k7YgNfQb5Ar7r5oByvhLmTEckE4Q7bVD9DQHK3L9LQXMR4K6D7MiBUo
2lPG7Xx2vFvBnvQ1KKJxc7PAE7WS5xPV0HPVLzQCwPjJZKXKJgLHdqVmDqWLHPJQ
kGJZQp3MJ7RpJvVcRjJZCPvSbLM=
-----END CERTIFICATE-----`

				fct := func() []string { return []string{pemCert, pemCert} }
				cert := GetRootCaCert(fct)
				Expect(cert).NotTo(BeNil())
			})
		})
	})

	Describe("Registration functions", func() {
		Context("Register function", func() {
			It("should register component in config", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx, nil, false, nil)
				key := "test-httpcli"

				Register(cfg, key, cpt)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})
		})

		Context("RegisterNew function", func() {
			It("should create and register component", func() {
				cfg := libcfg.New(vrs)
				key := "test-httpcli"

				RegisterNew(ctx, cfg, key, nil, false, nil)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
			})

			It("should register with custom root CA", func() {
				cfg := libcfg.New(vrs)
				key := "test-httpcli"
				rootCA := func() tlscas.Cert { return nil }

				RegisterNew(ctx, cfg, key, rootCA, true, nil)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
			})

			It("should register with message function", func() {
				cfg := libcfg.New(vrs)
				key := "test-httpcli"
				msg := func(s string) {}

				RegisterNew(ctx, cfg, key, nil, false, msg)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
			})
		})

		Context("Load function", func() {
			It("should return nil with nil getter", func() {
				loaded := Load(nil, "test")
				Expect(loaded).To(BeNil())
			})

			It("should return nil for non-existent key", func() {
				cfg := libcfg.New(vrs)
				loaded := Load(cfg.ComponentGet, "non-existent")
				Expect(loaded).To(BeNil())
			})

			It("should return nil for wrong component type", func() {
				cfg := libcfg.New(vrs)
				cfg.ComponentSet("wrong", &sharedWrongComponent{})
				loaded := Load(cfg.ComponentGet, "wrong")
				Expect(loaded).To(BeNil())
			})
		})
	})

	Describe("TransportWithTLS", func() {
		Context("before component start", func() {
			It("should return nil transport", func() {
				cpt := New(ctx, nil, false, nil)
				cfg := htcdns.TransportConfig{}
				tlsConfig := &tls.Config{}

				transport := cpt.TransportWithTLS(cfg, tlsConfig)
				Expect(transport).To(BeNil())
			})
		})

		Context("after component start", func() {
			It("should create transport with TLS config", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

				key := "test-httpcli"
				configData := map[string]interface{}{
					key: map[string]interface{}{
						"timeOut":   30,
						"keepAlive": 30,
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				cpt := New(ctx, nil, false, nil)
				vpr := func() libvpr.Viper {
					return &sharedMockViper{v: v}
				}

				cpt.Init(key, ctx, nil, vpr, vrs, nil)
				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())

				cfg := htcdns.TransportConfig{}
				tlsConfig := &tls.Config{
					InsecureSkipVerify: true,
				}

				transport := cpt.TransportWithTLS(cfg, tlsConfig)
				Expect(transport).NotTo(BeNil())
				Expect(transport.TLSClientConfig).NotTo(BeNil())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("nil component operations", func() {
			It("should handle operations on nil DNS mapper gracefully", func() {
				cpt := New(ctx, nil, false, nil)

				// All these should not panic even without DNS mapper
				Expect(func() {
					cpt.Add("test.com", "127.0.0.1")
					_ = cpt.Get("test.com")
					cpt.Del("test.com")
					_ = cpt.Len()
					cpt.Walk(func(from, to string) bool { return true })
				}).NotTo(Panic())
			})
		})

		Context("multiple start calls", func() {
			It("should handle multiple start calls", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

				key := "test-httpcli"
				configData := map[string]interface{}{
					key: map[string]interface{}{
						"timeOut": 30,
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				cpt := New(ctx, nil, false, nil)
				vpr := func() libvpr.Viper {
					return &sharedMockViper{v: v}
				}

				cpt.Init(key, ctx, nil, vpr, vrs, nil)

				// Multiple starts should not cause issues
				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())

				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Monitor pool registration", func() {
		Context("RegisterMonitorPool", func() {
			It("should not panic when registering monitor pool", func() {
				cpt := New(ctx, nil, false, nil)

				Expect(func() {
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})
		})
	})
})
