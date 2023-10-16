//go:build examples
// +build examples

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

package tools

import (
	"fmt"
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func MakeMPB(w io.Writer) *mpb.Progress {
	if w != nil {
		return mpb.New(
			mpb.WithWidth(80),
			mpb.WithOutput(w),
			mpb.WithRefreshRate(10*time.Millisecond),
		)
	} else {
		return mpb.New(
			mpb.WithWidth(80),
			mpb.WithRefreshRate(10*time.Millisecond),
		)
	}
}

func MakeBar(p *mpb.Progress, bar *mpb.Bar, name string) *mpb.Bar {
	tit := "Complex Bar"
	msg := "Done"

	pct := []decor.Decorator{
		decor.Percentage(decor.WC{W: 5, C: 0}),
		decor.Name(" | "),
		decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60),
	}

	pr := []decor.Decorator{
		decor.Name(tit, decor.WC{W: len(tit) + 1, C: decor.DidentRight}),
		decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight | decor.DextraSpace}),
		decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WC{W: 20, C: decor.DextraSpace}),
		decor.Name("  ", decor.WC{W: 3, C: decor.DidentRight | decor.DextraSpace}),
		decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: len(msg) + 1, C: 0}), msg),
	}

	if bar == nil {
		return p.AddBar(0,
			mpb.PrependDecorators(pr...),
			mpb.AppendDecorators(pct...),
		)
	} else {
		return p.AddBar(0,
			mpb.BarQueueAfter(bar),
			mpb.PrependDecorators(pr...),
			mpb.AppendDecorators(pct...),
		)
	}
}

func Run(bar *mpb.Bar, dur time.Duration, tot, inc int64, fct func(), log func(msg string)) {
	for _, f := range []func(){
		func() {
			bar.SetTotal(tot/2, false)
			time.Sleep(dur)
		},
		func() {
			for i := int64(0); i < (inc * 3); i++ {
				bar.IncrInt64(1)
				time.Sleep(dur / 10)
			}
			time.Sleep(dur)
		},
		func() {
			bar.SetTotal(tot, false)
			time.Sleep(dur)
		},
		func() {
			bar.EnableTriggerComplete()
			time.Sleep(dur)
		},
		func() {
			bar.SetTotal(tot+inc, true)
			time.Sleep(dur)
		},
	} {
		f()
		if bar.Completed() {
			fct()
		}
	}

	bar.IncrInt64(inc)
	time.Sleep(dur)

	if !bar.Completed() {
		fct()
	}

	if current := bar.Current(); current != tot {
		log(fmt.Sprintf("Expected current: %d, got: %d", tot, current))
	}

	time.Sleep(dur)
}
