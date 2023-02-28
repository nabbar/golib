//go:build !386 && !arm && !mips && !mipsle
// +build !386,!arm,!mips,!mipsle

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
	"github.com/nutsdb/nutsdb"
	"github.com/nutsdb/nutsdb/ds/zset"
)

type Commands interface {
	CommandTransaction
	CommandBPTree
	CommandSet
	CommandList
	CommandZSet
}

type CommandTransaction interface {
	// Put sets the value for a key in the bucket.
	Put(bucket string, key, value []byte, ttl uint32) error

	// PutWithTimestamp sets the value for a key in the bucket but allow capabilities to custom the timestamp for ttl.
	PutWithTimestamp(bucket string, key, value []byte, ttl uint32, timestamp uint64) error
}

type CommandBPTree interface {
	// Get retrieves the value for a key in the bucket.
	// The returned value is only valid for the life of the transaction.
	Get(bucket string, key []byte) (e *nutsdb.Entry, err error)

	//GetAll returns all keys and values of the bucket stored at given bucket.
	GetAll(bucket string) (entries nutsdb.Entries, err error)

	// RangeScan query a range at given bucket, start and end slice.
	RangeScan(bucket string, start, end []byte) (es nutsdb.Entries, err error)

	// PrefixScan iterates over a key prefix at given bucket, prefix and limitNum.
	// LimitNum will limit the number of entries return.
	PrefixScan(bucket string, prefix []byte, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error)

	// PrefixSearchScan iterates over a key prefix at given bucket, prefix, match regular expression and limitNum.
	// LimitNum will limit the number of entries return.
	PrefixSearchScan(bucket string, prefix []byte, reg string, offsetNum int, limitNum int) (es nutsdb.Entries, off int, err error)

	// Delete removes a key from the bucket at given bucket and key.
	Delete(bucket string, key []byte) error

	// FindTxIDOnDisk returns if txId on disk at given fid and txID.
	FindTxIDOnDisk(fID, txID uint64) (ok bool, err error)

	// FindOnDisk returns entry on disk at given fID, rootOff and key.
	FindOnDisk(fID uint64, rootOff uint64, key, newKey []byte) (entry *nutsdb.Entry, err error)

	// FindLeafOnDisk returns binary leaf node on disk at given fId, rootOff and key.
	FindLeafOnDisk(fID int64, rootOff int64, key, newKey []byte) (bn *nutsdb.BinaryNode, err error)
}

type CommandSet interface {
	// SAdd adds the specified members to the set stored int the bucket at given bucket,key and items.
	SAdd(bucket string, key []byte, items ...[]byte) error

	// SRem removes the specified members from the set stored int the bucket at given bucket,key and items.
	SRem(bucket string, key []byte, items ...[]byte) error

	// SAreMembers returns if the specified members are the member of the set int the bucket at given bucket,key and items.
	SAreMembers(bucket string, key []byte, items ...[]byte) (bool, error)

	// SIsMember returns if member is a member of the set stored int the bucket at given bucket,key and item.
	SIsMember(bucket string, key, item []byte) (bool, error)

	// SMembers returns all the members of the set value stored int the bucket at given bucket and key.
	SMembers(bucket string, key []byte) (list [][]byte, err error)

	// SHasKey returns if the set in the bucket at given bucket and key.
	SHasKey(bucket string, key []byte) (bool, error)

	// SPop removes and returns one or more random elements from the set value store in the bucket at given bucket and key.
	SPop(bucket string, key []byte) ([]byte, error)

	// SCard returns the set cardinality (number of elements) of the set stored in the bucket at given bucket and key.
	SCard(bucket string, key []byte) (int, error)

	// SDiffByOneBucket returns the members of the set resulting from the difference
	// between the first set and all the successive sets in one bucket.
	SDiffByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error)

	// SDiffByTwoBuckets returns the members of the set resulting from the difference
	// between the first set and all the successive sets in two buckets.
	SDiffByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error)

	// SMoveByOneBucket moves member from the set at source to the set at destination in one bucket.
	SMoveByOneBucket(bucket string, key1, key2, item []byte) (bool, error)

	// SMoveByTwoBuckets moves member from the set at source to the set at destination in two buckets.
	SMoveByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2, item []byte) (bool, error)

	// SUnionByOneBucket the members of the set resulting from the union of all the given sets in one bucket.
	SUnionByOneBucket(bucket string, key1, key2 []byte) (list [][]byte, err error)

	// SUnionByTwoBuckets the members of the set resulting from the union of all the given sets in two buckets.
	SUnionByTwoBuckets(bucket1 string, key1 []byte, bucket2 string, key2 []byte) (list [][]byte, err error)
}

