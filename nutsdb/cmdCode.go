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

type CmdCode uint32

const (
	// CmdUnknown is no Command.
	CmdUnknown CmdCode = iota
	// Command for transaction.
	CmdPut
	CmdPutWithTimestamp
	// Command for BPTree.
	CmdGet
	CmdGetAll
	CmdRangeScan
	CmdPrefixScan
	CmdPrefixSearchScan
	CmdDelete
	CmdFindTxIDOnDisk
	CmdFindOnDisk
	CmdFindLeafOnDisk
	// Command for Set.
	CmdSAdd
	CmdSRem
	CmdSAreMembers
	CmdSIsMember
	CmdSMembers
	CmdSHasKey
	CmdSPop
	CmdSCard
	CmdSDiffByOneBucket
	CmdSDiffByTwoBuckets
	CmdSMoveByOneBucket
	CmdSMoveByTwoBuckets
	CmdSUnionByOneBucket
	CmdSUnionByTwoBuckets
	// Command for List.
	CmdRPop
	CmdRPeek
	CmdRPush
	CmdLPush
	CmdLPop
	CmdLPeek
	CmdLSize
	CmdLRange
	CmdLRem
	CmdLSet
	CmdLTrim
	// Command for ZSet.
	CmdZAdd
	CmdZMembers
	CmdZCard
	CmdZCount
	CmdZPopMax
	CmdZPopMin
	CmdZPeekMax
	CmdZPeekMin
	CmdZRangeByScore
	CmdZRangeByRank
	CmdZRem
	CmdZRemRangeByRank
	CmdZRank
	CmdZRevRank
	CmdZScore
	CmdZGetByKey
)

// nolint #funlen
func CmdCodeFromName(name string) CmdCode {
	switch name {
	case CmdPut.Name():
		return CmdPut

	case CmdPutWithTimestamp.Name():
		return CmdPutWithTimestamp

	case CmdGet.Name():
		return CmdGet

	case CmdGetAll.Name():
		return CmdGetAll

	case CmdRangeScan.Name():
		return CmdRangeScan

	case CmdPrefixScan.Name():
		return CmdPrefixScan

	case CmdPrefixSearchScan.Name():
		return CmdPrefixSearchScan

	case CmdDelete.Name():
		return CmdDelete

	case CmdFindTxIDOnDisk.Name():
		return CmdFindTxIDOnDisk

	case CmdFindOnDisk.Name():
		return CmdFindOnDisk

	case CmdFindLeafOnDisk.Name():
		return CmdFindLeafOnDisk

	case CmdSAdd.Name():
		return CmdSAdd

	case CmdSRem.Name():
		return CmdSRem

	case CmdSAreMembers.Name():
		return CmdSAreMembers

	case CmdSIsMember.Name():
		return CmdSIsMember

	case CmdSMembers.Name():
		return CmdSMembers

	case CmdSHasKey.Name():
		return CmdSHasKey

	case CmdSPop.Name():
		return CmdSPop

	case CmdSCard.Name():
		return CmdSCard

	case CmdSDiffByOneBucket.Name():
		return CmdSDiffByOneBucket

	case CmdSDiffByTwoBuckets.Name():
		return CmdSDiffByTwoBuckets

	case CmdSMoveByOneBucket.Name():
		return CmdSMoveByOneBucket

	case CmdSMoveByTwoBuckets.Name():
		return CmdSMoveByTwoBuckets

	case CmdSUnionByOneBucket.Name():
		return CmdSUnionByOneBucket

	case CmdSUnionByTwoBuckets.Name():
		return CmdSUnionByTwoBuckets

	case CmdRPop.Name():
		return CmdRPop

	case CmdRPeek.Name():
		return CmdRPeek

	case CmdRPush.Name():
		return CmdRPush

	case CmdLPush.Name():
		return CmdLPush

	case CmdLPop.Name():
		return CmdLPop

	case CmdLPeek.Name():
		return CmdLPeek

	case CmdLSize.Name():
		return CmdLSize

	case CmdLRange.Name():
		return CmdLRange

	case CmdLRem.Name():
		return CmdLRem

	case CmdLSet.Name():
		return CmdLSet

	case CmdLTrim.Name():
		return CmdLTrim

	case CmdZAdd.Name():
		return CmdZAdd

	case CmdZMembers.Name():
		return CmdZMembers

	case CmdZCard.Name():
		return CmdZCard

	case CmdZCount.Name():
		return CmdZCount

	case CmdZPopMax.Name():
		return CmdZPopMax

	case CmdZPopMin.Name():
		return CmdZPopMin

	case CmdZPeekMax.Name():
		return CmdZPeekMax

	case CmdZPeekMin.Name():
		return CmdZPeekMin

	case CmdZRangeByScore.Name():
		return CmdZRangeByScore

	case CmdZRangeByRank.Name():
		return CmdZRangeByRank

	case CmdZRem.Name():
		return CmdZRem

	case CmdZRemRangeByRank.Name():
		return CmdZRemRangeByRank

	case CmdZRank.Name():
		return CmdZRank

	case CmdZRevRank.Name():
		return CmdZRevRank

	case CmdZScore.Name():
		return CmdZScore

	case CmdZGetByKey.Name():
		return CmdZGetByKey

	default:
		return CmdUnknown
	}
}

