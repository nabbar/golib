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

	liberr "github.com/nabbar/golib/errors"

	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	"github.com/xujiajun/nutsdb"
)

const (
	_BucketRaftAdmin           = "_raft"
	_AdminKey_RaftAppliedIndex = "_raft_applied_index"
	_BucketData                = "_data"
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

func (n *nutsNode) newTx(writable bool) (*nutsdb.Tx, liberr.Error) {
	if n == nil || n.d == nil {
		return nil, ErrorDatabaseClosed.Error(nil)
	}
	if i := n.d.Load(); i == nil {
		return nil, ErrorDatabaseClosed.Error(nil)
	} else if db, ok := i.(*nutsdb.DB); !ok {
		return nil, ErrorDatabaseClosed.Error(nil)
	} else if tx, e := db.Begin(writable); e != nil {
		return nil, ErrorTransactionInit.ErrorParent(e)
	} else {
		return tx, nil
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

func (n *nutsNode) applyRaftLogIndexLastApplied(idx uint64) error {
	var (
		b   []byte
		e   error
		tx  *nutsdb.Tx
		err liberr.Error
	)

	binary.LittleEndian.PutUint64(b, idx)

	if tx, err = n.newTx(true); err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if e = tx.Put(_BucketRaftAdmin, []byte(_AdminKey_RaftAppliedIndex), b, 0); e != nil {
		return ErrorTransactionPutKey.ErrorParent(e)
	} else if e = tx.Commit(); e != nil {
		return ErrorTransactionCommit.ErrorParent(e)
	} else {
		return nil
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

func (n *nutsNode) Update(logEntry []dgbstm.Entry) ([]dgbstm.Entry, error) {
	var (
		e error

		tx *nutsdb.Tx
		kv *DataKV

		err liberr.Error
		idx int
		ent dgbstm.Entry
	)

	if tx, err = n.newTx(true); err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	for idx, ent = range logEntry {
		if kv, err = DataKVFromJson(ent.Cmd); err != nil {
			logEntry[idx].Result = dgbstm.Result{
				Value: 0,
				Data:  nil,
			}
			return logEntry, err
		} else if err = kv.SetToTx(tx, _BucketData); err != nil {
			logEntry[idx].Result = dgbstm.Result{
				Value: 0,
				Data:  nil,
			}
			return logEntry, err
		} else {
			logEntry[idx].Result = dgbstm.Result{
				Value: uint64(idx),
				Data:  nil,
			}
		}
	}

	if e = tx.Commit(); e != nil {
		return logEntry, ErrorTransactionCommit.ErrorParent(e)
	}

	return logEntry, n.applyRaftLogIndexLastApplied(logEntry[len(logEntry)-1].Index)
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
		return nil, ErrorDatabaseKeyInvalid.Error(nil)
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
	return nil
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
