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

package bar

import semtps "github.com/nabbar/golib/semaphore/types"

func (o *bar) NewWorker() error {
	return o.s.NewWorker()
}

func (o *bar) NewWorkerTry() bool {
	return o.s.NewWorkerTry()
}

func (o *bar) DeferWorker() {
	o.Inc(1)
	o.s.DeferWorker()
}

func (o *bar) DeferMain() {
	if o.isMPB() {
		o.Complete()
		if !o.Completed() {
			o.b.Abort(o.d)
		}
	}

	o.s.DeferMain()
}

func (o *bar) WaitAll() error {
	return o.s.WaitAll()
}

func (o *bar) Weighted() int64 {
	return o.s.Weighted()
}

func (o *bar) New() semtps.Sem {
	return o.s.New()
}
