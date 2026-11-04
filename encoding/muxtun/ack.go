/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package muxtun

import (
	"context"
	"sync"
	"time"
)

const minValidId uint64 = 1 << 32

type ack struct {
	m sync.Mutex
	r map[uint64]int64 // map of packet id => time unix second
	s func()           // func to force RESYNC
}

func (o *ack) clear() {
	o.m.Lock()
	defer o.m.Unlock()

	o.r = make(map[uint64]int64, len(o.r))
}

func (o *ack) add(p *packet) {
	o.m.Lock()
	defer o.m.Unlock()

	o.r[p.id()] = time.Now().Unix()
}

func (o *ack) del(id ...uint64) {
	o.m.Lock()
	defer o.m.Unlock()

	for i := range id {
		delete(o.r, id[i])
	}
}

func (o *ack) ticker(x context.Context, d time.Duration) {
	var m = int64(2 * time.Second)

	for x.Err() == nil {
		select {
		case <-x.Done():
			return
		case <-time.After(d):
			if o.timeout(time.Now().Unix(), m) {
				o.s()
			}
		}
	}
}

func (o *ack) timeout(n, d int64) bool {
	o.m.Lock()
	defer o.m.Unlock()

	for id, ts := range o.r {
		if id < minValidId || ts < 1 {
			delete(o.r, id)
			continue
		}

		if n-ts > d {
			return true
		}
	}

	return false
}
