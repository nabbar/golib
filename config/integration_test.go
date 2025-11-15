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
	"encoding/json"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libver "github.com/nabbar/golib/version"
)

var _ = Describe("Config Integration", func() {
	var cfg libcfg.Config

	BeforeEach(func() {
		version := libver.NewVersion(
			libver.License_MIT,
			"testapp",
			"Test Application",
			"2024-01-01",
			"abc123",
			"v1.0.0",
			"Test Author",
			"testapp",
			struct{}{},
			0,
		)
		cfg = libcfg.New(version)
	})

	Describe("Full lifecycle with multiple components", func() {
		var (
			db    *mockDatabase
			cache *mockCache
			api   *mockAPI
		)

		BeforeEach(func() {
			db = &mockDatabase{}
			cache = &mockCache{}
			cache.deps = []string{"database"}
			api = &mockAPI{}
			api.deps = []string{"database", "cache"}

			cfg.ComponentSet("database", db)
			cfg.ComponentSet("cache", cache)
			cfg.ComponentSet("api", api)
		})

		It("should start all components in correct order", func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			// All should be started
			Expect(db.started).To(BeTrue())
			Expect(cache.started).To(BeTrue())
			Expect(api.started).To(BeTrue())

			// Verify order: db -> cache -> api
			Expect(db.startOrder).To(BeNumerically("<", cache.startOrder))
			Expect(cache.startOrder).To(BeNumerically("<", api.startOrder))
		})

		It("should reload all components", func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			err = cfg.Reload()
			Expect(err).ToNot(HaveOccurred())

			Expect(db.reloadCount).To(Equal(1))
			Expect(cache.reloadCount).To(Equal(1))
			Expect(api.reloadCount).To(Equal(1))
		})

		It("should stop all components in reverse order", func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			cfg.Stop()

			// All should be stopped
			Expect(db.started).To(BeFalse())
			Expect(cache.started).To(BeFalse())
			Expect(api.started).To(BeFalse())

			// Verify reverse order: api -> cache -> db
			Expect(api.stopOrder).To(BeNumerically("<", cache.stopOrder))
			Expect(cache.stopOrder).To(BeNumerically("<", db.stopOrder))
		})

		It("should handle component failure gracefully", func() {
			cache.startError = errors.New("cache connection failed")

			err := cfg.Start()
			Expect(err).To(HaveOccurred())
			// The error will be wrapped by the config system
			Expect(err.Error()).ToNot(BeEmpty())
		})

		It("should execute hooks in correct order", func() {
			var order []string

			cfg.RegisterFuncStartBefore(func() error {
				order = append(order, "config-start-before")
				return nil
			})

			cfg.RegisterFuncStartAfter(func() error {
				order = append(order, "config-start-after")
				return nil
			})

			db.onStartBefore = func(cpt cfgtps.Component) error {
				order = append(order, "db-start-before")
				return nil
			}

			db.onStartAfter = func(cpt cfgtps.Component) error {
				order = append(order, "db-start-after")
				return nil
			}

			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			// Verify hook execution order
			Expect(order).To(Equal([]string{
				"config-start-before",
				"db-start-before",
				"db-start-after",
				"config-start-after",
			}))
		})
	})

	Describe("Component dependency management", func() {
		It("should handle components without circular dependencies", func() {
			comp1 := &simpleComponent{name: "comp1"}
			comp2 := &simpleComponent{name: "comp2", deps: []string{"comp1"}}

			cfg.ComponentSet("comp1", comp1)
			cfg.ComponentSet("comp2", comp2)

			// Should work correctly
			Expect(func() {
				deps := cfg.(interface{ ComponentDependencies() []string }).ComponentDependencies()
				Expect(deps).ToNot(BeNil())
				Expect(len(deps)).To(Equal(2))
			}).ToNot(Panic())
		})

		It("should handle missing dependencies", func() {
			comp1 := &simpleComponent{name: "comp1", deps: []string{"missing"}}
			cfg.ComponentSet("comp1", comp1)

			// Should still work, just skip missing dependency
			Expect(func() {
				deps := cfg.(interface{ ComponentDependencies() []string }).ComponentDependencies()
				Expect(deps).To(ContainElement("comp1"))
			}).ToNot(Panic())
		})

		It("should handle deep dependency chains", func() {
			comp1 := &simpleComponent{name: "comp1"}
			comp2 := &simpleComponent{name: "comp2", deps: []string{"comp1"}}
			comp3 := &simpleComponent{name: "comp3", deps: []string{"comp2"}}
			comp4 := &simpleComponent{name: "comp4", deps: []string{"comp3"}}

			cfg.ComponentSet("comp4", comp4)
			cfg.ComponentSet("comp3", comp3)
			cfg.ComponentSet("comp2", comp2)
			cfg.ComponentSet("comp1", comp1)

			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			// All should be started
			Expect(comp1.started).To(BeTrue())
			Expect(comp2.started).To(BeTrue())
			Expect(comp3.started).To(BeTrue())
			Expect(comp4.started).To(BeTrue())
		})
	})

	Describe("Configuration generation", func() {
		It("should generate valid JSON configuration", func() {
			db := &mockDatabase{}
			cache := &mockCache{}

			cfg.ComponentSet("database", db)
			cfg.ComponentSet("cache", cache)

			reader := cfg.(interface{ DefaultConfig() io.Reader }).DefaultConfig()
			Expect(reader).ToNot(BeNil())

			var config map[string]interface{}
			decoder := json.NewDecoder(reader)
			err := decoder.Decode(&config)
			Expect(err).ToNot(HaveOccurred())

			Expect(config).To(HaveKey("database"))
			Expect(config).To(HaveKey("cache"))
		})

		It("should include component-specific configuration", func() {
			db := &mockDatabase{}
			cfg.ComponentSet("database", db)

			reader := cfg.(interface{ DefaultConfig() io.Reader }).DefaultConfig()
			var config map[string]interface{}
			decoder := json.NewDecoder(reader)
			err := decoder.Decode(&config)
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(HaveKey("database"))
		})
	})

	Describe("Error handling", func() {
		It("should aggregate errors from multiple components", func() {
			comp1 := &testComponent{startError: errors.New("comp1 error")}
			comp2 := &testComponent{startError: errors.New("comp2 error")}

			cfg.ComponentSet("comp1", comp1)
			cfg.ComponentSet("comp2", comp2)

			err := cfg.Start()
			Expect(err).To(HaveOccurred())
			// The error will be wrapped by the config system
			Expect(err.Error()).ToNot(BeEmpty())
		})

		It("should stop on first hook error", func() {
			var executed bool
			cfg.RegisterFuncStartBefore(func() error {
				return errors.New("hook error")
			})
			cfg.RegisterFuncStartAfter(func() error {
				executed = true
				return nil
			})

			comp := &simpleComponent{}
			cfg.ComponentSet("comp", comp)

			err := cfg.Start()
			Expect(err).To(HaveOccurred())
			Expect(executed).To(BeFalse())
			Expect(comp.started).To(BeFalse())
		})
	})

	Describe("State management", func() {
		It("should track component started state", func() {
			comp := &simpleComponent{}
			cfg.ComponentSet("comp", comp)

			Expect(cfg.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeFalse())

			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeTrue())

			cfg.Stop()
			Expect(cfg.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeFalse())
		})

		It("should track component running state", func() {
			comp := &simpleComponent{}
			cfg.ComponentSet("comp", comp)

			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.(interface{ ComponentIsRunning(bool) bool }).ComponentIsRunning(true)).To(BeTrue())

			cfg.Stop()
			Expect(cfg.(interface{ ComponentIsRunning(bool) bool }).ComponentIsRunning(true)).To(BeFalse())
		})
	})
})

