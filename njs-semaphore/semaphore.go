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

package njs_semaphore

import (
	"context"
	"runtime"
	"time"

	"golang.org/x/sync/semaphore"

	. "github.com/nabbar/golib/njs-errors"
)

type sem struct {
	m int64
	s *semaphore.Weighted
	x context.Context
	c context.CancelFunc
}

type Sem interface {
	NewWorker() Error
	NewWorkerTry() bool
	DeferWorker()
	DeferMain()

	WaitAll() Error
	Context() context.Context
	Cancel()
}

func GetMaxSimultaneous() int {
	return runtime.GOMAXPROCS(0)
}

func NewSemaphore(maxSimultaneous int, timeout time.Duration, deadline time.Time, parent context.Context) Sem {
	if maxSimultaneous < 1 {
		maxSimultaneous = GetMaxSimultaneous()
	}

	x, c := GetContext(timeout, deadline, parent)

	return &sem{
		m: int64(maxSimultaneous),
		s: semaphore.NewWeighted(int64(maxSimultaneous)),
		x: x,
		c: c,
	}
}

func (s *sem) NewWorker() Error {
	e := s.s.Acquire(s.x, 1)
	return NEW_WORKER.Iferror(e)
}

func (s *sem) NewWorkerTry() bool {
	return s.s.TryAcquire(1)
}

func (s *sem) WaitAll() Error {
	e := s.s.Acquire(s.Context(), s.m)
	return WAITALL.Iferror(e)
}

func (s *sem) DeferWorker() {
	s.s.Release(1)
}

func (s *sem) DeferMain() {
	s.Cancel()
}

func (s *sem) Cancel() {
	if s.c != nil {
		s.c()
	}
}

func (s *sem) Context() context.Context {
	if s.x == nil {
		s.x, s.c = GetContext(0, GetEmptyTime(), nil)
	}

	return s.x
}
