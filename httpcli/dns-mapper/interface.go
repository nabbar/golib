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
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libdur "github.com/nabbar/golib/duration"
)

type DNSMapper interface {
	Add(from, to string)
	Get(from string) string
	Del(from string)
	Len() int
	Walk(func(from, to string) bool)
	Clean(endpoint string) (host string, port string, err error)
	Search(endpoint string) (string, error)
	SearchWithCache(endpoint string) (string, error)

	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Transport(cfg TransportConfig) *http.Transport
	Client(cfg TransportConfig) *http.Client

	DefaultTransport() *http.Transport
	DefaultClient() *http.Client

	TimeCleaner(ctx context.Context, dur time.Duration)
}

func New(ctx context.Context, cfg *Config, fct libtls.FctRootCA) DNSMapper {
	if cfg == nil {
		cfg = &Config{
			DNSMapper:  make(map[string]string),
			TimerClean: libdur.ParseDuration(3 * time.Minute),
			Transport: TransportConfig{
				Proxy:     nil,
				TLSConfig: nil,
			},
		}
	}

	if fct == nil {
		fct = func() []string {
			return make([]string, 0)
		}
	}

	d := &dmp{
		d: new(sync.Map),
		z: new(sync.Map),
		c: new(atomic.Value),
		t: new(atomic.Value),
		f: fct,
	}

	for edp, adr := range cfg.DNSMapper {
		d.Add(edp, adr)
	}

	d.c.Store(cfg)
	_ = d.DefaultTransport()
	d.TimeCleaner(ctx, cfg.TimerClean.Time())

	return d
}
