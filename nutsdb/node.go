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
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	liblog "github.com/nabbar/golib/logger"

	dgbstm "github.com/lni/dragonboat/v3/statemachine"
	liberr "github.com/nabbar/golib/errors"
	"github.com/xujiajun/nutsdb"
)

const (
	_RaftBucket          = "_raft"
	_RaftKeyAppliedIndex = "_raft_applied_index"
	_WordByteTrue        = 0xff
)

func newNode(node uint64, cluster uint64, opt Options, fct func(state bool)) dgbstm.IOnDiskStateMachine {
	if fct == nil {
		fct = func(state bool) {}
	}

	o := new(atomic.Value)
	o.Store(opt)

	return &nutsNode{
		n: node,
		c: cluster,
		o: o,
		r: fct,
		d: new(atomic.Value),
	}
}

type nutsNode struct {
	m sync.Mutex       // mutex for struct var
	n uint64           // nodeId
	c uint64           // clusterId
	r func(state bool) // is running
	o *atomic.Value    // options nutsDB
	d *atomic.Value    // nutsDB database pointer
	l liblog.FuncLog   // logger
}

func (n *nutsNode) SetLogger(l liblog.FuncLog) {
	n.m.Lock()
	defer n.m.Unlock()

	if l != nil {
		n.l = l
	}
}

func (n *nutsNode) GetLogger() liblog.Logger {
	n.m.Lock()
	defer n.m.Unlock()

	if n.l != nil {
		return n.l()
	}

	return liblog.GetDefault()
}

func (n *nutsNode) setRunning(state bool) {
	n.m.Lock()
	defer n.m.Unlock()

	if n != nil && n.r != nil {
		n.r(state)
	}
}

func (n *nutsNode) newTx(writable bool) (*nutsdb.Tx, liberr.Error) {
	if db := n.getDb(); db == nil {
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

	if en, err = tx.Get(_RaftBucket, []byte(_RaftKeyAppliedIndex)); err != nil && errors.Is(err, nutsdb.ErrBucketNotFound) {
		return 0, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrBucketEmpty) {
		return 0, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrNotFoundKey) {
		return 0, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrKeyNotFound) {
		return 0, nil
	} else if err != nil && errors.Is(err, nutsdb.ErrKeyEmpty) {
		return 0, nil
	} else if err != nil && strings.HasPrefix(err.Error(), "not found bucket") {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else if en.IsZero() {
		return 0, nil
	} else {
		return n.btoi64(en.Value), nil
	}
}

func (n *nutsNode) i64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & _WordByteTrue)
	}
	return r
}

func (n *nutsNode) btoi64(val []byte) uint64 {
	r := uint64(0)
	for i := uint64(0); i < 8; i++ {
		r |= uint64(val[i]) << (8 * i)
	}
	return r
}

func (n *nutsNode) applyRaftLogIndexLastApplied(idx uint64) error {
	var (
		e   error
		tx  *nutsdb.Tx
		err liberr.Error
	)

	if tx, err = n.newTx(true); err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if e = tx.Put(_RaftBucket, []byte(_RaftKeyAppliedIndex), n.i64tob(idx), 0); e != nil {
		return ErrorTransactionPutKey.ErrorParent(e)
	} else if e = tx.Commit(); e != nil {
		return ErrorTransactionCommit.ErrorParent(e)
	} else {
		return nil
	}
}

// Open @TODO : analyze channel role !!
func (n *nutsNode) Open(stopc <-chan struct{}) (idxRaftlog uint64, err error) {
	var (
		opt Options
		db  *nutsdb.DB
	)

	if opt = n.getOptions(); opt == nil {
		return 0, ErrorValidateConfig.Error(nil)
	}

	if db, err = nutsdb.Open(opt.NutsDBOptions()); err != nil {
		return 0, err
	} else {
		n.setDb(db)
		n.setRunning(true)

		if idxRaftlog, err = n.getRaftLogIndexLastApplied(); err != nil {
			_ = n.Close()
		}

		return
	}
}

func (n *nutsNode) Close() error {
	defer n.setRunning(false)

	if db := n.getDb(); db != nil {
		err := db.Close()
		n.setDb(db)
		return err
	}

	return nil
}

