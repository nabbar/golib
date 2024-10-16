/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine BOU ARAM & Nicolas JUHEL
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
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	libarc "github.com/nabbar/golib/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
)

const chunkSize = 512

var (
	ErrInvalidSource    = errors.New("invalid source")
	ErrClosedResource   = errors.New("closed resource")
	ErrInvalidOperation = errors.New("invalid operation")
)

type Helper interface {
	io.ReadWriteCloser
}

func New(algo arccmp.Algorithm, ope Operation, src any) (h Helper, err error) {
	if r, k := src.(io.Reader); k {
		return NewReader(algo, ope, r)
	}
	if w, k := src.(io.Writer); k {
		return NewWriter(algo, ope, w)
	}
	return nil, ErrInvalidSource
}

func NewReader(algo arccmp.Algorithm, ope Operation, src io.Reader) (Helper, error) {
	switch ope {
	case Compress:
		return makeCompressReader(algo, src)
	case Decompress:
		return makeDeCompressReader(algo, src)
	}

	return nil, ErrInvalidOperation
}

func NewWriter(algo arccmp.Algorithm, ope Operation, dst io.Writer) (Helper, error) {
	switch ope {
	case Compress:
		return makeCompressWriter(algo, dst)
	case Decompress:
		return makeDeCompressWriter(algo, dst)
	}

	return nil, ErrInvalidOperation
}

func makeCompressWriter(algo arccmp.Algorithm, src io.Writer) (h Helper, err error) {
	wc, ok := src.(io.WriteCloser)

	if !ok {
		wc = libarc.NopWriteCloser(src)
	}

	if wc, err = algo.Writer(wc); err != nil {
		return nil, err
	} else {
		return &compressWriter{
			dst: wc,
		}, nil
	}
}

func makeCompressReader(algo arccmp.Algorithm, src io.Reader) (h Helper, err error) {
	rc, ok := src.(io.ReadCloser)

	if !ok {
		rc = io.NopCloser(src)
	}

	var (
		buf = bytes.NewBuffer(make([]byte, 0))
		wrt io.WriteCloser
	)

	wrt, err = algo.Writer(libarc.NopWriteCloser(buf))

	return &compressReader{
		src: rc,
		wrt: wrt,
		buf: buf,
		clo: new(atomic.Bool),
	}, nil
}

func makeDeCompressReader(algo arccmp.Algorithm, src io.Reader) (h Helper, err error) {
	rc, ok := src.(io.ReadCloser)

	if !ok {
		rc = io.NopCloser(src)
	}

	if rc, err = algo.Reader(rc); err != nil {
		return nil, err
	} else {
		return &deCompressReader{
			src: rc,
		}, nil
	}
}

func makeDeCompressWriter(algo arccmp.Algorithm, src io.Writer) (h Helper, err error) {
	wc, ok := src.(io.WriteCloser)

	if !ok {
		wc = libarc.NopWriteCloser(src)
	}

	return &deCompressWriter{
		alg: algo,
		wrt: wc,
		buf: &bufNoEOF{
			m: sync.Mutex{},
			b: bytes.NewBuffer(make([]byte, 0)),
			c: new(atomic.Bool),
		},
		clo: new(atomic.Bool),
		run: new(atomic.Bool),
	}, nil
}
