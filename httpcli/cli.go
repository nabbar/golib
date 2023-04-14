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

package httpcli

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
)

const (
	ClientTimeout30Sec = 30 * time.Second
	ClientTimeout10Sec = 10 * time.Second
	ClientTimeout5Sec  = 5 * time.Second
	ClientTimeout1Sec  = 1 * time.Second

	ClientNetworkTCP = "tcp"
	ClientNetworkUDP = "udp"
)

type FctHttpClient func() *http.Client

func GetClient(serverName string) *http.Client {
	c, e := GetClientTimeout(serverName, true, 0)

	if e != nil {
		c, _ = GetClientTimeout(serverName, false, 0)
	}

	return c
}

func GetClientError(serverName string) (*http.Client, liberr.Error) {
	return GetClientTimeout(serverName, true, 0)
}

func GetClientTimeout(serverName string, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	dl := &net.Dialer{}

	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DialContext:        dl.DialContext,
		DisableCompression: true,
		//nolint #staticcheck
		TLSClientConfig: libtls.GetTLSConfig(serverName),
	}

	return getclient(tr, http2Tr, GlobalTimeout)
}

func GetClientCustom(tr *http.Transport, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	return getclient(tr, http2Tr, GlobalTimeout)
}

func GetClientTls(serverName string, tls libtls.TLSConfig, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	dl := &net.Dialer{}

	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DialContext:        dl.DialContext,
		DisableCompression: true,
		//nolint #staticcheck
		TLSClientConfig: tls.TlsConfig(serverName),
	}

	return getclient(tr, http2Tr, GlobalTimeout)
}

func GetClientTlsForceIp(netw Network, ip string, serverName string, tls libtls.TLSConfig, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	u := &url.URL{
		Host: ip,
	}

	fctDial := func(ctx context.Context, network, address string) (net.Conn, error) {
		dl := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: net.ParseIP(u.Hostname()),
			},
		}
		return dl.DialContext(ctx, netw.Code(), ip)
	}

	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DialContext:        fctDial,
		DisableCompression: true,
		//nolint #staticcheck
		TLSClientConfig: tls.TlsConfig(serverName),
	}

	return getclient(tr, http2Tr, GlobalTimeout)
}

func getclient(tr *http.Transport, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	if http2Tr {
		if e := http2.ConfigureTransport(tr); e != nil {
			return nil, ErrorClientTransportHttp2.ErrorParent(e)
		}
	}

	c := &http.Client{
		Transport: tr,
	}

	if GlobalTimeout != 0 {
		c.Timeout = GlobalTimeout
	}

	return c, nil
}
