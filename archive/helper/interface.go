/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine Bou Aram
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

	"github.com/nabbar/golib/archive/compress"
)

type Mode uint8

const (
	ReaderMode = iota
	WriterMode
)

type Helper interface {
	Compress(any) error
	Decompress(any) error
	io.ReadWriteCloser
}

// NewHelper creates a new Helper instance for a specified algorithm.
func NewHelper(algo compress.Algorithm, mode Mode) (Helper, error) {
	if mode < 0 || mode > 1 {
		return nil, errors.New("unexpected mode argument")
	}

	return &engine{
		algo: algo,
		mode: mode,
	}, nil
}
