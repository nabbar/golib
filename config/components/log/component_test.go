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
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
)

// Component lifecycle tests verify the behavior of Init, Start, Reload, Stop
// and state management methods.
var _ = Describe("Component Lifecycle", func() {
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

	Describe("Initialization", func() {
		Context("Init method", func() {
			It("should initialize component with key", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				// Component should be initialized but not started
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should allow multiple initializations", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				// Should not panic or fail
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("State management", func() {
		Context("IsStarted method", func() {
			It("should return false before start", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should return false after initialization without start", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("IsRunning method", func() {
			It("should return false before start", func() {
				Expect(cpt.IsRunning()).To(BeFalse())
			})

			It("should match IsStarted state", func() {
				Expect(cpt.IsRunning()).To(Equal(cpt.IsStarted()))
			})
		})
	})

	Describe("Dependencies", func() {
		Context("Dependencies method", func() {
			It("should return empty slice by default", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
				Expect(deps).To(BeEmpty())
			})

			It("should return dependencies after SetDependencies", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				expectedDeps := []string{"dep1", "dep2"}
				err := cpt.SetDependencies(expectedDeps)
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal(expectedDeps))
			})

			It("should handle nil dependencies", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
			})
		})

		Context("SetDependencies method", func() {
			It("should set dependencies after Init", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())
			})

			It("should handle empty dependencies", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				err := cpt.SetDependencies([]string{})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})

			It("should allow updating dependencies", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				err := cpt.SetDependencies([]string{"dep1"})
				Expect(err).To(BeNil())

				err = cpt.SetDependencies([]string{"dep2", "dep3"})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal([]string{"dep2", "dep3"}))
			})
		})
	})

	Describe("Callback registration", func() {
		Context("RegisterFuncStart", func() {
			It("should not panic when registering callbacks", func() {
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
			It("should not panic when registering callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncReload(before, after)
				}).NotTo(Panic())
			})

			It("should accept nil callbacks", func() {
				Expect(func() {
					cpt.RegisterFuncReload(nil, nil)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Stop method", func() {
		Context("stopping component", func() {
			It("should not panic when called on uninitialized component", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should not panic when called on initialized component", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should allow multiple calls to Stop", func() {
				Expect(func() {
					cpt.Stop()
					cpt.Stop()
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Type method", func() {
		Context("component type", func() {
			It("should return 'log' as type", func() {
				Expect(cpt.Type()).To(Equal("log"))
			})

			It("should be consistent across calls", func() {
				type1 := cpt.Type()
				type2 := cpt.Type()
				Expect(type1).To(Equal(type2))
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("thread-safety", func() {
			It("should handle concurrent IsStarted calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.IsStarted()
						_ = cpt.IsRunning()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent Dependencies calls", func() {
				key := "test-log"
				getCpt := func(k string) cfgtps.Component { return nil }
				cpt.Init(key, ctx, getCpt, fv, vs, fl)
				_ = cpt.SetDependencies([]string{"dep1", "dep2"})

				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						deps := cpt.Dependencies()
						Expect(deps).NotTo(BeNil())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})
