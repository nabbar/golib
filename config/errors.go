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
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgConfig
	ErrorConfigMissingViper
	ErrorComponentNotFound
	ErrorComponentFlagError
	ErrorComponentConfigNotFound
	ErrorComponentConfigError
	ErrorComponentStart
	ErrorComponentReload
)

const (
	MinErrorComponentAws      = ErrorParamEmpty + 10
	MinErrorComponentDatabase = MinErrorComponentAws + 10
	MinErrorComponentHead     = MinErrorComponentDatabase + 10
	MinErrorComponentHttp     = MinErrorComponentHead + 10
	MinErrorComponentHttpCli  = MinErrorComponentHttp + 10
	MinErrorComponentLdap     = MinErrorComponentHttpCli + 10
	MinErrorComponentLog      = MinErrorComponentLdap + 10
	MinErrorComponentMail     = MinErrorComponentLog + 10
	MinErrorComponentNats     = MinErrorComponentMail + 10
	MinErrorComponentNutsDB   = MinErrorComponentNats + 10
	MinErrorComponentRequest  = MinErrorComponentNutsDB + 10
	MinErrorComponentSmtp     = MinErrorComponentRequest + 10
	MinErrorComponentTls      = MinErrorComponentSmtp + 10
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/config"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorConfigMissingViper:
		return "missing valid viper function"
	case ErrorComponentNotFound:
		return "component is not found"
	case ErrorComponentFlagError:
		return "component register flag occurs at least an error"
	case ErrorComponentConfigNotFound:
		return "config keys for component is not found"
	case ErrorComponentConfigError:
		return "config for component has occurs an error"
	case ErrorComponentStart:
		return "cannot start at least one component"
	case ErrorComponentReload:
		return "cannot reload at least one component"
	}

	return liberr.NullMessage
}
