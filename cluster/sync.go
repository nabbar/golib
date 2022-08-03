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
	"context"
	"time"

	dgbclt "github.com/lni/dragonboat/v3"
	dgbcli "github.com/lni/dragonboat/v3/client"
	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	liberr "github.com/nabbar/golib/errors"
)

func (c *cRaft) syncCtxTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	var (
		ctx context.Context
		cnl context.CancelFunc
	)

	if parent != nil {
		ctx, cnl = context.WithDeadline(parent, time.Now().Add(c.timeoutCmdSync))
	} else {
		ctx, cnl = context.WithDeadline(context.Background(), time.Now().Add(c.timeoutCmdSync))
	}

	return ctx, cnl
}

func (c *cRaft) syncCtxCancel(cancel context.CancelFunc) {
	if cancel != nil {
		cancel()
	}
}

func (c *cRaft) SyncPropose(parent context.Context, session *dgbcli.Session, cmd []byte) (dgbstm.Result, liberr.Error) {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	r, e := c.nodeHost.SyncPropose(ctx, session, cmd)

	if e != nil {
		return r, ErrorCommandSync.ErrorParent(c.getErrorCommand("Propose"), e)
	}

	return r, nil
}

func (c *cRaft) SyncRead(parent context.Context, query interface{}) (interface{}, liberr.Error) {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	r, e := c.nodeHost.SyncRead(ctx, c.config.ClusterID, query)

	if e != nil {
		return r, ErrorCommandSync.ErrorParent(c.getErrorCluster(), c.getErrorCommand("Read"), e)
	}

	return r, nil
}

func (c *cRaft) SyncGetClusterMembership(parent context.Context) (*dgbclt.Membership, liberr.Error) {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	r, e := c.nodeHost.SyncGetClusterMembership(ctx, c.config.ClusterID)

	if e != nil {
		return r, ErrorCommandSync.ErrorParent(c.getErrorCluster(), c.getErrorCommand("GetClusterMembership"), e)
	}

	return r, nil
}

func (c *cRaft) SyncGetSession(parent context.Context) (*dgbcli.Session, liberr.Error) {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	r, e := c.nodeHost.SyncGetSession(ctx, c.config.ClusterID)

	if e != nil {
		return r, ErrorCommandSync.ErrorParent(c.getErrorCluster(), c.getErrorCommand("GetSession"), e)
	}

	return r, nil
}

func (c *cRaft) SyncCloseSession(parent context.Context, cs *dgbcli.Session) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	e := c.nodeHost.SyncCloseSession(ctx, cs)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCommand("CloseSession"), e)
	}

	return nil
}

func (c *cRaft) SyncRequestSnapshot(parent context.Context, opt dgbclt.SnapshotOption) (uint64, liberr.Error) {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	r, e := c.nodeHost.SyncRequestSnapshot(ctx, c.config.ClusterID, opt)

	if e != nil {
		return r, ErrorCommandSync.ErrorParent(c.getErrorCluster(), c.getErrorCommand("RequestSnapshot"), e)
	}

	return r, nil
}

func (c *cRaft) SyncRequestDeleteNode(parent context.Context, nodeID uint64, configChangeIndex uint64) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	var en error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(nodeID)
	}

	e := c.nodeHost.SyncRequestDeleteNode(ctx, c.config.ClusterID, nodeID, configChangeIndex)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCluster(), en, c.getErrorCommand("RequestDeleteNode"), e)
	}

	return nil
}

// nolint #dupl
func (c *cRaft) SyncRequestAddNode(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	var en error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(nodeID)
	}

	e := c.nodeHost.SyncRequestAddNode(ctx, c.config.ClusterID, nodeID, target, configChangeIndex)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCluster(), en, c.getErrorCommand("RequestAddNode"), e)
	}

	return nil
}

// nolint #dupl
func (c *cRaft) SyncRequestAddObserver(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	var en error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(nodeID)
	}

	e := c.nodeHost.SyncRequestAddObserver(ctx, c.config.ClusterID, nodeID, target, configChangeIndex)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCluster(), en, c.getErrorCommand("RequestAddObserver"), e)
	}

	return nil
}

// nolint #dupl
func (c *cRaft) SyncRequestAddWitness(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	var en error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(nodeID)
	}

	e := c.nodeHost.SyncRequestAddWitness(ctx, c.config.ClusterID, nodeID, target, configChangeIndex)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCluster(), en, c.getErrorCommand("RequestAddWitness"), e)
	}

	return nil
}

func (c *cRaft) SyncRemoveData(parent context.Context, nodeID uint64) liberr.Error {
	ctx, cnl := c.syncCtxTimeout(parent)
	defer c.syncCtxCancel(cnl)

	var en error
	if nodeID == 0 {
		nodeID = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(nodeID)
	}

	e := c.nodeHost.SyncRemoveData(ctx, c.config.ClusterID, nodeID)

	if e != nil {
		return ErrorCommandSync.ErrorParent(c.getErrorCluster(), en, c.getErrorCommand("RemoveData"), e)
	}

	return nil
}
