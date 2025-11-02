/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package fields

import libctx "github.com/nabbar/golib/context"

func (o *fldModel) Clean() {
	o.c.Clean()
}

func (o *fldModel) Get(key string) (val interface{}, ok bool) {
	return o.c.Load(key)
}

func (o *fldModel) Store(key string, cfg interface{}) {
	o.c.Store(key, cfg)
}

func (o *fldModel) Delete(key string) Fields {
	o.c.Delete(key)
	return o
}

func (o *fldModel) Merge(f Fields) Fields {
	if f == nil || o == nil {
		return o
	}

	f.Walk(func(key string, val interface{}) bool {
		o.c.Store(key, val)
		return true
	})

	return o
}

func (o *fldModel) Walk(fct libctx.FuncWalk[string]) Fields {
	o.c.Walk(fct)
	return o
}

func (o *fldModel) WalkLimit(fct libctx.FuncWalk[string], validKeys ...string) Fields {
	o.c.WalkLimit(fct, validKeys...)
	return o
}

func (o *fldModel) LoadOrStore(key string, cfg interface{}) (val interface{}, loaded bool) {
	return o.c.LoadOrStore(key, cfg)
}

func (o *fldModel) LoadAndDelete(key string) (val interface{}, loaded bool) {
	return o.c.LoadAndDelete(key)
}
