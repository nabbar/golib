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

package mandatory

import (
	"sync/atomic"

	stsctr "github.com/nabbar/golib/status/control"
	"golang.org/x/exp/slices"
)

type model struct {
	Mode *atomic.Value
	Keys *atomic.Value
}

func (o *model) SetMode(m stsctr.Mode) {
	o.Mode.Store(m)
}

func (o *model) GetMode() stsctr.Mode {
	m := o.Mode.Load()

	if m != nil {
		return m.(stsctr.Mode)
	}

	return stsctr.Ignore
}

func (o *model) KeyHas(key string) bool {
	i := o.Keys.Load()

	if i == nil {
		return false
	} else if l, k := i.([]string); !k {
		return false
	} else {
		return slices.Contains(l, key)
	}
}

func (o *model) KeyAdd(keys ...string) {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		l = make([]string, 0)
	} else if l, k = i.([]string); !k {
		l = make([]string, 0)
	}

	for _, key := range keys {
		if !slices.Contains(l, key) {
			l = append(l, key)
		}
	}

	o.Keys.Store(l)
}

func (o *model) KeyDel(keys ...string) {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		o.Keys.Store(make([]string, 0))
		return
	} else if l, k = i.([]string); !k {
		o.Keys.Store(make([]string, 0))
		return
	}

	var res = make([]string, 0)

	for _, key := range l {
		if !slices.Contains(keys, key) {
			res = append(res, key)
		}
	}

	o.Keys.Store(res)
}

func (o *model) KeyList() []string {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		return make([]string, 0)
	} else if l, k = i.([]string); !k {
		return make([]string, 0)
	}

	return slices.Clone(l)
}
