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
	dgbcfg "github.com/lni/dragonboat/v3/config"
	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	liberr "github.com/nabbar/golib/errors"
)

type Cluster interface {
	GetConfig() dgbcfg.Config
	SetConfig(cfg dgbcfg.Config)
	GetNodeHostConfig() dgbcfg.NodeHostConfig

	GetFctCreate() dgbstm.CreateStateMachineFunc
	GetFctCreateConcurrent() dgbstm.CreateConcurrentStateMachineFunc
	GetFctCreateOnDisk() dgbstm.CreateOnDiskStateMachineFunc
	SetFctCreate(fctCreate interface{})
	SetFctCreateSTM(fctCreate dgbstm.CreateStateMachineFunc)
	SetFctCreateSTMConcurrent(fctCreate dgbstm.CreateConcurrentStateMachineFunc)
	SetFctCreateSTMOnDisk(fctCreate dgbstm.CreateOnDiskStateMachineFunc)

	GetMemberInit() map[uint64]dgbclt.Target
	SetMemberInit(memberList map[uint64]dgbclt.Target)

	SetTimeoutCommandSync(timeout time.Duration)
	SetTimeoutCommandASync(timeout time.Duration)

	HasNodeInfo(nodeId uint64) bool
	RaftAddress() string
	ID() string

	ClusterStart(join bool) liberr.Error
	ClusterStop(force bool) liberr.Error
	ClusterRestart(force bool) liberr.Error

	NodeStop(target uint64) liberr.Error
	NodeRestart(force bool) liberr.Error

	GetLeaderID() (leader uint64, valid bool, err liberr.Error)
	GetNoOPSession() *dgbcli.Session
	GetNodeHostInfo(opt dgbclt.NodeHostInfoOption) *dgbclt.NodeHostInfo
	RequestLeaderTransfer(targetNodeID uint64) liberr.Error

	StaleReadDangerous(query interface{}) (interface{}, error)

	SyncPropose(parent context.Context, session *dgbcli.Session, cmd []byte) (dgbstm.Result, liberr.Error)
	SyncRead(parent context.Context, query interface{}) (interface{}, liberr.Error)
	SyncGetClusterMembership(parent context.Context) (*dgbclt.Membership, liberr.Error)
	SyncGetSession(parent context.Context) (*dgbcli.Session, liberr.Error)
	SyncCloseSession(parent context.Context, cs *dgbcli.Session) liberr.Error
	SyncRequestSnapshot(parent context.Context, opt dgbclt.SnapshotOption) (uint64, liberr.Error)
	SyncRequestDeleteNode(parent context.Context, nodeID uint64, configChangeIndex uint64) liberr.Error
	SyncRequestAddNode(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error
	SyncRequestAddObserver(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error
	SyncRequestAddWitness(parent context.Context, nodeID uint64, target string, configChangeIndex uint64) liberr.Error
	SyncRemoveData(parent context.Context, nodeID uint64) liberr.Error

	AsyncPropose(session *dgbcli.Session, cmd []byte) (*dgbclt.RequestState, liberr.Error)
	AsyncProposeSession(session *dgbcli.Session) (*dgbclt.RequestState, liberr.Error)
	AsyncReadIndex() (*dgbclt.RequestState, liberr.Error)
	AsyncRequestCompaction(nodeID uint64) (*dgbclt.SysOpState, liberr.Error)

	LocalReadNode(rs *dgbclt.RequestState, query interface{}) (interface{}, liberr.Error)
	LocalNAReadNode(rs *dgbclt.RequestState, query []byte) ([]byte, liberr.Error)
}

func NewCluster(cfg Config, fctCreate interface{}) (Cluster, liberr.Error) {
	c := &cRaft{
		memberInit:      cfg.GetInitMember(),
		fctCreate:       fctCreate,
		config:          cfg.GetDGBConfigCluster(),
		nodeHost:        nil,
		timeoutCmdSync:  100 * time.Millisecond,
		timeoutCmdASync: 1 * time.Second,
	}

	if n, e := dgbclt.NewNodeHost(cfg.GetDGBConfigNode()); e != nil {
		return nil, ErrorNodeHostNew.ErrorParent(e)
	} else {
		c.nodeHost = n
	}

	return c, nil
}
