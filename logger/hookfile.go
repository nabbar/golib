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

package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/nabbar/golib/ioutils"
	"github.com/sirupsen/logrus"
)

type HookFile interface {
	logrus.Hook
	io.WriteCloser
	RegisterHook(log *logrus.Logger)
}

type _HookFile struct {
	f *os.File
	l []logrus.Level
	s bool
	d bool
	t bool
}

func NewHookFile(opt OptionsFile) (HookFile, error) {
	if opt.Filepath == "" {
		return nil, fmt.Errorf("missing file path")
	}

	var (
		flags = os.O_WRONLY | os.O_APPEND
		LVLs  = make([]logrus.Level, 0)
		hdl   *os.File
		err   error
	)

	if opt.FileMode == 0 {
		opt.FileMode = 0644
	}

	if opt.PathMode == 0 {
		opt.PathMode = 0755
	}

	if opt.CreatePath {
		if err = ioutils.PathCheckCreate(true, opt.Filepath, opt.FileMode, opt.PathMode); err != nil {
			return nil, err
		}
	} else if opt.Create {
		flags = os.O_CREATE | flags
	}

	if hdl, err = os.OpenFile(opt.Filepath, flags, opt.FileMode); err != nil {
		return nil, err
	}

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, GetLevelString(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	return &_HookFile{
		f: hdl,
		l: LVLs,
		s: opt.DisableStack,
		d: opt.DisableTimestamp,
		t: opt.EnableTrace,
	}, nil
}

func (o *_HookFile) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *_HookFile) Levels() []logrus.Level {
	return o.l
}

func (o *_HookFile) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()

	if o.s {
		ent.Data = o.filterKey(ent.Data, FieldStack)
	}

	if o.d {
		ent.Data = o.filterKey(ent.Data, FieldTime)
	}

	if !o.t {
		ent.Data = o.filterKey(ent.Data, FieldCaller)
		ent.Data = o.filterKey(ent.Data, FieldFile)
		ent.Data = o.filterKey(ent.Data, FieldLine)
	}

	if p, err := ent.Bytes(); err != nil {
		return err
	} else if _, err = o.Write(p); err != nil {
		return err
	}

	return nil
}

func (o *_HookFile) Write(p []byte) (n int, err error) {
	if o.f == nil {
		return 0, fmt.Errorf("logrus.hookfile: file not setup")
	}

	return o.f.Write(p)
}

func (o *_HookFile) Close() error {
	err := o.f.Close()
	o.f = nil
	return err
}

func (o *_HookFile) filterKey(f logrus.Fields, key string) logrus.Fields {
	if len(f) < 1 {
		return f
	}

	var res = make(map[string]interface{}, 0)

	for k, v := range f {
		if k == key {
			continue
		}
		res[k] = v
	}

	return res
}
