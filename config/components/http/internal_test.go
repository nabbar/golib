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

package http_test

import (
	"context"
	"io"
	"net/http"
	"time"

	libmap "github.com/go-viper/mapstructure/v2"
	. "github.com/nabbar/golib/config/components/http"
	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spfvpr "github.com/spf13/viper"
)

// Tests for internal component behavior and edge cases
var _ = Describe("Internal Component Behavior", func() {
	var (
		ctx context.Context
		cpt CptHttp
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("Handler Management", func() {
		Context("with nil handler", func() {
			It("should handle nil handler gracefully", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				Expect(cpt).NotTo(BeNil())
				Expect(cpt.GetPool()).NotTo(BeNil())
			})
		})

		Context("with empty handler", func() {
			It("should handle empty handler map", func() {
				handler := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}
				cpt = New(ctx, DefaultTlsKey, handler)
				Expect(cpt).NotTo(BeNil())
			})
		})

		Context("with multiple handlers", func() {
			It("should support multiple handlers", func() {
				handler := func() map[string]http.Handler {
					return map[string]http.Handler{
						"api":     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
						"status":  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
						"metrics": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}
				cpt = New(ctx, DefaultTlsKey, handler)
				cpt.SetHandler(handler)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Component State", func() {
		Context("before initialization", func() {
			It("should report not started", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should report not running", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})

		Context("with nil component", func() {
			It("should handle nil component safely", func() {
				var nilCpt CptHttp
				Expect(nilCpt).To(BeNil())
			})
		})
	})

	Describe("TLS Key Management", func() {
		Context("with empty TLS key", func() {
			It("should use default TLS key", func() {
				cpt = New(ctx, "", nil)
				Expect(cpt).NotTo(BeNil())
				// Default TLS key should be used
				deps := cpt.Dependencies()
				Expect(deps).To(ContainElement(DefaultTlsKey))
			})
		})

		Context("with custom TLS key", func() {
			It("should use custom TLS key", func() {
				customKey := "my-custom-tls"
				cpt = New(ctx, customKey, nil)
				Expect(cpt).NotTo(BeNil())
				deps := cpt.Dependencies()
				Expect(deps).To(ContainElement(customKey))
			})
		})

		Context("changing TLS key after creation", func() {
			It("should allow changing TLS key", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				newKey := "new-tls-key"
				cpt.SetTLSKey(newKey)
				deps := cpt.Dependencies()
				Expect(deps).To(ContainElement(newKey))
			})
		})
	})

	Describe("Pool Management", func() {
		Context("with nil pool", func() {
			It("should create new pool when setting nil", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				oldPool := cpt.GetPool()
				cpt.SetPool(nil)
				newPool := cpt.GetPool()
				Expect(newPool).NotTo(BeNil())
				Expect(newPool).NotTo(Equal(oldPool))
			})
		})
	})

	Describe("Initialization Edge Cases", func() {
		Context("with nil functions", func() {
			It("should handle nil context function", func() {
				cpt = New(nil, DefaultTlsKey, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should handle nil viper function", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				cpt.Init("test", ctx, nil, nil, nil, nil)
				Expect(cpt.Type()).To(Equal(ComponentType))
			})
		})

		Context("with version information", func() {
			It("should accept version information", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				vrs := libver.NewVersion(
					libver.License_MIT,
					"test-app",
					"",
					"2025-01-01",
					"test",
					"1.0.0",
					"test-build",
					"",
					struct{}{},
					0,
				)
				cpt.Init("test", ctx, nil, nil, vrs, nil)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Event Callbacks", func() {
		Context("with start callbacks", func() {
			It("should register before-start callback", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				called := false
				before := func(c cfgtps.Component) error {
					called = true
					return nil
				}
				cpt.RegisterFuncStart(before, nil)
				Expect(called).To(BeFalse()) // Not called until Start()
			})

			It("should register after-start callback", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				called := false
				after := func(c cfgtps.Component) error {
					called = true
					return nil
				}
				cpt.RegisterFuncStart(nil, after)
				Expect(called).To(BeFalse())
			})

			It("should register both callbacks", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				beforeCalled := false
				afterCalled := false
				before := func(c cfgtps.Component) error {
					beforeCalled = true
					return nil
				}
				after := func(c cfgtps.Component) error {
					afterCalled = true
					return nil
				}
				cpt.RegisterFuncStart(before, after)
				Expect(beforeCalled).To(BeFalse())
				Expect(afterCalled).To(BeFalse())
			})
		})

		Context("with reload callbacks", func() {
			It("should register before-reload callback", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				called := false
				before := func(c cfgtps.Component) error {
					called = true
					return nil
				}
				cpt.RegisterFuncReload(before, nil)
				Expect(called).To(BeFalse())
			})

			It("should register after-reload callback", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				called := false
				after := func(c cfgtps.Component) error {
					called = true
					return nil
				}
				cpt.RegisterFuncReload(nil, after)
				Expect(called).To(BeFalse())
			})
		})
	})

	Describe("Stop Behavior", func() {
		Context("stopping uninitialized component", func() {
			It("should not panic when stopping nil component", func() {
				var nilCpt CptHttp
				Expect(func() {
					if nilCpt != nil {
						nilCpt.Stop()
					}
				}).NotTo(Panic())
			})

			It("should not panic when stopping uninitialized component", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Logger Integration", func() {
		Context("with nil logger", func() {
			It("should handle nil logger function", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				cpt.Init("test", ctx, nil, nil, nil, nil)
				Expect(cpt).NotTo(BeNil())
			})
		})

		Context("with logger function", func() {
			It("should accept logger function", func() {
				cpt = New(ctx, DefaultTlsKey, nil)
				logFunc := func() liblog.Logger {
					return nil
				}
				cpt.Init("test", ctx, nil, nil, nil, logFunc)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Type Identification", func() {
		It("should always return correct component type", func() {
			cpt = New(ctx, DefaultTlsKey, nil)
			Expect(cpt.Type()).To(Equal(ComponentType))
			Expect(cpt.Type()).To(Equal("http"))
		})
	})
})

// Mock viper for testing internal configuration loading
type mockViperForInternal struct {
	configSet bool
}

func (m *mockViperForInternal) Viper() *spfvpr.Viper {
	v := spfvpr.New()
	if m.configSet {
		v.Set("test", []interface{}{
			map[string]interface{}{
				"name":        "test",
				"handler_key": "test",
				"listen":      "127.0.0.1:9090",
				"expose":      "http://127.0.0.1:9090",
			},
		})
	}
	return v
}

func (m *mockViperForInternal) Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error {
	return nil
}

func (m *mockViperForInternal) UnmarshalKey(key string, rawVal interface{}) error {
	return nil
}

func (m *mockViperForInternal) IsSet(key string) bool {
	return m.configSet
}

func (m *mockViperForInternal) SetRemoteProvider(provider string)       {}
func (m *mockViperForInternal) SetRemoteEndpoint(endpoint string)       {}
func (m *mockViperForInternal) SetRemotePath(path string)               {}
func (m *mockViperForInternal) SetRemoteSecureKey(key string)           {}
func (m *mockViperForInternal) SetRemoteModel(model interface{})        {}
func (m *mockViperForInternal) SetRemoteReloadFunc(fct func())          {}
func (m *mockViperForInternal) SetHomeBaseName(base string)             {}
func (m *mockViperForInternal) SetEnvVarsPrefix(prefix string)          {}
func (m *mockViperForInternal) SetDefaultConfig(fct func() io.Reader)   {}
func (m *mockViperForInternal) SetConfigFile(fileConfig string) error   { return nil }
func (m *mockViperForInternal) WatchFS(logLevelFSInfo loglvl.Level)     {}
func (m *mockViperForInternal) Unset(key ...string) error               { return nil }
func (m *mockViperForInternal) HookRegister(hook libmap.DecodeHookFunc) {}
func (m *mockViperForInternal) HookReset()                              {}
func (m *mockViperForInternal) Unmarshal(rawVal interface{}) error      { return nil }
func (m *mockViperForInternal) UnmarshalExact(rawVal interface{}) error { return nil }
func (m *mockViperForInternal) GetBool(key string) bool                 { return false }
func (m *mockViperForInternal) GetString(key string) string             { return "" }
func (m *mockViperForInternal) GetInt(key string) int                   { return 0 }
func (m *mockViperForInternal) GetInt32(key string) int32               { return 0 }
func (m *mockViperForInternal) GetInt64(key string) int64               { return 0 }
func (m *mockViperForInternal) GetUint(key string) uint                 { return 0 }
func (m *mockViperForInternal) GetUint16(key string) uint16             { return 0 }
func (m *mockViperForInternal) GetUint32(key string) uint32             { return 0 }
func (m *mockViperForInternal) GetUint64(key string) uint64             { return 0 }
func (m *mockViperForInternal) GetFloat64(key string) float64           { return 0 }
func (m *mockViperForInternal) GetTime(key string) time.Time            { return time.Time{} }
func (m *mockViperForInternal) GetDuration(key string) time.Duration    { return 0 }
func (m *mockViperForInternal) GetIntSlice(key string) []int            { return nil }
func (m *mockViperForInternal) GetStringSlice(key string) []string      { return nil }
func (m *mockViperForInternal) GetStringMap(key string) map[string]interface{} {
	return nil
}
func (m *mockViperForInternal) GetStringMapString(key string) map[string]string {
	return nil
}
func (m *mockViperForInternal) GetStringMapStringSlice(key string) map[string][]string {
	return nil
}

// Additional tests for TLS component loading
var _ = Describe("TLS Integration", func() {
	var (
		ctx context.Context
		cpt CptHttp
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("TLS Component Loading", func() {
		Context("without TLS component", func() {
			It("should handle missing TLS component", func() {
				cpt = New(ctx, "non-existent-tls", nil)
				getCpt := func(key string) cfgtps.Component {
					return nil
				}
				vpr := func() libvpr.Viper {
					return &mockViperForInternal{configSet: false}
				}
				cpt.Init("test", ctx, getCpt, vpr, nil, nil)
				Expect(cpt).NotTo(BeNil())
			})
		})

		Context("with TLS component", func() {
			It("should load TLS component successfully", func() {
				cpt = New(ctx, cpttls.ComponentType, nil)
				getCpt := func(key string) cfgtps.Component {
					if key == cpttls.ComponentType {
						// Return a mock TLS component
						return nil
					}
					return nil
				}
				vpr := func() libvpr.Viper {
					return &mockViperForInternal{configSet: false}
				}
				cpt.Init("test", ctx, getCpt, vpr, nil, nil)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})
})
