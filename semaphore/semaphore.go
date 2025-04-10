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

import semtps "github.com/nabbar/golib/semaphore/types"

func (o *sem) NewWorker() error {
	return o.s.NewWorker()
}

func (o *sem) NewWorkerTry() bool {
	return o.s.NewWorkerTry()
}

func (o *sem) DeferWorker() {
	o.s.DeferWorker()
}

func (o *sem) DeferMain() {
	if o.isMbp() {
		o.m.Shutdown()
	}

	o.s.DeferMain()
}

func (o *sem) WaitAll() error {
	return o.s.WaitAll()
}

func (o *sem) Weighted() int64 {
	return o.s.Weighted()
}

func (o *sem) Clone() Semaphore {
	return &sem{
		s: o.s.New(),
		m: o.m,
	}
}

func (o *sem) New() semtps.Sem {
	return o.s.New()
}
