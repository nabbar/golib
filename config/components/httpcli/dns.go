/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
	"time"

	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

func (o *componentHttpClient) Add(endpoint string, ip string) {
	if d := o.getDNSMapper(); d != nil {
		d.Add(endpoint, ip)
		o.setDNSMapper(d)
	}
}

func (o *componentHttpClient) Get(endpoint string) string {
	if d := o.getDNSMapper(); d != nil {
		return d.Get(endpoint)
	}
	return ""
}

func (o *componentHttpClient) Del(endpoint string) {
	if d := o.getDNSMapper(); d != nil {
		d.Del(endpoint)
		o.setDNSMapper(d)
	}
}

func (o *componentHttpClient) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d := o.getDNSMapper(); d != nil {
		return d.DialContext(ctx, network, address)
	}

	return nil, ErrorComponentNotInitialized.Error()
}

func (o *componentHttpClient) Transport(cfg htcdns.TransportConfig) *http.Transport {
	if d := o.getDNSMapper(); d != nil {
		return d.Transport(cfg)
	}

	return nil
}

func (o *componentHttpClient) Client(cfg htcdns.TransportConfig) *http.Client {
	if d := o.getDNSMapper(); d != nil {
		return d.Client(cfg)
	}

	return nil
}

func (o *componentHttpClient) DefaultTransport() *http.Transport {
	if d := o.getDNSMapper(); d != nil {
		return d.DefaultTransport()
	}

	return nil
}

func (o *componentHttpClient) DefaultClient() *http.Client {
	if d := o.getDNSMapper(); d != nil {
		return d.DefaultClient()
	}

	return nil
}

func (o *componentHttpClient) TimeCleaner(ctx context.Context, dur time.Duration) {
	if d := o.getDNSMapper(); d != nil {
		d.TimeCleaner(ctx, dur)
	}
}
