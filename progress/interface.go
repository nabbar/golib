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

	libsem "github.com/nabbar/golib/semaphore"
	libmpb "github.com/vbauerster/mpb/v5"
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

type Bar interface {
	libsem.SemBar

	DropOnDefer(flag bool)
	GetBarMPB() *libmpb.Bar
}

type ProgressBar interface {
	libsem.Sem

	GetMPB() *libmpb.Progress
	SetMaxThread(maxSimultaneous int)
	SetContext(ctx context.Context)

	MainProcessInit()

	NewBar(total int64, options ...libmpb.BarOption) Bar

	NewBarETA(name string, total int64, job string, parent Bar) Bar
	NewBarCounter(name string, total int64, job string, parent Bar) Bar
	NewBarKBits(name string, total int64, job string, parent Bar) Bar

	NewBarSimpleETA(name string, total int64) Bar
	NewBarSimpleCounter(name string, total int64) Bar
	NewBarSimpleKBits(name string, total int64) Bar
}

func NewProgressBar(options ...libmpb.ContainerOption) ProgressBar {
	return NewProgressBarWithContext(context.Background(), options...)
}

func NewProgressBarWithContext(ctx context.Context, options ...libmpb.ContainerOption) ProgressBar {
	if ctx == nil {
		ctx = context.Background()
	}

	return &progressBar{
		mpb:       libmpb.New(options...),
		ctx:       ctx,
		sem:       nil,
		sMaxSimul: libsem.GetMaxSimultaneous(),
	}
}