// nolint #funlen
func (c CmdCode) Name() string {
	switch c {
	case CmdPut:
		return "Put"

	case CmdPutWithTimestamp:
		return "PutWithTimestamp"

	case CmdGet:
		return "Get"

	case CmdGetAll:
		return "GetAll"

	case CmdRangeScan:
		return "RangeScan"

	case CmdPrefixScan:
		return "PrefixScan"

	case CmdPrefixSearchScan:
		return "PrefixSearchScan"

	case CmdDelete:
		return "Delete"

	case CmdFindTxIDOnDisk:
		return "FindTxIDOnDisk"

	case CmdFindOnDisk:
		return "FindOnDisk"

	case CmdFindLeafOnDisk:
		return "FindLeafOnDisk"

	case CmdSAdd:
		return "SAdd"

	case CmdSRem:
		return "SRem"

	case CmdSAreMembers:
		return "SAreMembers"

	case CmdSIsMember:
		return "SIsMember"

	case CmdSMembers:
		return "SMembers"

	case CmdSHasKey:
		return "SHasKey"

	case CmdSPop:
		return "SPop"

	case CmdSCard:
		return "SCard"

	case CmdSDiffByOneBucket:
		return "SDiffByOneBucket"

	case CmdSDiffByTwoBuckets:
		return "SDiffByTwoBuckets"

	case CmdSMoveByOneBucket:
		return "SMoveByOneBucket"

	case CmdSMoveByTwoBuckets:
		return "SMoveByTwoBuckets"

	case CmdSUnionByOneBucket:
		return "SUnionByOneBucket"

	case CmdSUnionByTwoBuckets:
		return "SUnionByTwoBuckets"

	case CmdRPop:
		return "RPop"

	case CmdRPeek:
		return "RPeek"

	case CmdRPush:
		return "RPush"

	case CmdLPush:
		return "LPush"

	case CmdLPop:
		return "LPop"

	case CmdLPeek:
		return "LPeek"

	case CmdLSize:
		return "LSize"

	case CmdLRange:
		return "LRange"

	case CmdLRem:
		return "LRem"

	case CmdLSet:
		return "LSet"

	case CmdLTrim:
		return "LTrim"

	case CmdZAdd:
		return "ZAdd"

	case CmdZMembers:
		return "ZMembers"

	case CmdZCard:
		return "ZCard"

	case CmdZCount:
		return "ZCount"

	case CmdZPopMax:
		return "ZPopMax"

	case CmdZPopMin:
		return "ZPopMin"

	case CmdZPeekMax:
		return "ZPeekMax"

	case CmdZPeekMin:
		return "ZPeekMin"

	case CmdZRangeByScore:
		return "ZRangeByScore"

	case CmdZRangeByRank:
		return "ZRangeByRank"

	case CmdZRem:
		return "ZRem"

	case CmdZRemRangeByRank:
		return "ZRemRangeByRank"

	case CmdZRank:
		return "ZRank"

	case CmdZRevRank:
		return "ZRevRank"

	case CmdZScore:
		return "ZScore"

	case CmdZGetByKey:
		return "ZGetByKey"

	default:
		return ""
	}
}

