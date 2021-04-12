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
	"time"

	"github.com/lni/dragonboat/v3/statemachine"
	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
	"github.com/xujiajun/nutsdb"
	"github.com/xujiajun/nutsdb/ds/zset"
)

type Client interface {
	Commands
}

type clientNutDB struct {
	x context.Context
	t time.Duration
	c func() libclu.Cluster
	w func(ctx context.Context, tick time.Duration)
}

func (c *clientNutDB) call(cmd *CommandRequest, read bool) (*CommandResponse, liberr.Error) {
	var (
		p []byte
		e liberr.Error
		i interface{}
		d statemachine.Result
		r *CommandResponse

		ok bool
	)

	if read {
		c.w(c.x, c.t)
		if i, e = c.c().SyncRead(c.x, cmd); e != nil {
			return nil, e
		} else if r, ok = i.(*CommandResponse); !ok {
			return nil, ErrorClientCommandResponseInvalid.Error(nil)
		} else {
			return r, nil
		}
	} else if p, e = cmd.EncodeRequest(); e != nil {
		return nil, e
	} else {
		c.w(c.x, c.t)
		if d, e = c.c().SyncPropose(c.x, c.c().GetNoOPSession(), p); e != nil {
			return nil, e
		} else if r, e = cmd.DecodeResult(d.Data); e != nil {
			return nil, e
		} else {
			return r, nil
		}
	}
}

func (c *clientNutDB) Put(bucket string, key, value []byte, ttl uint32) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, value, ttl)

	if r, f = c.call(d, false); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) PutWithTimestamp(bucket string, key, value []byte, ttl uint32, timestamp uint64) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, value, ttl, timestamp)

	if r, f = c.call(d, false); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) Get(bucket string, key []byte) (e *nutsdb.Entry, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if e, k = r.Value[0].(*nutsdb.Entry); !k {
		e = nil
	}

	return
}

func (c *clientNutDB) GetAll(bucket string) (entries nutsdb.Entries, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if entries, k = r.Value[0].(nutsdb.Entries); !k {
		entries = nil
	}

	return
}

func (c *clientNutDB) RangeScan(bucket string, start, end []byte) (es nutsdb.Entries, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, start, end)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if es, k = r.Value[0].(nutsdb.Entries); !k {
		es = nil
	}

	return
}

func (c *clientNutDB) PrefixScan(bucket string, prefix []byte, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, prefix, offsetNum, limitNum)

	if r, f = c.call(d, true); f != nil {
		return nil, 0, f
	} else if r == nil {
		return nil, 0, nil
	} else if r.Error != nil {
		return nil, 0, r.Error
	} else if len(r.Value) < 2 {
		return nil, 0, nil
	}

	if es, k = r.Value[0].(nutsdb.Entries); !k {
		es = nil
	}

	if off, k = r.Value[1].(int); !k {
		off = 0
	}

	return
}

func (c *clientNutDB) PrefixSearchScan(bucket string, prefix []byte, reg string, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, prefix, reg, offsetNum, limitNum)

	if r, f = c.call(d, true); f != nil {
		return nil, 0, f
	} else if r == nil {
		return nil, 0, nil
	} else if r.Error != nil {
		return nil, 0, r.Error
	} else if len(r.Value) < 2 {
		return nil, 0, nil
	}

	if es, k = r.Value[0].(nutsdb.Entries); !k {
		es = nil
	}

	if off, k = r.Value[1].(int); !k {
		off = 0
	}

	return
}

func (c *clientNutDB) Delete(bucket string, key []byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) FindTxIDOnDisk(fID, txID uint64) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(fID, txID)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) FindOnDisk(fID uint64, rootOff uint64, key, newKey []byte) (entry *nutsdb.Entry, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(fID, rootOff, key, newKey)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if entry, k = r.Value[0].(*nutsdb.Entry); !k {
		entry = nil
	}

	return
}

func (c *clientNutDB) FindLeafOnDisk(fID int64, rootOff int64, key, newKey []byte) (bn *nutsdb.BinaryNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(fID, rootOff, key, newKey)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if bn, k = r.Value[0].(*nutsdb.BinaryNode); !k {
		bn = nil
	}

	return
}

func (c *clientNutDB) SAdd(bucket string, key []byte, items ...[]byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, items)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) SRem(bucket string, key []byte, items ...[]byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, items)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) SAreMembers(bucket string, key []byte, items ...[]byte) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, items)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) SIsMember(bucket string, key, item []byte) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, item)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) SMembers(bucket string, key []byte) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) SHasKey(bucket string, key []byte) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) SPop(bucket string, key []byte) (val []byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if val, k = r.Value[0].([]byte); !k {
		val = nil
	}

	return
}

