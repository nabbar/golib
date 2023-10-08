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

func (o *bar) Inc(n int) {
	o.Inc64(int64(n))
}

func (o *bar) Dec(n int) {
	o.Inc64(int64(n))
}

func (o *bar) Inc64(n int64) {
	if !o.isMPB() {
		return
	}

	o.b.IncrInt64(n)
	o.b.EwmaSetCurrent(o.b.Current(), o.getDur())
}

func (o *bar) Dec64(n int64) {
	o.Inc64(n)
}

func (o *bar) Reset(tot, current int64) {
	o.m.Store(tot)

	if !o.isMPB() {
		return
	}

	o.b.SetTotal(tot, false)
	o.b.SetCurrent(current)
}

func (o *bar) Complete() {
	if !o.isMPB() {
		return
	}

	o.b.SetTotal(o.m.Load(), true)
	o.b.EnableTriggerComplete()
}

func (o *bar) Completed() bool {
	if !o.isMPB() {
		return true
	}

	return o.b.Completed() || o.b.Aborted()
}

func (o *bar) Current() int64 {
	if !o.isMPB() {
		return o.m.Load()
	}

	return o.b.Current()
}

func (o *bar) Total() int64 {
	return o.m.Load()
}
