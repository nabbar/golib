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

	semsem "github.com/nabbar/golib/semaphore/sem"
	semtps "github.com/nabbar/golib/semaphore/types"
	sdkmpb "github.com/vbauerster/mpb/v8"
)

type Semaphore interface {
	context.Context
	semtps.Sem
	semtps.Progress

	Clone() Semaphore
}

func MaxSimultaneous() int {
	return semsem.MaxSimultaneous()
}

func SetSimultaneous(n int) int64 {
	return semsem.SetSimultaneous(n)
}

func New(ctx context.Context, nbrSimultaneous int, progress bool, opt ...sdkmpb.ContainerOption) Semaphore {
	var (
		m *sdkmpb.Progress
	)

	if progress {
		m = sdkmpb.New(opt...)
	}

	return &sem{
		s: semsem.New(ctx, nbrSimultaneous),
		m: m,
	}
}
