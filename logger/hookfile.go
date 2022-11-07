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
	"strings"
	"sync"
	"time"

	"github.com/nabbar/golib/ioutils"
	"github.com/sirupsen/logrus"
)

type HookFile interface {
	logrus.Hook
	io.WriteCloser
	RegisterHook(log *logrus.Logger)
}

type _HookFile struct {
	m sync.Mutex
	h *os.File
	w time.Time
	r logrus.Formatter
	l []logrus.Level
	s bool
	d bool
	t bool
	a bool
	o _HookFileOptions
}

type _HookFileOptions struct {
	Create   bool
	FilePath string
	Flags    int
	ModeFile os.FileMode
	ModePath os.FileMode
}

func NewHookFile(opt OptionsFile, format logrus.Formatter) (HookFile, error) {
	if opt.Filepath == "" {
		return nil, fmt.Errorf("missing file path")
	}

	var (
		LVLs  = make([]logrus.Level, 0)
		flags = os.O_WRONLY | os.O_APPEND
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, GetLevelString(ls).Logrus())
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

	obj := &_HookFile{
		m: sync.Mutex{},
		h: nil,
		w: time.Time{},
		r: format,
		l: LVLs,
		s: opt.DisableStack,
		d: opt.DisableTimestamp,
		t: opt.EnableTrace,
		a: opt.EnableAccessLog,
		o: _HookFileOptions{
			Create:   opt.CreatePath,
			FilePath: opt.Filepath,
			Flags:    flags,
			ModeFile: opt.FileMode,
			ModePath: opt.PathMode,
		},
	}

	if h, e := obj.openCreate(); e != nil {
		return nil, e
	} else {
		_ = h.Close()
	}

	return obj, nil
}

func (o *_HookFile) openCreate() (*os.File, error) {
	var err error

	if o.o.Create {
		if err = ioutils.PathCheckCreate(true, o.o.FilePath, o.o.ModeFile, o.o.ModePath); err != nil {
			return nil, err
		}
	}

	if h, e := os.OpenFile(o.o.FilePath, o.o.Flags, o.o.ModeFile); e != nil {
		return nil, e
	} else if _, e = h.Seek(0, io.SeekEnd); e != nil {
		return nil, e
	} else {
		return h, nil
	}
}

func (o *_HookFile) isStack() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.s
}

func (o *_HookFile) isTimeStamp() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.d
}

func (o *_HookFile) isTrace() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.t
}

func (o *_HookFile) isAccessLog() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.a
}

func (o *_HookFile) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *_HookFile) Levels() []logrus.Level {
	return o.l
}

func (o *_HookFile) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if !o.isStack() {
		ent.Data = o.filterKey(ent.Data, FieldStack)
	}

	if !o.isTimeStamp() {
		ent.Data = o.filterKey(ent.Data, FieldTime)
	}

	if !o.isTrace() {
		ent.Data = o.filterKey(ent.Data, FieldCaller)
		ent.Data = o.filterKey(ent.Data, FieldFile)
		ent.Data = o.filterKey(ent.Data, FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.isAccessLog() {
		if len(entry.Message) > 0 {
			if !strings.HasSuffix(entry.Message, "\n") {
				entry.Message += "\n"
			}
			p = []byte(entry.Message)
		} else {
			return nil
		}
	} else {
		if len(ent.Data) < 1 {
			return nil
		} else if p, e = ent.Bytes(); e != nil {
			return e
		}
	}

	if _, e = o.Write(p); e != nil {
		return e
	}

	return nil
}

func (o *_HookFile) write(p []byte) (n int, err error) {
	o.m.Lock()
	defer o.m.Unlock()

	var e error

	if o.h == nil {
		if o.h, e = o.openCreate(); e != nil {
			return 0, fmt.Errorf("logrus.hookfile: cannot open '%s': %v", o.o.FilePath, e)
		}
	} else if _, e = o.h.Seek(0, io.SeekEnd); e != nil {
		return 0, fmt.Errorf("logrus.hookfile: cannot seek file '%s' to EOF: %v", o.o.FilePath, e)
	}

	return o.h.Write(p)
}

func (o *_HookFile) Write(p []byte) (n int, err error) {
	if n, err = o.write(p); err != nil {
		_ = o.Close()
		n, err = o.write(p)
	}

	if err != nil {
		return n, err
	}

	o.m.Lock()
	defer o.m.Unlock()

	if o.w.IsZero() {
		_ = o.h.Sync()
		o.w = time.Now()
		return n, err
	} else if time.Since(o.w) > 30*time.Second {
		_ = o.h.Sync()
		o.w = time.Now()
		return n, err
	}

	return n, err
}

func (o *_HookFile) Close() error {
	o.m.Lock()
	defer o.m.Unlock()

	if o.h != nil {
		var e error

		if er := o.h.Sync(); er != nil {
			e = fmt.Errorf("logrus.hookfile: sync file error '%s': %v", o.o.FilePath, er)
		}

		if er := o.h.Close(); er != nil {
			if e != nil {
				e = fmt.Errorf("%v, close file error '%s': %v", e, o.o.FilePath, er)
			} else {
				e = fmt.Errorf("logrus.hookfile: close file error '%s': %v", o.o.FilePath, er)
			}
		}

		o.h = nil
		return e
	}

	return nil
}

func (o *_HookFile) filterKey(f logrus.Fields, key string) logrus.Fields {
	if len(f) < 1 {
		return f
	}

	if _, ok := f[key]; !ok {
		return f
	} else {
		delete(f, key)
		return f
	}
}
