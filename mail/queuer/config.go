/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package queuer

import (
	"time"
)

// FuncCaller is a callback function type invoked when throttle events occur.
//
// This function is called in two scenarios:
//  1. When the rate limit is reached and the pooler needs to wait
//  2. When Reset() is called to reset the throttle counter
//
// The function can be used for:
//   - Logging throttle events
//   - Updating metrics/monitoring
//   - Implementing custom backoff strategies
//   - Injecting errors for testing purposes
//
// If the function returns an error, the throttle operation will fail and
// propagate the error to the caller. Returning nil allows the operation
// to proceed normally.
//
// Note: This function is called while holding an internal mutex, so it
// should be kept lightweight and avoid blocking operations that could
// cause deadlocks.
//
// Example:
//
//	cfg.SetFuncCaller(func() error {
//	    metrics.IncrementThrottleCounter()
//	    log.Info("Rate limit reached, throttling...")
//	    return nil
//	})
type FuncCaller func() error

// Config holds the configuration for rate limiting behavior.
//
// The rate limiter uses a sliding window approach where:
//   - Max specifies the maximum number of operations allowed
//   - Wait specifies the time window duration
//   - Operations beyond Max within Wait duration will be throttled
//
// Thread safety: Config instances should not be modified after being passed
// to New(). Create a new Config if different settings are needed.
type Config struct {
	// Max is the maximum number of emails that can be sent within the Wait duration.
	//
	// Special values:
	//   - 0 or negative: Throttling is disabled, unlimited sending
	//   - Positive: Enforces rate limit of Max emails per Wait period
	//
	// When the limit is reached, the pooler will sleep until the next time
	// window begins before allowing more emails to be sent.
	//
	// Example: Max=100 with Wait=1*time.Minute allows 100 emails per minute.
	Max int `json:"max" yaml:"max" toml:"max" mapstructure:"max"`

	// Wait is the time duration for the rate limiting window.
	//
	// Special values:
	//   - 0 or negative: Throttling is disabled, unlimited sending
	//   - Positive: Defines the time window for rate limiting
	//
	// The time window is measured from the first email sent in each period.
	// Once the window expires, a new window begins with a fresh counter.
	//
	// Example: Wait=5*time.Second with Max=10 allows 10 emails per 5 seconds.
	Wait time.Duration `json:"wait" yaml:"wait" toml:"wait" mapstructure:"wait"`

	// _fct is an optional callback function called during throttle events.
	// Use SetFuncCaller to set this field.
	_fct FuncCaller
}

// SetFuncCaller sets an optional callback function for throttle events.
//
// The provided function will be called when:
//  1. The rate limit is reached and throttling occurs
//  2. Reset() is called (if throttling is enabled)
//
// Parameters:
//   - fct: Callback function to invoke on throttle events. Can be nil to
//     disable callbacks.
//
// The callback is useful for monitoring, logging, or implementing custom
// behavior during throttling. If the callback returns an error, it will
// stop the throttling operation and return the error to the caller.
//
// Example with logging:
//
//	cfg.SetFuncCaller(func() error {
//	    log.Printf("Throttle event at %v", time.Now())
//	    return nil
//	})
//
// Example with error injection for testing:
//
//	cfg.SetFuncCaller(func() error {
//	    if testCondition {
//	        return errors.New("throttle error")
//	    }
//	    return nil
//	})
//
// Note: This method is not thread-safe. It should only be called during
// initialization before the Config is passed to New().
func (c *Config) SetFuncCaller(fct FuncCaller) {
	c._fct = fct
}
