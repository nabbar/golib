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
	"fmt"
	"io"
	"time"

	dgbclt "github.com/lni/dragonboat/v3"
	dgbcli "github.com/lni/dragonboat/v3/client"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	liberr "github.com/nabbar/golib/errors"
)

type cRaft struct {
	memberInit      map[uint64]dgbclt.Target
	fctCreate       interface{}
	config          dgbcfg.Config
	nodeHost        *dgbclt.NodeHost
	timeoutCmdSync  time.Duration
	timeoutCmdASync time.Duration
}

func (c *cRaft) getErrorCluster() error {
	//nolint #goerr113
	return fmt.Errorf("cluster: %v", c.config.ClusterID)
}

func (c *cRaft) getErrorNode() error {
	//nolint #goerr113
	return fmt.Errorf("node: %v", c.config.NodeID)
}

func (c *cRaft) getErrorNodeTarget(target uint64) error {
	//nolint #goerr113
	return fmt.Errorf("target node: %v", target)
}

func (c *cRaft) getErrorCommand(cmd string) error {
	//nolint #goerr113
	return fmt.Errorf("command: %v", cmd)
}

func (c *cRaft) GetConfig() dgbcfg.Config {
	return c.config
}

func (c *cRaft) SetConfig(cfg dgbcfg.Config) {
	c.config = cfg
}

func (c *cRaft) GetFctCreate() dgbstm.CreateStateMachineFunc {
	if f, ok := c.fctCreate.(dgbstm.CreateStateMachineFunc); ok {
		return f
	}
	return nil
}

func (c *cRaft) GetFctCreateConcurrent() dgbstm.CreateConcurrentStateMachineFunc {
	if f, ok := c.fctCreate.(dgbstm.CreateConcurrentStateMachineFunc); ok {
		return f
	}
	return nil
}

func (c *cRaft) GetFctCreateOnDisk() dgbstm.CreateOnDiskStateMachineFunc {
	if f, ok := c.fctCreate.(dgbstm.CreateOnDiskStateMachineFunc); ok {
		return f
	}
	return nil
}

func (c *cRaft) SetFctCreate(fctCreate interface{}) {
	c.fctCreate = fctCreate
}

func (c *cRaft) SetFctCreateSTM(fctCreate dgbstm.CreateStateMachineFunc) {
	c.fctCreate = fctCreate
}

func (c *cRaft) SetFctCreateSTMConcurrent(fctCreate dgbstm.CreateConcurrentStateMachineFunc) {
	c.fctCreate = fctCreate
}

func (c *cRaft) SetFctCreateSTMOnDisk(fctCreate dgbstm.CreateOnDiskStateMachineFunc) {
	c.fctCreate = fctCreate
}

func (c *cRaft) GetMemberInit() map[uint64]dgbclt.Target {
	return c.memberInit
}

func (c *cRaft) SetMemberInit(memberList map[uint64]dgbclt.Target) {
	c.memberInit = memberList
}

func (c *cRaft) SetTimeoutCommandSync(timeout time.Duration) {
	c.timeoutCmdSync = timeout
}

func (c *cRaft) SetTimeoutCommandASync(timeout time.Duration) {
	c.timeoutCmdASync = timeout
}

func (c *cRaft) GetNodeHostConfig() dgbcfg.NodeHostConfig {
	return c.nodeHost.NodeHostConfig()
}

func (c *cRaft) RaftAddress() string {
	return c.nodeHost.RaftAddress()
}

func (c *cRaft) ID() string {
	return c.nodeHost.ID()
}

