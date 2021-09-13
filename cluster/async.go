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

import (
	dgbclt "github.com/lni/dragonboat/v3"
	dgbcli "github.com/lni/dragonboat/v3/client"
	liberr "github.com/nabbar/golib/errors"
)

func (c *cRaft) AsyncPropose(session *dgbcli.Session, cmd []byte) (*dgbclt.RequestState, liberr.Error) {
	r, e := c.nodeHost.Propose(session, cmd, c.timeoutCmdASync)

	if e != nil {
		return r, ErrorCommandASync.ErrorParent(c.getErrorCommand("Propose"), e)
	}

	return r, nil
}

func (c *cRaft) AsyncProposeSession(session *dgbcli.Session) (*dgbclt.RequestState, liberr.Error) {
	r, e := c.nodeHost.ProposeSession(session, c.timeoutCmdASync)

	if e != nil {
		return r, ErrorCommandASync.ErrorParent(c.getErrorCommand("ProposeSession"), e)
	}

	return r, nil
}

func (c *cRaft) AsyncReadIndex() (*dgbclt.RequestState, liberr.Error) {
	r, e := c.nodeHost.ReadIndex(c.config.ClusterID, c.timeoutCmdASync)

	if e != nil {
		return r, ErrorCommandASync.ErrorParent(c.getErrorCluster(), c.getErrorCommand("ReadIndex"), e)
	}

	return r, nil
}

func (c *cRaft) AsyncRequestCompaction(nodeID uint64) (*dgbclt.SysOpState, liberr.Error) {
	var er error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		er = c.getErrorNode()
	} else {
		er = c.getErrorNodeTarget(nodeID)
	}

	r, e := c.nodeHost.RequestCompaction(c.config.ClusterID, nodeID)

	if e != nil {
		return r, ErrorCommandASync.ErrorParent(c.getErrorCluster(), er, c.getErrorCommand("RequestCompaction"), e)
	}

	return r, nil
}
