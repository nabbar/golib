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

package progress

import (
	"context"

	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/semaphore"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

/*
 https://github.com/vbauerster/mpb
*/

var (
	defaultStyle       = "[=>-]<+"
	defaultMessageDone = "done"
)

func SetDefaultStyle(style string) {
	defaultStyle = style
}

func SetDefaultMessageDone(message string) {
	defaultMessageDone = message
}

func GetDefaultStyle() string {
	return defaultStyle
}

func GetDefaultMessageDone() string {
	return defaultMessageDone
}

type progressBar struct {
	mpb       *mpb.Progress
	ctx       context.Context
	sMaxSimul int
	sem       semaphore.Sem
}

type ProgressBar interface {
	GetMPB() *mpb.Progress
	SetMaxThread(maxSimultaneous int)
	SetContext(ctx context.Context)

	UnicProcessInit()
	UnicProcessWait() errors.Error
	UnicProcessNewWorker() errors.Error
	UnicProcessDeferWorker()
	UnicProcessDefer()

	NewBar(total int64, options ...mpb.BarOption) Bar

	NewBarETA(name string, total int64, job string, parent Bar) Bar
	NewBarCounter(name string, total int64, job string, parent Bar) Bar
	NewBarKBits(name string, total int64, job string, parent Bar) Bar

	NewBarSimpleETA(name string, total int64) Bar
	NewBarSimpleCounter(name string, total int64) Bar
	NewBarSimpleKBits(name string, total int64) Bar
}

func NewProgressBar(options ...mpb.ContainerOption) ProgressBar {
	return NewProgressBarWithContext(context.Background(), options...)
}

func NewProgressBarWithContext(ctx context.Context, options ...mpb.ContainerOption) ProgressBar {
	if ctx == nil {
		ctx = context.Background()
	}

	return &progressBar{
		mpb:       mpb.New(options...),
		ctx:       ctx,
		sem:       nil,
		sMaxSimul: semaphore.GetMaxSimultaneous(),
	}
}

func (p *progressBar) GetMPB() *mpb.Progress {
	return p.mpb
}

func (p *progressBar) SetMaxThread(maxSimultaneous int) {
	p.sMaxSimul = maxSimultaneous
}

func (p *progressBar) UnicProcessInit() {
	p.sem = p.semaphore()
}

func (p *progressBar) UnicProcessWait() errors.Error {
	if p.sem != nil {
		return p.sem.WaitAll()
	}
	return nil
}

func (p *progressBar) UnicProcessNewWorker() errors.Error {
	if p.sem != nil {
		return p.sem.NewWorker()
	}
	return nil
}

func (p *progressBar) UnicProcessDeferWorker() {
	if p.sem != nil {
		p.sem.DeferWorker()
	}
}

func (p *progressBar) UnicProcessDefer() {
	if p.sem != nil {
		p.sem.DeferMain()
	}
}

func (p *progressBar) semaphore() semaphore.Sem {
	return semaphore.NewSemaphoreWithContext(p.ctx, p.sMaxSimul)
}

func (p *progressBar) NewBar(total int64, options ...mpb.BarOption) Bar {
	return newBar(
		p.mpb.AddBar(0, options...),
		p.semaphore(),
		total,
		p.sem != nil,
	)
}

func (p *progressBar) NewBarSimpleETA(name string, total int64) Bar {
	return p.NewBarETA(name, total, "", nil)
}

func (p *progressBar) NewBarSimpleCounter(name string, total int64) Bar {
	return p.NewBarCounter(name, total, "", nil)
}

func (p *progressBar) NewBarSimpleKBits(name string, total int64) Bar {
	return p.NewBarKBits(name, total, "", nil)
}

func (p *progressBar) NewBarETA(name string, total int64, job string, parent Bar) Bar {
	if parent != nil && job != "" {
		return newBar(p.addBarJob(total, name, job, nil, nil, parent.GetBarMPB()), p.semaphore(), total, p.sem != nil)
	} else {
		return newBar(p.addBarSimple(total, name, nil, nil), p.semaphore(), total, p.sem != nil)
	}
}

func (p *progressBar) NewBarCounter(name string, total int64, job string, parent Bar) Bar {
	d := decor.CountersNoUnit("[%d / %d] ", decor.WCSyncWidth)
	if parent != nil && job != "" {
		return newBar(p.addBarJob(total, name, job, d, nil, parent.GetBarMPB()), p.semaphore(), total, p.sem != nil)
	} else {
		return newBar(p.addBarSimple(total, name, d, nil), p.semaphore(), total, p.sem != nil)
	}
}

func (p *progressBar) NewBarKBits(name string, total int64, job string, parent Bar) Bar {
	//nolint #gomnd
	d := decor.Counters(decor.UnitKiB, "% .2f / % .2f", decor.WC{W: 20, C: decor.DextraSpace})
	a := []decor.Decorator{
		//nolint #gomnd
		decor.Percentage(decor.WC{W: 5, C: 0}),
		decor.Name(" | "),
		//nolint #gomnd
		decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
	}

	if parent != nil && job != "" {
		return newBar(p.addBarJob(total, name, job, d, a, parent.GetBarMPB()), p.semaphore(), total, p.sem != nil)
	} else {
		return newBar(p.addBarSimple(total, name, d, a), p.semaphore(), total, p.sem != nil)
	}
}

func (p *progressBar) addBarSimple(total int64, name string, counter decor.Decorator, pct []decor.Decorator) *mpb.Bar {
	pr := make([]decor.Decorator, 0)
	// display our name with one space on the right
	pr = append(pr, decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}))
	if counter != nil {
		// use counter (no ETA)
		pr = append(pr, counter)
	}
	//nolint #gomnd
	pr = append(pr, decor.Name("  ", decor.WC{W: 3, C: decor.DidentRight | decor.DextraSpace}))
	// replace ETA decorator with "done" message, OnComplete event
	pr = append(pr, decor.OnComplete(
		// nolint: gomnd
		decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: len(defaultMessageDone) + 1, C: 0}), defaultMessageDone,
	))

	if pct == nil {
		pct = make([]decor.Decorator, 0)
		//nolint #gomnd
		pct = append(pct, decor.Percentage(decor.WC{W: 5, C: 0}))
	}

	return p.mpb.AddBar(total,
		mpb.BarStyle(defaultStyle),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(pr...),
		mpb.AppendDecorators(pct...),
	)
}

