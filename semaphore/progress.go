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

import (
	sdkmpb "github.com/vbauerster/mpb/v8"
	mpbdec "github.com/vbauerster/mpb/v8/decor"

	sembar "github.com/nabbar/golib/semaphore/bar"
	semnbr "github.com/nabbar/golib/semaphore/nobar"
	semtps "github.com/nabbar/golib/semaphore/types"
)

const done = "Done"

// isMbp returns true if MPB progress container is enabled.
func (o *sem) isMbp() bool {
	return o.m != nil
}

// defOpts creates default MPB bar options with decorators for name, job, and unit.
// This is an internal helper for creating consistent progress bar displays.
func (o *sem) defOpts(unit interface{}, name, job string, bar semtps.Bar) []sdkmpb.BarOption {
	var opt = make([]sdkmpb.BarOption, 0)

	if bar != nil {
		if b, k := bar.(semtps.BarMPB); k {
			opt = append(opt, sdkmpb.BarQueueAfter(b.GetMPB()))
		}
	}

	var dec = make([]mpbdec.Decorator, 0)

	if len(name) > 0 {
		dec = append(dec,
			mpbdec.Name(name, mpbdec.WC{W: len(name) + 1, C: mpbdec.DindentRight}),
		)
	}

	if len(job) > 0 {
		if len(dec) > 0 {
			dec = append(dec,
				mpbdec.Name(" | "),
			)
		}
		dec = append(dec,
			mpbdec.Name(job, mpbdec.WC{W: len(job) + 1, C: mpbdec.DindentRight | mpbdec.DextraSpace}),
		)
	}

	if unit != nil {
		if len(dec) > 0 {
			dec = append(dec,
				mpbdec.Name(" | "),
			)
		}
		dec = append(dec,
			mpbdec.Counters(unit, "", mpbdec.WCSyncWidth),
		)
	}

	opt = append(opt, sdkmpb.PrependDecorators(dec...))

	dec = append(make([]mpbdec.Decorator, 0),
		mpbdec.Percentage(mpbdec.WC{W: 5, C: 0}),
		mpbdec.Name(" | "),
		mpbdec.AverageETA(mpbdec.ET_STYLE_GO, mpbdec.WCSyncWidth),
	)

	if unit != nil {
		dec = append(dec,
			mpbdec.Name(" | "),
			mpbdec.AverageSpeed(unit, "% .2f", mpbdec.WCSyncWidth),
		)
	}

	return append(opt, sdkmpb.AppendDecorators(append(dec, mpbdec.OnComplete(mpbdec.Name(""), " | "+done))...))
}

// BarBytes creates a progress bar for tracking byte quantities.
// The bar displays sizes in human-readable format (KB, MB, GB, etc.).
//
// See: github.com/nabbar/golib/semaphore/types.Progress
func (o *sem) BarBytes(name, job string, tot int64, drop bool, bar semtps.SemBar) semtps.SemBar {
	return o.BarOpts(tot, drop, o.defOpts(mpbdec.SizeB1024(0), name, job, bar)...)
}

// BarTime creates a progress bar for time-based operations.
// Displays progress with time estimates (ETA, elapsed time).
//
// See: github.com/nabbar/golib/semaphore/types.Progress
func (o *sem) BarTime(name, job string, tot int64, drop bool, bar semtps.SemBar) semtps.SemBar {
	return o.BarOpts(tot, drop, o.defOpts(nil, name, job, bar)...)
}

// BarNumber creates a progress bar for tracking numeric quantities.
// Displays progress as a simple counter (e.g., "45/100").
//
// See: github.com/nabbar/golib/semaphore/types.Progress
func (o *sem) BarNumber(name, job string, tot int64, drop bool, bar semtps.SemBar) semtps.SemBar {
	return o.BarOpts(tot, drop, o.defOpts(int64(0), name, job, bar)...)
}

// BarOpts creates a progress bar with custom MPB options.
// Use this for full control over bar appearance and behavior.
//
// See: github.com/nabbar/golib/semaphore/types.Progress
func (o *sem) BarOpts(tot int64, drop bool, opts ...sdkmpb.BarOption) semtps.SemBar {
	if o.isMbp() {
		return sembar.New(o, tot, drop, opts...)
	}
	return semnbr.New(o, tot, drop, opts...)
}

// GetMPB returns the underlying MPB progress container.
// Returns nil if progress is disabled.
//
// See: github.com/nabbar/golib/semaphore/types.ProgressMPB
func (o *sem) GetMPB() *sdkmpb.Progress {
	return o.m
}
