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
	"context"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

const (
	DefaultLevel = loglvl.InfoLevel
)

type CptLog interface {
	cfgtps.Component

	Log() liblog.Logger
	LogClone() liblog.Logger

	SetLevel(lvl loglvl.Level)
	GetLevel() loglvl.Level

	SetField(fields logfld.Fields)
	GetField() logfld.Fields

	SetOptions(opt *logcfg.Options) error
	GetOptions() *logcfg.Options
}

func New(ctx context.Context, lvl loglvl.Level) CptLog {
	c := &mod{
		x: libctx.New[uint8](ctx),
		l: libatm.NewValue[liblog.Logger](),
		r: new(atomic.Bool),
		v: new(atomic.Uint32),
	}

	c.r.Store(false)
	c.v.Store(lvl.Uint32())

	return c
}

func Register(cfg libcfg.Config, key string, cpt CptLog) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx context.Context, cfg libcfg.Config, key string, lvl loglvl.Level) {
	cfg.ComponentSet(key, New(ctx, lvl))
}

func Load(getCpt cfgtps.FuncCptGet, key string) CptLog {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptLog); !ok {
		return nil
	} else {
		return h
	}
}
