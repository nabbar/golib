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
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	libdur "github.com/nabbar/golib/duration"
)

type FuncMessage func(msg string)

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
	Close() error
}

func GetRootCaCert(fct libtls.FctRootCA) tlscas.Cert {
	var res tlscas.Cert

	for _, c := range fct() {
		if res == nil {
			res, _ = tlscas.Parse(c)
		} else {
			_ = res.AppendString(c)
		}
	}

	return res
}

func New(ctx context.Context, cfg *Config, fct libtls.FctRootCACert, msg FuncMessage) DNSMapper {
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
		fct = func() tlscas.Cert {
			return nil
		}
	}

	if msg == nil {
		msg = func(msg string) {}
	}

	d := &dmp{
		d: new(sync.Map),
		z: new(sync.Map),
		c: libatm.NewValue[*Config](),
		t: libatm.NewValue[*http.Transport](),
		f: fct,
		i: msg,
		n: libatm.NewValue[context.CancelFunc](),
		x: libatm.NewValue[context.Context](),
	}

	for edp, adr := range cfg.DNSMapper {
		d.Add(edp, adr)
	}

	d.c.Store(cfg)
	_ = d.DefaultTransport()
	d.TimeCleaner(ctx, cfg.TimerClean.Time())

	return d
}
