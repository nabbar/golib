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

package head

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	librtr "github.com/nabbar/golib/router/header"
)

type CptHead interface {
	cfgtps.Component

	GetHeaders() librtr.Headers
	SetHeaders(head librtr.Headers)
}

func New(ctx context.Context) CptHead {
	return &mod{
		x: libctx.New[uint8](ctx),
		h: libatm.NewValue[librtr.Headers](),
	}
}

func Register(cfg libcfg.Config, key string, cpt CptHead) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx context.Context, cfg libcfg.Config, key string) {
	cfg.ComponentSet(key, New(ctx))
}

func Load(getCpt cfgtps.FuncCptGet, key string) CptHead {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptHead); !ok {
		return nil
	} else {
		return h
	}
}
