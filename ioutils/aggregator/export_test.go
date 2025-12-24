/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
 */

package aggregator

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
	librun "github.com/nabbar/golib/runner/startStop"
)

// Helper functions exported for testing purposes (black-box testing support)

func InternalCtx(ctx1, ctx2 context.Context, cfg Config) (Aggregator, error) {
	a, e := New(ctx1, cfg)
	if e != nil {
		return nil, e
	} else if i, k := a.(*agg); k {
		if ctx2 == nil {
			i.x = libatm.NewValue[context.Context]()
		} else {
			i.x.Store(ctx2)
		}
		return i, nil
	}
	return nil, ErrInvalidInstance
}

func InternalRunner(a Aggregator, r librun.StartStop) {
	if i, k := a.(*agg); k {
		if r == nil {
			i.r = libatm.NewValue[librun.StartStop]()
		} else {
			i.setRunner(r)
		}
	}
}

func InternalOp(a Aggregator, v bool) {
	if i, k := a.(*agg); k {
		i.op.Store(v)
	}
}

func InternalGetRunner(a Aggregator) librun.StartStop {
	if i, k := a.(*agg); k {
		return i.getRunner()
	}
	return nil
}

func InternalGetOp(a Aggregator) bool {
	if i, k := a.(*agg); k {
		return i.op.Load()
	}
	return false
}
