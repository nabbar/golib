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
	"os"

	liberr "github.com/nabbar/golib/errors"
)

func (c *configModel) Start() liberr.Error {
	if err := c.runFuncStartBefore(); err != nil {
		return err
	}

	if err := c.ComponentStart(); err != nil {
		return err
	}

	if err := c.runFuncStartAfter(); err != nil {
		return err
	}

	return nil
}

func (c *configModel) Reload() liberr.Error {
	if err := c.runFuncReloadBefore(); err != nil {
		return err
	}

	if err := c.ComponentReload(); err != nil {
		return err
	}

	if err := c.runFuncReloadAfter(); err != nil {
		return err
	}

	return nil
}

func (c *configModel) Stop() {
	_ = c.runFuncStopBefore()
	c.ComponentStop()
	_ = c.runFuncStopAfter()
}

func (c *configModel) Shutdown(code int) {
	c.cancel()
	os.Exit(code)
}
