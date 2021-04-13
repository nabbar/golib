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
	c, e := GetClientTimeout(serverName, true, 0)

	if e != nil {
		c, _ = GetClientTimeout(serverName, false, 0)
	}

	return c
}

func GetClientError(serverName string) (*http.Client, errors.Error) {
	return GetClientTimeout(serverName, true, 0)
}

func GetClientTimeout(serverName string, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, errors.Error) {
	dl := &net.Dialer{}

	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DialContext:        dl.DialContext,
		DisableCompression: true,
		//nolint #staticcheck
		TLSClientConfig: certificates.GetTLSConfig(serverName),
	}

	return getclient(tr, http2Tr, GlobalTimeout)
}

func GetClientCustom(tr *http.Transport, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, errors.Error) {
	return getclient(tr, http2Tr, GlobalTimeout)
}

func getclient(tr *http.Transport, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, errors.Error) {
	if http2Tr {
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
