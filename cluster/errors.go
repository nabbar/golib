//go:build amd64 || arm64 || arm64be || ppc64 || ppc64le || mips64 || mips64le || riscv64 || s390x || sparc64 || wasm
// +build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

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

import liberr "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty liberr.CodeError = iota + liberr.MinPkgCluster
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
	ErrorValidateConfig
	ErrorValidateCluster
	ErrorValidateNode
	ErrorValidateGossip
	ErrorValidateExpert
	ErrorValidateEngine
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = liberr.ExistInMapMessage(ErrorParamsEmpty)
	liberr.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case liberr.UNK_ERROR:
		return ""
	case ErrorParamsEmpty:
		return "at least one given parameter is empty"
	case ErrorParamsMissing:
		return "at least one given parameter is missing"
	case ErrorParamsMismatching:
		return "at least one given parameter does not match the awaiting type"
	case ErrorLeader:
		return "unable to retrieve cluster leader"
	case ErrorLeaderTransfer:
		return "unable to transfer cluster leader"
	case ErrorNodeUser:
		return "unable to retrieve node user"
	case ErrorNodeHostNew:
		return "unable to init new NodeHost"
	case ErrorNodeHostStart:
		return "unable to start cluster"
	case ErrorNodeHostJoin:
		return "unable to join cluster"
	case ErrorNodeHostStop:
		return "unable to stop cluster or node"
	case ErrorNodeHostRestart:
		return "unable to restart node properly"
	case ErrorCommandSync:
		return "unable to call synchronous command"
	case ErrorCommandASync:
		return "unable to call asynchronous command"
	case ErrorCommandLocal:
		return "unable to call local command"
	case ErrorValidateConfig:
		return "config seems to be invalid"
	case ErrorValidateCluster:
		return "cluster config seems to be invalid"
	case ErrorValidateNode:
		return "node config seems to be invalid"
	case ErrorValidateGossip:
		return "gossip config seems to be invalid"
	case ErrorValidateExpert:
		return "expert config seems to be invalid"
	case ErrorValidateEngine:
		return "engine config seems to be invalid"
	}

	return ""
}
