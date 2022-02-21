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

package viper

import (
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

func (v *viper) Config(logLevelRemoteKO, logLevelRemoteOK liblog.Level) liberr.Error {
	if err := v.initAddRemote(); err == nil {
		v.initWatchRemote(logLevelRemoteKO, logLevelRemoteOK)
		return nil
	} else if !err.IsCodeError(ErrorParamMissing) {
		return err
	}

	err := v.v.ReadInConfig()

	if err == nil {
		return nil
	}

	if v.deft != nil {
		v.v.SetConfigType("json")
		if err = v.v.ReadConfig(v.deft()); err != nil {
			return ErrorConfigReadDefault.ErrorParent(err)
		}

		return ErrorConfigIsDefault.Error(nil)
	}

	return ErrorConfigRead.ErrorParent(err)
}