func (p *progressBar) addBarJob(total int64, name, job string, counter decor.Decorator, pct []decor.Decorator, bar *mpb.Bar) *mpb.Bar {
	pr := make([]decor.Decorator, 0)
	// display our name with one space on the right
	pr = append(pr, decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}))
	// display our job task with one space on the right
	pr = append(pr, decor.Name(job, decor.WC{W: len(job) + 1, C: decor.DidentRight | decor.DextraSpace}))
	if counter != nil {
		// use counter (no ETA)
		pr = append(pr, counter)
	}
	//nolint #gomnd
	pr = append(pr, decor.Name("  ", decor.WC{W: 3, C: decor.DidentRight | decor.DextraSpace}))
	if bar != nil {
		pr = append(pr, decor.OnComplete(
			// replace ETA decorator with "done" message, OnComplete event
			decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: len(defaultMessageDone) + 1, C: 0}), defaultMessageDone,
		))
	}

	if pct == nil {
		pct = make([]decor.Decorator, 0)
		//nolint #gomnd
		pct = append(pct, decor.Percentage(decor.WC{W: 5, C: 0}))
	}

	if bar == nil {
		return p.mpb.AddBar(total,
			mpb.BarStyle(defaultStyle),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(pr...),
			mpb.AppendDecorators(pct...),
		)
	} else {
		return p.mpb.AddBar(total,
			mpb.BarStyle(defaultStyle),
			mpb.BarQueueAfter(bar),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(pr...),
			mpb.AppendDecorators(pct...),
		)
	}
}

func (p *progressBar) SetContext(ctx context.Context) {
	p.ctx = ctx
}
