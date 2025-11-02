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

package viper

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	libmap "github.com/go-viper/mapstructure/v2"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	spfvpr "github.com/spf13/viper"
)

// FuncViper is a function type that returns a Viper instance.
// This is commonly used for dependency injection and lazy initialization.
type FuncViper func() Viper

// FuncSPFViper is a function type that returns the underlying spf13/viper instance.
// This allows direct access to the viper library when needed.
type FuncSPFViper func() *spfvpr.Viper

// FuncConfigGet is a function type for retrieving configuration values.
// It takes a key string and a model interface to unmarshal the value into.
type FuncConfigGet func(key string, model interface{}) error

// Viper is the main interface for configuration management.
// It provides methods for setting up configuration sources, reading values,
// and managing decode hooks for custom type conversions.
//
// All methods are safe for concurrent use.
type Viper interface {
	// SetRemoteProvider sets the remote configuration provider (e.g., "etcd").
	SetRemoteProvider(provider string)

	// SetRemoteEndpoint sets the endpoint URL for the remote configuration provider.
	SetRemoteEndpoint(endpoint string)

	// SetRemotePath sets the path to the configuration in the remote provider.
	SetRemotePath(path string)

	// SetRemoteSecureKey sets the encryption key for secure remote connections.
	SetRemoteSecureKey(key string)

	// SetRemoteModel sets the model struct for unmarshalling remote configuration.
	SetRemoteModel(model interface{})

	// SetRemoteReloadFunc sets a callback function to be called when configuration is reloaded.
	SetRemoteReloadFunc(fct func())

	// SetHomeBaseName sets the base name for configuration file in home directory.
	// The actual file will be named ".<basename>" (e.g., ".myapp").
	SetHomeBaseName(base string)

	// SetEnvVarsPrefix sets the prefix for environment variables.
	// Environment variables will be read as PREFIX_KEY_NAME.
	SetEnvVarsPrefix(prefix string)

	// SetDefaultConfig sets a function that returns a default configuration reader.
	// This is used as fallback when no configuration file is found.
	SetDefaultConfig(fct func() io.Reader)

	// SetConfigFile sets the path to the configuration file.
	// If empty, it will search for config in home directory using SetHomeBaseName.
	SetConfigFile(fileConfig string) error

	// Config initializes the configuration from file or remote provider.
	// logLevelRemoteKO is used for remote errors, logLevelRemoteOK for success messages.
	Config(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) error

	// Viper returns the underlying spf13/viper instance for advanced operations.
	Viper() *spfvpr.Viper

	// WatchFS starts watching the configuration file for changes.
	// When changes are detected, the reload function set by SetRemoteReloadFunc is called.
	WatchFS(logLevelFSInfo loglvl.Level)

	// Unset removes one or more configuration keys.
	// Supports nested keys using dot notation (e.g., "database.host").
	Unset(key ...string) error

	// HookRegister registers a custom decode hook for type conversion during unmarshalling.
	HookRegister(hook libmap.DecodeHookFunc)

	// HookReset removes all registered decode hooks.
	HookReset()

	// UnmarshalKey unmarshals a specific configuration key into the provided struct.
	UnmarshalKey(key string, rawVal interface{}) error

	// Unmarshal unmarshals the entire configuration into the provided struct.
	Unmarshal(rawVal interface{}) error

	// UnmarshalExact is like Unmarshal but returns an error if the config contains
	// fields that are not present in the target struct.
	UnmarshalExact(rawVal interface{}) error

	// GetBool returns the value associated with the key as a boolean.
	GetBool(key string) bool

	// GetString returns the value associated with the key as a string.
	GetString(key string) string

	// GetInt returns the value associated with the key as an integer.
	GetInt(key string) int

	// GetInt32 returns the value associated with the key as an int32.
	GetInt32(key string) int32

	// GetInt64 returns the value associated with the key as an int64.
	GetInt64(key string) int64

	// GetUint returns the value associated with the key as an unsigned integer.
	GetUint(key string) uint

	// GetUint16 returns the value associated with the key as a uint16.
	GetUint16(key string) uint16

	// GetUint32 returns the value associated with the key as a uint32.
	GetUint32(key string) uint32

	// GetUint64 returns the value associated with the key as a uint64.
	GetUint64(key string) uint64

	// GetFloat64 returns the value associated with the key as a float64.
	GetFloat64(key string) float64

	// GetTime returns the value associated with the key as time.Time.
	GetTime(key string) time.Time

	// GetDuration returns the value associated with the key as a time.Duration.
	GetDuration(key string) time.Duration

	// GetIntSlice returns the value associated with the key as a slice of int.
	GetIntSlice(key string) []int

	// GetStringSlice returns the value associated with the key as a slice of strings.
	GetStringSlice(key string) []string

	// GetStringMap returns the value associated with the key as a map of interfaces.
	GetStringMap(key string) map[string]any

	// GetStringMapString returns the value associated with the key as a map of strings.
	GetStringMapString(key string) map[string]string

	// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
	GetStringMapStringSlice(key string) map[string][]string
}

// New creates a new Viper instance with the provided context and logger.
//
// If log is nil, a default logger will be created using the provided context.
// The returned Viper instance is ready to use and thread-safe.
//
// Example:
//
//	ctx := func() context.Context { return context.Background() }
//	log := func() logger.Logger { return logger.New(ctx) }
//	v := viper.New(ctx, log)
func New(ctx context.Context, log liblog.FuncLog) Viper {
	if log == nil {
		l := liblog.New(ctx)
		log = func() liblog.Logger {
			return l
		}
	}
	v := &viper{
		v: spfvpr.New(),
		i: new(atomic.Uint32),
		l: log,
		h: libctx.New[uint8](ctx),
	}

	v.i.Store(0)

	return v
}
