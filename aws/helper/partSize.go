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

package helper

import (
	"errors"
	"io"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type PartSize int64

const (
	SizeBytes     PartSize = 1
	SizeKiloBytes          = 1024 * SizeBytes
	SizeMegaBytes          = 1024 * SizeKiloBytes
	SizeGigaBytes          = 1024 * SizeMegaBytes
	SizeTeraBytes          = 1024 * SizeGigaBytes
	SizePetaBytes          = 1024 * SizeTeraBytes
)

func SetSize(val int) PartSize {
	return PartSize(val)
}

func SetSizeInt64(val int64) PartSize {
	return PartSize(val)
}

func (p PartSize) Int() int {
	return int(p)
}

func (p PartSize) Int64() int64 {
	return int64(p)
}

func (p PartSize) String() string {
	switch p {
	case SizePetaBytes:
		return "PB"
	case SizeTeraBytes:
		return "TB"
	case SizeGigaBytes:
		return "GB"
	case SizeMegaBytes:
		return "MB"
	case SizeKiloBytes:
		return "KB"
	case SizeBytes:
		return "B"
	}

	return ""
}

type ReaderPartSize interface {
	io.Reader
	NextPart(eTag *string)
	CurrPart() int32
	CompPart() *sdktps.CompletedMultipartUpload
	IeOEF() bool
}

func NewReaderPartSize(rd io.Reader, p PartSize) ReaderPartSize {
	return &readerPartSize{
		b: rd,
		p: p.Int64(),
		i: 1,
		j: 0,
		e: false,
		c: nil,
	}
}

type readerPartSize struct {
	// buffer
	b io.Reader
	// partsize
	p int64
	// partNumber
	i int64
	// current part counter
	j int64
	// Is EOF
	e bool
	// complete part slice
	c *sdktps.CompletedMultipartUpload
}

func (r *readerPartSize) NextPart(eTag *string) {
	if r.c == nil {
		r.c = &sdktps.CompletedMultipartUpload{
			Parts: nil,
		}
	}

	if r.c.Parts == nil {
		r.c.Parts = make([]sdktps.CompletedPart, 0)
	}

	r.c.Parts = append(r.c.Parts, sdktps.CompletedPart{
		ETag:       sdkaws.String(strings.Replace(*eTag, "\"", "", -1)),
		PartNumber: int32(r.i),
	})

	r.i++
	r.j = 0
}

func (r readerPartSize) CurrPart() int32 {
	return int32(r.i)
}

func (r readerPartSize) CompPart() *sdktps.CompletedMultipartUpload {
	return r.c
}

func (r readerPartSize) IeOEF() bool {
	return r.e
}

func (r *readerPartSize) Read(p []byte) (n int, err error) {
	if r.e || r.j >= r.p {
		return 0, io.EOF
	}

	if len(p) > int(r.p-r.j) {
		p = make([]byte, int(r.p-r.j))
	}

	n, e := r.b.Read(p)

	if errors.Is(e, io.EOF) {
		r.e = true
	}

	return n, e
}