func (c *cRaft) ClusterStart(join bool) liberr.Error {
	err := ErrorNodeHostStart.Error(nil)

	if join {
		err = ErrorNodeHostJoin.Error(nil)
	}

	if f, ok := c.fctCreate.(dgbstm.CreateStateMachineFunc); ok {
		err.AddParent(c.nodeHost.StartCluster(c.memberInit, join, f, c.config))
	} else if f, ok := c.fctCreate.(dgbstm.CreateConcurrentStateMachineFunc); ok {
		err.AddParent(c.nodeHost.StartConcurrentCluster(c.memberInit, join, f, c.config))
	} else if f, ok := c.fctCreate.(dgbstm.CreateOnDiskStateMachineFunc); ok {
		err.AddParent(c.nodeHost.StartOnDiskCluster(c.memberInit, join, f, c.config))
	} else {
		//nolint #goerr113
		return ErrorParamsMismatching.ErrorParent(fmt.Errorf("create function is not one of type of CreateStateMachineFunc, CreateConcurrentStateMachineFunc, CreateOnDiskStateMachineFunc"))
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *cRaft) ClusterStop(force bool) liberr.Error {
	e := c.nodeHost.StopCluster(c.config.ClusterID)

	if e != nil && !force {
		return ErrorNodeHostStop.ErrorParent(c.getErrorCluster(), e)
	}

	c.nodeHost.Stop()
	return nil
}

func (c *cRaft) ClusterRestart(force bool) liberr.Error {
	if err := c.ClusterStop(force); err != nil {
		return ErrorNodeHostRestart.Error(err)
	}

	return c.ClusterStart(false)
}

func (c *cRaft) NodeStop(target uint64) liberr.Error {
	var en error
	if target == 0 {
		target = c.config.NodeID
		en = c.getErrorNode()
	} else {
		en = c.getErrorNodeTarget(target)
	}

	e := c.nodeHost.StopNode(c.config.ClusterID, target)

	if e != nil {
		return ErrorNodeHostStop.ErrorParent(c.getErrorCluster(), en, e)
	}

	return nil
}

func (c *cRaft) NodeRestart(force bool) liberr.Error {
	var join = false

	if l, ok, err := c.GetLeaderID(); err == nil && ok && l == c.config.NodeID {
		join = true
		var sErr = ErrorNodeHostRestart.ErrorParent(c.getErrorCluster(), c.getErrorNode())
		for id, nd := range c.memberInit {
			if id == c.config.NodeID {
				continue
			}
			if nd == c.RaftAddress() {
				continue
			}
			if err = c.RequestLeaderTransfer(id); err == nil {
				sErr.AddParentError(err)
				break
			}
		}
		if l, ok, err = c.GetLeaderID(); err == nil && ok && l == c.config.NodeID && !force {
			return sErr
		}
	} else if err == nil && ok {
		join = true
	}

	if err := c.NodeStop(0); err != nil && !force {
		return ErrorNodeHostRestart.Error(err)
	} else if err != nil && force {
		join = false
		_ = c.ClusterStop(true)
	}

	return c.ClusterStart(join)
}

func (c *cRaft) GetLeaderID() (leader uint64, valid bool, err liberr.Error) {
	var e error

	leader, valid, e = c.nodeHost.GetLeaderID(c.config.ClusterID)

	if e != nil {
		err = ErrorLeader.ErrorParent(c.getErrorCluster(), e)
	}

	return
}

func (c *cRaft) GetNoOPSession() *dgbcli.Session {
	return c.nodeHost.GetNoOPSession(c.config.ClusterID)
}

func (c *cRaft) GetNodeUser() (dgbclt.INodeUser, liberr.Error) {
	r, e := c.nodeHost.GetNodeUser(c.config.ClusterID)

	if e != nil {
		return nil, ErrorNodeUser.ErrorParent(c.getErrorCluster())
	}

	return r, nil
}

func (c *cRaft) HasNodeInfo(nodeId uint64) bool {
	if nodeId == 0 {
		nodeId = c.config.NodeID
	}

	return c.nodeHost.HasNodeInfo(c.config.ClusterID, nodeId)
}

func (c *cRaft) GetNodeHostInfo(opt dgbclt.NodeHostInfoOption) *dgbclt.NodeHostInfo {
	return c.nodeHost.GetNodeHostInfo(opt)
}

func (c *cRaft) StaleReadDangerous(query interface{}) (interface{}, error) {
	return c.nodeHost.StaleRead(c.config.ClusterID, query)
}

func (c *cRaft) RequestLeaderTransfer(targetNodeID uint64) liberr.Error {
	e := c.nodeHost.RequestLeaderTransfer(c.config.ClusterID, targetNodeID)

	if e != nil {
		return ErrorLeaderTransfer.ErrorParent(c.getErrorCluster(), c.getErrorNodeTarget(targetNodeID), e)
	}

	return nil
}

func (c *cRaft) HandlerMetrics(w io.Writer) {
	dgbclt.WriteHealthMetrics(w)
}