type CommandList interface {
	// RPop removes and returns the last element of the list stored in the bucket at given bucket and key.
	RPop(bucket string, key []byte) (item []byte, err error)

	// RPeek returns the last element of the list stored in the bucket at given bucket and key.
	RPeek(bucket string, key []byte) (item []byte, err error)

	// RPush inserts the values at the tail of the list stored in the bucket at given bucket,key and values.
	RPush(bucket string, key []byte, values ...[]byte) error

	// LPush inserts the values at the head of the list stored in the bucket at given bucket,key and values.
	LPush(bucket string, key []byte, values ...[]byte) error

	// LPop removes and returns the first element of the list stored in the bucket at given bucket and key.
	LPop(bucket string, key []byte) (item []byte, err error)

	// LPeek returns the first element of the list stored in the bucket at given bucket and key.
	LPeek(bucket string, key []byte) (item []byte, err error)

	// LSize returns the size of key in the bucket in the bucket at given bucket and key.
	LSize(bucket string, key []byte) (int, error)

	// LRange returns the specified elements of the list stored in the bucket at given bucket,key, start and end.
	// The offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list),
	// 1 being the next element and so on.
	// Start and end can also be negative numbers indicating offsets from the end of the list,
	// where -1 is the last element of the list, -2 the penultimate element and so on.
	LRange(bucket string, key []byte, start, end int) (list [][]byte, err error)

	// LRem removes the first count occurrences of elements equal to value from the list stored in the bucket at given bucket,key,count.
	// The count argument influences the operation in the following ways:
	// count > 0: Remove elements equal to value moving from head to tail.
	// count < 0: Remove elements equal to value moving from tail to head.
	// count = 0: Remove all elements equal to value.
	LRem(bucket string, key []byte, count int, value []byte) (removedNum int, err error)

	// LSet sets the list element at index to value.
	LSet(bucket string, key []byte, index int, value []byte) error

	// LTrim trims an existing list so that it will contain only the specified range of elements specified.
	// the offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list),
	// 1 being the next element and so on.
	// start and end can also be negative numbers indicating offsets from the end of the list,
	// where -1 is the last element of the list, -2 the penultimate element and so on.
	LTrim(bucket string, key []byte, start, end int) error
}

type CommandZSet interface {
	// ZAdd adds the specified member key with the specified score and specified val to the sorted set stored at bucket.
	ZAdd(bucket string, key []byte, score float64, val []byte) error

	// ZMembers returns all the members of the set value stored at bucket.
	ZMembers(bucket string) (map[string]*zset.SortedSetNode, error)

	// ZCard returns the sorted set cardinality (number of elements) of the sorted set stored at bucket.
	ZCard(bucket string) (int, error)

	// ZCount returns the number of elements in the sorted set at bucket with a score between min and max and opts.
	// opts includes the following parameters:
	// Limit        int  // limit the max nodes to return
	// ExcludeStart bool // exclude start value, so it search in interval (start, end] or (start, end)
	// ExcludeEnd   bool // exclude end value, so it search in interval [start, end) or (start, end)
	ZCount(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) (int, error)

	// ZPopMax removes and returns the member with the highest score in the sorted set stored at bucket.
	ZPopMax(bucket string) (*zset.SortedSetNode, error)

	// ZPopMin removes and returns the member with the lowest score in the sorted set stored at bucket.
	ZPopMin(bucket string) (*zset.SortedSetNode, error)

	// ZPeekMax returns the member with the highest score in the sorted set stored at bucket.
	ZPeekMax(bucket string) (*zset.SortedSetNode, error)

	// ZPeekMin returns the member with the lowest score in the sorted set stored at bucket.
	ZPeekMin(bucket string) (*zset.SortedSetNode, error)

	// ZRangeByScore returns all the elements in the sorted set at bucket with a score between min and max.
	ZRangeByScore(bucket string, start, end float64, opts *zset.GetByScoreRangeOptions) ([]*zset.SortedSetNode, error)

	// ZRangeByRank returns all the elements in the sorted set in one bucket and key
	// with a rank between start and end (including elements with rank equal to start or end).
	ZRangeByRank(bucket string, start, end int) ([]*zset.SortedSetNode, error)

	// ZRem removes the specified members from the sorted set stored in one bucket at given bucket and key.
	ZRem(bucket, key string) error

	// ZRemRangeByRank removes all elements in the sorted set stored in one bucket at given bucket with rank between start and end.
	// the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node.
	ZRemRangeByRank(bucket string, start, end int) error

	// ZRank returns the rank of member in the sorted set stored in the bucket at given bucket and key,
	// with the scores ordered from low to high.
	ZRank(bucket string, key []byte) (int, error)

	// ZRevRank returns the rank of member in the sorted set stored in the bucket at given bucket and key,
	// with the scores ordered from high to low.
	ZRevRank(bucket string, key []byte) (int, error)

	// ZScore returns the score of member in the sorted set in the bucket at given bucket and key.
	ZScore(bucket string, key []byte) (float64, error)

	// ZGetByKey returns node in the bucket at given bucket and key.
	ZGetByKey(bucket string, key []byte) (*zset.SortedSetNode, error)
}
