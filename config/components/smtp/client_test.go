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

package smtp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	. "github.com/nabbar/golib/config/components/smtp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libmap "github.com/go-viper/mapstructure/v2"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

// Client tests verify the internal client operations, configuration loading,
// and callback execution.
var _ = Describe("Client Operations", func() {
	var (
		cpt CptSMTP
		ctx context.Context
		vrs libver.Version
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		key = "test-smtp"
		cpt = New(ctx, "")
	})

	Describe("Start with configuration", func() {
		Context("with valid viper configuration", func() {
			It("should handle viper config", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

				configData := map[string]interface{}{
					key: map[string]interface{}{
						"host": "localhost",
						"port": 1025,
						"from": "test@example.com",
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &testMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				// Start may fail if no SMTP server or TLS, but shouldn't panic
				Expect(func() {
					_ = cpt.Start()
				}).NotTo(Panic())
			})

			It("should handle invalid config gracefully", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

				configData := map[string]interface{}{
					key: map[string]interface{}{
						"invalid": "config",
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &testMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err = cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Callback execution", func() {
		BeforeEach(func() {
			v := spfvpr.New()
			v.SetConfigType("json")

			configData := map[string]interface{}{
				key: map[string]interface{}{
					"host": "localhost",
					"port": 1025,
					"from": "test@example.com",
				},
			}

			configJSON, _ := json.Marshal(configData)
			_ = v.ReadConfig(bytes.NewReader(configJSON))

			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper {
				return &testMockViper{v: v}
			}
			log := func() liblog.Logger { return nil }

			cpt.Init(key, ctx, getCpt, vpr, vrs, log)
		})

		Context("Start callbacks", func() {
			It("should register and potentially call start callbacks", func() {
				beforeStart := func(c cfgtps.Component) error {
					return nil
				}
				afterStart := func(c cfgtps.Component) error {
					return nil
				}

				// Register callbacks
				Expect(func() {
					cpt.RegisterFuncStart(beforeStart, afterStart)
				}).NotTo(Panic())

				// Start may or may not succeed, but should not panic
				Expect(func() {
					_ = cpt.Start()
				}).NotTo(Panic())
			})

			It("should handle callback errors", func() {
				beforeStart := func(c cfgtps.Component) error {
					return ErrorComponentNotInitialized.Error(nil)
				}

				cpt.RegisterFuncStart(beforeStart, nil)
				err := cpt.Start()

				Expect(err).To(HaveOccurred())
			})
		})

		Context("Reload callbacks", func() {
			It("should register and potentially call reload callbacks", func() {
				beforeReload := func(c cfgtps.Component) error {
					return nil
				}
				afterReload := func(c cfgtps.Component) error {
					return nil
				}

				// Register callbacks
				Expect(func() {
					cpt.RegisterFuncReload(beforeReload, afterReload)
				}).NotTo(Panic())

				// Reload may or may not succeed, but should not panic
				Expect(func() {
					_ = cpt.Reload()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Configuration loading", func() {
		Context("with missing viper", func() {
			It("should handle nil viper", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with empty config", func() {
			It("should handle empty configuration", func() {
				v := spfvpr.New()
				v.SetConfigType("json")

				configData := map[string]interface{}{}
				configJSON, _ := json.Marshal(configData)
				_ = v.ReadConfig(bytes.NewReader(configJSON))

				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper {
					return &testMockViper{v: v}
				}
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

// testMockViper is a minimal mock for client tests
type testMockViper struct {
	v *spfvpr.Viper
}

func (m *testMockViper) Viper() *spfvpr.Viper { return m.v }
func (m *testMockViper) UnmarshalKey(key string, rawVal interface{}) error {
	return m.v.UnmarshalKey(key, rawVal)
}
func (m *testMockViper) Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error { return nil }
func (m *testMockViper) SetRemoteProvider(provider string)                            {}
func (m *testMockViper) SetRemoteEndpoint(endpoint string)                            {}
func (m *testMockViper) SetRemotePath(path string)                                    {}
func (m *testMockViper) SetRemoteSecureKey(key string)                                {}
func (m *testMockViper) SetRemoteModel(model interface{})                             {}
func (m *testMockViper) SetRemoteReloadFunc(fct func())                               {}
func (m *testMockViper) SetHomeBaseName(base string)                                  {}
func (m *testMockViper) SetEnvVarsPrefix(prefix string)                               {}
func (m *testMockViper) SetDefaultConfig(fct func() io.Reader)                        {}
func (m *testMockViper) SetConfigFile(fileConfig string) error                        { return nil }
func (m *testMockViper) WatchFS(logLevelFSInfo loglvl.Level)                          {}
func (m *testMockViper) Unset(key ...string) error                                    { return nil }
func (m *testMockViper) HookRegister(hook libmap.DecodeHookFunc)                      {}
func (m *testMockViper) HookReset()                                                   {}
func (m *testMockViper) Unmarshal(rawVal interface{}) error                           { return nil }
func (m *testMockViper) UnmarshalExact(rawVal interface{}) error                      { return nil }
func (m *testMockViper) GetBool(key string) bool                                      { return m.v.GetBool(key) }
func (m *testMockViper) GetString(key string) string                                  { return m.v.GetString(key) }
func (m *testMockViper) GetInt(key string) int                                        { return m.v.GetInt(key) }
func (m *testMockViper) GetInt32(key string) int32                                    { return m.v.GetInt32(key) }
func (m *testMockViper) GetInt64(key string) int64                                    { return m.v.GetInt64(key) }
func (m *testMockViper) GetUint(key string) uint                                      { return m.v.GetUint(key) }
func (m *testMockViper) GetUint16(key string) uint16                                  { return m.v.GetUint16(key) }
func (m *testMockViper) GetUint32(key string) uint32                                  { return m.v.GetUint32(key) }
func (m *testMockViper) GetUint64(key string) uint64                                  { return m.v.GetUint64(key) }
func (m *testMockViper) GetFloat64(key string) float64                                { return m.v.GetFloat64(key) }
func (m *testMockViper) GetTime(key string) time.Time                                 { return m.v.GetTime(key) }
func (m *testMockViper) GetDuration(key string) time.Duration                         { return m.v.GetDuration(key) }
func (m *testMockViper) GetIntSlice(key string) []int                                 { return m.v.GetIntSlice(key) }
func (m *testMockViper) GetStringSlice(key string) []string                           { return m.v.GetStringSlice(key) }
func (m *testMockViper) GetStringMap(key string) map[string]any                       { return m.v.GetStringMap(key) }
func (m *testMockViper) GetStringMapString(key string) map[string]string {
	return m.v.GetStringMapString(key)
}
func (m *testMockViper) GetStringMapStringSlice(key string) map[string][]string {
	return m.v.GetStringMapStringSlice(key)
}
