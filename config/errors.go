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

// Error codes for the config package.
// These errors are used throughout the config package and its components
// to provide standardized error reporting with proper error chaining.
const (
	// ErrorParamEmpty indicates that required parameters were not provided.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgConfig

	// ErrorConfigMissingViper indicates that the Viper configuration provider is not registered.
	// Register one using Config.RegisterFuncViper() before component initialization.
	ErrorConfigMissingViper

	// ErrorComponentNotFound indicates that a requested component was not found in the registry.
	// Verify the component key and ensure the component is registered via ComponentSet().
	ErrorComponentNotFound

	// ErrorComponentFlagError indicates that at least one component failed during flag registration.
	// Check individual component RegisterFlag() implementations for specific errors.
	ErrorComponentFlagError

	// ErrorComponentConfigNotFound indicates that configuration keys for a component are missing.
	// Ensure the configuration file contains the required component section.
	ErrorComponentConfigNotFound

	// ErrorComponentConfigError indicates that a component's configuration is invalid or malformed.
	// Review the component's configuration structure and validation rules.
	ErrorComponentConfigError

	// ErrorComponentStart indicates that at least one component failed to start.
	// Check component logs for specific startup errors.
	ErrorComponentStart

	// ErrorComponentReload indicates that at least one component failed to reload.
	// Check component logs for specific reload errors.
	ErrorComponentReload
)

// Error code ranges reserved for component-specific errors.
// Each component package has a reserved range of error codes to avoid collisions.
// Components should define their error codes starting from their MinError constant.
const (
	// MinErrorComponentAws is the starting error code for AWS component errors.
	MinErrorComponentAws = ErrorParamEmpty + 10

	// MinErrorComponentDatabase is the starting error code for Database component errors.
	MinErrorComponentDatabase = MinErrorComponentAws + 10

	// MinErrorComponentHead is the starting error code for Header component errors.
	MinErrorComponentHead = MinErrorComponentDatabase + 10

	// MinErrorComponentHttp is the starting error code for HTTP server component errors.
	MinErrorComponentHttp = MinErrorComponentHead + 10

	// MinErrorComponentHttpCli is the starting error code for HTTP client component errors.
	MinErrorComponentHttpCli = MinErrorComponentHttp + 10

	// MinErrorComponentLdap is the starting error code for LDAP component errors.
	MinErrorComponentLdap = MinErrorComponentHttpCli + 10

	// MinErrorComponentLog is the starting error code for Logger component errors.
	MinErrorComponentLog = MinErrorComponentLdap + 10

	// MinErrorComponentMail is the starting error code for Mail component errors.
	MinErrorComponentMail = MinErrorComponentLog + 10

	// MinErrorComponentNats is the starting error code for NATS component errors.
	MinErrorComponentNats = MinErrorComponentMail + 10

	// MinErrorComponentNutsDB is the starting error code for NutsDB component errors.
	MinErrorComponentNutsDB = MinErrorComponentNats + 10

	// MinErrorComponentRequest is the starting error code for Request component errors.
	MinErrorComponentRequest = MinErrorComponentNutsDB + 10

	// MinErrorComponentSmtp is the starting error code for SMTP component errors.
	MinErrorComponentSmtp = MinErrorComponentRequest + 10

	// MinErrorComponentTls is the starting error code for TLS component errors.
	MinErrorComponentTls = MinErrorComponentSmtp + 10
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
