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

package hookstderr

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
	logcfg "github.com/nabbar/golib/logger/config"
	loghkw "github.com/nabbar/golib/logger/hookwriter"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type HookStdErr interface {
	logtps.Hook
}

// New returns a new HookStdErr with the given options.
// If opt is nil, or opt.DisableStandard is true, it will return nil.
// If len(lvls) is less than 1, it will use all levels.
// If opt.DisableColor is true, it will use os.Stderr, otherwise it will use a colorable stderr.
// It will return nil and an error if there is an issue creating the HookStdErr.
func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error) {
	return NewWithWriter(nil, opt, lvls, f)
}

func NewWithWriter(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error) {
	if w == nil {
		w = os.Stderr
	}

	if opt == nil || opt.DisableStandard {
		return nil, nil
	} else if opt.DisableColor {
		w = colorable.NewColorableStdout()
	}

	return loghkw.New(w, opt, lvls, f)
}
