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
)

func (v *viper) initAddRemote() liberr.Error {
	if v.remote.provider == "" || v.remote.endpoint == "" || v.remote.path == "" {
		return ErrorParamMissing.Error(nil)
	}

	if v.remote.secure != "" {
		if err := v.v.AddSecureRemoteProvider(v.remote.provider, v.remote.endpoint, v.remote.path, v.remote.secure); err != nil {
			return ErrorRemoteProviderSecure.Error(err)
		}
	} else if err := v.v.AddRemoteProvider(v.remote.provider, v.remote.endpoint, v.remote.path); err != nil {
		return ErrorRemoteProvider.Error(err)
	}

	return v.initSetRemote()
}

func (v *viper) initSetRemote() liberr.Error {
	v.v.SetConfigType("json")

	if err := v.v.ReadRemoteConfig(); err != nil {
		return ErrorRemoteProviderRead.Error(err)
	}

	if err := v.v.Unmarshal(v.remote.model); err != nil {
		return ErrorRemoteProviderMarshall.Error(err)
	}

	return nil
}
