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
 */

package semaphore

import (
	"context"
	"sync/atomic"

	liberr "github.com/nabbar/golib/errors"
	"golang.org/x/sync/semaphore"
)

type sem struct {
	d *atomic.Value // compatibility with SemBar => isCompleted
	i int64         // max process simultaneous
	s *semaphore.Weighted
	x context.Context
	c context.CancelFunc
}

func (s *sem) getCompleted() bool {
	if s.d == nil {
		s.d = new(atomic.Value)
	}

	if i := s.d.Load(); i == nil {
		return false
	} else if o, ok := i.(bool); !ok {
		return false
	} else {
		return o
	}
}

func (s *sem) setCompleted(flag bool) {
	if s.d == nil {
		s.d = new(atomic.Value)
	}

	s.d.Store(flag)
}

func (s *sem) NewWorker() liberr.Error {
	e := s.s.Acquire(s.context(), 1)
	return ErrorWorkerNew.IfError(e)
}

func (s *sem) NewWorkerTry() bool {
	return s.s.TryAcquire(1)
}

func (s *sem) WaitAll() liberr.Error {
	e := s.s.Acquire(s.context(), s.i)
	s.setCompleted(true)
	return ErrorWorkerWaitAll.IfError(e)
}

func (s *sem) DeferWorker() {
	s.s.Release(1)
}

func (s *sem) DeferMain() {
	s.setCompleted(true)
	if s.c != nil {
		s.c()
	}
}

func (s *sem) context() context.Context {
	if s.x == nil {
		if s.c != nil {
			s.c()
		}
		s.x, s.c = NewContext(context.Background(), 0, EmptyTime())
	}
	return s.x
}

func (s *sem) Current() int64 {
	return -1
}

func (s *sem) Completed() bool {
	return s.getCompleted()
}

func (s *sem) Reset(total, current int64) {
	return
}

func (s *sem) ResetDefined(current int64) {
	return
}

func (s *sem) Done() {
	return
}

func (s *sem) Increment(n int) {
	return
}

func (s *sem) Increment64(n int64) {
	return
}
