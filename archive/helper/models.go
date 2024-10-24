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
	"fmt"
	"io"

	libarc "github.com/nabbar/golib/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
)

// engine manages both compression and decompression operations.
type engine struct {
	operation    string
	compressor   *compressor
	decompressor io.ReadCloser
	algo         arccmp.Algorithm
	mode         Mode
}

func (e *engine) Write(p []byte) (n int, err error) {

	if e.mode != WriterMode {
		return 0, errors.New("write func can't be invoked for non writer mode")
	}

	if e.operation == "" {
		return 0, fmt.Errorf("operation mode not set, please call Compress or Decompress first")
	}

	switch e.operation {

	case "compress":

		defer func() {
			err = e.compressor.writer.Close()
		}()

		return e.compressor.writer.Write(p)

	case "decompress":
		var rc io.ReadCloser

		rc, err = e.algo.Reader(bytes.NewReader(p))

		buf := make([]byte, 512)

		for {

			n, err = rc.Read(buf)

			if n == 0 && err == io.EOF {

				if err = rc.Close(); err != nil {
					return 0, err

				} else {
					return 0, nil

				}
			}

			_, errW := e.compressor.writer.Write(buf[:n])

			if errW != nil {
				return 0, errW
			}
		}

	default:
		return 0, io.EOF
	}

}

// Compress initializes the compressor.
func (e *engine) Compress(source any) error {

	switch src := source.(type) {

	case io.ReadWriter:
		if e.mode == ReaderMode {
			return e.setupRCompress(src)
		} else {
			return e.setupWCompress(src)
		}
	case io.Reader:
		if e.mode != ReaderMode {
			return errors.New("unexpected reader argument for non reader mode")
		}
		return e.setupRCompress(src)
	case io.Writer:
		if e.mode != WriterMode {
			return errors.New("unexpected writer argument for non writer mode")
		}
		return e.setupWCompress(src)
	default:
		return errors.New("unsupported source type")
	}
}

func (e *engine) setupRCompress(source io.Reader) error {
	buffer := bytes.NewBuffer(make([]byte, 0))

	writer, err := e.algo.Writer(libarc.NopWriteCloser(buffer))

	if err != nil {
		return err
	}

	e.compressor = &compressor{
		source: source,
		writer: writer,
		buffer: buffer,
		closed: false,
	}

	e.operation = "compress"

	return nil
}

func (e *engine) setupWCompress(source io.Writer) error {
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

	e.operation = "compress"

	return nil
}

func (e *engine) setupWDecompress(source io.Writer) error {
	if e.mode != WriterMode {
		return errors.New("unexpected reader argument for non reader mode")
	}

	e.compressor = &compressor{
		source: nil,
		buffer: bytes.NewBuffer(make([]byte, 0)),
		writer: libarc.NopWriteCloser(source),
		closed: false,
	}

	e.operation = "decompress"

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

	e.decompressor = reader
	e.operation = "decompress"

	return nil
}

// Decompress initializes the decompressor.
func (e *engine) Decompress(source any) error {

	switch src := source.(type) {

	case io.ReadWriter:

		if e.mode == ReaderMode {
			return e.setupRDecompress(src)
		} else {
			return e.setupWDecompress(src)
		}

	case io.Reader:
		return e.setupRDecompress(src)

	case io.Writer:
		return e.setupWDecompress(src)

	}

	return nil
}

// Read handles reading from the compressor or decompressor.
func (e *engine) Read(p []byte) (int, error) {

	if e.mode != ReaderMode {
		return 0, errors.New("read func can't be invoked for non reader mode")
	}

	if e.operation == "" {
		return 0, fmt.Errorf("operation mode not set, please call Compress or Decompress first")
	}

	switch e.operation {
	case "compress":
		return e.compressor.Read(p)
	case "decompress":
		return e.decompressor.Read(p)
	default:
		return 0, io.EOF
	}
}

// Close handles closing the compressor or decompressor.
func (e *engine) Close() error {
	switch e.operation {
	case "compress":
		if e.compressor != nil {
			return e.compressor.Close()
		}
	case "decompress":
		if e.decompressor != nil {
			return e.decompressor.Close()
		}
	}
	return nil
}
