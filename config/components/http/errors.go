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

package http

import (
	"fmt"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the HTTP component package.
// These codes are based on the MinErrorComponentHttp offset defined in the config package.
const (
	// ErrorParamEmpty indicates that one or more required parameters are empty.
	ErrorParamEmpty liberr.CodeError = iota + libcfg.MinErrorComponentHttp

	// ErrorParamInvalid indicates that one or more parameters have invalid values.
	ErrorParamInvalid

	// ErrorComponentNotInitialized indicates that the component has not been properly initialized.
	// This typically means Init() was not called or required dependencies are missing.
	ErrorComponentNotInitialized

	// ErrorConfigInvalid indicates that the component configuration is invalid.
	// This can occur when configuration validation fails or required fields are missing.
	ErrorConfigInvalid

	// ErrorComponentStart indicates a failure during component startup.
	// This can be caused by configuration errors, missing dependencies, or server startup failures.
	ErrorComponentStart

	// ErrorComponentReload indicates a failure during component reload.
	// This can occur when new configuration is invalid or servers fail to restart.
	ErrorComponentReload

	// ErrorDependencyTLSDefault indicates that the TLS component could not be retrieved.
	// This typically means the TLS component is not registered or the TLS key is incorrect.
	ErrorDependencyTLSDefault
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/config/components/http"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "at least one given parameters is empty"
	case ErrorParamInvalid:
		return "at least one given parameters is invalid"
	case ErrorComponentNotInitialized:
		return "this component seems to not be correctly initialized"
	case ErrorConfigInvalid:
		return "invalid component config"
	case ErrorComponentStart:
		return "cannot start component with config"
	case ErrorComponentReload:
		return "cannot reload component with new config"
	case ErrorDependencyTLSDefault:
		return "cannot retrieve TLS component"
	}

	return liberr.NullMessage
}
