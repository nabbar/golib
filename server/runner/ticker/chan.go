/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package ticker

var closeChan = make(chan struct{})

func init() {
	close(closeChan)
}

func (o *run) IsRunning() bool {
	if e := o.checkMe(); e != nil {
		return false
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.c == nil || o.c == closeChan {
		return false
	}

	select {
	case <-o.c:
		return false
	default:
		return true
	}
}

func (o *run) chanInit() {
	if e := o.checkMe(); e != nil {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	o.c = make(chan struct{})
}

func (o *run) chanDone() <-chan struct{} {
	if e := o.checkMe(); e != nil {
		return nil
	}

	o.m.RLock()
	defer o.m.RUnlock()

	return o.c
}

func (o *run) chanSend() {
	if e := o.checkMe(); e != nil {
		return
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.c != nil && o.c != closeChan {
		o.c <- struct{}{}
	}
}

func (o *run) chanClose() {
	if e := o.checkMe(); e != nil {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	if o.c != nil && o.c != closeChan {
		o.c = closeChan
	}
}
