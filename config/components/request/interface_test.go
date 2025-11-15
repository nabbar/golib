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

package request_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/request"
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
		It("should create a valid Request component with nil client", func() {
			cpt := New(ctx, nil)
			Expect(cpt).NotTo(BeNil())
		})

		It("should create component with custom client", func() {
			cpt := New(ctx, nil)
			Expect(cpt).NotTo(BeNil())
		})

		It("should not be started initially", func() {
			cpt := New(ctx, nil)
			Expect(cpt.IsStarted()).To(BeFalse())
		})
	})

	Describe("Register function", func() {
		It("should register component in config", func() {
			cfg := libcfg.New(vrs)
			cpt := New(ctx, nil)

			Register(cfg, "test-request", cpt)

			loaded := Load(cfg.ComponentGet, "test-request")
			Expect(loaded).NotTo(BeNil())
			Expect(loaded).To(Equal(cpt))
		})
	})

	Describe("RegisterNew function", func() {
		It("should create and register component", func() {
			cfg := libcfg.New(vrs)

			RegisterNew(ctx, cfg, "test-request", nil)

			loaded := Load(cfg.ComponentGet, "test-request")
			Expect(loaded).NotTo(BeNil())
		})
	})

	Describe("Load function", func() {
		It("should load registered component", func() {
			cfg := libcfg.New(vrs)
			cpt := New(ctx, nil)
			Register(cfg, "test-request", cpt)

			loaded := Load(cfg.ComponentGet, "test-request")
			Expect(loaded).NotTo(BeNil())
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
	})

	Describe("Interface compliance", func() {
		It("should implement cfgtps.Component", func() {
			var _ cfgtps.Component = New(ctx, nil)
		})

		It("should implement ComponentRequest interface", func() {
			var _ ComponentRequest = New(ctx, nil)
		})

		It("should have all required methods", func() {
			cpt := New(ctx, nil)

			Expect(cpt.Type).NotTo(BeNil())
			Expect(cpt.Init).NotTo(BeNil())
			Expect(cpt.Start).NotTo(BeNil())
			Expect(cpt.Stop).NotTo(BeNil())
			Expect(cpt.Reload).NotTo(BeNil())
			Expect(cpt.IsStarted).NotTo(BeNil())
			Expect(cpt.IsRunning).NotTo(BeNil())
			Expect(cpt.SetHTTPClient).NotTo(BeNil())
			Expect(cpt.Request).NotTo(BeNil())
		})
	})
})

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
