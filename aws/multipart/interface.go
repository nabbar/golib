/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package multipart

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libsiz "github.com/nabbar/golib/size"
)

const (
	DefaultPartSize       = 5 * libsiz.SizeMega
	MaxPartSize           = 5 * libsiz.SizeGiga
	MaxObjectSize         = 5 * libsiz.SizeTera
	MaxNumberPart   int32 = 10000
)

type FuncClientS3 func() *sdksss.Client

type MultiPart interface {
	io.WriteCloser

	RegisterContext(ctx context.Context)
	RegisterClientS3(fct FuncClientS3)
	RegisterMultipartID(id string)
	RegisterWorkingFile(file string, truncate bool) error
	RegisterFuncOnPushPart(fct func(eTag string, e error))
	RegisterFuncOnAbort(fct func(nPart int, obj string, e error))
	RegisterFuncOnComplete(fct func(nPart int, obj string, e error))

	StartMPU() error
	StopMPU(abort bool) error

	Copy(fromBucket, fromObject, fromVersionId string) error
	AddPart(r io.Reader) (n int64, e error)
	SendPart() error
	CurrentSizePart() int64
	AddToPart(p []byte) (n int, e error)
	RegisterPart(etag string)

	IsStarted() bool
	Counter() int32
	CounterLeft() int32
}

func New(partSize libsiz.Size, object string, bucket string) MultiPart {
	return &mpu{
		m: sync.RWMutex{},
		c: nil,
		s: partSize,
		i: "",
		o: object,
		b: bucket,
		n: 0,
	}
}

func GetOptimalPartSize(objectSize, partSize libsiz.Size) (libsiz.Size, error) {
	var (
		lim = math.MaxFloat64
		nbr int64
		prt int64
		obj int64
	)

	if partSize <= DefaultPartSize {
		prt = DefaultPartSize.Int64()
	} else {
		prt = partSize.Int64()
	}

	obj = objectSize.Int64()

	if obj > (int64(MaxNumberPart) * MaxPartSize.Int64()) {
		return 0, fmt.Errorf("object size need exceed the maximum number of part with the maximum size of part")
	} else if objectSize > MaxObjectSize {
		return 0, fmt.Errorf("object size is over allowed maximum size of object")
	} else if int64Uin64(obj) > flot64Uint64(lim) || int64Uin64(prt) > flot64Uint64(lim) || int64Uin64(obj/prt) > flot64Uint64(lim) {
		return GetOptimalPartSize(objectSize, libsiz.SizeFromInt64(prt*2))
	} else if nbr = int64(math.Ceil(float64(obj) / float64(prt))); nbr > int64(MaxNumberPart) {
		return GetOptimalPartSize(objectSize, libsiz.SizeFromInt64(prt*2))
	} else if prt > MaxPartSize.Int64() {
		return GetOptimalPartSize(objectSize, libsiz.SizeFromInt64((prt/4)*3))
	}

	return libsiz.SizeFromInt64(prt), nil
}

func int64Uin64(i int64) uint64 {
	if i <= 0 {
		return uint64(0)
	} else {
		return uint64(i)
	}
}

func flot64Uint64(i float64) uint64 {
	if i <= 0 {
		return uint64(0)
	} else if i > float64(math.MaxUint64) {
		return uint64(math.MaxUint64)
	} else {
		return uint64(i)
	}
}
