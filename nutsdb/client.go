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

	"github.com/lni/dragonboat/v3/statemachine"

	liberr "github.com/nabbar/golib/errors"

	"github.com/lni/dragonboat/v3/client"
	"github.com/nabbar/golib/cluster"
	"github.com/xujiajun/nutsdb"
	"github.com/xujiajun/nutsdb/ds/zset"
)

type Client interface {
	Commands
}

type clientNutDB struct {
	x context.Context
	s *client.Session
	c cluster.Cluster
	r bool
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

	if p, e = cmd.EncodeRequest(); e != nil {
		return nil, e
	}

	if read {
		if i, e = c.c.SyncRead(c.x, p); e != nil {
			return nil, e
		} else if r, ok = i.(*CommandResponse); !ok {
			return nil, ErrorClientCommandResponseInvalid.Error(nil)
		} else {
			return r, nil
		}
	} else {
		if d, e = c.c.SyncPropose(c.x, c.s, p); e != nil {
			return nil, e
		} else if r, e = cmd.DecodeResult(d.Data); e != nil {
			return nil, e
		} else {
			return r, nil
		}

	}
}

func (c *clientNutDB) Put(bucket string, key, value []byte, ttl uint32) error {
	cmd := NewCommand()
	cmd.Cmd = CmdPut
	cmd.Params = make([]interface{}, 4)
	cmd.Params[0] = bucket
	cmd.Params[1] = key
	cmd.Params[2] = value
	cmd.Params[3] = ttl

	if res, err := c.call(cmd, false); err != nil {
		return err
	} else if res == nil {
		return nil
	} else if res.Error != nil {
		return res.Error
	}

	return nil
}

func (c *clientNutDB) PutWithTimestamp(bucket string, key, value []byte, ttl uint32, timestamp uint64) error {
	cmd := NewCommand()
	cmd.Cmd = CmdPutWithTimestamp
	cmd.Params = make([]interface{}, 5)
	cmd.Params[0] = bucket
	cmd.Params[1] = key
	cmd.Params[2] = value
	cmd.Params[3] = ttl
	cmd.Params[4] = timestamp

	if res, err := c.call(cmd, false); err != nil {
		return err
	} else if res == nil {
		return nil
	} else if res.Error != nil {
		return res.Error
	}

	return nil
}

func (c *clientNutDB) Get(bucket string, key []byte) (e *nutsdb.Entry, err error) {
	cmd := NewCommand()
	cmd.Cmd = CmdGet
	cmd.Params = make([]interface{}, 2)
	cmd.Params[0] = bucket
	cmd.Params[1] = key

	if res, err := c.call(cmd, false); err != nil {
		return nil, err
	} else if res == nil {
		return nil, nil
	} else if res.Error != nil {
		return nil, res.Error
	} else if len(res.Value) < 1 || res.Value[0] == nil {
		return nil, nil
	} else if e, ok := res.Value[0].(*nutsdb.Entry); !ok {
		return nil, nil
	} else {
		return e, nil
	}
}

func (c *clientNutDB) GetAll(bucket string) (entries nutsdb.Entries, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) RangeScan(bucket string, start, end []byte) (es nutsdb.Entries, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) PrefixScan(bucket string, prefix []byte, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) PrefixSearchScan(bucket string, prefix []byte, reg string, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) Delete(bucket string, key []byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) FindTxIDOnDisk(fID, txID uint64) (ok bool, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) FindOnDisk(fID uint64, rootOff uint64, key, newKey []byte) (entry *nutsdb.Entry, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) FindLeafOnDisk(fID int64, rootOff int64, key, newKey []byte) (bn *nutsdb.BinaryNode, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SAdd(bucket string, key []byte, items ...byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SRem(bucket string, key []byte, items ...byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SAreMembers(bucket string, key []byte, items ...byte) (bool, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SIsMember(bucket string, key, item []byte) (bool, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SMembers(bucket string, key []byte) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SHasKey(bucket string, key []byte) (bool, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SPop(bucket string, key []byte) ([]byte, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SCard(bucket string, key []byte) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SDiffByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SDiffByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SMoveByOneBucket(bucket string, key1, key2, item []byte) (bool, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SMoveByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2, item []byte) (bool, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SUnionByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) SUnionByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) RPop(bucket string, key []byte) (item []byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) RPeek(bucket string, key []byte) (item []byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) RPush(bucket string, key []byte, values ...byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LPush(bucket string, key []byte, values ...byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LPop(bucket string, key []byte) (item []byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LPeek(bucket string, key []byte) (item []byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LSize(bucket string, key []byte) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LRange(bucket string, key []byte, start, end int) (list [][]byte, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LRem(bucket string, key []byte, count int, value []byte) (removedNum int, err error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LSet(bucket string, key []byte, index int, value []byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) LTrim(bucket string, key []byte, start, end int) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZAdd(bucket string, key []byte, score float64, val []byte) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZMembers(bucket string) (map[string]*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZCard(bucket string) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZCount(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZPopMax(bucket string) (*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZPopMin(bucket string) (*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZPeekMax(bucket string) (*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZPeekMin(bucket string) (*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRangeByScore(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) ([]*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRangeByRank(bucket string, start, end int) ([]*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRem(bucket, key string) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRemRangeByRank(bucket string, start, end int) error {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRank(bucket string, key []byte) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZRevRank(bucket string, key []byte) (int, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZScore(bucket string, key []byte) (float64, error) {
	//@TODO : implement me !!
	panic("implement me")
}

func (c *clientNutDB) ZGetByKey(bucket string, key []byte) (*zset.SortedSetNode, error) {
	//@TODO : implement me !!
	panic("implement me")
}
