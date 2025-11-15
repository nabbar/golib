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

/*
Package viper provides a wrapper around spf13/viper for configuration management
with enhanced features including remote configuration support, custom decode hooks,
and file watching capabilities.

# Overview

This package extends the functionality of github.com/spf13/viper by providing:
  - Remote configuration support (ETCD)
  - Custom decode hooks for type conversion
  - File system watching with reload callbacks
  - Configuration key unsetting
  - Thread-safe operations

# Basic Usage

Create a new Viper instance and load configuration from a file:

	import (
		"context"
		"github.com/nabbar/golib/logger"
		"github.com/nabbar/golib/viper"
	)

	ctx := func() context.Context { return context.Background() }
	log := func() logger.Logger { return logger.New(ctx) }

	v := viper.New(ctx, log)
	v.SetConfigFile("/path/to/config.yaml")
	err := v.Config(logger.ErrorLevel, logger.InfoLevel)
	if err != nil {
		// Handle error
	}

	// Read values
	name := v.GetString("app.name")
	port := v.GetInt("app.port")

# Configuration Sources

The package supports multiple configuration sources:

1. File-based configuration (JSON, YAML, TOML, etc.)
2. Environment variables with prefix
3. Remote configuration providers (ETCD)
4. Default configuration via io.Reader

# Remote Configuration

Configure remote provider for dynamic configuration updates:

	v.SetRemoteProvider("etcd")
	v.SetRemoteEndpoint("http://localhost:2379")
	v.SetRemotePath("/config/myapp")
	v.SetRemoteModel(&MyConfig{})

	err := v.Config(logger.ErrorLevel, logger.InfoLevel)
	// Configuration will be automatically reloaded from remote

# Custom Decode Hooks

Register custom decode hooks for type conversion during unmarshalling:

	import "github.com/go-viper/mapstructure/v2"

	hook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// Custom conversion logic
		return data, nil
	}

	v.HookRegister(hook)

# File Watching

Watch for configuration file changes and trigger reload:

	v.SetRemoteReloadFunc(func() {
		// Reload application configuration
	})
	v.WatchFS(logger.InfoLevel)

# Unmarshalling

Unmarshal configuration into structs:

	type AppConfig struct {
		Name    string
		Port    int
		Debug   bool
	}

	var config AppConfig
	err := v.Unmarshal(&config)

	// Or unmarshal a specific key
	var dbConfig DatabaseConfig
	err := v.UnmarshalKey("database", &dbConfig)

# Configuration Management

Unset configuration keys dynamically:

	// Unset single key
	err := v.Unset("app.debug")

	// Unset multiple keys
	err := v.Unset("app.debug", "app.verbose", "cache.enabled")

	// Unset nested keys
	err := v.Unset("database.connection.pool")

# Error Handling

The package defines specific error codes for different failure scenarios:

  - ErrorParamEmpty: Required parameter is empty
  - ErrorParamMissing: Required parameter is missing
  - ErrorHomePathNotFound: Cannot retrieve user home path
  - ErrorBasePathNotFound: Cannot retrieve base config path
  - ErrorRemoteProvider: Cannot define remote provider
  - ErrorRemoteProviderSecure: Cannot define secure remote provider
  - ErrorRemoteProviderRead: Cannot read config from remote provider
  - ErrorRemoteProviderMarshall: Cannot marshall config from remote provider
  - ErrorConfigRead: Cannot read config from file
  - ErrorConfigReadDefault: Cannot read default config
  - ErrorConfigIsDefault: Using default config (warning)

# Thread Safety

All operations are thread-safe and can be called concurrently from multiple goroutines.
The underlying viper instance and custom hooks are properly synchronized.

# Dependencies

This package depends on:
  - github.com/spf13/viper: Core configuration management
  - github.com/go-viper/mapstructure/v2: Struct decoding with hooks
  - github.com/nabbar/golib/context: Context management
  - github.com/nabbar/golib/logger: Logging functionality
  - github.com/nabbar/golib/errors: Error handling
  - github.com/fsnotify/fsnotify: File system notifications

# See Also

For more information on the underlying viper library:
https://github.com/spf13/viper

For mapstructure decode hooks:
https://github.com/go-viper/mapstructure
*/
package viper
