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
	"bytes"
	"context"
	"encoding/json"
	"io"

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

var _ = Describe("Config Components", func() {
	var cfg libcfg.Config

	BeforeEach(func() {
		cfg = libcfg.New(nil)
	})

	Describe("ComponentSet", func() {
		It("should register a component", func() {
			cpt := &simpleComponent{name: "comp1"}
			cfg.ComponentSet("comp1", cpt)
			Expect(cfg.ComponentHas("comp1")).To(BeTrue())
		})

		It("should initialize component with key", func() {
			cpt := &simpleComponent{name: "comp1"}
			cfg.ComponentSet("mykey", cpt)
			Expect(cpt.key).To(Equal("mykey"))
		})

		It("should replace existing component", func() {
			cpt1 := &simpleComponent{name: "comp1"}
			cpt2 := &simpleComponent{name: "comp2"}
			cfg.ComponentSet("key", cpt1)
			cfg.ComponentSet("key", cpt2)

			retrieved := cfg.ComponentGet("key").(*simpleComponent)
			Expect(retrieved.name).To(Equal("comp2"))
		})
	})

	Describe("ComponentGet", func() {
		It("should return component by key", func() {
			cpt := &simpleComponent{name: "comp1"}
			cfg.ComponentSet("comp1", cpt)
			retrieved := cfg.ComponentGet("comp1")
			Expect(retrieved).To(Equal(cpt))
		})

		It("should return nil for non-existent component", func() {
			retrieved := cfg.ComponentGet("nonexistent")
			Expect(retrieved).To(BeNil())
		})
	})

	Describe("ComponentHas", func() {
		It("should return true for existing component", func() {
			cpt := &simpleComponent{name: "comp1"}
			cfg.ComponentSet("comp1", cpt)
			Expect(cfg.ComponentHas("comp1")).To(BeTrue())
		})

		It("should return false for non-existent component", func() {
			Expect(cfg.ComponentHas("nonexistent")).To(BeFalse())
		})
	})

	Describe("ComponentType", func() {
		It("should return component type", func() {
			cpt := &simpleComponent{name: "comp1", typ: "database"}
			cfg.ComponentSet("comp1", cpt)
			Expect(cfg.ComponentType("comp1")).To(Equal("database"))
		})

		It("should return empty string for non-existent component", func() {
			Expect(cfg.ComponentType("nonexistent")).To(Equal(""))
		})
	})

	Describe("ComponentDel", func() {
		It("should delete component", func() {
			cpt := &simpleComponent{name: "comp1"}
			cfg.ComponentSet("comp1", cpt)
			cfg.ComponentDel("comp1")
			Expect(cfg.ComponentHas("comp1")).To(BeFalse())
		})

		It("should not panic when deleting non-existent component", func() {
			Expect(func() {
				cfg.ComponentDel("nonexistent")
			}).ToNot(Panic())
		})
	})

	Describe("ComponentList", func() {
		It("should return empty map when no components", func() {
			list := cfg.ComponentList()
			Expect(list).To(HaveLen(0))
		})

		It("should return all registered components", func() {
			cpt1 := &simpleComponent{name: "comp1"}
			cpt2 := &simpleComponent{name: "comp2"}
			cfg.ComponentSet("comp1", cpt1)
			cfg.ComponentSet("comp2", cpt2)

			list := cfg.ComponentList()
			Expect(list).To(HaveLen(2))
			Expect(list["comp1"]).To(Equal(cpt1))
			Expect(list["comp2"]).To(Equal(cpt2))
		})
	})

	Describe("ComponentKeys", func() {
		It("should return empty slice when no components", func() {
			keys := cfg.ComponentKeys()
			Expect(keys).To(HaveLen(0))
		})

		It("should return all component keys", func() {
			cfg.ComponentSet("comp1", &simpleComponent{name: "comp1"})
			cfg.ComponentSet("comp2", &simpleComponent{name: "comp2"})

			keys := cfg.ComponentKeys()
			Expect(keys).To(HaveLen(2))
			Expect(keys).To(ContainElement("comp1"))
			Expect(keys).To(ContainElement("comp2"))
		})
	})

	Describe("ComponentDependencies", func() {
		It("should return components in dependency order", func() {
			// comp2 depends on comp1
			comp1 := &simpleComponent{name: "comp1", deps: []string{}}
			comp2 := &simpleComponent{name: "comp2", deps: []string{"comp1"}}

			cfg.ComponentSet("comp2", comp2)
			cfg.ComponentSet("comp1", comp1)

			deps := cfg.(interface{ ComponentDependencies() []string }).ComponentDependencies()
			// comp1 should come before comp2
			comp1Idx := -1
			comp2Idx := -1
			for i, k := range deps {
				if k == "comp1" {
					comp1Idx = i
				}
				if k == "comp2" {
					comp2Idx = i
				}
			}
			Expect(comp1Idx).To(BeNumerically("<", comp2Idx))
		})

		It("should handle multiple dependencies", func() {
			// comp3 depends on comp1 and comp2
			comp1 := &simpleComponent{name: "comp1", deps: []string{}}
			comp2 := &simpleComponent{name: "comp2", deps: []string{}}
			comp3 := &simpleComponent{name: "comp3", deps: []string{"comp1", "comp2"}}

			cfg.ComponentSet("comp3", comp3)
			cfg.ComponentSet("comp2", comp2)
			cfg.ComponentSet("comp1", comp1)

			deps := cfg.(interface{ ComponentDependencies() []string }).ComponentDependencies()
			// comp3 should come last
			comp3Idx := -1
			for i, k := range deps {
				if k == "comp3" {
					comp3Idx = i
				}
			}
			Expect(comp3Idx).To(Equal(len(deps) - 1))
		})
	})

	Describe("DefaultConfig", func() {
		It("should generate default config JSON", func() {
			comp1 := &simpleComponent{name: "comp1", defaultConfig: `{"enabled": true}`}
			comp2 := &simpleComponent{name: "comp2", defaultConfig: `{"port": 8080}`}

			cfg.ComponentSet("comp1", comp1)
			cfg.ComponentSet("comp2", comp2)

			reader := cfg.(interface{ DefaultConfig() io.Reader }).DefaultConfig()
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())

			content := buf.String()
			Expect(content).To(ContainSubstring("comp1"))
			Expect(content).To(ContainSubstring("comp2"))
			Expect(content).To(ContainSubstring("enabled"))
			Expect(content).To(ContainSubstring("port"))
		})

		It("should return valid JSON structure", func() {
			comp1 := &simpleComponent{name: "comp1", defaultConfig: `{"enabled": true}`}
			cfg.ComponentSet("comp1", comp1)

			reader := cfg.(interface{ DefaultConfig() io.Reader }).DefaultConfig()
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())

			// Should be valid JSON
			var result map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &result)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

// simpleComponent is a minimal component for testing
type simpleComponent struct {
	key           string
	name          string
	typ           string
	deps          []string
	defaultConfig string
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

func (s *simpleComponent) Type() string {
	if s.typ == "" {
		return "simple"
	}
	return s.typ
}

func (s *simpleComponent) Init(key string, _ context.Context, _ cfgtps.FuncCptGet, _ libvpr.FuncViper, _ libver.Version, _ liblog.FuncLog) {
	s.key = key
}

func (s *simpleComponent) DefaultConfig(indent string) []byte {
	if s.defaultConfig == "" {
		return []byte("{}")
	}
	return []byte(s.defaultConfig)
}

func (s *simpleComponent) Dependencies() []string {
	return append([]string{}, s.deps...)
}

func (s *simpleComponent) SetDependencies(d []string) error {
	s.deps = append([]string{}, d...)
	return nil
}

func (s *simpleComponent) RegisterFlag(_ *spfcbr.Command) error {
	return nil
}

func (s *simpleComponent) RegisterMonitorPool(_ montps.FuncPool) {}

func (s *simpleComponent) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {}

func (s *simpleComponent) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {}

func (s *simpleComponent) IsStarted() bool {
	return s.started
}

func (s *simpleComponent) IsRunning() bool {
	return s.running
}

func (s *simpleComponent) Start() error {
	if s.onStartBefore != nil {
		_ = s.onStartBefore(s)
	}
	if s.startError != nil {
		return s.startError
	}
	s.started = true
	s.running = true
	s.startCount++
	if s.onStartAfter != nil {
		_ = s.onStartAfter(s)
	}
	return nil
}

func (s *simpleComponent) Reload() error {
	if s.onRelBefore != nil {
		_ = s.onRelBefore(s)
	}
	if s.reloadError != nil {
		return s.reloadError
	}
	s.reloadCount++
	if s.onRelAfter != nil {
		_ = s.onRelAfter(s)
	}
	return nil
}

func (s *simpleComponent) Stop() {
	s.started = false
	s.running = false
	s.stopped = true
	s.stopCount++
}
