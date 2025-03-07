/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package dns_mapper

import (
	"context"
	"net/http"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
)

type dmp struct {
	d *sync.Map
	z *sync.Map
	c libatm.Value[*Config]
	t libatm.Value[*http.Transport]
	n libatm.Value[context.CancelFunc]
	x libatm.Value[context.Context]
	f libtls.FctRootCACert
	i func(msg string)
}

func (o *dmp) Close() error {
	if i := o.n.Swap(func() {}); i != nil {
		i()
	}

	return nil
}

func (o *dmp) config() *Config {
	var cfg = &Config{}

	if c := o.c.Load(); c == nil {
		return cfg
	} else {
		*cfg = *c
		return cfg
	}
}

func (o *dmp) GetConfig() Config {
	var cfg = Config{}

	if c := o.config(); c != nil {
		cfg = *c
	}

	return cfg
}

func (o *dmp) configDialerTimeout() time.Duration {
	if cfg := o.config(); cfg == nil {
		return 30 * time.Second
	} else if cfg.Transport.TimeoutGlobal == 0 {
		return 30 * time.Second
	} else {
		return cfg.Transport.TimeoutGlobal.Time()
	}
}

func (o *dmp) configDialerKeepAlive() time.Duration {
	if cfg := o.config(); cfg == nil {
		return 15 * time.Second
	} else if cfg.Transport.TimeoutKeepAlive == 0 {
		return 15 * time.Second
	} else {
		return cfg.Transport.TimeoutKeepAlive.Time()
	}
}

func (o *dmp) TimeCleaner(ctx context.Context, dur time.Duration) {
	if dur < 5*time.Second {
		dur = 5 * time.Minute
	}

	var (
		x context.Context
		n context.CancelFunc
	)

	if ctx != nil {
		x, n = context.WithCancel(ctx)
		o.x.Store(x)
		if i := o.n.Swap(n); i != nil {
			i()
		}
	}

	go func() {
		var (
			tk = time.NewTicker(dur)
			cx = o.x.Load()
		)

		defer func() {
			tk.Stop()
		}()

		for {
			select {
			case <-cx.Done():
				return
			case <-tk.C:
				o.DefaultTransport().CloseIdleConnections()
			}
		}
	}()
}

func (o *dmp) Message(msg string) {
	if o.i != nil {
		o.i(msg)
	}
}
