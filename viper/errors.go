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
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgViper
	ErrorParamMissing
	ErrorHomePathNotFound
	ErrorBasePathNotFound
	ErrorRemoteProvider
	ErrorRemoteProviderSecure
	ErrorRemoteProviderRead
	ErrorRemoteProviderMarshall
	ErrorConfigRead
	ErrorConfigReadDefault
	ErrorConfigIsDefault
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/viper"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamMissing:
		return "at least one parameter is missing"
	case ErrorHomePathNotFound:
		return "cannot retrieve user home path"
	case ErrorBasePathNotFound:
		return "cannot retrieve base config path"
	case ErrorRemoteProvider:
		return "cannot define remote provider"
	case ErrorRemoteProviderSecure:
		return "cannot define secure remote provider"
	case ErrorRemoteProviderRead:
		return "cannot read config from remote provider"
	case ErrorRemoteProviderMarshall:
		return "cannot marshall config from remote provider"
	case ErrorConfigRead:
		return "cannot read config from file"
	case ErrorConfigReadDefault:
		return "cannot read default config"
	case ErrorConfigIsDefault:
		return "cannot read config, use default config"
	}

	return liberr.NullMessage
}
