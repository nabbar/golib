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

// Inc increments the progress bar by n.
func (o *bar) Inc(n int) {
	o.Inc64(int64(n))
}

// Dec decrements the progress bar by n.
// Note: This delegates to Inc64 with a negative value for proper decrement behavior.
func (o *bar) Dec(n int) {
	o.Dec64(int64(n))
}

// Inc64 increments the progress bar by n (64-bit version).
func (o *bar) Inc64(n int64) {
	if !o.isMPB() {
		return
	}

	o.b.IncrInt64(n)
	o.b.EwmaSetCurrent(o.b.Current(), o.getDur())
}

// Dec64 decrements the progress bar by n (64-bit version).
// This is implemented by incrementing with a negative value.
func (o *bar) Dec64(n int64) {
	o.Inc64(-n)
}

// Reset resets the progress bar with new total and current values.
// This updates both the internal total counter and the MPB bar if present.
func (o *bar) Reset(tot, current int64) {
	o.m.Store(tot)

	if !o.isMPB() {
		return
	}

	o.b.SetTotal(tot, false)
	o.b.SetCurrent(current)
}

// Complete marks the progress bar as complete.
// If MPB is enabled, this triggers the completion animation.
func (o *bar) Complete() {
	if !o.isMPB() {
		return
	}

	o.b.SetTotal(o.m.Load(), true)
	o.b.EnableTriggerComplete()
}

// Completed returns true if the progress bar is completed or aborted.
// Without MPB, this always returns true.
func (o *bar) Completed() bool {
	if !o.isMPB() {
		return true
	}

	return o.b.Completed() || o.b.Aborted()
}

// Current returns the current progress value.
// Without MPB, this returns the total value.
func (o *bar) Current() int64 {
	if !o.isMPB() {
		return o.m.Load()
	}

	return o.b.Current()
}

// Total returns the total/maximum value of the progress bar.
func (o *bar) Total() int64 {
	return o.m.Load()
}
