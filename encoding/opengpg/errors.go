/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package opengpg

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the OpenGPG package.
// Each error code is registered with liberr to provide standardized, human-readable
// messages and proper error chaining across the OpenGPG subsystem.
//
// ErrorParamEmpty — A required parameter (e.g., Identity pointer) was nil or zero.
// ErrorIdentityInvalid — The Identity struct is nil, uninitialized, or yields an
// invalid key pair (e.g., no usable public or private key material).
// ErrorPublicKeyInvalid — The public key material could not be parsed or is corrupt.
// ErrorPrivateKeyInvalid — The private key material could not be parsed or is corrupt.
const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgEncodingOpenGPG
	ErrorIdentityInvalid
	ErrorPublicKeyInvalid
	ErrorPrivateKeyInvalid
)

// init verifies that the base error code range does not collide with other packages
// and registers the error-to-message mapping with the global error registry. A panic
// halts startup if a collision is detected, forcing a developer fix before the
// binary links.
func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/encoding/opengpg"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage maps an error code to its human-readable description. If an unknown
// code is supplied, liberr.NullMessage is returned as a fallback.
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "required parameter is nil or empty"
	case ErrorIdentityInvalid:
		return "identity is empty or invalid"
	case ErrorPublicKeyInvalid:
		return "public key is empty or invalid"
	case ErrorPrivateKeyInvalid:
		return "private key is empty or invalid"
	}

	return liberr.NullMessage
}
