package randRead

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

type remote struct {
	f *atomic.Value
	r io.ReadCloser
}

func (o *remote) readReader(p []byte) (n int, err error) {
	if o.r == nil {
		return 0, io.EOF
	}

	if n, err = o.r.Read(p); err != nil {
		return n, err
	} else {
		return n, nil
	}
}

func (o *remote) readRemote() error {
	if o.f == nil {
		return fmt.Errorf("invalid reader")
	} else if i := o.f.Load(); i == nil {
		return fmt.Errorf("invalid reader")
	} else if f, ok := i.(FuncRemote); !ok {
		return fmt.Errorf("invalid reader")
	} else if r, err := f(); err != nil {
		return err
	} else {
		if o.r != nil {
			_ = o.r.Close()
		}

		o.r = r
		return nil
	}
}

func (o *remote) Read(p []byte) (n int, err error) {
	if n, err = o.readReader(p); err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if err = o.readRemote(); err != nil {
		return 0, err
	}

	return o.readReader(p)
}

func (o *remote) Close() error {
	if o.r != nil {
		_ = o.r.Close()
	}

	return nil
}

type prnd struct {
	b *bufio.Reader
	r *remote
}

func (o *prnd) Read(p []byte) (n int, err error) {
	if o.b != nil {
		return o.b.Read(p)
	}

	if o.r != nil {
		o.b = bufio.NewReader(o.r)
		return o.b.Read(p)
	} else {
		return 0, fmt.Errorf("invalid reader")
	}
}

func (o *prnd) Close() error {
	if o.b != nil {
		o.b.Reset(nil)
	}

	if o.r != nil {
		_ = o.r.Close()
	}

	return nil
}
