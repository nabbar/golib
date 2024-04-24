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

package sem

import (
	"context"
	"runtime"
	"sync"

	semtps "github.com/nabbar/golib/semaphore/types"
	goxsem "golang.org/x/sync/semaphore"
)

func MaxSimultaneous() int {
	return runtime.GOMAXPROCS(0)
}

func SetSimultaneous(n int) int64 {
	m := MaxSimultaneous()
	if n < 1 {
		return int64(m)
	} else if m < n {
		return int64(m)
	} else {
		return int64(n)
	}
}

func New(ctx context.Context, nbrSimultaneous int) semtps.Sem {
	var (
		x context.Context
		n context.CancelFunc
	)

	x, n = context.WithCancel(ctx)

	if nbrSimultaneous == 0 {
		b := SetSimultaneous(nbrSimultaneous)
		return &sem{
			c: n,
			x: x,
			s: goxsem.NewWeighted(b),
			n: b,
		}
	} else if nbrSimultaneous > 0 {
		return &sem{
			c: n,
			x: x,
			s: goxsem.NewWeighted(int64(nbrSimultaneous)),
			n: int64(nbrSimultaneous),
		}
	} else {
		return &wg{
			c: n,
			x: x,
			w: sync.WaitGroup{},
		}
	}
}
