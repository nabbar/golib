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

package compress

import (
	"bytes"
	"errors"
	"io"
	"sync/atomic"
)

type operation uint8

const (
	Compress operation = iota
	Decompress
)

const chunkSize = 856

type Helper interface {
	SetReader(io.Reader) error
	SetWriter(io.Writer) error
	io.ReadWriter
}

func NewHelper(algo Algorithm, operation operation) (Helper, error) {
	var eng *engine

	if operation < 0 || operation > 1 {
		return nil, errors.New("invalid operation: choose 'compress' or 'decompress'")
	}

	eng = &engine{
		state:     new(atomic.Bool),
		algo:      algo,
		buffer:    bytes.NewBuffer(make([]byte, 0)),
		operation: operation,
		closed:    new(atomic.Bool),
		writer:    nil,
		reader:    nil,
	}

	return eng, nil

}
