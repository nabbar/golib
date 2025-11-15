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

	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tlscas "github.com/nabbar/golib/certificates/ca"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

// Lifecycle tests verify component initialization, Start/Reload/Stop operations,
// and callback execution.
var _ = Describe("Component Lifecycle", func() {
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
	})

	Describe("Component initialization", func() {
		Context("creating new component", func() {
			It("should create component with all parameters", func() {
				rootCA := func() tlscas.Cert { return nil }
				msg := func(s string) {}

				cpt = New(ctx, rootCA, true, msg)

				Expect(cpt).NotTo(BeNil())
				Expect(cpt.Type()).To(Equal("tls"))
			})

			It("should handle nil root CA function", func() {
				cpt = New(ctx, nil, false, nil)

				Expect(cpt).NotTo(BeNil())
			})

			It("should handle nil message function", func() {
				cpt = New(ctx, nil, false, nil)

				Expect(cpt).NotTo(BeNil())
			})
		})

		Context("Init method", func() {
			It("should initialize without error", func() {
				cpt = New(ctx, nil, false, nil)
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }

				Expect(func() {
					cpt.Init(key, ctx, getCpt, vpr, vrs, log)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Lifecycle operations", func() {
		BeforeEach(func() {
			cpt = New(ctx, nil, false, nil)
		})

		Context("IsStarted and IsRunning", func() {
			It("should return false before Start", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})

		Context("Start operation", func() {
			It("should fail without proper initialization", func() {
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})

			It("should fail without viper", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})

			It("should start with valid configuration", func() {
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

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &sharedMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())
				Expect(cpt.IsStarted()).To(BeTrue())
			})
		})

		Context("Reload operation", func() {
			It("should reload configuration", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

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

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &sharedMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())

				err = cpt.Reload()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Stop operation", func() {
			It("should not panic on Stop", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Callback registration", func() {
		BeforeEach(func() {
			cpt = New(ctx, nil, false, nil)
		})

		Context("RegisterFuncStart", func() {
			It("should register start callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncStart(before, after)
				}).NotTo(Panic())
			})

			It("should accept nil callbacks", func() {
				Expect(func() {
					cpt.RegisterFuncStart(nil, nil)
				}).NotTo(Panic())
			})
		})

		Context("RegisterFuncReload", func() {
			It("should register reload callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncReload(before, after)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Dependencies", func() {
		BeforeEach(func() {
			cpt = New(ctx, nil, false, nil)
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }
			cpt.Init(key, ctx, getCpt, vpr, vrs, log)
		})

		Context("SetDependencies and Dependencies", func() {
			It("should return empty slice by default", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
				Expect(deps).To(BeEmpty())
			})

			It("should set and retrieve dependencies", func() {
				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal([]string{"dep1", "dep2"}))
			})

			It("should handle empty dependencies", func() {
				err := cpt.SetDependencies([]string{})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})
		})
	})

	Describe("Message function", func() {
		Context("SetFuncMessage", func() {
			It("should set message function", func() {
				cpt = New(ctx, nil, false, nil)
				msg := func(s string) {}

				cpt.SetFuncMessage(msg)

				// Message function is stored
				Expect(cpt).NotTo(BeNil())
			})

			It("should not panic with nil message function", func() {
				cpt = New(ctx, nil, false, nil)

				Expect(func() {
					cpt.SetFuncMessage(nil)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Default HTTP client", func() {
		BeforeEach(func() {
			cpt = New(ctx, nil, false, nil)
		})

		Context("SetAsDefaultHTTPClient", func() {
			It("should set default flag to true", func() {
				Expect(func() {
					cpt.SetAsDefaultHTTPClient(true)
				}).NotTo(Panic())
			})

			It("should set default flag to false", func() {
				Expect(func() {
					cpt.SetAsDefaultHTTPClient(false)
				}).NotTo(Panic())
			})
		})

		Context("SetDefault", func() {
			It("should not panic when called", func() {
				Expect(func() {
					cpt.SetDefault()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Config method", func() {
		BeforeEach(func() {
			cpt = New(ctx, nil, false, nil)
		})

		Context("Config retrieval", func() {
			It("should return empty config before start", func() {
				cfg := cpt.Config()
				Expect(cfg).NotTo(BeNil())
			})

			It("should return config after start", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

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

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &sharedMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())

				cfg := cpt.Config()
				Expect(cfg).NotTo(BeNil())
			})
		})
	})
})
