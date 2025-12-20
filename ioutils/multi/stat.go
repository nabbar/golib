/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package multi

import "math"

// Stats contains performance statistics for adaptive writer mode.
// These metrics are used to monitor and debug adaptive mode behavior.
type Stats struct {
	// AdaptiveMode indicates whether adaptive mode is enabled.
	// false = sticky mode (sequential or parallel), true = adaptive mode
	AdaptiveMode bool

	// WriterMode indicates the current write strategy.
	// false = sequential writes, true = parallel writes
	WriterMode bool

	// AverageLatency is the mean write latency in nanoseconds over the sample window.
	// This value is used to determine mode switching in adaptive mode.
	AverageLatency int64

	// SampleCollected is the number of write operations sampled in the current window.
	SampleCollected int64

	// WriterCounter is the current number of registered writers.
	WriterCounter int
}

// Stats returns current performance statistics for the Multi instance.
// This includes adaptive mode status, current write mode, latency metrics,
// and writer count. Returns zero values if no samples have been collected yet.
func (o *mlt) Stats() Stats {
	var w int
	if i := o.lst.Load(); i > math.MaxInt {
		w = math.MaxInt
	} else {
		w = int(i)
	}

	if i := o.d.Load(); i != nil {
		s := i.Smp()
		return Stats{
			AdaptiveMode:    o.adp.Load(),
			WriterMode:      o.par.Load(),
			AverageLatency:  i.Lat() / s,
			SampleCollected: s,
			WriterCounter:   w,
		}
	} else {
		return Stats{
			AdaptiveMode:    o.adp.Load(),
			WriterMode:      o.par.Load(),
			AverageLatency:  0,
			SampleCollected: 0,
			WriterCounter:   w,
		}
	}
}
