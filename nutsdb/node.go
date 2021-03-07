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
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"

	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	"github.com/xujiajun/nutsdb"
)

const (
	_BucketRaftAdmin           = "_raft"
	_AdminKey_RaftAppliedIndex = "_raft_applied_index"
)

var (
	ErrClosed     = errors.New("database node is closed")
	ErrKeyInvalid = errors.New("given key is invalid")
)

func newNode(node uint64, cluster uint64, opt nutsdb.Options, fct func(state bool)) dgbstm.IOnDiskStateMachine {
	if fct == nil {
		fct = func(state bool) {}
	}

	return &nutsNode{
		n: node,
		c: cluster,
		o: opt,
		r: fct,
		d: new(atomic.Value),
	}
}

type nutsNode struct {
	n uint64 // nodeId
	c uint64 // clusterId
	o nutsdb.Options
	r func(state bool)
	d *atomic.Value
}

func (n *nutsNode) setRunning(state bool) {
	if n != nil && n.r != nil {
		n.r(state)
	}
}

func (n *nutsNode) newTx(writable bool) (*nutsdb.Tx, error) {
	if n == nil || n.d == nil {
		return nil, ErrClosed
	}
	if i := n.d.Load(); i == nil {
		return nil, ErrClosed
	} else if db, ok := i.(*nutsdb.DB); !ok {
		return nil, ErrClosed
	} else {
		return db.Begin(writable)
	}
}

func (n *nutsNode) getRaftLogIndexLastApplied() (idxRaftlog uint64, err error) {
	var (
		tx *nutsdb.Tx
		en *nutsdb.Entry
	)

	if tx, err = n.newTx(false); err != nil {
		return 0, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if en, err = tx.Get(_BucketRaftAdmin, []byte(_AdminKey_RaftAppliedIndex)); err != nil && errors.Is(err, nutsdb.ErrBucketEmpty) {
		return 0, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrNotFoundKey) {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else if en.IsZero() {
		return 0, nil
	} else {
		return binary.LittleEndian.Uint64(en.Value), nil
	}
}

func (n *nutsNode) Open(stopc <-chan struct{}) (idxRaftlog uint64, err error) {
	var db *nutsdb.DB

	if db, err = nutsdb.Open(n.o); err != nil {
		return 0, err
	} else {
		n.d.Store(db)
		n.setRunning(true)

		if idxRaftlog, err = n.getRaftLogIndexLastApplied(); err != nil {
			_ = n.Close()
		}

		return
	}
}

func (n *nutsNode) Close() error {
	defer n.setRunning(false)

	if n == nil || n.d == nil {

		return nil
	} else if i := n.d.Load(); i == nil {
		return nil
	} else if db, ok := i.(*nutsdb.DB); !ok {
		return nil
	} else {
		return db.Close()
	}
}

func (n *nutsNode) Update([]dgbstm.Entry) ([]dgbstm.Entry, error) {
	// utiliser cbor marshal
	panic("implement me")
}

func (n *nutsNode) Lookup(key interface{}) (value interface{}, err error) {
	var (
		tx *nutsdb.Tx
		en *nutsdb.Entry
		bk []byte
	)

	if sk, ok := key.(string); ok {
		bk = []byte(sk)
	} else if bk, ok = key.([]byte); !ok {
		return nil, ErrKeyInvalid
	}

	if tx, err = n.newTx(false); err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if en, err = tx.Get(_AdminKey_RaftAppliedIndex, bk); err != nil && errors.Is(err, nutsdb.ErrNotFoundKey) {
		return nil, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrBucketEmpty) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return en.Value, nil
	}
}

func (n *nutsNode) Sync() error {
	panic("implement me")
}

func (n *nutsNode) PrepareSnapshot() (interface{}, error) {
	panic("implement me")
}

func (n *nutsNode) SaveSnapshot(i interface{}, writer io.Writer, i2 <-chan struct{}) error {
	panic("implement me")
}

func (n *nutsNode) RecoverFromSnapshot(reader io.Reader, i <-chan struct{}) error {
	panic("implement me")
}
