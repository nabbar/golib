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
	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

func (l *ListMandatory) Walk(fct func(m stsmdt.Mandatory) bool) {
	if l == nil {
		return
	} else if len(*l) < 1 {
		return
	}

	var (
		k bool
		m stsmdt.Mandatory
		n = *l
	)

	for i := range n {
		m = n[i]
		k = fct(m)
		n[i] = m

		if !k {
			break
		}
	}

	*l = n
}

func (l *ListMandatory) Add(m ...stsmdt.Mandatory) {
	if l == nil {
		return
	}

	var n = *l

	if len(n) < 1 {
		n = make([]stsmdt.Mandatory, 0)
	}

	*l = append(n, m...)
}

func (l *ListMandatory) Del(m stsmdt.Mandatory) {
	if l == nil {
		return
	}

	var (
		n = *l
		r = make([]stsmdt.Mandatory, 0)
	)

	if len(n) < 1 {
		*l = make([]stsmdt.Mandatory, 0)
		return
	}

	for i := range n {
		if n[i] != m {
			r = append(r, n[i])
		}
	}

	*l = r
}

func (l ListMandatory) GetMode(key string) stsctr.Mode {
	if len(l) < 1 {
		return stsctr.Ignore
	}

	for i := range l {
		if l[i].KeyHas(key) {
			return l[i].GetMode()
		}
	}

	return stsctr.Ignore
}

func (l *ListMandatory) SetMode(key string, mod stsctr.Mode) {
	if len(*l) < 1 {
		return
	}

	var n = *l

	for i := range n {
		if n[i].KeyHas(key) {
			n[i].SetMode(mod)
			*l = n
			return
		}
	}
}
