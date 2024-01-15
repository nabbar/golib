/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package certificates

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgCertificate
	ErrorFileStat
	ErrorFileRead
	ErrorFileEmpty
	ErrorCertAppend
	ErrorCertKeyPairLoad
	ErrorCertKeyPairParse
	ErrorValidatorError
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/certificates"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorFileStat:
		return "cannot get file stat"
	case ErrorFileRead:
		return "cannot read file"
	case ErrorFileEmpty:
		return "given file is empty"
	case ErrorCertAppend:
		return "cannot append PEM file"
	case ErrorCertKeyPairLoad:
		return "cannot X509 parsing certificate string"
	case ErrorCertKeyPairParse:
		return "cannot x509 loading certificate file"
	case ErrorValidatorError:
		return "tls : invalid config"
	}

	return liberr.NullMessage
}
