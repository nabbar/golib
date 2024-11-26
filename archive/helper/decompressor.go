package helper

import (
	"bytes"
	"io"
	"sync/atomic"
	"time"
)

type deCompressReader struct {
	src io.ReadCloser
}

func (o *deCompressReader) Read(p []byte) (n int, err error) {
	return o.src.Read(p)
}

func (o *deCompressReader) Write(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

func (o *deCompressReader) Close() error {
	return o.src.Close()
}

type deCompressWriter struct {
	src io.ReadCloser
	wrt io.WriteCloser
	buf *bytes.Buffer
	clo *atomic.Bool
	run *atomic.Bool
}

func (o *deCompressWriter) Read(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

func (o *deCompressWriter) Write(p []byte) (n int, err error) {
	if o.clo.Load() {
		return 0, ErrClosedResource
	} else if !o.run.Load() {
		o.run.Store(true)
		go func() {
			var b = make([]byte, chunkSize)
			for o.run.Load() {
				clear(b)
				if i, _ := o.src.Read(b); i > 0 {
					_, _ = o.wrt.Write(b[:i])
				}
				if o.buf.Len() < 1 {
					o.buf.Reset()
				}
			}
		}()
	}

	return o.buf.Write(p)
}

func (o *deCompressWriter) Close() error {
	o.clo.Store(true)
	o.run.Store(false)

	var tc = time.NewTicker(500 * time.Millisecond)
	select {
	case <-tc.C:
		tc.Stop()
		if o.buf.Len() > 0 {
			_, _ = io.Copy(o.wrt, o.src)
		}
	}

	erc := o.src.Close()
	err := o.wrt.Close()

	if err != nil {
		return err
	} else if erc != nil {
		return erc
	}

	return nil
}
