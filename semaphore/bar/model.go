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

import (
	"sync/atomic"
	"time"

	semtps "github.com/nabbar/golib/semaphore/types"

	"github.com/vbauerster/mpb/v8"
)

// bar is the internal implementation of the SemBar interface.
// It wraps a semaphore with progress bar functionality.
type bar struct {
	s semtps.Sem // Underlying semaphore
	d bool       // Drop bar on complete flag

	b *mpb.Bar      // MPB progress bar (nil if progress disabled)
	m *atomic.Int64 // Total value (atomic for thread safety)
	t *atomic.Int64 // Last update timestamp in Unix nanoseconds (atomic for thread safety)
}

// isMPB returns true if MPB progress bar is enabled.
func (o *bar) isMPB() bool {
	return o.b != nil
}

// GetMPB returns the underlying MPB bar instance.
// Returns nil if progress bar is disabled.
func (o *bar) GetMPB() *mpb.Bar {
	return o.b
}

// getDur calculates the duration since the last update.
// Returns time.Millisecond as default if no previous timestamp exists.
// This is used for EWMA (Exponentially Weighted Moving Average) calculations.
func (o *bar) getDur() time.Duration {
	now := time.Now().UnixNano()
	prev := o.t.Swap(now)

	if prev == 0 {
		return time.Millisecond
	}

	dur := time.Duration(now - prev)
	if dur <= 0 {
		return time.Millisecond
	}

	return dur
}
