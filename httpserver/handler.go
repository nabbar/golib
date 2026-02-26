/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

// Handler registers or updates the handler function that provides HTTP handlers.
// If h is nil, an empty handler map is used as default.
func (o *srv) Handler(h srvtps.FuncHandler) {
	if h == nil {
		h = func() map[string]http.Handler {
			return map[string]http.Handler{}
		}
	}

	o.h.Store(h)
}

// HandlerHas checks if a handler is registered for the specified key.
// Returns true if the handler exists, false otherwise.
func (o *srv) HandlerHas(key string) bool {
	if l := o.getHandler(); len(l) < 1 {
		return false
	} else {
		_, k := l[key]
		return k
	}
}

// HandlerGet retrieves the handler registered for the specified key.
// Returns BadHandler if no handler is found for the key.
func (o *srv) HandlerGet(key string) http.Handler {
	if l := o.getHandler(); len(l) < 1 {
		return srvtps.NewBadHandler()
	} else if h, k := l[key]; !k {
		return srvtps.NewBadHandler()
	} else {
		return h
	}
}

// HandlerGetValidKey returns the currently active handler key.
// Returns BadHandlerName if no valid handler is configured.
func (o *srv) HandlerGetValidKey() string {
	if i, l := o.c.Load(cfgHandler); !l {
		return srvtps.BadHandlerName
	} else if _, f := i.(*srvtps.BadHandler); f {
		return srvtps.BadHandlerName
	} else if i == nil {
		return srvtps.BadHandlerName
	} else if i, l = o.c.Load(cfgHandlerKey); !l {
		return srvtps.BadHandlerName
	} else if v, k := i.(string); !k || len(v) < 1 {
		return srvtps.BadHandlerName
	} else {
		return v
	}
}

// HandlerStoreFct stores a handler function reference for the specified key.
// This is used internally to cache the handler function.
func (o *srv) HandlerStoreFct(key string) {
	o.c.Store(cfgHandler, func() http.Handler {
		return o.HandlerGet(key)
	})
	o.c.Store(cfgHandlerKey, key)
}

// HandlerLoadFct loads and executes the stored handler function.
// Returns BadHandler if no valid handler function is stored.
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

func (o *srv) getHandler() map[string]http.Handler {
	if o == nil || o.h == nil {
		return nil
	} else if f := o.h.Load(); f == nil {
		return nil
	} else {
		return f()
	}
}
