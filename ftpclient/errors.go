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

package ftpclient

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamsEmpty liberr.CodeError = iota + liberr.MinPkgFTPClient
	ErrorValidatorError
	ErrorEndpointParser
	ErrorNotInitialized
	ErrorFTPConnection
	ErrorFTPConnectionCheck
	ErrorFTPLogin
	ErrorFTPCommand
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamsEmpty) {
		panic(fmt.Errorf("error code collision with package golib/ftpclient"))
	}
	liberr.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamsEmpty:
		return "given parameters is empty"
	case ErrorValidatorError:
		return "ftp client : invalid config"
	case ErrorEndpointParser:
		return "ftp client : cannot understand given endpoint"
	case ErrorNotInitialized:
		return "ftp client : instance seems to not be initialized"
	case ErrorFTPLogin:
		return "ftp client : cannot login to server"
	case ErrorFTPConnectionCheck:
		return "ftp client : check connection (NOOP) to server trigger an error"
	case ErrorFTPCommand:
		return "ftp client : command to server trigger an error"
	}

	return liberr.NullMessage
}
