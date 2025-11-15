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

package httpcli_test

import (
	"context"
	"io"
	"time"

	libmap "github.com/go-viper/mapstructure/v2"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

// sharedMockViper is a shared minimal mock implementation for testing
type sharedMockViper struct {
	v *spfvpr.Viper
}

func (m *sharedMockViper) Viper() *spfvpr.Viper {
	return m.v
}

func (m *sharedMockViper) UnmarshalKey(key string, rawVal interface{}) error {
	return m.v.UnmarshalKey(key, rawVal)
}

func (m *sharedMockViper) Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error {
	return nil
}

func (m *sharedMockViper) SetRemoteProvider(provider string)       {}
func (m *sharedMockViper) SetRemoteEndpoint(endpoint string)       {}
func (m *sharedMockViper) SetRemotePath(path string)               {}
func (m *sharedMockViper) SetRemoteSecureKey(key string)           {}
func (m *sharedMockViper) SetRemoteModel(model interface{})        {}
func (m *sharedMockViper) SetRemoteReloadFunc(fct func())          {}
func (m *sharedMockViper) SetHomeBaseName(base string)             {}
func (m *sharedMockViper) SetEnvVarsPrefix(prefix string)          {}
func (m *sharedMockViper) SetDefaultConfig(fct func() io.Reader)   {}
func (m *sharedMockViper) SetConfigFile(fileConfig string) error   { return nil }
func (m *sharedMockViper) WatchFS(logLevelFSInfo loglvl.Level)     {}
func (m *sharedMockViper) Unset(key ...string) error               { return nil }
func (m *sharedMockViper) HookRegister(hook libmap.DecodeHookFunc) {}
func (m *sharedMockViper) HookReset()                              {}
func (m *sharedMockViper) Unmarshal(rawVal interface{}) error      { return nil }
func (m *sharedMockViper) UnmarshalExact(rawVal interface{}) error { return nil }
func (m *sharedMockViper) GetBool(key string) bool                 { return m.v.GetBool(key) }
func (m *sharedMockViper) GetString(key string) string             { return m.v.GetString(key) }
func (m *sharedMockViper) GetInt(key string) int                   { return m.v.GetInt(key) }
func (m *sharedMockViper) GetInt32(key string) int32               { return m.v.GetInt32(key) }
func (m *sharedMockViper) GetInt64(key string) int64               { return m.v.GetInt64(key) }
func (m *sharedMockViper) GetUint(key string) uint                 { return m.v.GetUint(key) }
func (m *sharedMockViper) GetUint16(key string) uint16             { return m.v.GetUint16(key) }
func (m *sharedMockViper) GetUint32(key string) uint32             { return m.v.GetUint32(key) }
func (m *sharedMockViper) GetUint64(key string) uint64             { return m.v.GetUint64(key) }
func (m *sharedMockViper) GetFloat64(key string) float64           { return m.v.GetFloat64(key) }
func (m *sharedMockViper) GetTime(key string) time.Time            { return m.v.GetTime(key) }
func (m *sharedMockViper) GetDuration(key string) time.Duration    { return m.v.GetDuration(key) }
func (m *sharedMockViper) GetIntSlice(key string) []int            { return m.v.GetIntSlice(key) }
func (m *sharedMockViper) GetStringSlice(key string) []string      { return m.v.GetStringSlice(key) }
func (m *sharedMockViper) GetStringMap(key string) map[string]any  { return m.v.GetStringMap(key) }
func (m *sharedMockViper) GetStringMapString(key string) map[string]string {
	return m.v.GetStringMapString(key)
}
func (m *sharedMockViper) GetStringMapStringSlice(key string) map[string][]string {
	return m.v.GetStringMapStringSlice(key)
}

// sharedWrongComponent for testing type safety
type sharedWrongComponent struct{}

func (w *sharedWrongComponent) Type() string { return "wrong" }
func (w *sharedWrongComponent) Init(string, context.Context, cfgtps.FuncCptGet, libvpr.FuncViper, libver.Version, liblog.FuncLog) {
}
func (w *sharedWrongComponent) RegisterFuncStart(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent)  {}
func (w *sharedWrongComponent) RegisterFuncReload(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {}
func (w *sharedWrongComponent) IsStarted() bool                                             { return false }
func (w *sharedWrongComponent) IsRunning() bool                                             { return false }
func (w *sharedWrongComponent) Start() error                                                { return nil }
func (w *sharedWrongComponent) Reload() error                                               { return nil }
func (w *sharedWrongComponent) Stop()                                                       {}
func (w *sharedWrongComponent) Dependencies() []string                                      { return nil }
func (w *sharedWrongComponent) SetDependencies([]string) error                              { return nil }
func (w *sharedWrongComponent) DefaultConfig(string) []byte                                 { return nil }
func (w *sharedWrongComponent) RegisterFlag(*spfcbr.Command) error                          { return nil }
func (w *sharedWrongComponent) RegisterMonitorPool(montps.FuncPool)                         {}
