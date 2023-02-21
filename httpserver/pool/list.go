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

package pool

import (
	"regexp"
	"strings"

	liblog "github.com/nabbar/golib/logger"

	libhtp "github.com/nabbar/golib/httpserver"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

func (o *pool) Clean() {
	o.p.Clean()
}

func (o *pool) Walk(fct FuncWalk) bool {
	return o.WalkLimit(fct)
}

func (o *pool) WalkLimit(fct FuncWalk, onlyBindAddress ...string) bool {
	if fct == nil {
		return false
	}

	return o.p.WalkLimit(func(key string, val interface{}) bool {
		if v, k := val.(libhtp.Server); !k {
			return true
		} else {
			return fct(key, v)
		}
	}, onlyBindAddress...)
}

func (o *pool) Load(bindAddress string) libhtp.Server {
	if i, l := o.p.Load(bindAddress); !l {
		return nil
	} else if v, k := i.(libhtp.Server); !k {
		return nil
	} else {
		return v
	}
}

func (o *pool) Store(srv libhtp.Server) {
	o.p.Store(srv.GetBindable(), srv)
}

func (o *pool) StoreNew(cfg libhtp.Config, defLog liblog.FuncLog) error {
	if s, e := libhtp.New(cfg, defLog); e != nil {
		return e
	} else {
		o.Store(s)
		return nil
	}
}

func (o *pool) Delete(bindAddress string) {
	o.p.Delete(bindAddress)
}

func (o *pool) LoadAndDelete(bindAddress string) (val libhtp.Server, loaded bool) {
	if i, l := o.p.LoadAndDelete(bindAddress); !l {
		return nil, false
	} else if v, k := i.(libhtp.Server); !k {
		return nil, false
	} else {
		return v, true
	}
}

func (o *pool) Has(bindAddress string) bool {
	if i, l := o.p.Load(bindAddress); !l {
		return false
	} else {
		_, ok := i.(libhtp.Server)
		return ok
	}
}

func (o *pool) Len() int {
	var cnt int

	o.p.Walk(func(key string, val interface{}) bool {
		cnt++
		return true
	})

	return cnt
}

func (o *pool) Filter(field srvtps.FieldType, pattern, regex string) Pool {
	var (
		r = o.Clone(nil)
		f string
	)

	r.Clean()
	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		switch field {
		case srvtps.FieldBind:
			f = srv.GetBindable()
		case srvtps.FieldExpose:
			f = srv.GetExpose()
		default:
			f = srv.GetName()
		}

		var found = false

		if len(pattern) > 0 && strings.EqualFold(f, pattern) {
			found = true
		} else if len(regex) > 0 {
			if ok, err := regexp.MatchString(regex, f); err == nil && ok {
				found = true
			}
		}

		if found {
			r.Store(srv)
		}

		return true
	})

	return r
}

func (o *pool) List(fieldFilter, fieldReturn srvtps.FieldType, pattern, regex string) []string {
	var r = make([]string, 0)

	o.Filter(fieldFilter, pattern, regex).Walk(func(bindAddress string, srv libhtp.Server) bool {
		switch fieldReturn {
		case srvtps.FieldBind:
			r = append(r, srv.GetBindable())
		case srvtps.FieldExpose:
			r = append(r, srv.GetExpose())
		default:
			r = append(r, srv.GetName())
		}
		return true
	})

	return r
}
