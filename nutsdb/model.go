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

package nutsdb

import (
	"context"
	"sync/atomic"
	"time"

	libsh "github.com/nabbar/golib/shell"

	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
)

type ndb struct {
	c Config
	t *atomic.Value
	r *atomic.Value
}

func (n *ndb) createNodeMachine(node uint64, cluster uint64) dgbstm.IOnDiskStateMachine {
	var (
		err liberr.Error
		opt Options
	)

	if opt, err = n.c.GetOptions(); err != nil {
		panic(err)
	}

	return newNode(node, cluster, opt, n.setRunning)
}

func (n *ndb) newCluster() liberr.Error {
	var (
		clu libclu.Cluster
		err liberr.Error
		cfg libclu.Config
	)

	if i := n.t.Load(); i != nil {
		if err = n.Shutdown(); err != nil {
			return err
		}
	}

	if cfg, err = n.c.GetConfigCluster(); err != nil {
		return err
	}

	clu, err = libclu.NewCluster(cfg, nil)

	if err != nil {
		return err
	}

	clu.SetFctCreateSTMOnDisk(n.createNodeMachine)
	n.t.Store(clu)
	return nil
}

func (n *ndb) IsRunning() bool {
	if i := n.r.Load(); i == nil {
		return false
	} else if b, ok := i.(bool); !ok {
		return false
	} else {
		return b
	}
}

func (n *ndb) setRunning(state bool) {
	if n == nil || n.r == nil {
		return
	} else {
		n.r.Store(state)
	}
}

func (n *ndb) IsReady(ctx context.Context) bool {
	if m, e := n.Cluster().SyncGetClusterMembership(ctx); e != nil || m == nil || len(m.Nodes) < 1 {
		return false
	}

	if _, ok, e := n.Cluster().GetLeaderID(); e != nil || !ok {
		return false
	} else {
		return true
	}
}

func (n *ndb) WaitReady(ctx context.Context, tick time.Duration) {
	for {
		if n.IsRunning() && n.IsReady(ctx) {
			return
		}

		time.Sleep(tick)
	}
}

func (n *ndb) Listen() liberr.Error {
	var (
		c libclu.Cluster
		e liberr.Error
	)

	if c = n.Cluster(); c == nil {
		if e = n.newCluster(); e != nil {
			return e
		} else if c = n.Cluster(); c == nil {
			return ErrorClusterInit.Error(nil)
		}
	}

	if e = c.ClusterStart(len(n.c.Cluster.InitMember) < 1); e != nil {
		return e
	}

	n.t.Store(c)
	return nil
}

func (n *ndb) Restart() liberr.Error {
	return n.Listen()
}

func (n *ndb) ForceRestart() {
	n.ForceShutdown()
	_ = n.Listen()
}

func (n *ndb) Shutdown() liberr.Error {
	if i := n.t.Load(); i == nil {
		return nil
	} else if c, ok := i.(libclu.Cluster); !ok {
		return nil
	} else if !c.HasNodeInfo(0) {
		return nil
	} else if err := c.NodeStop(0); err != nil {
		return err
	} else {
		return nil
	}
}

func (n *ndb) ForceShutdown() {
	if err := n.Shutdown(); err == nil {
		return
	} else if i := n.t.Load(); i == nil {
		return
	} else if c, ok := i.(libclu.Cluster); !ok {
		return
	} else if !c.HasNodeInfo(0) {
		return
	} else {
		_ = c.ClusterStop(true)
	}
}

func (n *ndb) Cluster() libclu.Cluster {
	if i := n.t.Load(); i == nil {
		return nil
	} else if c, ok := i.(libclu.Cluster); !ok {
		return nil
	} else {
		return c
	}
}

func (n *ndb) Client(ctx context.Context, tickSync time.Duration) Client {
	return &clientNutDB{
		x: ctx,
		t: tickSync,
		c: n.Cluster,
		w: n.WaitReady,
	}
}

