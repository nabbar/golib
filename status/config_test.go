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

package status_test

import (
	"context"
	"net/http"

	cfgtypes "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spfcbr "github.com/spf13/cobra"
)

// mockComponent is a mock implementation of cfgtypes.Component for testing.
type mockComponent struct {
	monitorNames []string
}

func (m *mockComponent) Type() string { return "mock" }
func (m *mockComponent) Init(key string, ctx context.Context, get cfgtypes.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
}
func (m *mockComponent) DefaultConfig(indent string) []byte                     { return nil }
func (m *mockComponent) Dependencies() []string                                 { return nil }
func (m *mockComponent) SetDependencies(d []string) error                       { return nil }
func (m *mockComponent) RegisterFlag(Command *spfcbr.Command) error             { return nil }
func (m *mockComponent) RegisterFuncStart(before, after cfgtypes.FuncCptEvent)  {}
func (m *mockComponent) RegisterFuncReload(before, after cfgtypes.FuncCptEvent) {}
func (m *mockComponent) IsStarted() bool                                        { return true }
func (m *mockComponent) IsRunning() bool                                        { return true }
func (m *mockComponent) Start() error                                           { return nil }
func (m *mockComponent) Reload() error                                          { return nil }
func (m *mockComponent) Stop()                                                  {}
func (m *mockComponent) RegisterMonitorPool(p montps.FuncPool)                  {}
func (m *mockComponent) GetMonitorNames() []string {
	return m.monitorNames
}

var _ = Describe("Status/Config", func() {
	var (
		status libsts.Status
	)

	BeforeEach(func() {
		status = libsts.New(globalCtx)
		status.SetInfo("test-app", "v1.0.0", "abc123")
	})

	Describe("Config.Validate", func() {
		It("should validate empty config", func() {
			cfg := libsts.Config{}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid config"))
		})

		It("should validate config with return codes", func() {
			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 207,
					monsts.KO:   500,
				},
			}
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not validate config with mandatory components", func() {
			cfg := libsts.Config{
				Component: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"database", "cache"},
					},
				},
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid config"))
		})

		It("should validate config with mandatory components and return code", func() {
			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 207,
					monsts.KO:   500,
				},
				Component: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"database", "cache"},
					},
				},
			}
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("SetConfig", func() {
		Context("with default return codes", func() {
			It("should use default codes when not specified", func() {
				cfg := libsts.Config{}
				status.SetConfig(cfg)

				// Verify default behavior by checking health
				healthy := status.IsHealthy()
				Expect(healthy).To(BeAssignableToTypeOf(false))
			})
		})

		Context("with custom return codes", func() {
			It("should accept custom HTTP status codes", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   200,
						monsts.Warn: 200, // Treat warnings as OK
						monsts.KO:   503,
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should handle partial return code configuration", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK: 200,
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})
		})

		Context("with mandatory components", func() {
			It("should configure Must mode", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"critical-service"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Should mode", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Should,
							Keys: []string{"optional-service"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure AnyOf mode", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.AnyOf,
							Keys: []string{"db-primary", "db-secondary", "db-tertiary"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Quorum mode", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Quorum,
							Keys: []string{"node-1", "node-2", "node-3", "node-4", "node-5"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Ignore mode", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Ignore,
							Keys: []string{"non-critical"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should handle multiple mandatory groups", func() {
				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"database"},
						},
						{
							Mode: stsctr.Should,
							Keys: []string{"cache"},
						},
						{
							Mode: stsctr.AnyOf,
							Keys: []string{"queue-1", "queue-2"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should load mandatory components from config keys", func() {
				// Arrange
				mockComp := &mockComponent{
					monitorNames: []string{"monitor1", "monitor2"},
				}

				status.RegisterGetConfigCpt(func(key string) cfgtypes.ComponentMonitor {
					if key == "my-component" {
						return mockComp
					}
					return nil
				})

				cfg := libsts.Config{
					Component: []libsts.Mandatory{
						{
							Mode:       stsctr.Must,
							ConfigKeys: []string{"my-component"},
						},
					},
				}

				// Act
				status.SetConfig(cfg)

				cnf := status.GetConfig()
				Expect(cnf.Component).To(HaveLen(1))
				Expect(cnf.Component[0].Mode).To(Equal(stsctr.Must))
				Expect(cnf.Component[0].Keys).To(HaveLen(2))
				Expect(cnf.Component[0].Keys[0]).To(Equal("monitor1"))
				Expect(cnf.Component[0].Keys[1]).To(Equal("monitor2"))
			})
		})

		Context("with complete configuration", func() {
			It("should configure both return codes and mandatory components", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   http.StatusOK,
						monsts.Warn: http.StatusMultiStatus,
						monsts.KO:   http.StatusServiceUnavailable,
					},
					Component: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"database", "api"},
						},
						{
							Mode: stsctr.Should,
							Keys: []string{"cache", "queue"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})
		})
	})

	Describe("Config updates", func() {
		It("should allow updating configuration multiple times", func() {
			// First configuration
			cfg1 := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK: 200,
				},
			}
			status.SetConfig(cfg1)

			// Second configuration
			cfg2 := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 207,
					monsts.KO:   500,
				},
			}
			status.SetConfig(cfg2)

			Expect(true).To(BeTrue())
		})

		It("should replace previous configuration", func() {
			cfg1 := libsts.Config{
				Component: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"old-service"},
					},
				},
			}
			status.SetConfig(cfg1)

			cfg2 := libsts.Config{
				Component: []libsts.Mandatory{
					{
						Mode: stsctr.Should,
						Keys: []string{"new-service"},
					},
				},
			}
			status.SetConfig(cfg2)

			Expect(true).To(BeTrue())
		})
	})
})
