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
	"sync/atomic"

	dgbstm "github.com/lni/dragonboat/v3/statemachine"

	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
)

type ndb struct {
	c Config
	t *atomic.Value
	r *atomic.Value
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

func (n *ndb) Listen() liberr.Error {
	var (
		clu libclu.Cluster
		err liberr.Error
		opt Options
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

	if opt, err = n.c.GetOptions(); err != nil {
		return err
	}

	clu, err = libclu.NewCluster(cfg, func(node uint64, cluster uint64) dgbstm.IOnDiskStateMachine {
		return newNode(node, cluster, opt, n.setRunning)
	})

	if err != nil {
		return err
	}

	n.t.Store(clu)
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

func (n *ndb) Client() Client {
	//@TODO : implement me !!
	panic("implement me")
}
