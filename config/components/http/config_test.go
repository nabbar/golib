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
	"time"

	. "github.com/nabbar/golib/config/components/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"

	libmap "github.com/go-viper/mapstructure/v2"
	spfvpr "github.com/spf13/viper"
)

// Configuration tests verify config handling and validation
var _ = Describe("Configuration", func() {
	var (
		ctx context.Context
		cpt CptHttp
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, DefaultTlsKey, nil)
	})

	Describe("RegisterFlag method", func() {
		Context("with cobra command", func() {
			It("should accept valid command", func() {
				cmd := &spfcbr.Command{
					Use: "test",
				}

				err := cpt.RegisterFlag(cmd)
				Expect(err).To(BeNil())
			})

			It("should accept nil command", func() {
				err := cpt.RegisterFlag(nil)
				Expect(err).To(BeNil())
			})

			It("should not modify command", func() {
				cmd := &spfcbr.Command{
					Use:   "test",
					Short: "test command",
				}

				initialFlags := cmd.Flags().NFlag()
				err := cpt.RegisterFlag(cmd)

				Expect(err).To(BeNil())
				// RegisterFlag returns nil without adding flags
				Expect(cmd.Flags().NFlag()).To(Equal(initialFlags))
			})
		})

		Context("multiple calls", func() {
			It("should handle multiple RegisterFlag calls", func() {
				cmd1 := &spfcbr.Command{Use: "cmd1"}
				cmd2 := &spfcbr.Command{Use: "cmd2"}

				err1 := cpt.RegisterFlag(cmd1)
				err2 := cpt.RegisterFlag(cmd2)

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
			})
		})
	})

	Describe("Configuration retrieval", func() {
		Context("without initialization", func() {
			It("should fail to get config without viper", func() {
				// Component not initialized, no viper set
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
				// Error message can be "initialized" or "start" depending on the error path
				Expect(err.Error()).To(Or(ContainSubstring("initialized"), ContainSubstring("start")))
			})
		})

		Context("with initialization but no viper", func() {
			It("should fail to get config", func() {
				key := "http-server"
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with empty key", func() {
			It("should fail to get config with empty key", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init("", ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Configuration validation", func() {
		Context("missing required fields", func() {
			It("should return error for missing config key", func() {
				key := "http-server"
				getCpt := func(k string) cfgtps.Component { return nil }

				// Mock viper that returns nil (no config set)
				mockViper := &mockViperEmpty{}
				vpr := func() libvpr.Viper { return mockViper }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("invalid configuration", func() {
			It("should return error for invalid config", func() {
				key := "http-server"
				getCpt := func(k string) cfgtps.Component { return nil }

				// Mock viper with invalid config
				mockViper := &mockViperInvalid{}
				vpr := func() libvpr.Viper { return mockViper }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("nil component", func() {
			It("should panic on RegisterFlag with nil component", func() {
				var nilCpt CptHttp
				cmd := &spfcbr.Command{Use: "test"}

				Expect(func() {
					_ = nilCpt.RegisterFlag(cmd)
				}).To(Panic())
			})
		})

		Context("concurrent RegisterFlag calls", func() {
			It("should handle concurrent RegisterFlag calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(index int) {
						defer GinkgoRecover()
						cmd := &spfcbr.Command{Use: "test"}
						err := cpt.RegisterFlag(cmd)
						Expect(err).To(BeNil())
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})

// mockViperEmpty is a mock Viper that returns nil/empty values
type mockViperEmpty struct{}

func (m *mockViperEmpty) Viper() *spfvpr.Viper {
	return spfvpr.New()
}
func (m *mockViperEmpty) Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error {
	return nil
}
func (m *mockViperEmpty) UnmarshalKey(key string, rawVal interface{}) error {
	return nil
}
func (m *mockViperEmpty) IsSet(key string) bool {
	return false
}
func (m *mockViperEmpty) SetRemoteProvider(provider string)       {}
func (m *mockViperEmpty) SetRemoteEndpoint(endpoint string)       {}
func (m *mockViperEmpty) SetRemotePath(path string)               {}
func (m *mockViperEmpty) SetRemoteSecureKey(key string)           {}
func (m *mockViperEmpty) SetRemoteModel(model interface{})        {}
func (m *mockViperEmpty) SetRemoteReloadFunc(fct func())          {}
func (m *mockViperEmpty) SetHomeBaseName(base string)             {}
func (m *mockViperEmpty) SetEnvVarsPrefix(prefix string)          {}
func (m *mockViperEmpty) SetDefaultConfig(fct func() io.Reader)   {}
func (m *mockViperEmpty) SetConfigFile(fileConfig string) error   { return nil }
func (m *mockViperEmpty) WatchFS(logLevelFSInfo loglvl.Level)     {}
func (m *mockViperEmpty) Unset(key ...string) error               { return nil }
func (m *mockViperEmpty) HookRegister(hook libmap.DecodeHookFunc) {}
func (m *mockViperEmpty) HookReset()                              {}
func (m *mockViperEmpty) Unmarshal(rawVal interface{}) error      { return nil }
func (m *mockViperEmpty) UnmarshalExact(rawVal interface{}) error { return nil }
func (m *mockViperEmpty) GetBool(key string) bool                 { return false }
func (m *mockViperEmpty) GetString(key string) string             { return "" }
func (m *mockViperEmpty) GetInt(key string) int                   { return 0 }
func (m *mockViperEmpty) GetInt32(key string) int32               { return 0 }
func (m *mockViperEmpty) GetInt64(key string) int64               { return 0 }
func (m *mockViperEmpty) GetUint(key string) uint                 { return 0 }
func (m *mockViperEmpty) GetUint16(key string) uint16             { return 0 }
func (m *mockViperEmpty) GetUint32(key string) uint32             { return 0 }
func (m *mockViperEmpty) GetUint64(key string) uint64             { return 0 }
func (m *mockViperEmpty) GetFloat64(key string) float64           { return 0 }
func (m *mockViperEmpty) GetTime(key string) time.Time            { return time.Time{} }
func (m *mockViperEmpty) GetDuration(key string) time.Duration    { return 0 }
func (m *mockViperEmpty) GetIntSlice(key string) []int            { return nil }
func (m *mockViperEmpty) GetStringSlice(key string) []string      { return nil }
func (m *mockViperEmpty) GetStringMap(key string) map[string]any  { return nil }
func (m *mockViperEmpty) GetStringMapString(key string) map[string]string {
	return nil
}
func (m *mockViperEmpty) GetStringMapStringSlice(key string) map[string][]string {
	return nil
}

// mockSpfViper is a minimal mock of spf13/viper.Viper
type mockSpfViper struct{}

func (m *mockSpfViper) IsSet(key string) bool {
	return false
}

// mockViperInvalid is a mock Viper that returns invalid config
type mockViperInvalid struct{}

func (m *mockViperInvalid) Viper() *spfvpr.Viper {
	return spfvpr.New()
}
func (m *mockViperInvalid) Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error {
	return nil
}
func (m *mockViperInvalid) UnmarshalKey(key string, rawVal interface{}) error {
	return nil
}
func (m *mockViperInvalid) IsSet(key string) bool {
	return true
}
func (m *mockViperInvalid) SetRemoteProvider(provider string)       {}
func (m *mockViperInvalid) SetRemoteEndpoint(endpoint string)       {}
func (m *mockViperInvalid) SetRemotePath(path string)               {}
func (m *mockViperInvalid) SetRemoteSecureKey(key string)           {}
func (m *mockViperInvalid) SetRemoteModel(model interface{})        {}
func (m *mockViperInvalid) SetRemoteReloadFunc(fct func())          {}
func (m *mockViperInvalid) SetHomeBaseName(base string)             {}
func (m *mockViperInvalid) SetEnvVarsPrefix(prefix string)          {}
func (m *mockViperInvalid) SetDefaultConfig(fct func() io.Reader)   {}
func (m *mockViperInvalid) SetConfigFile(fileConfig string) error   { return nil }
func (m *mockViperInvalid) WatchFS(logLevelFSInfo loglvl.Level)     {}
func (m *mockViperInvalid) Unset(key ...string) error               { return nil }
func (m *mockViperInvalid) HookRegister(hook libmap.DecodeHookFunc) {}
func (m *mockViperInvalid) HookReset()                              {}
func (m *mockViperInvalid) Unmarshal(rawVal interface{}) error      { return nil }
func (m *mockViperInvalid) UnmarshalExact(rawVal interface{}) error { return nil }
func (m *mockViperInvalid) GetBool(key string) bool                 { return false }
func (m *mockViperInvalid) GetString(key string) string             { return "" }
func (m *mockViperInvalid) GetInt(key string) int                   { return 0 }
func (m *mockViperInvalid) GetInt32(key string) int32               { return 0 }
func (m *mockViperInvalid) GetInt64(key string) int64               { return 0 }
func (m *mockViperInvalid) GetUint(key string) uint                 { return 0 }
func (m *mockViperInvalid) GetUint16(key string) uint16             { return 0 }
func (m *mockViperInvalid) GetUint32(key string) uint32             { return 0 }
func (m *mockViperInvalid) GetUint64(key string) uint64             { return 0 }
func (m *mockViperInvalid) GetFloat64(key string) float64           { return 0 }
func (m *mockViperInvalid) GetTime(key string) time.Time            { return time.Time{} }
func (m *mockViperInvalid) GetDuration(key string) time.Duration    { return 0 }
func (m *mockViperInvalid) GetIntSlice(key string) []int            { return nil }
func (m *mockViperInvalid) GetStringSlice(key string) []string      { return nil }
func (m *mockViperInvalid) GetStringMap(key string) map[string]any  { return nil }
func (m *mockViperInvalid) GetStringMapString(key string) map[string]string {
	return nil
}
func (m *mockViperInvalid) GetStringMapStringSlice(key string) map[string][]string {
	return nil
}

// mockSpfViperInvalid is a mock that reports IsSet true
type mockSpfViperInvalid struct{}

func (m *mockSpfViperInvalid) IsSet(key string) bool {
	return true
}
