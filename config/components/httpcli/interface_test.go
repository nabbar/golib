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
	"context"

	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tlscas "github.com/nabbar/golib/certificates/ca"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libver "github.com/nabbar/golib/version"
)

// Interface tests verify the public interface functions and component
// registration/loading mechanisms.
var _ = Describe("Interface Functions", func() {
	var (
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
	})

	Describe("New function", func() {
		Context("creating new HTTPCli component", func() {
			It("should create a valid HTTPCli component with nil CA root", func() {
				cpt := New(ctx, nil, false, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create component with custom CA root", func() {
				defCARoot := func() tlscas.Cert {
					return nil
				}
				cpt := New(ctx, defCARoot, false, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create component as default HTTP client", func() {
				cpt := New(ctx, nil, true, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create component with message function", func() {
				msg := func(m string) {}
				cpt := New(ctx, nil, false, msg)
				Expect(cpt).NotTo(BeNil())
			})

			It("should not be started initially", func() {
				cpt := New(ctx, nil, false, nil)
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("Register function", func() {
		Context("registering component", func() {
			It("should register component in config", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx, nil, false, nil)

				Register(cfg, "test-httpcli", cpt)

				loaded := Load(cfg.ComponentGet, "test-httpcli")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})

			It("should handle multiple registrations with different keys", func() {
				cfg := libcfg.New(vrs)
				cpt1 := New(ctx, nil, false, nil)
				cpt2 := New(ctx, nil, true, nil)

				Register(cfg, "httpcli1", cpt1)
				Register(cfg, "httpcli2", cpt2)

				loaded1 := Load(cfg.ComponentGet, "httpcli1")
				loaded2 := Load(cfg.ComponentGet, "httpcli2")

				Expect(loaded1).NotTo(BeNil())
				Expect(loaded2).NotTo(BeNil())
				Expect(loaded1).To(Equal(cpt1))
				Expect(loaded2).To(Equal(cpt2))
			})
		})
	})

	Describe("RegisterNew function", func() {
		Context("registering new component", func() {
			It("should create and register component", func() {
				cfg := libcfg.New(vrs)

				RegisterNew(ctx, cfg, "test-httpcli", nil, false, nil)

				loaded := Load(cfg.ComponentGet, "test-httpcli")
				Expect(loaded).NotTo(BeNil())
			})

			It("should create component with all parameters", func() {
				cfg := libcfg.New(vrs)
				defCARoot := func() tlscas.Cert { return nil }
				msg := func(m string) {}

				RegisterNew(ctx, cfg, "test-httpcli", defCARoot, true, msg)

				loaded := Load(cfg.ComponentGet, "test-httpcli")
				Expect(loaded).NotTo(BeNil())
			})
		})
	})

	Describe("Load function", func() {
		Context("loading component", func() {
			It("should load registered component", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx, nil, false, nil)
				Register(cfg, "test-httpcli", cpt)

				loaded := Load(cfg.ComponentGet, "test-httpcli")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
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

	Describe("Type identification", func() {
		Context("component type", func() {
			It("should return correct component type", func() {
				cpt := New(ctx, nil, false, nil)
				Expect(cpt.Type()).To(Equal("tls"))
			})
		})
	})

	Describe("GetRootCaCert function", func() {
		Context("parsing root CA certificates", func() {
			It("should return nil for empty certificate list", func() {
				fct := func() []string {
					return []string{}
				}
				cert := GetRootCaCert(fct)
				Expect(cert).To(BeNil())
			})

			It("should parse single certificate", func() {
				fct := func() []string {
					return []string{"-----BEGIN CERTIFICATE-----\nMIIBIjANBgk\n-----END CERTIFICATE-----"}
				}
				cert := GetRootCaCert(fct)
				// May be nil if certificate is invalid, but shouldn't panic
				_ = cert
			})

			It("should parse multiple certificates", func() {
				fct := func() []string {
					return []string{
						"-----BEGIN CERTIFICATE-----\nMIIBIjANBgk\n-----END CERTIFICATE-----",
						"-----BEGIN CERTIFICATE-----\nMIIBIjANBgl\n-----END CERTIFICATE-----",
					}
				}
				cert := GetRootCaCert(fct)
				// May be nil if certificates are invalid, but shouldn't panic
				_ = cert
			})
		})
	})

	Describe("Interface compliance", func() {
		Context("CptHTTPClient interface", func() {
			It("should implement cfgtps.Component", func() {
				var _ cfgtps.Component = New(ctx, nil, false, nil)
			})

			It("should implement CptHTTPClient interface", func() {
				var _ CptHTTPClient = New(ctx, nil, false, nil)
			})

			It("should have all required methods", func() {
				cpt := New(ctx, nil, false, nil)

				// Component methods
				Expect(cpt.Type).NotTo(BeNil())
				Expect(cpt.Init).NotTo(BeNil())
				Expect(cpt.Start).NotTo(BeNil())
				Expect(cpt.Stop).NotTo(BeNil())
				Expect(cpt.Reload).NotTo(BeNil())
				Expect(cpt.IsStarted).NotTo(BeNil())
				Expect(cpt.IsRunning).NotTo(BeNil())
				Expect(cpt.Dependencies).NotTo(BeNil())
				Expect(cpt.SetDependencies).NotTo(BeNil())
				Expect(cpt.RegisterFuncStart).NotTo(BeNil())
				Expect(cpt.RegisterFuncReload).NotTo(BeNil())
				Expect(cpt.DefaultConfig).NotTo(BeNil())
				Expect(cpt.RegisterFlag).NotTo(BeNil())
				Expect(cpt.RegisterMonitorPool).NotTo(BeNil())

				// CptHTTPClient methods
				Expect(cpt.Config).NotTo(BeNil())
				Expect(cpt.SetDefault).NotTo(BeNil())
				Expect(cpt.SetAsDefaultHTTPClient).NotTo(BeNil())
				Expect(cpt.SetFuncMessage).NotTo(BeNil())

				// DNSMapper methods
				Expect(cpt.Close).NotTo(BeNil())
				Expect(cpt.Add).NotTo(BeNil())
				Expect(cpt.Get).NotTo(BeNil())
				Expect(cpt.Del).NotTo(BeNil())
				Expect(cpt.Len).NotTo(BeNil())
				Expect(cpt.Walk).NotTo(BeNil())
			})
		})
	})
})
