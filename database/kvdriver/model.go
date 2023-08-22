/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package kvdriver

func (o *Driver[K, M]) Get(key K, model *M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.FctGet == nil {
		return ErrorGetFunction.Error(nil)
	} else {
		m, e := o.FctGet(key)
		*model = m
		return e
	}
}

func (o *Driver[K, M]) Set(key K, model M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.FctSet == nil {
		return ErrorSetFunction.Error(nil)
	} else {
		return o.FctSet(key, model)
	}
}

func (o *Driver[K, M]) List() ([]K, error) {
	if o == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.FctList == nil {
		return nil, ErrorListFunction.Error(nil)
	} else {
		return o.FctList()
	}
}

func (o *Driver[K, M]) Walk(fct FctWalk[K, M]) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if fct == nil {
		return ErrorFunctionParams.Error(nil)
	} else if o.FctWalk == nil {
		return o.fakeWalk(fct)
	} else {
		return o.FctWalk(fct)
	}
}

func (o *Driver[K, M]) fakeWalk(fct FctWalk[K, M]) error {
	if l, e := o.List(); e != nil {
		return e
	} else {
		for _, k := range l {
			var m = *(new(M))

			if er := o.Get(k, &m); er != nil {
				return er
			}

			if !fct(k, m) {
				return nil
			}
		}
	}

	return nil
}
