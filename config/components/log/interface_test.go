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

package log_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
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
		cnl context.CancelFunc
		cpt CptLog
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, loglvl.NilLevel)
		cpt.Init(kd, ctx, nil, fv, vs, fl)

		v.Viper().SetConfigType("json")

		configData := map[string]interface{}{
			kd: map[string]interface{}{
				"stdout": map[string]interface{}{
					"disableStandard": true,
				},
			},
		}

		configJSON, err := json.Marshal(configData)
		Expect(err).To(BeNil())

		err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if cpt != nil {
			cpt.Stop()
		}
		cnl()
	})

	Describe("New function", func() {
		Context("creating new Log component", func() {
			It("should create a valid Log component", func() {
				cpt := New(ctx, DefaultLevel)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create component with specified level", func() {
				cpt := New(ctx, loglvl.DebugLevel)
				Expect(cpt).NotTo(BeNil())
				Expect(cpt.GetLevel()).To(Equal(loglvl.DebugLevel))
			})

			It("should create component with info level", func() {
				cpt := New(ctx, loglvl.InfoLevel)
				Expect(cpt).NotTo(BeNil())
				Expect(cpt.GetLevel()).To(Equal(loglvl.InfoLevel))
			})

			It("should not be started initially", func() {
				cpt := New(ctx, DefaultLevel)
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("Register function", func() {
		Context("registering component", func() {
			It("should register component in config", func() {
				cfg := libcfg.New(vs)
				cpt := New(ctx, DefaultLevel)

				Register(cfg, "test-log", cpt)

				loaded := Load(cfg.ComponentGet, "test-log")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})

			It("should handle multiple registrations with different keys", func() {
				cfg := libcfg.New(vs)
				cpt1 := New(ctx, loglvl.DebugLevel)
				cpt2 := New(ctx, loglvl.InfoLevel)

				Register(cfg, "log1", cpt1)
				Register(cfg, "log2", cpt2)

				loaded1 := Load(cfg.ComponentGet, "log1")
				loaded2 := Load(cfg.ComponentGet, "log2")

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
				cfg := libcfg.New(vs)

				RegisterNew(ctx, cfg, "test-log", DefaultLevel)

				loaded := Load(cfg.ComponentGet, "test-log")
				Expect(loaded).NotTo(BeNil())
			})

			It("should create component with specified level", func() {
				cfg := libcfg.New(vs)

				RegisterNew(ctx, cfg, "test-log", loglvl.ErrorLevel)

				loaded := Load(cfg.ComponentGet, "test-log")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.GetLevel()).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("Load function", func() {
		Context("loading component", func() {
			It("should load registered component", func() {
				cfg := libcfg.New(vs)
				cpt := New(ctx, DefaultLevel)
				Register(cfg, "test-log", cpt)

				loaded := Load(cfg.ComponentGet, "test-log")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})

			It("should return nil for non-existent key", func() {
				cfg := libcfg.New(vs)
				loaded := Load(cfg.ComponentGet, "non-existent")
				Expect(loaded).To(BeNil())
			})

			It("should return nil for wrong component type", func() {
				cfg := libcfg.New(vs)
				cfg.ComponentSet("wrong", &wrongComponent{})
				loaded := Load(cfg.ComponentGet, "wrong")
				Expect(loaded).To(BeNil())
			})
		})
	})

	Describe("Type identification", func() {
		Context("component type", func() {
			It("should return correct component type", func() {
				cpt := New(ctx, DefaultLevel)
				Expect(cpt.Type()).To(Equal("log"))
			})
		})
	})

	Describe("DefaultLevel constant", func() {
		Context("default level value", func() {
			It("should be InfoLevel", func() {
				Expect(DefaultLevel).To(Equal(loglvl.InfoLevel))
			})
		})
	})

	Describe("Interface compliance", func() {
		Context("CptLog interface", func() {
			It("should implement cfgtps.Component", func() {
				var _ cfgtps.Component = New(ctx, DefaultLevel)
			})

			It("should implement CptLog interface", func() {
				var _ CptLog = New(ctx, DefaultLevel)
			})

			It("should have all required methods", func() {
				cpt := New(ctx, DefaultLevel)

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

				// CptLog methods
				Expect(cpt.Log).NotTo(BeNil())
				Expect(cpt.SetLevel).NotTo(BeNil())
				Expect(cpt.GetLevel).NotTo(BeNil())
				Expect(cpt.SetField).NotTo(BeNil())
				Expect(cpt.GetField).NotTo(BeNil())
				Expect(cpt.SetOptions).NotTo(BeNil())
				Expect(cpt.GetOptions).NotTo(BeNil())
			})
		})
	})
})

// wrongComponent is a mock component type used for testing type safety
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
