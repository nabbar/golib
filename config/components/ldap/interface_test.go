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

package ldap_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/ldap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

// Interface tests verify the public interface functions, component registration,
// and loading mechanisms for the LDAP component.
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
		Context("creating new LDAP component", func() {
			It("should create a new component with valid context", func() {
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should initialize with empty attributes", func() {
				cpt := New(ctx)
				attrs := cpt.GetAttributes()
				Expect(attrs).NotTo(BeNil())
				Expect(attrs).To(BeEmpty())
			})

			It("should initialize with empty config", func() {
				cpt := New(ctx)
				cfg := cpt.GetConfig()
				Expect(cfg).To(BeNil()) // Empty config returns nil
			})

			It("should return correct component type", func() {
				cpt := New(ctx)
				Expect(cpt.Type()).To(Equal("LDAP"))
			})
		})
	})

	Describe("Register function", func() {
		Context("registering component", func() {
			It("should register component in config", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx)
				key := "test-ldap"

				Register(cfg, key, cpt)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})

			It("should allow multiple registrations with different keys", func() {
				cfg := libcfg.New(vrs)
				cpt1 := New(ctx)
				cpt2 := New(ctx)

				Register(cfg, "ldap1", cpt1)
				Register(cfg, "ldap2", cpt2)

				loaded1 := Load(cfg.ComponentGet, "ldap1")
				loaded2 := Load(cfg.ComponentGet, "ldap2")

				Expect(loaded1).To(Equal(cpt1))
				Expect(loaded2).To(Equal(cpt2))
			})
		})
	})

	Describe("RegisterNew function", func() {
		Context("registering new component", func() {
			It("should create and register new component", func() {
				cfg := libcfg.New(vrs)
				key := "test-ldap"

				RegisterNew(ctx, cfg, key)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("LDAP"))
			})
		})
	})

	Describe("Load function", func() {
		Context("loading component", func() {
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
				cfg.ComponentSet("wrong", &wrongComponent{})
				loaded := Load(cfg.ComponentGet, "wrong")
				Expect(loaded).To(BeNil())
			})

			It("should load registered component", func() {
				cfg := libcfg.New(vrs)
				cpt := New(ctx)
				key := "test-ldap"

				Register(cfg, key, cpt)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("LDAP"))
			})
		})
	})

	Describe("Type identification", func() {
		Context("component type", func() {
			It("should return correct component type", func() {
				cpt := New(ctx)
				Expect(cpt.Type()).To(Equal("LDAP"))
			})
		})
	})

	Describe("Interface compliance", func() {
		Context("CptLDAP interface", func() {
			It("should implement cfgtps.Component", func() {
				var _ cfgtps.Component = New(ctx)
			})

			It("should implement CptLDAP interface", func() {
				var _ CptLDAP = New(ctx)
			})

			It("should have all required methods", func() {
				cpt := New(ctx)

				// Component methods
				Expect(cpt.Type).NotTo(BeNil())
				Expect(cpt.Init).NotTo(BeNil())
				Expect(cpt.Start).NotTo(BeNil())
				Expect(cpt.Reload).NotTo(BeNil())
				Expect(cpt.Stop).NotTo(BeNil())
				Expect(cpt.IsStarted).NotTo(BeNil())
				Expect(cpt.IsRunning).NotTo(BeNil())
				Expect(cpt.Dependencies).NotTo(BeNil())
				Expect(cpt.SetDependencies).NotTo(BeNil())

				// LDAP specific methods
				Expect(cpt.GetConfig).NotTo(BeNil())
				Expect(cpt.SetConfig).NotTo(BeNil())
				Expect(cpt.GetLDAP).NotTo(BeNil())
				Expect(cpt.SetLDAP).NotTo(BeNil())
				Expect(cpt.SetAttributes).NotTo(BeNil())
			})
		})
	})
})

// wrongComponent for testing type safety
type wrongComponent struct{}

func (w *wrongComponent) Type() string { return "wrong" }
func (w *wrongComponent) Init(string, context.Context, cfgtps.FuncCptGet, libvpr.FuncViper, libver.Version, liblog.FuncLog) {
}
func (w *wrongComponent) RegisterFuncStart(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent)  {}
func (w *wrongComponent) RegisterFuncReload(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {}
func (w *wrongComponent) IsStarted() bool                                             { return false }
func (w *wrongComponent) IsRunning() bool                                             { return false }
func (w *wrongComponent) Start() error                                                { return nil }
func (w *wrongComponent) Reload() error                                               { return nil }
func (w *wrongComponent) Stop()                                                       {}
func (w *wrongComponent) Dependencies() []string                                      { return nil }
func (w *wrongComponent) SetDependencies([]string) error                              { return nil }
func (w *wrongComponent) DefaultConfig(string) []byte                                 { return nil }
func (w *wrongComponent) RegisterFlag(*spfcbr.Command) error                          { return nil }
func (w *wrongComponent) RegisterMonitorPool(montps.FuncPool)                         {}
