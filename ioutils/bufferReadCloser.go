package ioutils

import (
	"bytes"
	"io"
)

type brc struct {
	b *bytes.Buffer
}

func NewBufferReadCloser(b *bytes.Buffer) io.ReadCloser {
	return &brc{
		b: b,
	}
}

func (b *brc) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

func (b *brc) Close() error {
	b.b.Reset()
	return nil
}
