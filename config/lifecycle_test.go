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

package config_test

import (
	"context"
	"errors"

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

var _ = Describe("Config Lifecycle", func() {
	var (
		cfg libcfg.Config
		cpt *testComponent
	)

	BeforeEach(func() {
		cfg = libcfg.New(nil)
		cpt = &testComponent{}
		cfg.ComponentSet("test", cpt)
	})

	Describe("Start", func() {
		Context("when components are registered", func() {
			It("should start all components successfully", func() {
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpt.started).To(BeTrue())
				Expect(cpt.running).To(BeTrue())
				Expect(cpt.startCount).To(Equal(1))
			})

			It("should mark config as started", func() {
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeTrue())
			})

			It("should mark config as running", func() {
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.(interface{ ComponentIsRunning(bool) bool }).ComponentIsRunning(true)).To(BeTrue())
			})
		})

		Context("when start fails", func() {
			It("should return error from component", func() {
				cpt.startError = errors.New("start failed")
				err := cfg.Start()
				Expect(err).To(HaveOccurred())
				// The error will be wrapped by the config system
				Expect(err.Error()).ToNot(BeEmpty())
			})
		})

		Context("with start hooks", func() {
			var beforeCalled, afterCalled bool

			BeforeEach(func() {
				beforeCalled = false
				afterCalled = false
				cfg.RegisterFuncStartBefore(func() error {
					beforeCalled = true
					return nil
				})
				cfg.RegisterFuncStartAfter(func() error {
					afterCalled = true
					return nil
				})
			})

			It("should call before and after hooks", func() {
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})

			It("should call before hook before starting components", func() {
				var beforeComponentStarted bool
				cfg.RegisterFuncStartBefore(func() error {
					beforeComponentStarted = cpt.started
					return nil
				})
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(beforeComponentStarted).To(BeFalse())
			})

			It("should call after hook after starting components", func() {
				var afterComponentStarted bool
				cfg.RegisterFuncStartAfter(func() error {
					afterComponentStarted = cpt.started
					return nil
				})
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(afterComponentStarted).To(BeTrue())
			})

			It("should fail if before hook returns error", func() {
				cfg.RegisterFuncStartBefore(func() error {
					return errors.New("before hook error")
				})
				err := cfg.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("before hook error"))
			})

			It("should fail if after hook returns error", func() {
				cfg.RegisterFuncStartAfter(func() error {
					return errors.New("after hook error")
				})
				err := cfg.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("after hook error"))
			})
		})
	})

	Describe("Reload", func() {
		Context("when components are started", func() {
			BeforeEach(func() {
				err := cfg.Start()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should reload all components successfully", func() {
				err := cfg.Reload()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpt.reloadCount).To(Equal(1))
				Expect(cpt.started).To(BeTrue())
			})

			It("should increment reload counter", func() {
				err := cfg.Reload()
				Expect(err).ToNot(HaveOccurred())
				err = cfg.Reload()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpt.reloadCount).To(Equal(2))
			})
		})

		Context("when reload fails", func() {
			It("should return error from component", func() {
				cpt.reloadError = errors.New("reload failed")
				err := cfg.Reload()
				Expect(err).To(HaveOccurred())
				// The error will be wrapped by the config system
				Expect(err.Error()).ToNot(BeEmpty())
			})
		})

		Context("with reload hooks", func() {
			var beforeCalled, afterCalled bool

			BeforeEach(func() {
				beforeCalled = false
				afterCalled = false
				cfg.RegisterFuncReloadBefore(func() error {
					beforeCalled = true
					return nil
				})
				cfg.RegisterFuncReloadAfter(func() error {
					afterCalled = true
					return nil
				})
			})

			It("should call before and after hooks", func() {
				err := cfg.Reload()
				Expect(err).ToNot(HaveOccurred())
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})

			It("should fail if before hook returns error", func() {
				cfg.RegisterFuncReloadBefore(func() error {
					return errors.New("before reload error")
				})
				err := cfg.Reload()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("before reload error"))
			})

			It("should fail if after hook returns error", func() {
				cfg.RegisterFuncReloadAfter(func() error {
					return errors.New("after reload error")
				})
				err := cfg.Reload()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("after reload error"))
			})
		})
	})

	Describe("Stop", func() {
		BeforeEach(func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop all components", func() {
			cfg.Stop()
			Expect(cpt.started).To(BeFalse())
			Expect(cpt.running).To(BeFalse())
			Expect(cpt.stopCount).To(Equal(1))
		})

		It("should mark config as stopped", func() {
			cfg.Stop()
			Expect(cfg.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeFalse())
		})

		Context("with stop hooks", func() {
			var beforeCalled, afterCalled bool

			BeforeEach(func() {
				beforeCalled = false
				afterCalled = false
				cfg.RegisterFuncStopBefore(func() error {
					beforeCalled = true
					return nil
				})
				cfg.RegisterFuncStopAfter(func() error {
					afterCalled = true
					return nil
				})
			})

			It("should call before and after hooks", func() {
				cfg.Stop()
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})

			It("should call hooks even if they return errors", func() {
				cfg.RegisterFuncStopBefore(func() error {
					return errors.New("stop before error")
				})
				cfg.RegisterFuncStopAfter(func() error {
					return errors.New("stop after error")
				})
				// Should not panic
				cfg.Stop()
				Expect(cpt.stopped).To(BeTrue())
			})
		})
	})
})

