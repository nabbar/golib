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
	"time"

	"github.com/vbauerster/mpb/v5/decor"

	njs_semaphore "github.com/nabbar/golib/semaphore"

	"github.com/vbauerster/mpb/v5"
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
	cnl       context.CancelFunc
	sTimeOut  time.Duration
	sMaxSimul int
}

type ProgressBar interface {
	GetMPB() *mpb.Progress

	SetSemaphoreOption(maxSimultaneous int, timeout time.Duration)

	NewBar(parent context.Context, total int64, options ...mpb.BarOption) Bar
	NewBarSimpleETA(name string) Bar
	NewBarSimpleCounter(name string, total int64) Bar
}

func NewProgressBar(timeout time.Duration, deadline time.Time, parent context.Context, options ...mpb.ContainerOption) ProgressBar {
	x, c := njs_semaphore.GetContext(timeout, deadline, parent)

	return &progressBar{
		mpb:       mpb.New(options...),
		ctx:       x,
		cnl:       c,
		sTimeOut:  timeout,
		sMaxSimul: njs_semaphore.GetMaxSimultaneous(),
	}
}

func (p *progressBar) GetMPB() *mpb.Progress {
	return p.mpb
}

func (p *progressBar) SetSemaphoreOption(maxSimultaneous int, timeout time.Duration) {
	p.sMaxSimul = maxSimultaneous
	p.sTimeOut = timeout
}

func (p progressBar) NewBar(parent context.Context, total int64, options ...mpb.BarOption) Bar {
	if parent == nil {
		parent = p.ctx
	}

	return newBar(
		p.mpb.AddBar(0, options...),
		njs_semaphore.NewSemaphore(p.sMaxSimul, p.sTimeOut, njs_semaphore.GetEmptyTime(), parent),
		total,
	)
}

func (p progressBar) NewBarSimpleETA(name string) Bar {
	return newBar(
		p.mpb.AddBar(0,
			mpb.BarStyle(defaultStyle),
			mpb.PrependDecorators(
				// display our name with one space on the right
				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), defaultMessageDone,
				),
			),
			mpb.AppendDecorators(decor.Percentage()),
		),
		njs_semaphore.NewSemaphore(p.sMaxSimul, p.sTimeOut, njs_semaphore.GetEmptyTime(), p.ctx),
		0,
	)
}

func (p progressBar) NewBarSimpleCounter(name string, total int64) Bar {
	return newBar(
		p.mpb.AddBar(total,
			mpb.BarStyle(defaultStyle),
			mpb.PrependDecorators(
				// display our name with one space on the right
				decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
				// use counter (no ETA)
				decor.CountersNoUnit("[%d / %d] ", decor.WCSyncWidth),
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), defaultMessageDone,
				),
			),
			mpb.AppendDecorators(decor.Percentage()),
		),
		njs_semaphore.NewSemaphore(p.sMaxSimul, p.sTimeOut, njs_semaphore.GetEmptyTime(), p.ctx),
		total,
	)
}
