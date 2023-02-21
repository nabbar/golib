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

package httpserver

import (
	"net/http"

	srvtps "github.com/nabbar/golib/httpserver/types"
)

func (o *srv) Handler(h srvtps.FuncHandler) {
	o.m.Lock()
	defer o.m.Unlock()
	o.h = h
}

func (o *srv) HandlerGet(key string) http.Handler {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.h == nil {
		return srvtps.NewBadHandler()
	} else if l := o.h(); len(l) < 1 {
		return srvtps.NewBadHandler()
	} else if h, k := l[key]; !k {
		return srvtps.NewBadHandler()
	} else {
		return h
	}
}

func (o *srv) HandlerGetValidKey() string {
	if i, l := o.c.Load(cfgHandler); !l {
		return srvtps.BadHandlerName
	} else if _, f := i.(*srvtps.BadHandler); f {
		return srvtps.BadHandlerName
	} else if i == nil {
		return srvtps.BadHandlerName
	} else if i, l = o.c.Load(cfgHandlerKey); !l {
		return srvtps.BadHandlerName
	} else if v, k := i.(string); !k {
		return srvtps.BadHandlerName
	} else {
		return v
	}
}

func (o *srv) HandlerHas(key string) bool {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.h == nil {
		return false
	} else if l := o.h(); len(l) < 1 {
		return false
	} else {
		_, k := l[key]
		return k
	}
}

func (o *srv) HandlerStoreFct(key string) {
	o.c.Store(cfgHandler, func() http.Handler {
		return o.HandlerGet(key)
	})
	o.c.Store(cfgHandlerKey, key)
}

func (o *srv) HandlerLoadFct() http.Handler {
	if i, l := o.c.Load(cfgHandler); !l {
		return srvtps.NewBadHandler()
	} else if v, k := i.(func() http.Handler); !k {
		return srvtps.NewBadHandler()
	} else if h := v(); h == nil {
		return srvtps.NewBadHandler()
	} else {
		return h
	}
}
