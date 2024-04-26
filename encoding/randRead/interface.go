package randRead

import (
	"bufio"
	"io"
	"sync/atomic"
)

type FuncRemote func() (io.ReadCloser, error)

/*
New return an interface for a random reader from remote source.
New returns io.ReadCloser, will buffer data from remote source.
The input function should return an io.ReadCloser and an error.
The http request or any other implementation could be used.
*/
func New(fct FuncRemote) io.ReadCloser {
	if fct == nil {
		return nil
	}

	f := new(atomic.Value)
	f.Store(fct)

	r := &remote{
		f: f,
	}

	return &prnd{
		b: bufio.NewReader(r),
		r: r,
	}
}
