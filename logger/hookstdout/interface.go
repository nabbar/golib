/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hookstdout

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
	logcfg "github.com/nabbar/golib/logger/config"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type HookStdOut interface {
	logtps.Hook
}

func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error) {
	if opt == nil || opt.DisableStandard {
		return nil, nil
	}

	if len(lvls) < 1 {
		lvls = logrus.AllLevels
	}

	var w io.Writer

	if opt.DisableColor {
		w = os.Stdout
	} else {
		w = colorable.NewColorableStdout()
	}

	n := &hkstd{
		w: w,
		l: lvls,
		f: f,
		s: opt.DisableStack,
		d: opt.DisableTimestamp,
		t: opt.EnableTrace,
		c: opt.DisableColor,
		a: opt.EnableAccessLog,
	}

	return n, nil
}