// testComponent is a minimal test implementation of cfgtps.Component
type testComponent struct {
	key           string
	deps          []string
	started       bool
	running       bool
	stopped       bool
	startCount    int
	reloadCount   int
	stopCount     int
	startError    error
	reloadError   error
	onStartBefore cfgtps.FuncCptEvent
	onStartAfter  cfgtps.FuncCptEvent
	onRelBefore   cfgtps.FuncCptEvent
	onRelAfter    cfgtps.FuncCptEvent
}

func (t *testComponent) Type() string { return "test" }

func (t *testComponent) Init(key string, _ context.Context, _ cfgtps.FuncCptGet, _ libvpr.FuncViper, _ libver.Version, _ liblog.FuncLog) {
	t.key = key
}

func (t *testComponent) DefaultConfig(indent string) []byte {
	return []byte("{}")
}

func (t *testComponent) Dependencies() []string {
	return append([]string{}, t.deps...)
}

func (t *testComponent) SetDependencies(d []string) error {
	t.deps = append([]string{}, d...)
	return nil
}

func (t *testComponent) RegisterFlag(_ *spfcbr.Command) error {
	return nil
}

func (t *testComponent) RegisterMonitorPool(_ montps.FuncPool) {}

func (t *testComponent) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {
	t.onStartBefore, t.onStartAfter = before, after
}

func (t *testComponent) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {
	t.onRelBefore, t.onRelAfter = before, after
}

func (t *testComponent) IsStarted() bool {
	return t.started
}

func (t *testComponent) IsRunning() bool {
	return t.running
}

func (t *testComponent) Start() error {
	if t.onStartBefore != nil {
		_ = t.onStartBefore(t)
	}
	if t.startError != nil {
		return t.startError
	}
	t.started = true
	t.running = true
	t.startCount++
	if t.onStartAfter != nil {
		_ = t.onStartAfter(t)
	}
	return nil
}

func (t *testComponent) Reload() error {
	if t.onRelBefore != nil {
		_ = t.onRelBefore(t)
	}
	if t.reloadError != nil {
		return t.reloadError
	}
	t.started = true
	t.reloadCount++
	if t.onRelAfter != nil {
		_ = t.onRelAfter(t)
	}
	return nil
}

func (t *testComponent) Stop() {
	t.running = false
	t.started = false
	t.stopped = true
	t.stopCount++
}
