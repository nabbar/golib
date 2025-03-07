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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	libtls "github.com/nabbar/golib/certificates"
)

func (o *dmp) dialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   o.configDialerTimeout(),
		DualStack: true,
		KeepAlive: o.configDialerKeepAlive(),
	}
}

func (o *dmp) Dial(network, address string) (net.Conn, error) {
	return o.DialContext(context.Background(), network, address)
}

func (o *dmp) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var (
		e   error
		d   = o.dialer()
		dst string
	)

	if dst, e = o.SearchWithCache(address); e != nil {
		return nil, e
	} else {
		o.Message(fmt.Sprintf("Dialing '%s %s' => '%s %s'", network, address, network, dst))
		o.CacheSet(address, dst)
		return d.DialContext(ctx, network, dst)
	}
}

func (o *dmp) Transport(cfg TransportConfig) *http.Transport {
	return o.TransportWithTLS(cfg, nil)
}

func (o *dmp) TransportWithTLS(cfg TransportConfig, ssl *tls.Config) *http.Transport {
	var prx func(*http.Request) (*url.URL, error)
	if cfg.Proxy == nil {
		prx = http.ProxyFromEnvironment
	} else {
		prx = http.ProxyURL(cfg.Proxy)
	}

	if ssl != nil {
		ssl = o.getTransportTLS(cfg)
	}

	return &http.Transport{
		Proxy:                 prx,
		Dial:                  o.Dial,
		DialContext:           o.DialContext,
		TLSClientConfig:       ssl,
		TLSHandshakeTimeout:   cfg.TimeoutTLSHandshake.Time(),
		DisableKeepAlives:     cfg.DisableKeepAlive,
		DisableCompression:    cfg.DisableCompression,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.MaxConnsPerHost,
		IdleConnTimeout:       cfg.TimeoutIdleConn.Time(),
		ResponseHeaderTimeout: cfg.TimeoutResponseHeader.Time(),
		ExpectContinueTimeout: cfg.TimeoutExpectContinue.Time(),
		ForceAttemptHTTP2:     !cfg.DisableHTTP2,
	}
}

func (o *dmp) getTransportTLS(cfg TransportConfig) *tls.Config {
	var ssl libtls.TLSConfig

	if cfg.TLSConfig == nil {
		ssl = libtls.New()
		ssl.SetVersionMin(tls.VersionTLS12)
		ssl.SetVersionMax(tls.VersionTLS13)
	} else if ssl = cfg.TLSConfig.New(); ssl == nil {
		ssl = libtls.New()
		ssl.SetVersionMin(tls.VersionTLS12)
		ssl.SetVersionMax(tls.VersionTLS13)
	}

	if v := o.f(); v != nil && v.Len() > 0 {
		ssl.AddRootCA(v)
	}

	return ssl.TlsConfig("")
}

func (o *dmp) Client(cfg TransportConfig) *http.Client {
	return &http.Client{
		Transport: o.Transport(cfg),
	}
}

func (o *dmp) RegisterTransport(t *http.Transport) {
	if t != nil {
		o.t.Store(t)
	}
}

func (o *dmp) DefaultTransport() *http.Transport {
	i := o.t.Load()
	if i != nil {
		return i
	}

	c := o.config()

	t := o.TransportWithTLS(c.Transport, c.TLSConfig)
	o.t.Store(t)
	return t
}

func (o *dmp) DefaultClient() *http.Client {
	return &http.Client{
		Transport: o.DefaultTransport(),
	}
}