func (c *clientNutDB) SCard(bucket string, key []byte) (card int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if card, k = r.Value[0].(int); !k {
		card = 0
	}

	return
}

func (c *clientNutDB) SDiffByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key1, key2)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) SDiffByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket1, key1, bucket2, key2)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) SMoveByOneBucket(bucket string, key1, key2, item []byte) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key1, key2, item)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) SMoveByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2, item []byte) (ok bool, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket1, key1, bucket2, key2, item)

	if r, f = c.call(d, true); f != nil {
		return false, f
	} else if r == nil {
		return false, nil
	} else if r.Error != nil {
		return false, r.Error
	} else if len(r.Value) < 1 {
		return false, nil
	}

	if ok, k = r.Value[0].(bool); !k {
		ok = false
	}

	return
}

func (c *clientNutDB) SUnionByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key1, key2)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) SUnionByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket1, key1, bucket2, key2)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) RPop(bucket string, key []byte) (item []byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].([]byte); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) RPeek(bucket string, key []byte) (item []byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].([]byte); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) RPush(bucket string, key []byte, values ...[]byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, values)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) LPush(bucket string, key []byte, values ...[]byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, values)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) LPop(bucket string, key []byte) (item []byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].([]byte); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) LPeek(bucket string, key []byte) (item []byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].([]byte); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) LSize(bucket string, key []byte) (size int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if size, k = r.Value[0].(int); !k {
		size = 0
	}

	return
}

func (c *clientNutDB) LRange(bucket string, key []byte, start, end int) (list [][]byte, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, start, end)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([][]byte); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) LRem(bucket string, key []byte, count int, value []byte) (removedNum int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, count, value)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if removedNum, k = r.Value[0].(int); !k {
		removedNum = 0
	}

	return
}

func (c *clientNutDB) LSet(bucket string, key []byte, index int, value []byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, index, value)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) LTrim(bucket string, key []byte, start, end int) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, start, end)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) ZAdd(bucket string, key []byte, score float64, val []byte) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key, score, val)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) ZMembers(bucket string) (list map[string]*zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].(map[string]*zset.SortedSetNode); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) ZCard(bucket string) (card int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if card, k = r.Value[0].(int); !k {
		card = 0
	}

	return
}

func (c *clientNutDB) ZCount(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) (number int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, start, end, opts)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if number, k = r.Value[0].(int); !k {
		number = 0
	}

	return
}

func (c *clientNutDB) ZPopMax(bucket string) (item *zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].(*zset.SortedSetNode); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) ZPopMin(bucket string) (item *zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].(*zset.SortedSetNode); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) ZPeekMax(bucket string) (item *zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].(*zset.SortedSetNode); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) ZPeekMin(bucket string) (item *zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].(*zset.SortedSetNode); !k {
		item = nil
	}

	return
}

func (c *clientNutDB) ZRangeByScore(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) (list []*zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, start, end, opts)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([]*zset.SortedSetNode); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) ZRangeByRank(bucket string, start, end int) (list []*zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, start, end)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if list, k = r.Value[0].([]*zset.SortedSetNode); !k {
		list = nil
	}

	return
}

func (c *clientNutDB) ZRem(bucket, key string) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) ZRemRangeByRank(bucket string, start, end int) error {
	var (
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, start, end)

	if r, f = c.call(d, true); f != nil {
		return f
	} else if r == nil {
		return nil
	} else if r.Error != nil {
		return r.Error
	}

	return nil
}

func (c *clientNutDB) ZRank(bucket string, key []byte) (rank int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if rank, k = r.Value[0].(int); !k {
		rank = 0
	}

	return
}

func (c *clientNutDB) ZRevRank(bucket string, key []byte) (rank int, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if rank, k = r.Value[0].(int); !k {
		rank = 0
	}

	return
}

func (c *clientNutDB) ZScore(bucket string, key []byte) (score float64, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return 0, f
	} else if r == nil {
		return 0, nil
	} else if r.Error != nil {
		return 0, r.Error
	} else if len(r.Value) < 1 {
		return 0, nil
	}

	if score, k = r.Value[0].(float64); !k {
		score = 0
	}

	return
}

func (c *clientNutDB) ZGetByKey(bucket string, key []byte) (item *zset.SortedSetNode, err error) {
	var (
		k bool
		f liberr.Error
		r *CommandResponse
		d *CommandRequest
	)

	d = NewCommandByCaller(bucket, key)

	if r, f = c.call(d, true); f != nil {
		return nil, f
	} else if r == nil {
		return nil, nil
	} else if r.Error != nil {
		return nil, r.Error
	} else if len(r.Value) < 1 {
		return nil, nil
	}

	if item, k = r.Value[0].(*zset.SortedSetNode); !k {
		item = nil
	}

	return
}
