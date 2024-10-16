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
	"bytes"
	"errors"
	"io"

	libarc "github.com/nabbar/golib/archive"
)

func (e *engine) setupWDecompress(source io.Writer) error {

	if e.mode != WriterMode {
		return errors.New("unexpected reader argument for non reader mode")
	}

	wc, ok := source.(io.WriteCloser)

	if !ok {
		wc = libarc.NopWriteCloser(source)
	}

	e.decompressor = &decompressor{
		source: nil,
		buffer: bytes.NewBuffer(make([]byte, 0)),
		writer: wc,
		closed: false,
	}

	e.operation = Decompress

	return nil
}

func (e *engine) setupRDecompress(source io.Reader) error {

	if e.mode != ReaderMode {
		return errors.New("unexpected reader argument for non reader mode")
	}

	var (
		reader io.ReadCloser
		err    error
	)

	reader, err = e.algo.Reader(source)

	if err != nil {
		return err
	}

	e.decompressor = &decompressor{
		source: reader,
		writer: nil,
		buffer: nil,
		closed: false,
	}

	e.operation = Decompress

	return nil
}

func (e *engine) setupWCompress(source io.Writer) error {

	if e.mode != WriterMode {
		return errors.New("unexpected reader argument for non reader mode")
	}

	wc, ok := source.(io.WriteCloser)

	if !ok {
		wc = libarc.NopWriteCloser(source)
	}

	cw, err := e.algo.Writer(wc)

	if err != nil {
		return err
	}

	e.compressor = &compressor{
		source: nil,
		writer: cw,
		closed: false,
	}

	e.operation = Compress

	return nil
}

func (e *engine) setupRCompress(source io.Reader) error {

	if e.mode != ReaderMode {
		return errors.New("unexpected reader argument for non reader mode")
	}

	buffer := bytes.NewBuffer(make([]byte, 0))

	writer, err := e.algo.Writer(libarc.NopWriteCloser(buffer))

	if err != nil {
		return err
	}

	e.compressor = &compressor{
		source: io.NopCloser(source),
		writer: writer,
		buffer: buffer,
		closed: false,
	}

	e.operation = Compress

	return nil
}
