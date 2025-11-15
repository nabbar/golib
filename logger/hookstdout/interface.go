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
	loghkw "github.com/nabbar/golib/logger/hookwriter"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type HookStdOut interface {
	logtps.Hook
}

// New returns a new HookStdOut instance.
//
// If opt is nil or opt.DisableStandard is true, the function returns nil and no error.
//
// If len(lvls) is less than 1, the function sets lvls to logrus.AllLevels.
//
// The function then creates a new io.Writer instance. If opt.DisableColor is true, it sets w to os.Stdout.
// Otherwise, it sets w to colorable.NewColorableStdout().
//
// The function then creates a new hkstd instance and sets its fields to w, lvls, f, opt.DisableStack,
// opt.DisableTimestamp, opt.EnableTrace, opt.DisableColor, and opt.EnableAccessLog.
//
// Finally, the function returns the new hkstd instance and no error.
func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error) {
	return NewWithWriter(nil, opt, lvls, f)
}

func NewWithWriter(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error) {
	if w == nil {
		w = os.Stdout
	}

	if opt == nil || opt.DisableStandard {
		return nil, nil
	} else if opt.DisableColor {
		w = colorable.NewColorableStdout()
	}

	return loghkw.New(w, opt, lvls, f)
}
