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

package listmandatory

import (
	"sort"
	"sync"
	"sync/atomic"

	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
	"golang.org/x/exp/slices"
)

type model struct {
	l sync.Map
	k *atomic.Int32
}

func (o *model) inc() int32 {
	i := o.k.Load()
	i++
	o.k.Store(i)
	return i
}

func (o *model) Len() int {
	var l int

	o.l.Range(func(key, value any) bool {
		var k bool

		if _, k = key.(int32); !k {
			o.l.Delete(key)
			return true
		} else if _, k = value.(stsmdt.Mandatory); !k {
			o.l.Delete(key)
			return true
		}

		l++
		return true
	})

	return l
}

func (o *model) Walk(fct func(m stsmdt.Mandatory) bool) {
	o.l.Range(func(key, value any) bool {
		var (
			k bool
			v stsmdt.Mandatory
		)

		if _, k = key.(int32); !k {
			o.l.Delete(key)
			return true
		} else if v, k = value.(stsmdt.Mandatory); !k {
			o.l.Delete(key)
			return true
		} else {
			k = fct(v)
			o.l.Store(k, v)
			return k
		}
	})
}

func (o *model) Add(m ...stsmdt.Mandatory) {
	for _, v := range m {
		o.l.Store(o.inc(), v)
	}
}

func (o *model) Del(m stsmdt.Mandatory) {
	o.l.Range(func(key, value any) bool {
		var (
			k bool
			v stsmdt.Mandatory
		)

		if _, k = key.(int32); !k {
			o.l.Delete(key)
			return true
		} else if v, k = value.(stsmdt.Mandatory); !k {
			o.l.Delete(key)
			return true
		} else {
			u := v.KeyList()
			sort.Strings(u)

			n := m.KeyList()
			sort.Strings(n)

			if slices.Compare(u, n) != 0 {
				return true
			}

			o.l.Delete(key)
			return true
		}
	})
}

func (o *model) GetMode(key string) stsctr.Mode {
	var res = stsctr.Ignore

	o.l.Range(func(ref, value any) bool {
		var (
			k bool
			v stsmdt.Mandatory
		)

		if _, k = ref.(int32); !k {
			o.l.Delete(ref)
			return true
		} else if v, k = value.(stsmdt.Mandatory); !k {
			o.l.Delete(ref)
			return true
		} else if v.KeyHas(key) {
			res = v.GetMode()
			return false
		} else {
			return true
		}
	})

	return res
}

func (o *model) SetMode(key string, mod stsctr.Mode) {
	o.l.Range(func(ref, value any) bool {
		var (
			k bool
			v stsmdt.Mandatory
		)

		if _, k = ref.(int32); !k {
			o.l.Delete(ref)
			return true
		} else if v, k = value.(stsmdt.Mandatory); !k {
			o.l.Delete(ref)
			return true
		} else if v.KeyHas(key) {
			v.SetMode(mod)
			return false
		} else {
			return true
		}
	})
}
