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
	"encoding/json"
	"net/http"

	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

// DNS Mapper tests verify DNS mapping functionality, transport creation,
// and client operations after component start.
var _ = Describe("DNS Mapper Functionality", func() {
	var (
		cpt CptHTTPClient
		ctx context.Context
		vrs libver.Version
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		key = "test-httpcli"
		cpt = New(ctx, nil, false, nil)
	})

	// Helper to start component with valid config
	startComponent := func() {
		v := spfvpr.New()
		v.SetConfigType("json")

		configData := map[string]interface{}{
			key: map[string]interface{}{
				"timeOut":         30,
				"keepAlive":       30,
				"maxIdleConns":    100,
				"idleConnTimeout": 90,
			},
		}

		configJSON, err := json.Marshal(configData)
		Expect(err).To(BeNil())

		err = v.ReadConfig(bytes.NewReader(configJSON))
		Expect(err).To(BeNil())

		vpr := func() libvpr.Viper {
			return &sharedMockViper{v: v}
		}

		cpt.Init(key, ctx, nil, vpr, vrs, nil)
		err = cpt.Start()
		Expect(err).NotTo(HaveOccurred())
	}

	Describe("DNS mapping operations", func() {
		Context("before component start", func() {
			It("should handle Add without panic", func() {
				Expect(func() {
					cpt.Add("example.com", "127.0.0.1")
				}).NotTo(Panic())
			})

			It("should return empty string for Get", func() {
				result := cpt.Get("example.com")
				Expect(result).To(Equal(""))
			})

			It("should handle Del without panic", func() {
				Expect(func() {
					cpt.Del("example.com")
				}).NotTo(Panic())
			})

			It("should return zero for Len", func() {
				length := cpt.Len()
				Expect(length).To(Equal(0))
			})
		})

		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should add DNS mapping", func() {
				cpt.Add("example.com", "192.168.1.1")
				result := cpt.Get("example.com")
				Expect(result).To(Equal("192.168.1.1"))
			})

			It("should delete DNS mapping", func() {
				cpt.Add("example.com", "192.168.1.1")
				cpt.Del("example.com")
				result := cpt.Get("example.com")
				Expect(result).To(Equal(""))
			})

			It("should return correct length", func() {
				cpt.Add("example.com", "192.168.1.1")
				cpt.Add("test.com", "192.168.1.2")
				length := cpt.Len()
				Expect(length).To(Equal(2))
			})

			It("should walk through mappings", func() {
				cpt.Add("example.com", "192.168.1.1")
				cpt.Add("test.com", "192.168.1.2")

				count := 0
				cpt.Walk(func(from, to string) bool {
					count++
					Expect(from).NotTo(BeEmpty())
					Expect(to).NotTo(BeEmpty())
					return true
				})

				Expect(count).To(Equal(2))
			})
		})
	})

	Describe("Transport and Client operations", func() {
		Context("before component start", func() {
			It("should return nil for Transport", func() {
				cfg := htcdns.TransportConfig{}
				transport := cpt.Transport(cfg)
				Expect(transport).To(BeNil())
			})

			It("should return nil for Client", func() {
				cfg := htcdns.TransportConfig{}
				client := cpt.Client(cfg)
				Expect(client).To(BeNil())
			})

			It("should return nil for DefaultTransport", func() {
				transport := cpt.DefaultTransport()
				Expect(transport).To(BeNil())
			})

			It("should return nil for DefaultClient", func() {
				client := cpt.DefaultClient()
				Expect(client).To(BeNil())
			})
		})

		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should create custom transport", func() {
				cfg := htcdns.TransportConfig{
					DisableKeepAlive:    false,
					DisableCompression:  false,
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 10,
					MaxConnsPerHost:     0,
				}

				transport := cpt.Transport(cfg)
				Expect(transport).NotTo(BeNil())
				Expect(transport.MaxIdleConns).To(Equal(100))
			})

			It("should create custom client", func() {
				cfg := htcdns.TransportConfig{}
				client := cpt.Client(cfg)
				Expect(client).NotTo(BeNil())
			})

			It("should create default transport", func() {
				transport := cpt.DefaultTransport()
				Expect(transport).NotTo(BeNil())
			})

			It("should create default client", func() {
				client := cpt.DefaultClient()
				Expect(client).NotTo(BeNil())
			})

			It("should register transport", func() {
				transport := &http.Transport{}
				Expect(func() {
					cpt.RegisterTransport(transport)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Address resolution", func() {
		Context("before component start", func() {
			It("should return error for DialContext", func() {
				_, err := cpt.DialContext(context.Background(), "tcp", "example.com:80")
				Expect(err).To(HaveOccurred())
			})

			It("should return error for Clean", func() {
				_, _, err := cpt.Clean("http://example.com:80")
				Expect(err).To(HaveOccurred())
			})

			It("should return error for Search", func() {
				_, err := cpt.Search("example.com:80")
				Expect(err).To(HaveOccurred())
			})

			It("should return error for SearchWithCache", func() {
				_, err := cpt.SearchWithCache("example.com:80")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should clean endpoint", func() {
				host, port, err := cpt.Clean("example.com:8080")
				Expect(err).NotTo(HaveOccurred())
				Expect(host).To(Equal("example.com"))
				Expect(port).To(Equal("8080"))
			})

			It("should search for endpoint", func() {
				cpt.Add("example.com", "192.168.1.1")
				result, err := cpt.Search("example.com:80")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeEmpty())
			})

			It("should search with cache", func() {
				cpt.Add("example.com", "192.168.1.1")
				result, err := cpt.SearchWithCache("example.com:80")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeEmpty())
			})
		})
	})

	Describe("GetConfig method", func() {
		Context("before component start", func() {
			It("should return empty config", func() {
				cfg := cpt.GetConfig()
				Expect(cfg).NotTo(BeNil())
			})
		})

		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should return valid config", func() {
				cfg := cpt.GetConfig()
				Expect(cfg).NotTo(BeNil())
			})
		})
	})

	Describe("Close operation", func() {
		Context("before component start", func() {
			It("should not error when closing unstarted component", func() {
				err := cpt.Close()
				Expect(err).To(BeNil())
			})
		})

		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should close successfully", func() {
				err := cpt.Close()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("TimeCleaner", func() {
		Context("after component start", func() {
			BeforeEach(func() {
				startComponent()
			})

			It("should not panic when calling TimeCleaner", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				Expect(func() {
					// Call TimeCleaner with a very short duration for testing
					go cpt.TimeCleaner(ctx, 1)
					cancel() // Cancel immediately to avoid long-running goroutine
				}).NotTo(Panic())
			})
		})
	})
})
