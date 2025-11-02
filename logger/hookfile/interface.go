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
	libsiz "github.com/nabbar/golib/size"
	"github.com/sirupsen/logrus"
)

type HookFile interface {
	logtps.Hook

	// Done returns a channel that will be closed when the hook is finished.
	// Use this channel to wait until the hook is finished and then flush
	// the buffer before exit function.
	//
	Done() <-chan struct{}
}

// New returns a new HookFile instance.
//
// The `opt` parameter is required and cannot be nil. If the `opt.Filepath`
// field is empty, an error will be returned.
//
// The `format` parameter is optional and can be nil. If it is nil, the
// default logrus formatter will be used.
//
// If the `opt.LogLevel` field is not empty, the levels will be parsed and set
// on the hook. If the `opt.LogLevel` field is empty, all levels will be enabled.
//
// The `opt.Create` field is optional and can be false. If it is true, the
// file will be created if it does not exist. The file mode is set to the value
// of `opt.FileMode`. If the `opt.FileMode` field is empty, the file mode
// will be set to 0644.
//
// The `opt.PathMode` field is optional and can be false. If it is true, the
// parent directories of the file will be created if they do not exist. The parent
// directory mode is set to the value of `opt.PathMode`. If the `opt.PathMode`
// field is empty, the parent directory mode will be set to 0755.
//
// The `opt.FileBufferSize` field is optional and can be set to a value greater
// than zero. If it is set to zero or a negative value, the default buffer size
// will be used. The default buffer size is 4KB.
//
// The `opt.DisableStack` field is optional and can be false. If it is true,
// the stack trace will be disabled on the hook.
//
// The `opt.DisableTimestamp` field is optional and can be false. If it is true,
// the timestamp will be disabled on the hook.
//
// The `opt.EnableTrace` field is optional and can be false. If it is true,
// the trace will be enabled on the hook.
//
// The `opt.EnableAccessLog` field is optional and can be false. If it is true,
// the access log will be enabled on the hook.
//
// The returned hook is safe for use in multiple goroutines.
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
		b: new(atomic.Int64),
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
			fileMode:         opt.FileMode.FileMode(),
			pathMode:         opt.PathMode.FileMode(),
		},
		r: new(atomic.Bool),
	}

	if opt.FileBufferSize <= libsiz.SizeKilo {
		n.b.Store(opt.FileBufferSize.Int64())
	} else {
		n.b.Store(sizeBuffer)
	}

	if opt.CreatePath {
		if e := libiot.PathCheckCreate(true, opt.Filepath, opt.FileMode.FileMode(), opt.PathMode.FileMode()); e != nil {
			return nil, e
		}
	}

	// #nosec
	h, e := os.OpenFile(opt.Filepath, flags, opt.FileMode.FileMode())

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
