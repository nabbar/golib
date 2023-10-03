//go:build examples
// +build examples

/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	var (
		tot = int64(1000)
		inc = int64(1)
		nbb = 5
		bar = make([]*mpb.Bar, 0)
		wgp = sync.WaitGroup{}
		msg = "done"
		pct = []decor.Decorator{
			decor.Percentage(decor.WC{W: 5, C: 0}),
			decor.Name(" | "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60),
		}
		cnt = decor.Counters(decor.SizeB1024(0), "% .2f / % .2f", decor.WC{W: 20, C: decor.DextraSpace})
	)

	pb := mpb.New(
		mpb.WithWidth(64),
		mpb.WithRefreshRate(200*time.Millisecond),
	)

	for i := 0; i < nbb; i++ {
		name := "Complex Bar"
		job := fmt.Sprintf(" | step %d | ", i)

		if i > 0 {
			pr := []decor.Decorator{
				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
				decor.Name(job, decor.WC{W: len(job) + 1, C: decor.DidentRight | decor.DextraSpace}),
				cnt,
				decor.Name("  ", decor.WC{W: 3, C: decor.DidentRight | decor.DextraSpace}),
				decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: len(msg) + 1, C: 0}), msg),
			}

			if i%2 != 1 {
				bar = append(bar, pb.AddBar(tot,
					mpb.BarQueueAfter(bar[i-1]),
					mpb.BarFillerClearOnComplete(),
					mpb.PrependDecorators(pr...),
					mpb.AppendDecorators(pct...),
				))
			} else {
				bar = append(bar, pb.AddBar(0,
					mpb.BarQueueAfter(bar[i-1]),
					mpb.BarFillerClearOnComplete(),
					mpb.PrependDecorators(pr...),
					mpb.AppendDecorators(pct...),
				))
			}
		} else {
			pr := []decor.Decorator{
				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
				decor.Name(job, decor.WC{W: len(job) + 1, C: decor.DidentRight | decor.DextraSpace}),
				cnt,
				decor.Name("  ", decor.WC{W: 3, C: decor.DidentRight | decor.DextraSpace}),
			}
			bar = append(bar, pb.AddBar(0,
				mpb.BarFillerClearOnComplete(),
				mpb.PrependDecorators(pr...),
				mpb.AppendDecorators(pct...),
			))
		}
	}

	for i := range bar {
		wgp.Add(1)
		go func(nb int, done func()) {
			defer done()

			rand.Seed(999)

			for {
				if nb > 0 && !(bar[nb-1].Aborted() || bar[nb-1].Completed()) {
					time.Sleep(time.Second)
				} else if bar[nb].Current() == 0 {
					bar[nb].SetTotal(tot, false)
					bar[nb].SetRefill(0)
				}

				if bar[nb].Current() < tot {
					time.Sleep(time.Duration(rand.Intn(9)+1) * time.Millisecond)
					bar[nb].IncrInt64(inc)
				} else {
					bar[nb].EnableTriggerComplete()
					bar[nb].Abort(false)
					return
				}
			}
		}(i, wgp.Done)
	}

	wgp.Wait()
	time.Sleep(500 * time.Millisecond)

	println("finish complex...")
}
