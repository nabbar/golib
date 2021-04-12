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
	// No Command
	CmdUnknown CmdCode = iota
	// Transaction
	CmdPut
	CmdPutWithTimestamp
	// BPTree
	CmdGet
	CmdGetAll
	CmdRangeScan
	CmdPrefixScan
	CmdPrefixSearchScan
	CmdDelete
	CmdFindTxIDOnDisk
	CmdFindOnDisk
	CmdFindLeafOnDisk
	// Set
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
	// List
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
	// ZSet
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
