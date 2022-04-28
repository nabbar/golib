/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package configAws

import (
	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorAwsError liberr.CodeError = iota + liberr.MinPkgAws + 40
	ErrorConfigLoader
	ErrorConfigValidator
	ErrorConfigJsonUnmarshall
	ErrorEndpointInvalid
	ErrorRegionInvalid
	ErrorRegionEndpointNotFound
	ErrorCredentialsInvalid
)

var isErrInit = liberr.ExistInMapMessage(ErrorAwsError)

func init() {
	liberr.RegisterIdFctMessage(ErrorAwsError, getMessage)
}

func IsErrorInit() bool {
	return isErrInit
}

func getMessage(code liberr.CodeError) string {
	switch code {
	case ErrorAwsError:
		return "calling aws api occurred a response error"
	case ErrorConfigLoader:
		return "calling AWS Default config Loader has occurred an error"
	case ErrorConfigValidator:
		return "invalid config, validation error"
	case ErrorConfigJsonUnmarshall:
		return "invalid json config, unmarshall error"
	case ErrorEndpointInvalid:
		return "the specified endpoint seems to be invalid"
	case ErrorRegionInvalid:
		return "the specified region seems to be invalid"
	case ErrorRegionEndpointNotFound:
		return "cannot find the endpoint for the specify region"
	case ErrorCredentialsInvalid:
		return "the specified credentials seems to be incorrect"
	}

	return liberr.UNK_MESSAGE
}
