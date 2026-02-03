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

package logger

import (
	"bytes"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	libctx "github.com/nabbar/golib/context"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

const (
	keyContext = iota
	KeyCancel
	keyLevel
	keyOptions
	keyLogrus
	keyWriter
	keyFilter
	keyFctUpdLog
	keyFctUpdLvl
)

var self = path.Base(reflect.TypeOf(lgr{}).PkgPath())

type lgr struct {
	m sync.RWMutex
	x libctx.Config[uint8] // cf const key...
	f logfld.Fields        // fields map
	c *atomic.Value        // closer
}

func defaultFormatter() logrus.TextFormatter {
	return logrus.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                true,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          true,
		FullTimestamp:             false,
		TimestampFormat:           time.RFC3339,
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    true,
		PadLevelText:              true,
		QuoteEmptyFields:          true,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}
}

func (o *lgr) defaultFormatter(opt *logcfg.OptionsStd) logrus.Formatter {
	f := defaultFormatter()

	if opt != nil && opt.DisableColor {
		f.ForceColors = false
		f.EnvironmentOverrideColors = false
		f.DisableColors = true
	} else {
		f.ForceColors = true
		f.DisableColors = false
	}

	return &f
}

func (o *lgr) defaultFormatterNoColor() logrus.Formatter {
	f := defaultFormatter()
	f.ForceColors = false
	f.EnvironmentOverrideColors = false
	f.DisableColors = true
	return &f
}

func (o *lgr) optionsMerge(opt *logcfg.Options) {
	var oo logcfg.Options

	if i, l := o.x.Load(keyOptions); !l {
		return
	} else if v, k := i.(*logcfg.Options); !k || v == nil {
		return
	} else {
		oo = *v
	}

	oo.Merge(opt)
	*opt = oo
}

func (o *lgr) setLogrusLevel(lvl loglvl.Level) {
	if i, l := o.x.Load(keyLogrus); !l {
		return
	} else if v, k := i.(*logrus.Logger); !k {
		return
	} else {
		v.SetLevel(lvl.Logrus())
		o.x.Store(keyLogrus, v)
	}
}

func (o *lgr) getLogrus() *logrus.Logger {
	if i, l := o.x.Load(keyLogrus); !l {
		return nil
	} else if v, k := i.(*logrus.Logger); !k {
		return nil
	} else {
		return v
	}
}

func (o *lgr) getStack() uint64 {
	b := make([]byte, 64)

	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	//nolint #nosec
	/* #nosec */
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

func (o *lgr) getCaller() runtime.Frame {
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need.
	programCounters := make([]uintptr, 10, 255)
	n := runtime.Callers(1, programCounters)

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		more := true

		for more {
			var frame runtime.Frame
			frame, more = frames.Next()

			if strings.Contains(frame.Function, self) {
				continue
			}

			return frame
		}
	}

	return runtime.Frame{Function: "unknown", File: "unknown", Line: 0}
}
