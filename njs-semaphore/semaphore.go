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
)

type sem struct {
	m int64
	s *semaphore.Weighted
	x context.Context
	c context.CancelFunc
}

type Sem interface {
	NewWorker() error
	NewWorkerTry() bool
	DeferWorker()
	DeferMain()

	WaitAll() error
	Context() context.Context
	Cancel()
}

func GetMaxSimultaneous() int {
	return runtime.GOMAXPROCS(0)
}

func NewSemaphore(maxSimultaneous int, timeout time.Duration, deadline time.Time) Sem {
	if maxSimultaneous < 1 {
		maxSimultaneous = GetMaxSimultaneous()
	}

	var (
		c context.CancelFunc
		x context.Context
	)

	if timeout > 0 {
		x, c = context.WithTimeout(context.Background(), timeout)
	} else if deadline.Unix() > 0 {
		x, c = context.WithDeadline(context.Background(), deadline)
	} else {
		x, c = context.WithCancel(context.Background())
	}

	return &sem{
		m: int64(maxSimultaneous),
		s: semaphore.NewWeighted(int64(maxSimultaneous)),
		x: x,
		c: c,
	}
}

func (s *sem) NewWorker() error {
	if e := s.s.Acquire(s.x, 1); e != nil {
		return e
	}

	return nil
}

func (s *sem) NewWorkerTry() bool {
	return s.s.TryAcquire(1)
}

func (s *sem) WaitAll() error {
	return s.s.Acquire(s.Context(), s.m)
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
	}

	return s.x
}