// Mock components for integration testing
var startOrderCounter = 0
var stopOrderCounter = 0

type mockDatabase struct {
	simpleComponent
	startOrder    int
	stopOrder     int
	startError    error
	onStartBefore cfgtps.FuncCptEvent
	onStartAfter  cfgtps.FuncCptEvent
}

func (m *mockDatabase) Type() string { return "database" }

func (m *mockDatabase) Start() error {
	if m.onStartBefore != nil {
		_ = m.onStartBefore(m)
	}
	if m.startError != nil {
		return m.startError
	}
	m.started = true
	m.running = true
	m.startCount++
	startOrderCounter++
	m.startOrder = startOrderCounter
	if m.onStartAfter != nil {
		_ = m.onStartAfter(m)
	}
	return nil
}

func (m *mockDatabase) Stop() {
	m.started = false
	m.running = false
	m.stopped = true
	m.stopCount++
	stopOrderCounter++
	m.stopOrder = stopOrderCounter
}

type mockCache struct {
	simpleComponent
	startOrder    int
	stopOrder     int
	startError    error
	onStartBefore cfgtps.FuncCptEvent
	onStartAfter  cfgtps.FuncCptEvent
}

func (m *mockCache) Type() string { return "cache" }

func (m *mockCache) Start() error {
	if m.onStartBefore != nil {
		_ = m.onStartBefore(m)
	}
	if m.startError != nil {
		return m.startError
	}
	m.started = true
	m.running = true
	m.startCount++
	startOrderCounter++
	m.startOrder = startOrderCounter
	if m.onStartAfter != nil {
		_ = m.onStartAfter(m)
	}
	return nil
}

func (m *mockCache) Stop() {
	m.started = false
	m.running = false
	m.stopped = true
	m.stopCount++
	stopOrderCounter++
	m.stopOrder = stopOrderCounter
}

type mockAPI struct {
	simpleComponent
	startOrder    int
	stopOrder     int
	startError    error
	onStartBefore cfgtps.FuncCptEvent
	onStartAfter  cfgtps.FuncCptEvent
}

func (m *mockAPI) Type() string { return "api" }

func (m *mockAPI) Start() error {
	if m.onStartBefore != nil {
		_ = m.onStartBefore(m)
	}
	if m.startError != nil {
		return m.startError
	}
	m.started = true
	m.running = true
	m.startCount++
	startOrderCounter++
	m.startOrder = startOrderCounter
	if m.onStartAfter != nil {
		_ = m.onStartAfter(m)
	}
	return nil
}

func (m *mockAPI) Stop() {
	m.started = false
	m.running = false
	m.stopped = true
	m.stopCount++
	stopOrderCounter++
	m.stopOrder = stopOrderCounter
}
