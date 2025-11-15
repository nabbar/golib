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

package monitor

import (
	"context"
	"fmt"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner/ticker"
)

// Monitor is the interface that wraps all monitor functionalities.
// It extends montps.Monitor and provides health check monitoring capabilities.
type Monitor interface {
	montps.Monitor
}

// New creates a new Monitor instance with the given context provider and info.
// The ctx parameter provides a function that returns the current context.
// The info parameter provides metadata about the monitored component.
// Returns an error if info is nil.
//
// Example:
//
//	inf, _ := info.New("my-service")
//	mon, err := monitor.New(context.Background, inf)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(ctx context.Context, info montps.Info) (montps.Monitor, error) {
	if info == nil {
		return nil, fmt.Errorf("info cannot be nil")
	} else if ctx == nil {
		ctx = context.Background()
	}

	m := &mon{
		x: libctx.New[string](ctx),
		i: libatm.NewValue[montps.Info](),
		r: libatm.NewValue[librun.Ticker](),
	}

	m.i.Store(info)
	_ = m.Stop(ctx)

	return m, nil
}
