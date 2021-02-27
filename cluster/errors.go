/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package cluster

import "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty errors.CodeError = iota + errors.MinPkgCluster
	ErrorParamsMissing
	ErrorParamsMismatching
	ErrorLeader
	ErrorLeaderTransfer
	ErrorNodeUser
	ErrorNodeHostNew
	ErrorNodeHostStart
	ErrorNodeHostJoin
	ErrorNodeHostStop
	ErrorNodeHostRestart
	ErrorCommandSync
	ErrorCommandASync
	ErrorCommandLocal
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = errors.ExistInMapMessage(ErrorParamsEmpty)
	errors.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case errors.UNK_ERROR:
		return ""
	case ErrorParamsEmpty:
		return "at least on given parameters is empty"
	case ErrorParamsMissing:
		return "at least on given parameters is missing"
	case ErrorParamsMismatching:
		return "at least on given parameters is mismatching awaiting type"
	case ErrorLeader:
		return "enable to retrieve cluster leader"
	case ErrorLeaderTransfer:
		return "enable to transfer cluster leader"
	case ErrorNodeUser:
		return "enable to retrieve node user"
	case ErrorNodeHostNew:
		return "enable to init new NodeHost"
	case ErrorNodeHostStart:
		return "enable to start cluster"
	case ErrorNodeHostJoin:
		return "enable to join cluster"
	case ErrorNodeHostStop:
		return "enable to stop cluster or node"
	case ErrorNodeHostRestart:
		return "enable to restart node properly"
	case ErrorCommandSync:
		return "enable to call synchronous command"
	case ErrorCommandASync:
		return "enable to call asynchronous command"
	case ErrorCommandLocal:
		return "enable to call local command"
	}

	return ""
}
