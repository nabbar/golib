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

package aws

import (
	"context"
	"sync/atomic"

	libaws "github.com/nabbar/golib/aws"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
)

type CptAws interface {
	cfgtps.Component

	GetAws() libaws.AWS
	SetAws(a libaws.AWS)

	RegisterHTTPClient(cli libhtc.HttpClient)
}

func New(ctx context.Context, drv ConfigDriver) CptAws {
	return &mod{
		x: libctx.NewConfig[uint8](ctx),
		d: drv,
		p: new(atomic.Value),
		c: new(atomic.Value),
		a: new(atomic.Value),
		r: new(atomic.Value),
		s: new(atomic.Bool),
	}
}

func Register(cfg libcfg.Config, key string, cpt CptAws) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx context.Context, cfg libcfg.Config, drv ConfigDriver, key string) {
	cfg.ComponentSet(key, New(ctx, drv))
}

func Load(getCpt cfgtps.FuncCptGet, key string) CptAws {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptAws); !ok {
		return nil
	} else {
		return h
	}
}
