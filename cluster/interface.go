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
	"github.com/lni/dragonboat/v3/client"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	sm "github.com/lni/dragonboat/v3/statemachine"
)

type Cluster interface {
	Config() dgbcfg.NodeHostConfig
	RaftAddress() string
	ID() string

	StartCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateStateMachineFunc, cfg dgbcfg.Config) error
	StartConcurrentCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateConcurrentStateMachineFunc, cfg dgbcfg.Config) error
	StartOnDiskCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateOnDiskStateMachineFunc, cfg dgbcfg.Config) error

	Stop()
	StopCluster(clusterID uint64) error
	StopNode(clusterID uint64, nodeID uint64) error

	SyncPropose(ctx context.Context, session *client.Session, cmd []byte) (sm.Result, error)
	SyncRead(ctx context.Context, clusterID uint64, query interface{}) (interface{}, error)

	GetClusterMembership(ctx context.Context, clusterID uint64) (*dgbclt.Membership, error)
	GetLeaderID(clusterID uint64) (uint64, bool, error)

	GetNoOPSession(clusterID uint64) *client.Session
	GetNewSession(ctx context.Context, clusterID uint64) (*client.Session, error)
	CloseSession(ctx context.Context, session *client.Session) error
	SyncGetSession(ctx context.Context, clusterID uint64) (*client.Session, error)
	SyncCloseSession(ctx context.Context, cs *client.Session) error

	Propose(session *client.Session, cmd []byte, timeout time.Duration) (*dgbclt.RequestState, error)
	ProposeSession(session *client.Session, timeout time.Duration) (*dgbclt.RequestState, error)

	ReadIndex(clusterID uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	ReadLocalNode(rs *dgbclt.RequestState, query interface{}) (interface{}, error)
	NAReadLocalNode(rs *dgbclt.RequestState, query []byte) ([]byte, error)

	StaleRead(clusterID uint64, query interface{}) (interface{}, error)

	SyncRequestSnapshot(ctx context.Context, clusterID uint64, opt dgbclt.SnapshotOption) (uint64, error)
	RequestSnapshot(clusterID uint64, opt dgbclt.SnapshotOption, timeout time.Duration) (*dgbclt.RequestState, error)

	RequestCompaction(clusterID uint64, nodeID uint64) (*dgbclt.SysOpState, error)

	SyncRequestDeleteNode(ctx context.Context, clusterID uint64, nodeID uint64, configChangeIndex uint64) error
	SyncRequestAddNode(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error
	SyncRequestAddObserver(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error
	SyncRequestAddWitness(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error

	//@todo continue at line 1148
}

func NewCluster(cfg *dgbcfg.NodeHostConfig) (Cluster, error) {
	if c, e := dgbclt.NewNodeHost(*cfg); e != nil {
		return nil, e
	} else {
		return &cRaft{
			c: c,
		}, nil
	}
}