func (n *ndb) ShellCommand(ctx func() context.Context, tickSync time.Duration) []libsh.Command {
	var (
		res = make([]libsh.Command, 0)
		cli func() Client
	)

	cli = func() Client {
		x := ctx()

		if x.Err() != nil {
			return nil
		}

		return n.Client(x, tickSync)
	}

	res = append(res, newShellCommand(CmdPut, cli))
	res = append(res, newShellCommand(CmdPutWithTimestamp, cli))
	res = append(res, newShellCommand(CmdGet, cli))
	res = append(res, newShellCommand(CmdGetAll, cli))
	res = append(res, newShellCommand(CmdRangeScan, cli))
	res = append(res, newShellCommand(CmdPrefixScan, cli))
	res = append(res, newShellCommand(CmdPrefixSearchScan, cli))
	res = append(res, newShellCommand(CmdDelete, cli))
	res = append(res, newShellCommand(CmdFindTxIDOnDisk, cli))
	res = append(res, newShellCommand(CmdFindOnDisk, cli))
	res = append(res, newShellCommand(CmdFindLeafOnDisk, cli))
	res = append(res, newShellCommand(CmdSAdd, cli))
	res = append(res, newShellCommand(CmdSRem, cli))
	res = append(res, newShellCommand(CmdSAreMembers, cli))
	res = append(res, newShellCommand(CmdSIsMember, cli))
	res = append(res, newShellCommand(CmdSMembers, cli))
	res = append(res, newShellCommand(CmdSHasKey, cli))
	res = append(res, newShellCommand(CmdSPop, cli))
	res = append(res, newShellCommand(CmdSCard, cli))
	res = append(res, newShellCommand(CmdSDiffByOneBucket, cli))
	res = append(res, newShellCommand(CmdSDiffByTwoBuckets, cli))
	res = append(res, newShellCommand(CmdSMoveByOneBucket, cli))
	res = append(res, newShellCommand(CmdSMoveByTwoBuckets, cli))
	res = append(res, newShellCommand(CmdSUnionByOneBucket, cli))
	res = append(res, newShellCommand(CmdSUnionByTwoBuckets, cli))
	res = append(res, newShellCommand(CmdRPop, cli))
	res = append(res, newShellCommand(CmdRPeek, cli))
	res = append(res, newShellCommand(CmdRPush, cli))
	res = append(res, newShellCommand(CmdLPush, cli))
	res = append(res, newShellCommand(CmdLPop, cli))
	res = append(res, newShellCommand(CmdLPeek, cli))
	res = append(res, newShellCommand(CmdLSize, cli))
	res = append(res, newShellCommand(CmdLRange, cli))
	res = append(res, newShellCommand(CmdLRem, cli))
	res = append(res, newShellCommand(CmdLSet, cli))
	res = append(res, newShellCommand(CmdLTrim, cli))
	res = append(res, newShellCommand(CmdZAdd, cli))
	res = append(res, newShellCommand(CmdZMembers, cli))
	res = append(res, newShellCommand(CmdZCard, cli))
	res = append(res, newShellCommand(CmdZCount, cli))
	res = append(res, newShellCommand(CmdZPopMax, cli))
	res = append(res, newShellCommand(CmdZPopMin, cli))
	res = append(res, newShellCommand(CmdZPeekMax, cli))
	res = append(res, newShellCommand(CmdZPeekMin, cli))
	res = append(res, newShellCommand(CmdZRangeByScore, cli))
	res = append(res, newShellCommand(CmdZRangeByRank, cli))
	res = append(res, newShellCommand(CmdZRem, cli))
	res = append(res, newShellCommand(CmdZRemRangeByRank, cli))
	res = append(res, newShellCommand(CmdZRank, cli))
	res = append(res, newShellCommand(CmdZRevRank, cli))
	res = append(res, newShellCommand(CmdZScore, cli))
	res = append(res, newShellCommand(CmdZGetByKey, cli))

	return res
}