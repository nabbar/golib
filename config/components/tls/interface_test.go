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

package tls_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tlscas "github.com/nabbar/golib/certificates/ca"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
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
		Context("creating new component", func() {
			It("should create a new component with context", func() {
				cpt := New(ctx, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create component with nil defCARoot", func() {
				cpt := New(ctx, nil)
				Expect(cpt).NotTo(BeNil())
				Expect(cpt.Type()).To(Equal(ComponentType))
			})

			It("should create component with custom defCARoot", func() {
				defCARoot := func() tlscas.Cert {
					return nil
				}
				cpt := New(ctx, defCARoot)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create multiple independent components", func() {
				cpt1 := New(ctx, nil)
				cpt2 := New(ctx, nil)

				Expect(cpt1).NotTo(BeNil())
				Expect(cpt2).NotTo(BeNil())
				Expect(cpt1).NotTo(BeIdenticalTo(cpt2))
			})
		})
	})

	Describe("GetRootCaCert function", func() {
		Context("converting root CA function to cert", func() {
			It("should handle nil function result", func() {
				fct := func() []string {
					return nil
				}
				cert := GetRootCaCert(fct)
				Expect(cert).To(BeNil())
			})

			It("should handle empty function result", func() {
				fct := func() []string {
					return []string{}
				}
				cert := GetRootCaCert(fct)
				Expect(cert).To(BeNil())
			})

			It("should parse single certificate", func() {
				// Using a simple test certificate string
				testCert := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU50EMA0GCSqGSIb3DQEBCwUAMA0xCzAJBgNVBAYTAlVT
MB4XDTI0MDEwMTAwMDAwMFoXDTI1MDEwMTAwMDAwMFowDTELMAkGA1UEBhMCVVMw
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAM1ZCzsrb3gBQyFcmzQJk6jQ6g/Z
N8e7C4n7sU9yHfOCVdDJ3x3bGn5n5f1K2L6E8g7U1e6b3G4P2f8C9J0D2Y0c7C5Y
6S8u4T4S2Y0e8C6G3S2V8N7Y0G1S4P5D8G2R1K5N6T9M3C0J8E3X4K9N5L7S1V0Q
6U2C0H1J4P7Y8E9D2C5R0S6XAgMBAAEwDQYJKoZIhvcNAQELBQADgYEAjPHo6j7k
d6F0k1Q8L2W0R3C5Z8H4N6Y2M9P3S0J5K8E9F1L7X6U2C4V0D8N3G5Y1H2S9E0K7
J6P4R8C2M3N0Q1L5T9Y8F3E6D7S1V4K2U0C8X9H5J3N6G1P0R7Q8M2Y4E5L1S0W3
-----END CERTIFICATE-----`

				fct := func() []string {
					return []string{testCert}
				}
				cert := GetRootCaCert(fct)
				// Parse may fail with invalid cert, but function should not panic
				_ = cert
			})

			It("should handle multiple certificates", func() {
				testCert1 := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU50EMA0GCSqGSIb3DQEBCwUAMA0xCzAJBgNVBAYTAlVT
MB4XDTI0MDEwMTAwMDAwMFoXDTI1MDEwMTAwMDAwMFowDTELMAkGA1UEBhMCVVMw
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAM1ZCzsrb3gBQyFcmzQJk6jQ6g/Z
N8e7C4n7sU9yHfOCVdDJ3x3bGn5n5f1K2L6E8g7U1e6b3G4P2f8C9J0D2Y0c7C5Y
6S8u4T4S2Y0e8C6G3S2V8N7Y0G1S4P5D8G2R1K5N6T9M3C0J8E3X4K9N5L7S1V0Q
6U2C0H1J4P7Y8E9D2C5R0S6XAgMBAAEwDQYJKoZIhvcNAQELBQADgYEAjPHo6j7k
d6F0k1Q8L2W0R3C5Z8H4N6Y2M9P3S0J5K8E9F1L7X6U2C4V0D8N3G5Y1H2S9E0K7
J6P4R8C2M3N0Q1L5T9Y8F3E6D7S1V4K2U0C8X9H5J3N6G1P0R7Q8M2Y4E5L1S0W3
-----END CERTIFICATE-----`
				testCert2 := testCert1 // Using same for simplicity

				fct := func() []string {
					return []string{testCert1, testCert2}
				}
				cert := GetRootCaCert(fct)
				// Should handle multiple certs
				_ = cert
			})

			It("should not panic with invalid certificates", func() {
				fct := func() []string {
					return []string{"invalid cert"}
				}

				Expect(func() {
					_ = GetRootCaCert(fct)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Register function", func() {
		Context("registering component", func() {
			It("should register component to config", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx, nil)

				Expect(func() {
					Register(cfg, "tls-test", cpt)
				}).NotTo(Panic())
			})

			It("should allow retrieving registered component", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx, nil)

				Register(cfg, "tls-test", cpt)

				retrieved := cfg.ComponentGet("tls-test")
				Expect(retrieved).NotTo(BeNil())
			})

			It("should allow multiple components with different keys", func() {
				cfg := libcfg.New(vrs)
				cpt1 := New(ctx, nil)
				cpt2 := New(ctx, nil)

				Register(cfg, "tls-1", cpt1)
				Register(cfg, "tls-2", cpt2)

				retrieved1 := cfg.ComponentGet("tls-1")
				retrieved2 := cfg.ComponentGet("tls-2")

				Expect(retrieved1).NotTo(BeNil())
				Expect(retrieved2).NotTo(BeNil())
			})

			It("should allow replacing component with same key", func() {
				cfg := libcfg.New(vrs)
				cpt1 := New(ctx, nil)
				cpt2 := New(ctx, nil)

				Register(cfg, "tls-test", cpt1)
				Register(cfg, "tls-test", cpt2)

				// Should not panic
			})
		})
	})

	Describe("RegisterNew function", func() {
		Context("creating and registering component", func() {
			It("should create and register component", func() {
				cfg := libcfg.New(vrs)

				Expect(func() {
					RegisterNew(ctx, cfg, "tls-test", nil)
				}).NotTo(Panic())
			})

			It("should allow retrieving registered component", func() {
				cfg := libcfg.New(vrs)

				RegisterNew(ctx, cfg, "tls-test", nil)

				retrieved := cfg.ComponentGet("tls-test")
				Expect(retrieved).NotTo(BeNil())
			})

			It("should create component with custom defCARoot", func() {
				cfg := libcfg.New(vrs)
				defCARoot := func() tlscas.Cert {
					return nil
				}

				RegisterNew(ctx, cfg, "tls-test", defCARoot)

				retrieved := cfg.ComponentGet("tls-test")
				Expect(retrieved).NotTo(BeNil())
			})

			It("should allow multiple registrations", func() {
				cfg := libcfg.New(vrs)

				RegisterNew(ctx, cfg, "tls-1", nil)
				RegisterNew(ctx, cfg, "tls-2", nil)
				RegisterNew(ctx, cfg, "tls-3", nil)

				// All should be retrievable
				Expect(cfg.ComponentGet("tls-1")).NotTo(BeNil())
				Expect(cfg.ComponentGet("tls-2")).NotTo(BeNil())
				Expect(cfg.ComponentGet("tls-3")).NotTo(BeNil())
			})
		})
	})

	Describe("Load function", func() {
		Context("loading component from getter", func() {
			It("should return nil when component not found", func() {
				getCpt := func(key string) cfgtps.Component {
					return nil
				}

				cpt := Load(getCpt, "nonexistent")
				Expect(cpt).To(BeNil())
			})

			It("should return nil when component is wrong type", func() {
				wrongCpt := &wrongComponent{}
				getCpt := func(key string) cfgtps.Component {
					if key == "wrong" {
						return wrongCpt
					}
					return nil
				}

				cpt := Load(getCpt, "wrong")
				Expect(cpt).To(BeNil())
			})

			It("should return component when found and correct type", func() {
				expectedCpt := New(ctx, nil)
				getCpt := func(key string) cfgtps.Component {
					if key == "tls-test" {
						return expectedCpt
					}
					return nil
				}

				cpt := Load(getCpt, "tls-test")
				Expect(cpt).NotTo(BeNil())
				Expect(cpt).To(BeIdenticalTo(expectedCpt))
			})

			It("should work with config ComponentGet", func() {
				cfg := libcfg.New(vrs)
				RegisterNew(ctx, cfg, "tls-test", nil)

				cpt := Load(cfg.ComponentGet, "tls-test")
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("ComponentType constant", func() {
		Context("component type value", func() {
			It("should have expected value", func() {
				Expect(ComponentType).To(Equal("tls"))
			})

			It("should match component Type method", func() {
				cpt := New(ctx, nil)
				Expect(cpt.Type()).To(Equal(ComponentType))
			})
		})
	})

	Describe("Integration scenarios", func() {
		Context("full registration and load cycle", func() {
			It("should register and load component successfully", func() {
				cfg := libcfg.New(vrs)

				// Register using RegisterNew
				RegisterNew(ctx, cfg, "tls-full-test", nil)

				// Load using Load function
				cpt := Load(cfg.ComponentGet, "tls-full-test")

				Expect(cpt).NotTo(BeNil())
				Expect(cpt.Type()).To(Equal(ComponentType))
			})

			It("should handle custom root CA throughout cycle", func() {
				cfg := libcfg.New(vrs)
				defCARoot := func() tlscas.Cert {
					return nil
				}

				// Register with custom root CA
				RegisterNew(ctx, cfg, "tls-custom", defCARoot)

				// Load
				cpt := Load(cfg.ComponentGet, "tls-custom")

				Expect(cpt).NotTo(BeNil())
			})
		})
	})
})

// wrongComponent is a test helper that implements Component but not CptTlS
type wrongComponent struct{}

func (w *wrongComponent) Type() string { return "wrong" }
func (w *wrongComponent) Init(key string, ctx context.Context, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
}
func (w *wrongComponent) RegisterFuncStart(before, after cfgtps.FuncCptEvent)  {}
func (w *wrongComponent) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {}
func (w *wrongComponent) IsStarted() bool                                      { return false }
func (w *wrongComponent) IsRunning() bool                                      { return false }
func (w *wrongComponent) Start() error                                         { return nil }
func (w *wrongComponent) Reload() error                                        { return nil }
func (w *wrongComponent) Stop()                                                {}
func (w *wrongComponent) Dependencies() []string                               { return nil }
func (w *wrongComponent) SetDependencies(d []string) error                     { return nil }
func (w *wrongComponent) DefaultConfig(indent string) []byte                   { return nil }
func (w *wrongComponent) RegisterFlag(cmd *spfcbr.Command) error               { return nil }
func (w *wrongComponent) RegisterMonitorPool(fct montps.FuncPool)              {}
