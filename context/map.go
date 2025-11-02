/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"slices"
)

func (c *ccx[T]) Clean() {
	c.m.Range(func(key T, _ any) bool {
		c.m.Delete(key)
		return true
	})
}

func (c *ccx[T]) Load(key T) (val interface{}, ok bool) {
	return c.m.Load(key)
}

func (c *ccx[T]) Store(key T, cfg interface{}) {
	if c.Err() != nil {
		c.Clean()
		return
	} else if cfg != nil {
		c.m.Store(key, cfg)
	}
}

func (c *ccx[T]) Delete(key T) {
	if c.Err() != nil {
		c.Clean()
		return
	}

	c.m.Delete(key)
}

func (c *ccx[T]) LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool) {
	if c.Err() != nil {
		c.Clean()
		return nil, false
	}

	return c.m.LoadOrStore(key, cfg)
}

func (c *ccx[T]) LoadAndDelete(key T) (val interface{}, loaded bool) {
	if c.Err() != nil {
		c.Clean()
		return nil, false
	}
	return c.m.LoadAndDelete(key)
}

func (c *ccx[T]) Walk(fct FuncWalk[T]) {
	c.WalkLimit(fct)
}

func (c *ccx[T]) WalkLimit(fct FuncWalk[T], validKeys ...T) {
	c.m.Range(func(key T, val any) bool {
		if val == nil {
			c.m.Delete(key)
		} else if len(validKeys) < 1 {
			return fct(key, val)
		} else if slices.Contains(validKeys, key) {
			return fct(key, val)
		}
		return true
	})
}
