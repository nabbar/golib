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
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/nabbar/golib/certificates"
	"github.com/nabbar/golib/errors"
	"golang.org/x/net/http2"
)

const (
	TIMEOUT_30_SEC = 30 * time.Second
	TIMEOUT_10_SEC = 10 * time.Second
	TIMEOUT_5_SEC  = 5 * time.Second
	TIMEOUT_1_SEC  = 1 * time.Second
)

func GetClient(serverName string) *http.Client {
	c, e := getClient(true, TIMEOUT_30_SEC, TIMEOUT_10_SEC, TIMEOUT_30_SEC, TIMEOUT_30_SEC, TIMEOUT_5_SEC, TIMEOUT_5_SEC, certificates.GetTLSConfig(serverName))

	if e != nil {
		c, _ = getClient(false, TIMEOUT_30_SEC, TIMEOUT_10_SEC, TIMEOUT_30_SEC, TIMEOUT_30_SEC, TIMEOUT_5_SEC, TIMEOUT_5_SEC, certificates.GetTLSConfig(serverName))
	}

	return c
}

func GetClientError(serverName string) (*http.Client, errors.Error) {
	return getClient(true, TIMEOUT_30_SEC, TIMEOUT_30_SEC, TIMEOUT_30_SEC, TIMEOUT_30_SEC, TIMEOUT_5_SEC, TIMEOUT_5_SEC, certificates.GetTLSConfig(serverName))
}

func GetClientTimeout(serverName string, GlobalTimeout, DialTimeOut, DialKeepAlive, IdleConnTimeout, TLSHandshakeTimeout, ExpectContinueTimeout time.Duration) (*http.Client, errors.Error) {
	return getClient(true, GlobalTimeout, DialTimeOut, DialKeepAlive, IdleConnTimeout, TLSHandshakeTimeout, ExpectContinueTimeout, certificates.GetTLSConfig(serverName))
}

func getClient(http2Transport bool, GlobalTimeout, DialTimeOut, DialKeepAlive, IdleConnTimeout, TLSHandshakeTimeout, ExpectContinueTimeout time.Duration, tlsConfig *tls.Config) (*http.Client, errors.Error) {

	dl := &net.Dialer{}

	if DialTimeOut != 0 {
		dl.Timeout = DialTimeOut
	}

	if DialKeepAlive != 0 {
		dl.KeepAlive = DialKeepAlive
	}

	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DialContext:        dl.DialContext,
		DisableCompression: true,
	}

	if tlsConfig != nil {
		tr.TLSClientConfig = tlsConfig
	}

	if TLSHandshakeTimeout != 0 {
		tr.TLSHandshakeTimeout = TLSHandshakeTimeout
	}

	if IdleConnTimeout != 0 {
		tr.IdleConnTimeout = IdleConnTimeout
	}

	if ExpectContinueTimeout != 0 {
		tr.ExpectContinueTimeout = ExpectContinueTimeout
	}

	if http2Transport {
		if e := http2.ConfigureTransport(tr); e != nil {
			return nil, HTTP2_CONFIGURE.ErrorParent(e)
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
