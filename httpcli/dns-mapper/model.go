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
	"sync"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
)

type dmp struct {
	d *sync.Map
	z *sync.Map
	c *atomic.Value // *Config
	t *atomic.Value // *http transport
	f libtls.FctRootCA
}

func (o *dmp) config() *Config {
	var cfg = &Config{}

	if i := o.c.Load(); i == nil {
		return cfg
	} else if c, k := i.(*Config); !k {
		return cfg
	} else {
		*cfg = *c
		return cfg
	}
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

	go func() {
		var tck = time.NewTicker(dur)
		defer tck.Stop()

		for {
			if ctx.Err() != nil {
				return
			}

			select {
			case <-tck.C:
				o.DefaultTransport().CloseIdleConnections()
			case <-ctx.Done():
				return
			}
		}
	}()
}