// nolint #funlen
func (c CmdCode) Desc() string {
	switch c {
	case CmdPut:
		return "Sets the value for a key in the bucket."

	case CmdPutWithTimestamp:
		return "Sets the value for a key in the bucket but allow capabilities to custom the timestamp for ttl"

	case CmdGet:
		return "Retrieves the value for a key in the bucket"

	case CmdGetAll:
		return "Returns all keys and values of the bucket stored at given bucket"

	case CmdRangeScan:
		return "Query a range at given bucket, start and end slice."

	case CmdPrefixScan:
		return "Iterates over a key prefix at given bucket, prefix and limitNum. LimitNum will limit the number of entries return."

	case CmdPrefixSearchScan:
		return "Iterates over a key prefix at given bucket, prefix, match regular expression and limitNum. LimitNum will limit the number of entries return."

	case CmdDelete:
		return "Removes a key from the bucket at given bucket and key."

	case CmdFindTxIDOnDisk:
		return "Returns if txId on disk at given fid and txID."

	case CmdFindOnDisk:
		return "Returns entry on disk at given fID, rootOff and key."

	case CmdFindLeafOnDisk:
		return "Returns binary leaf node on disk at given fId, rootOff and key."

	case CmdSAdd:
		return "Adds the specified members to the set stored int the bucket at given bucket,key and items."

	case CmdSRem:
		return "Removes the specified members from the set stored int the bucket at given bucket,key and items."

	case CmdSAreMembers:
		return "Returns if the specified members are the member of the set int the bucket at given bucket,key and items."

	case CmdSIsMember:
		return "Returns if member is a member of the set stored int the bucket at given bucket,key and item."

	case CmdSMembers:
		return "Returns all the members of the set value stored int the bucket at given bucket and key."

	case CmdSHasKey:
		return "Returns if the set in the bucket at given bucket and key."

	case CmdSPop:
		return "Removes and returns one or more random elements from the set value store in the bucket at given bucket and key."

	case CmdSCard:
		return "Returns the set cardinality (number of elements) of the set stored in the bucket at given bucket and key."

	case CmdSDiffByOneBucket:
		return "Returns the members of the set resulting from the difference between the first set and all the successive sets in one bucket."

	case CmdSDiffByTwoBuckets:
		return "Returns the members of the set resulting from the difference between the first set and all the successive sets in two buckets."

	case CmdSMoveByOneBucket:
		return "Moves member from the set at source to the set at destination in one bucket."

	case CmdSMoveByTwoBuckets:
		return "Moves member from the set at source to the set at destination in two buckets."

	case CmdSUnionByOneBucket:
		return "The members of the set resulting from the union of all the given sets in one bucket."

	case CmdSUnionByTwoBuckets:
		return "The members of the set resulting from the union of all the given sets in two buckets."

	case CmdRPop:
		return "Removes and returns the last element of the list stored in the bucket at given bucket and key."

	case CmdRPeek:
		return "Returns the last element of the list stored in the bucket at given bucket and key."

	case CmdRPush:
		return "Inserts the values at the tail of the list stored in the bucket at given bucket,key and values."

	case CmdLPush:
		return "Inserts the values at the head of the list stored in the bucket at given bucket,key and values."

	case CmdLPop:
		return "Removes and returns the first element of the list stored in the bucket at given bucket and key."

	case CmdLPeek:
		return "Returns the first element of the list stored in the bucket at given bucket and key."

	case CmdLSize:
		return "Returns the size of key in the bucket in the bucket at given bucket and key."

	case CmdLRange:
		return "Returns the specified elements of the list stored in the bucket at given bucket,key, start and end. \n" +
			"The offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list), 1 being the next element and so on. \n" +
			"Start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on."

	case CmdLRem:
		return "Removes the first count occurrences of elements equal to value from the list stored in the bucket at given bucket,key,count. \n" +
			"The count argument influences the operation in the following ways: \n" +
			"count > 0: Remove elements equal to value moving from head to tail. \n" +
			"count < 0: Remove elements equal to value moving from tail to head. \n" +
			"count = 0: Remove all elements equal to value."

	case CmdLSet:
		return "Sets the list element at index to value."

	case CmdLTrim:
		return "Trims an existing list so that it will contain only the specified range of elements specified. \n" +
			"The offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list), 1 being the next element and so on. \n" +
			"Start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on."

	case CmdZAdd:
		return "Adds the specified member key with the specified score and specified val to the sorted set stored at bucket."

	case CmdZMembers:
		return "Returns all the members of the set value stored at bucket."

	case CmdZCard:
		return "Returns the sorted set cardinality (number of elements) of the sorted set stored at bucket."

	case CmdZCount:
		return "Returns the number of elements in the sorted set at bucket with a score between min and max and opts. \n" +
			"Options includes the following parameters: \n" +
			"Limit: (int) the max nodes to return. \n" +
			"ExcludeStart: (bool) exclude start value, so it search in interval (start, end] or (start, end). \n" +
			"ExcludeEnd: (bool) exclude end value, so it search in interval [start, end) or (start, end)."

	case CmdZPopMax:
		return "Removes and returns the member with the highest score in the sorted set stored at bucket."

	case CmdZPopMin:
		return "Removes and returns the member with the lowest score in the sorted set stored at bucket."

	case CmdZPeekMax:
		return "Returns the member with the highest score in the sorted set stored at bucket."

	case CmdZPeekMin:
		return "Returns the member with the lowest score in the sorted set stored at bucket."

	case CmdZRangeByScore:
		return "Returns all the elements in the sorted set at bucket with a score between min and max."

	case CmdZRangeByRank:
		return "Returns all the elements in the sorted set in one bucket and key with a rank between start and end (including elements with rank equal to start or end)."

	case CmdZRem:
		return "Removes the specified members from the sorted set stored in one bucket at given bucket and key."

	case CmdZRemRangeByRank:
		return "Removes all elements in the sorted set stored in one bucket at given bucket with rank between start and end. \n" +
			"The rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node."

	case CmdZRank:
		return "Returns the rank of member in the sorted set stored in the bucket at given bucket and key, with the scores ordered from low to high."

	case CmdZRevRank:
		return "Returns the rank of member in the sorted set stored in the bucket at given bucket and key, with the scores ordered from high to low."

	case CmdZScore:
		return "Returns the score of member in the sorted set in the bucket at given bucket and key."

	case CmdZGetByKey:
		return "Returns node in the bucket at given bucket and key."

	default:
		return ""
	}
}

