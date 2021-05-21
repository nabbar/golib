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

	"github.com/sirupsen/logrus"
)

type HookStandard interface {
	logrus.Hook
	io.WriteCloser
	RegisterHook(log *logrus.Logger)
}

type _HookStd struct {
	w io.Writer
	l []logrus.Level
	s bool
	d bool
	t bool
	c bool
}

func NewHookStandard(opt Options, w io.Writer, lvls []logrus.Level) HookFile {
	if len(lvls) < 1 {
		lvls = logrus.AllLevels
	}

	return &_HookStd{
		w: w,
		l: lvls,
		s: opt.DisableStack,
		d: opt.DisableTimestamp,
		t: opt.EnableTrace,
	}
}

func (o *_HookStd) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *_HookStd) Levels() []logrus.Level {
	return o.l
}

func (o *_HookStd) Fire(entry *logrus.Entry) error {
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

func (o *_HookStd) Write(p []byte) (n int, err error) {
	if o.w == nil {
		return 0, fmt.Errorf("logrus.hookstd: writer not setup")
	}

	return o.w.Write(p)
}

func (o *_HookStd) Close() error {
	return nil
}

func (o *_HookStd) filterKey(f logrus.Fields, key string) logrus.Fields {
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
