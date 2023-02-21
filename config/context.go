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

package config

import (
	libctx "github.com/nabbar/golib/context"
)

func (c *configModel) Context() libctx.Config[string] {
	return c.ctx
}

func (c *configModel) CancelAdd(fct ...func()) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fcnl = append(c.fcnl, fct...)
}

func (c *configModel) CancelClean() {
	c.m.Lock()
	defer c.m.Unlock()

	c.fcnl = make([]func(), 0)
}

func (c *configModel) cancel() {
	if l := c.getCancelCustom(); len(l) > 0 {
		for _, f := range l {
			f()
		}
	}

	c.Stop()
}

func (c *configModel) getCancelCustom() []func() {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.fcnl
}