func (n *nutsNode) Update(logEntry []dgbstm.Entry) ([]dgbstm.Entry, error) {
	var (
		e error

		tx *nutsdb.Tx
		kv *CommandRequest

		err liberr.Error
		res []byte
		idx int
		ent dgbstm.Entry
	)

	if tx, err = n.newTx(true); err != nil {
		return nil, err
	}

	for idx, ent = range logEntry {
		if kv, err = NewCommandByDecode(n.GetLogger, ent.Cmd); err != nil {
			logEntry[idx].Result = dgbstm.Result{
				Value: 0,
				Data:  nil,
			}

			_ = tx.Rollback()
			return logEntry, err
		}

		if res, err = kv.Run(tx); err != nil {
			logEntry[idx].Result = dgbstm.Result{
				Value: 0,
				Data:  nil,
			}

			_ = tx.Rollback()
			return logEntry, err
		} else {
			logEntry[idx].Result = dgbstm.Result{
				Value: uint64(idx),
				Data:  res,
			}
		}
	}

	if e = tx.Commit(); e != nil {
		_ = tx.Rollback()
		return logEntry, ErrorTransactionCommit.ErrorParent(e)
	}

	return logEntry, n.applyRaftLogIndexLastApplied(logEntry[len(logEntry)-1].Index)
}

func (n *nutsNode) Lookup(query interface{}) (value interface{}, err error) {
	var (
		t *nutsdb.Tx
		r *CommandResponse
		c *CommandRequest
		e liberr.Error

		ok bool
	)

	if t, e = n.newTx(true); e != nil {
		return nil, e
	}

	defer func() {
		if t != nil {
			_ = t.Rollback()
		}
	}()

	if c, ok = query.(*CommandRequest); !ok {
		return nil, ErrorCommandInvalid.Error(nil)
	} else if r, e = c.RunLocal(t); e != nil {
		return nil, e
	} else {
		return r, nil
	}
}

func (n *nutsNode) Sync() error {
	return nil
}

func (n *nutsNode) PrepareSnapshot() (interface{}, error) {
	var sh = newSnap()

	if opt := n.getOptions(); opt == nil {
		return nil, ErrorValidateConfig.Error(nil)
	} else if db := n.getDb(); db == nil {
		sh.Finish()
		return nil, ErrorDatabaseClosed.Error(nil)
	} else if err := sh.Prepare(opt, db); err != nil {
		sh.Finish()
		return nil, ErrorDatabaseBackup.ErrorParent(err)
	} else {
		return sh, nil
	}
}

func (n *nutsNode) SaveSnapshot(i interface{}, writer io.Writer, c <-chan struct{}) error {
	if i == nil {
		return ErrorParamsEmpty.Error(nil)
	} else if sh, ok := snapCast(i); !ok {
		return ErrorParamsMismatching.Error(nil)
	} else if opt := n.getOptions(); opt == nil {
		return ErrorValidateConfig.Error(nil)
	} else if err := sh.Save(opt, writer); err != nil {
		sh.Finish()
		return err
	} else {
		sh.Finish()
	}

	return nil
}

func (n *nutsNode) RecoverFromSnapshot(reader io.Reader, c <-chan struct{}) error {
	var (
		sh  = newSnap()
		opt = n.getOptions()
	)

	defer sh.Finish()

	if opt == nil {
		return ErrorValidateConfig.Error(nil)
	}

	if err := sh.Load(opt, reader); err != nil {
		return err
	}

	if db := n.getDb(); db != nil {
		_ = db.Close()
		n.setDb(db)
	}

	if err := sh.Apply(opt); err != nil {
		return err
	}

	//@TODO : check channel is ok....
	if _, err := n.Open(nil); err != nil {
		return err
	}

	return nil
}

func (n *nutsNode) getDb() *nutsdb.DB {
	n.m.Lock()
	defer n.m.Unlock()

	if n.d == nil {
		return nil
	} else if i := n.d.Load(); i == nil {
		return nil
	} else if db, ok := i.(*nutsdb.DB); !ok {
		return nil
	} else {
		return db
	}
}

func (n *nutsNode) setDb(db *nutsdb.DB) {
	n.m.Lock()
	defer n.m.Unlock()

	if n.d == nil {
		n.d = new(atomic.Value)
	}

	n.d.Store(db)
}

func (n *nutsNode) getOptions() Options {
	n.m.Lock()
	defer n.m.Unlock()

	if n.o == nil {
		return nil
	} else if i := n.o.Load(); i == nil {
		return nil
	} else if opt, ok := i.(Options); !ok {
		return nil
	} else {
		return opt
	}
}
