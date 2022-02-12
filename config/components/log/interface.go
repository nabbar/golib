/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package log

import (
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

const (
	DefaultLevel  = liblog.InfoLevel
	ComponentType = "log"
)

type ComponentLog interface {
	libcfg.Component

	Log() liblog.Logger

	SetLevel(lvl liblog.Level)
	SetField(fields liblog.Fields)
	SetOptions(opt *liblog.Options) liberr.Error
}

func New(lvl liblog.Level, defLogger func() liblog.Logger) ComponentLog {
	return &componentLog{
		d: defLogger,
		l: nil,
		v: lvl,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentLog) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key string, lvl liblog.Level, defLogger func() liblog.Logger) {
	cfg.ComponentSet(key, New(lvl, defLogger))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentLog {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentLog); !ok {
		return nil
	} else {
		return h
	}
}
