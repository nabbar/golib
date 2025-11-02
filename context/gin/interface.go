/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

// Package gin provides a wrapper around Gin context that integrates with Go's context.Context.
// It offers type-safe value getters, signal handling, logger integration, and request-scoped storage.
//
// The GinTonic interface extends context.Context with Gin-specific functionality while maintaining
// full compatibility with the standard library.
//
// Features:
//   - Full context.Context compatibility
//   - Direct access to underlying gin.Context
//   - Type-safe getters for common types (string, int, bool, time.Time, etc.)
//   - Signal-based context cancellation (SIGTERM, SIGINT)
//   - Logger integration support
//   - Zero additional overhead
//
// Example usage:
//
//	import (
//	    "github.com/gin-gonic/gin"
//	    ginlib "github.com/nabbar/golib/context/gin"
//	)
//
//	func handler(c *gin.Context) {
//	    gtx := ginlib.New(c, nil)
//	    gtx.Set("user_id", "12345")
//	    userID := gtx.GetString("user_id")
//	    c.JSON(200, gin.H{"user_id": userID})
//	}
package gin

import (
	"context"
	"os"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
)

type GinTonic interface {
	context.Context

	// GinContext returns the underlying Gin context.
	//
	// This method is useful for extracting the original Gin context from a GinTonic context.
	// For example, when you want to use the Gin context in a middleware.
	//
	// Returns an instance of *ginsdk.Context.
	GinContext() *ginsdk.Context
	// CancelOnSignal will cancel the context when one of the provided signals is received.
	//
	// This method is useful for canceling the context when a signal is received.
	// For example, when you want to cancel the context when a SIGINT signal is received.
	//
	// The provided signals are monitored concurrently and the context will be canceled
	// when one of the signals is received.
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// ctx.CancelOnSignal(os.Interrupt, os.Kill)
	//
	// Parameters:
	//   s ...os.Signal - signals to monitor
	CancelOnSignal(s ...os.Signal)

	// Set sets the value for the given key.
	//
	// It is a part of the GinTonic interface.
	//
	// Parameters:
	//   key any - key to store the value
	//   value any - value to store
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// ctx.Set("key", "value")
	Set(key any, value any)
	// Get returns the value for the given key if it exists in the context.
	//
	// If the key does not exist, Get returns (nil, false).
	//
	// Parameters:
	//   key any - key to retrieve the value
	//
	// Returns:
	//   value any - value associated with the given key
	//   exists bool - whether the key exists in the context
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value, exists := ctx.Get("key")
	// if exists {
	//     // Do something with the value
	// }
	Get(key any) (value any, exists bool)
	// MustGet is a helper function which returns the value for the given key if it exists in the context.
	// If the key does not exist, it panics with a message indicating that the key does not exist in the context.
	//
	// Parameters:
	//   key any - key to retrieve the value
	//
	// Returns:
	//   value any - value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.MustGet("key")
	// if value != nil {
	//     // Do something with the value
	// }
	MustGet(key any) any
	// GetString returns the value associated with the given key as a string.
	// If the key does not exist, GetString returns an empty string.
	//
	// Parameters:
	//   key any - key to retrieve the string value
	//
	// Returns:
	//   s string - string value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetString("key")
	// if value != "" {
	//     // Do something with the value
	// }
	GetString(key any) (s string)
	// GetBool returns the value associated with the given key as a boolean.
	// If the key does not exist, GetBool returns false.
	//
	// Parameters:
	//   key any - key to retrieve the boolean value
	//
	// Returns:
	//   b bool - boolean value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetBool("key")
	// if value {
	//     // Do something with the value
	// }
	GetBool(key any) (b bool)
	// GetInt returns the value associated with the given key as an integer.
	//
	// If the key does not exist, GetInt returns 0.
	//
	// Parameters:
	//   key any - key to retrieve the integer value
	//
	// Returns:
	//   i int - integer value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetInt("key")
	// if value != 0 {
	//     // Do something with the value
	// }
	GetInt(key any) (i int)
	// GetInt64 returns the value associated with the given key as an int64.
	// If the key does not exist, GetInt64 returns 0.
	//
	// Parameters:
	//   key any - key to retrieve the int64 value
	//
	// Returns:
	//   i64 int64 - int64 value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetInt64("key")
	// if value != 0 {
	//     // Do something with the value
	// }
	GetInt64(key any) (i64 int64)
	// GetFloat64 returns the value associated with the given key as a float64.
	//
	// If the key does not exist, GetFloat64 returns 0.
	//
	// Parameters:
	//   key any - key to retrieve the float64 value
	//
	// Returns:
	//   f64 float64 - float64 value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetFloat64("key")
	// if value != 0 {
	//     // Do something with the value
	// }
	GetFloat64(key any) (f64 float64)
	// GetTime returns the value associated with the given key as a time.Time.
	//
	// If the key does not exist, GetTime returns the zero time.
	//
	// Parameters:
	//   key any - key to retrieve the time.Time value
	//
	// Returns:
	//   t time.Time - time.Time value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetTime("key")
	// if !value.IsZero() {
	//     // Do something with the value
	// }
	GetTime(key any) (t time.Time)
	// GetDuration returns the value associated with the given key as a time.Duration.
	//
	// If the key does not exist, GetDuration returns the zero duration.
	//
	// Parameters:
	//   key any - key to retrieve the time.Duration value
	//
	// Returns:
	//   d time.Duration - time.Duration value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetDuration("key")
	// if value != 0 {
	//     // Do something with the value
	// }
	GetDuration(key any) (d time.Duration)
	// GetStringSlice returns the value associated with the given key as a string slice.
	//
	// If the key does not exist, GetStringSlice returns an empty string slice.
	//
	// Parameters:
	//   key any - key to retrieve the string slice value
	//
	// Returns:
	//   ss []string - string slice value associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetStringSlice("key")
	// if len(value) > 0 {
	//     // Do something with the value
	// }
	GetStringSlice(key any) (ss []string)
	// GetStringMap returns the value associated with the given key as a map of string keys to any values.
	//
	// If the key does not exist, GetStringMap returns an empty map.
	//
	// Parameters:
	//   key any - key to retrieve the map value
	//
	// Returns:
	//   sm map[string]any - map of string keys to any values associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetStringMap("key")
	// if len(value) > 0 {
	//     // Do something with the value
	// }
	GetStringMap(key any) (sm map[string]any)
	// GetStringMapString returns the value associated with the given key as a map of string keys to string values.
	//
	// If the key does not exist, GetStringMapString returns an empty map.
	//
	// Parameters:
	//   key any - key to retrieve the map of string keys to string values
	//
	// Returns:
	//   sms map[string]string - map of string keys to string values associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = NewGinTonic(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetStringMapString("key")
	// if len(value) > 0 {
	//     // Do something with the value
	// }
	GetStringMapString(key any) (sms map[string]string)
	// GetStringMapStringSlice returns the value associated with the given key as a map of string keys to slices of string values.
	//
	// If the key does not exist, GetStringMapStringSlice returns an empty map.
	//
	// Parameters:
	//   key any - key to retrieve the map of string keys to slices of string values
	//
	// Returns:
	//   smss map[string][]string - map of string keys to slices of string values associated with the given key
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = New(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// value := ctx.GetStringMapStringSlice("key")
	// if len(value) > 0 {
	//     // Do something with the value
	// }
	GetStringMapStringSlice(key any) (smss map[string][]string)

	// SetLogger sets the logger function for the GinTonic context.
	//
	// The logger function is used to log messages at different log levels.
	//
	// Parameters:
	//   log liblog.FuncLog - logger function to set
	//
	// Example:
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx = New(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	// ctx.SetLogger(log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
	SetLogger(log liblog.FuncLog)
}

// New returns a new instance of GinTonic.
//
// Parameters:
//
//	c *ginsdk.Context - underlying Gin context
//	log liblog.FuncLog - logger function for the GinTonic context
//
// Returns:
//
//	*ctxGinTonic - new instance of GinTonic
//
// Example:
// ctx, cancel := context.WithCancel(context.Background())
// defer cancel()
// ctx = New(ctx, log.NewLogFunc(os.Stderr, loglvl.InfoLevel))
//
// If the underlying Gin context is nil, a new Gin context is created.
// The logger function is used to log messages at different log levels.
// The underlying Gin context is wrapped in a context which is canceled
// when the context is canceled.
func New(c *ginsdk.Context, log liblog.FuncLog) GinTonic {
	if c == nil {
		c = &ginsdk.Context{
			Request:  nil,
			Writer:   nil,
			Params:   make(ginsdk.Params, 0),
			Keys:     make(map[any]any),
			Errors:   make([]*ginsdk.Error, 0),
			Accepted: make([]string, 0),
		}
	}

	var (
		x context.Context
		l context.CancelFunc
	)

	if c.Request != nil && c.Request.Context() != nil {
		x, l = context.WithCancel(c.Request.Context())
	} else {
		x, l = context.WithCancel(c)
	}

	return &ctxGinTonic{
		l: log,
		g: c,
		x: x,
		c: l,
	}
}
