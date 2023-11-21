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

package hookfile

import (
	"io"
	"os"
	"sync/atomic"

	libiot "github.com/nabbar/golib/ioutils"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type HookFile interface {
	logtps.Hook

	Done() <-chan struct{}
}

func New(opt logcfg.OptionsFile, format logrus.Formatter) (HookFile, error) {
	if opt.Filepath == "" {
		return nil, errMissingFilePath
	}

	var (
		LVLs  = make([]logrus.Level, 0)
		flags = os.O_WRONLY | os.O_APPEND
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, loglvl.Parse(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	if opt.Create {
		flags = os.O_CREATE | flags
	}

	if opt.FileMode == 0 {
		opt.FileMode = 0644
	}

	if opt.PathMode == 0 {
		opt.PathMode = 0755
	}

	n := &hkf{
		s: new(atomic.Value),
		d: new(atomic.Value),
		o: ohkf{
			format:           format,
			flags:            flags,
			levels:           LVLs,
			disableStack:     opt.DisableStack,
			disableTimestamp: opt.DisableTimestamp,
			enableTrace:      opt.EnableTrace,
			enableAccessLog:  opt.EnableAccessLog,
			createPath:       opt.CreatePath,
			filepath:         opt.Filepath,
			fileMode:         opt.FileMode,
			pathMode:         opt.PathMode,
		},
	}

	if opt.CreatePath {
		if e := libiot.PathCheckCreate(true, opt.Filepath, opt.FileMode, opt.PathMode); e != nil {
			return nil, e
		}
	}

	// #nosec
	h, e := os.OpenFile(opt.Filepath, flags, opt.FileMode)

	if e != nil {
		return nil, e
	} else if _, e = h.Seek(0, io.SeekEnd); e != nil {
		_ = h.Close()
		return nil, e
	} else if e = h.Close(); e != nil {
		return nil, e
	}

	return n, nil
}