// nolint #funlen
func (c CmdCode) Usage() string {
	switch c {
	case CmdPut:
		return c.Name() + " <key> <value>"

	case CmdPutWithTimestamp:
		return "Sets the value for a key in the bucket but allow capabilities to custom the timestamp for ttl"

	case CmdGet:
		return "Retrieves the value for a key in the bucket"

	case CmdGetAll:
		return "Returns all keys and values of the bucket stored at given bucket"

	case CmdRangeScan:
		return "Query a range at given bucket, start and end slice."

	case CmdPrefixScan:
		return "Iterates over a key prefix at given bucket, prefix and limitNum. LimitNum will limit the number of entries return."

	case CmdPrefixSearchScan:
		return "Iterates over a key prefix at given bucket, prefix, match regular expression and limitNum. LimitNum will limit the number of entries return."

	case CmdDelete:
		return "Removes a key from the bucket at given bucket and key."

	case CmdFindTxIDOnDisk:
		return "Returns if txId on disk at given fid and txID."

	case CmdFindOnDisk:
		return "Returns entry on disk at given fID, rootOff and key."

	case CmdFindLeafOnDisk:
		return "Returns binary leaf node on disk at given fId, rootOff and key."

	case CmdSAdd:
		return "Adds the specified members to the set stored int the bucket at given bucket,key and items."

	case CmdSRem:
		return "Removes the specified members from the set stored int the bucket at given bucket,key and items."

	case CmdSAreMembers:
		return "Returns if the specified members are the member of the set int the bucket at given bucket,key and items."

	case CmdSIsMember:
		return "Returns if member is a member of the set stored int the bucket at given bucket,key and item."

	case CmdSMembers:
		return "Returns all the members of the set value stored int the bucket at given bucket and key."

	case CmdSHasKey:
		return "Returns if the set in the bucket at given bucket and key."

	case CmdSPop:
		return "Removes and returns one or more random elements from the set value store in the bucket at given bucket and key."

	case CmdSCard:
		return "Returns the set cardinality (number of elements) of the set stored in the bucket at given bucket and key."

	case CmdSDiffByOneBucket:
		return "Returns the members of the set resulting from the difference between the first set and all the successive sets in one bucket."

	case CmdSDiffByTwoBuckets:
		return "Returns the members of the set resulting from the difference between the first set and all the successive sets in two buckets."

	case CmdSMoveByOneBucket:
		return "Moves member from the set at source to the set at destination in one bucket."

	case CmdSMoveByTwoBuckets:
		return "Moves member from the set at source to the set at destination in two buckets."

	case CmdSUnionByOneBucket:
		return "The members of the set resulting from the union of all the given sets in one bucket."

	case CmdSUnionByTwoBuckets:
		return "The members of the set resulting from the union of all the given sets in two buckets."

	case CmdRPop:
		return "Removes and returns the last element of the list stored in the bucket at given bucket and key."

	case CmdRPeek:
		return "Returns the last element of the list stored in the bucket at given bucket and key."

	case CmdRPush:
		return "Inserts the values at the tail of the list stored in the bucket at given bucket,key and values."

	case CmdLPush:
		return "Inserts the values at the head of the list stored in the bucket at given bucket,key and values."

	case CmdLPop:
		return "Removes and returns the first element of the list stored in the bucket at given bucket and key."

	case CmdLPeek:
		return "Returns the first element of the list stored in the bucket at given bucket and key."

	case CmdLSize:
		return "Returns the size of key in the bucket in the bucket at given bucket and key."

	case CmdLRange:
		return "Returns the specified elements of the list stored in the bucket at given bucket,key, start and end. \n" +
			"The offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list), 1 being the next element and so on. \n" +
			"Start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on."

	case CmdLRem:
		return "Removes the first count occurrences of elements equal to value from the list stored in the bucket at given bucket,key,count. \n" +
			"The count argument influences the operation in the following ways: \n" +
			"count > 0: Remove elements equal to value moving from head to tail. \n" +
			"count < 0: Remove elements equal to value moving from tail to head. \n" +
			"count = 0: Remove all elements equal to value."

	case CmdLSet:
		return "Sets the list element at index to value."

	case CmdLTrim:
		return "Trims an existing list so that it will contain only the specified range of elements specified. \n" +
			"The offsets start and stop are zero-based indexes 0 being the first element of the list (the head of the list), 1 being the next element and so on. \n" +
			"Start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on."

	case CmdZAdd:
		return "Adds the specified member key with the specified score and specified val to the sorted set stored at bucket."

	case CmdZMembers:
		return "Returns all the members of the set value stored at bucket."

	case CmdZCard:
		return "Returns the sorted set cardinality (number of elements) of the sorted set stored at bucket."

	case CmdZCount:
		return "Returns the number of elements in the sorted set at bucket with a score between min and max and opts. \n" +
			"Options includes the following parameters: \n" +
			"Limit: (int) the max nodes to return. \n" +
			"ExcludeStart: (bool) exclude start value, so it search in interval (start, end] or (start, end). \n" +
			"ExcludeEnd: (bool) exclude end value, so it search in interval [start, end) or (start, end)."

	case CmdZPopMax:
		return "Removes and returns the member with the highest score in the sorted set stored at bucket."

	case CmdZPopMin:
		return "Removes and returns the member with the lowest score in the sorted set stored at bucket."

	case CmdZPeekMax:
		return "Returns the member with the highest score in the sorted set stored at bucket."

	case CmdZPeekMin:
		return "Returns the member with the lowest score in the sorted set stored at bucket."

	case CmdZRangeByScore:
		return "Returns all the elements in the sorted set at bucket with a score between min and max."

	case CmdZRangeByRank:
		return "Returns all the elements in the sorted set in one bucket and key with a rank between start and end (including elements with rank equal to start or end)."

	case CmdZRem:
		return "Removes the specified members from the sorted set stored in one bucket at given bucket and key."

	case CmdZRemRangeByRank:
		return "Removes all elements in the sorted set stored in one bucket at given bucket with rank between start and end. \n" +
			"The rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node."

	case CmdZRank:
		return "Returns the rank of member in the sorted set stored in the bucket at given bucket and key, with the scores ordered from low to high."

	case CmdZRevRank:
		return "Returns the rank of member in the sorted set stored in the bucket at given bucket and key, with the scores ordered from high to low."

	case CmdZScore:
		return "Returns the score of member in the sorted set in the bucket at given bucket and key."

	case CmdZGetByKey:
		return "Returns node in the bucket at given bucket and key."

	default:
		return ""
	}
}
