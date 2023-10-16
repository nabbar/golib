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

package semaphore

import (
	"context"
	"runtime"

	semtps "github.com/nabbar/golib/semaphore/types"
	"github.com/vbauerster/mpb/v8"
	"golang.org/x/sync/semaphore"
)

type Semaphore interface {
	context.Context
	semtps.Sem
	semtps.Progress
}

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

func New(ctx context.Context, nbrSimultaneous int, progress bool, opt ...mpb.ContainerOption) Semaphore {
	nbr := SetSimultaneous(nbrSimultaneous)
	ctx, cnl := context.WithCancel(ctx)

	var m *mpb.Progress

	if progress {
		m = mpb.New(opt...)
	}

	return &sem{
		c: cnl,
		x: ctx,
		s: semaphore.NewWeighted(nbr),
		n: nbr,
		m: m,
	}
}
