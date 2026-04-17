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

package ticker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	liberr "github.com/nabbar/golib/errors"
	errpol "github.com/nabbar/golib/errors/pool"
	libsrv "github.com/nabbar/golib/runner"
)

// Ticker defines the public capabilities of the ticker runner.
// It is a composite interface that brings together:
//   - libsrv.Runner: Standard methods to manage the execution lifecycle (Start, Stop, Restart, etc.).
//   - liberr.Errors: Methods to inspect and manage errors collected during periodic executions.
//
// Any implementation of Ticker is guaranteed to be thread-safe for all its public methods.
type Ticker interface {
	libsrv.Runner
	liberr.Errors
}

// New initializes and returns a new Ticker instance.
//
// The Ticker will execute the provided 'fct' at every interval defined by 'tick'.
//
// Parameters:
//   - tick: The time.Duration between two consecutive executions of the ticker function.
//     If 'tick' is less than 1 millisecond, the function will default to 'defaultDuration' (30 seconds)
//     to prevent accidental high-frequency execution that could saturate CPU or resources.
//   - fct: The libsrv.FuncTicker to be executed. This function must accept a context.Context and
//     a libsrv.TickUpdate. If 'fct' is nil, New returns a ticker that will always record an
//     "invalid function ticker" error upon every tick.
//
// Returns:
//   - A Ticker interface implementation, specifically the internal 'run' structure.
//
// Thread-Safety:
//
//	The returned Ticker uses internal synchronization primitives (Mutex, Atomic values) to
//	allow safe usage from multiple goroutines.
func New(tick time.Duration, fct libsrv.FuncTicker) Ticker {
	// Security check for interval duration.
	// Very small intervals (< 1ms) are often a configuration mistake and can lead to performance issues.
	if tick < time.Millisecond {
		tick = defaultDuration
	}

	// Safety check for the ticker function.
	// If no function is provided, we use a placeholder that reports the configuration error.
	if fct == nil {
		fct = func(_ context.Context, _ libsrv.TickUpdate) error {
			return fmt.Errorf("invalid function ticker")
		}
	}

	// Construct the internal runner.
	// We use specialized atomic types from 'github.com/nabbar/golib/atomic' for better type safety.
	return &run{
		m: sync.Mutex{},
		e: errpol.New(),
		t: libatm.NewValue[time.Time](),          // Start time for uptime calculation.
		n: libatm.NewValue[context.CancelFunc](), // Storage for the current execution context's cancel function.
		s: new(atomic.Uint32),                    // State storage for the FSM.

		f: fct,
		d: tick,
		k: time.NewTicker(tick),
		w: time.NewTicker(pollChange), // Internal ticker for state change monitoring.
	}
}
