/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package njs_httpcli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	njs_certif "github.com/nabbar/golib/njs-certif"

	njs_logger "github.com/nabbar/golib/njs-logger"
)

type httpClient struct {
	url *url.URL
	cli *http.Client
}

type HTTP interface {
	Check() bool
	Call(file *bytes.Buffer) (bool, *bytes.Buffer)
}

func NewClient(uri string) HTTP {
	var (
		pUri *url.URL
		err  error
		host string
	)

	if uri != "" {
		pUri, err = url.Parse(uri)
		njs_logger.PanicLevel.LogErrorCtx(njs_logger.NilLevel, fmt.Sprintf("parsing url '%s'", uri), err)
		host = pUri.Host
	} else {
		pUri = nil
		host = ""
	}

	return &httpClient{
		url: pUri,
		cli: GetClient(host),
	}
}

func GetClient(serverName string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableCompression:    true,
			TLSClientConfig:       njs_certif.GetTLSConfig(serverName),
		},
	}
}

func (obj *httpClient) Check() bool {
	obj.doRequest(obj.newRequest(http.MethodHead, nil))
	return true
}

func (obj *httpClient) Call(body *bytes.Buffer) (bool, *bytes.Buffer) {
	return obj.checkResponse(
		obj.doRequest(
			obj.newRequest(http.MethodPost, body),
		),
	)
}

func (obj *httpClient) newRequest(method string, body *bytes.Buffer) *http.Request {
	var reader *bytes.Reader

	if body != nil && body.Len() > 0 {
		reader = bytes.NewReader(body.Bytes())
	}

	req, err := http.NewRequest(method, obj.url.String(), reader)
	njs_logger.PanicLevel.LogErrorCtx(njs_logger.NilLevel, fmt.Sprintf("creating '%s' request to '%s'", method, obj.url.Host), err)

	return req
}

func (obj *httpClient) doRequest(req *http.Request) *http.Response {
	res, err := obj.cli.Do(req)
	njs_logger.PanicLevel.LogErrorCtx(njs_logger.NilLevel, fmt.Sprintf("running request '%s:%s'", req.Method, req.URL.Host), err)

	return res
}

func (obj *httpClient) checkResponse(res *http.Response) (bool, *bytes.Buffer) {
	var buf *bytes.Buffer

	if res.Body != nil {
		bdy, err := ioutil.ReadAll(res.Body)

		if err == nil {
			_, err = buf.Write(bdy)
		}

		njs_logger.DebugLevel.LogError(err)
	}

	njs_logger.InfoLevel.Logf("Calling '%s:%s' result %s (Body : %d bytes)", res.Request.Method, res.Request.URL.Host, res.Status, buf.Len())

	return strings.HasPrefix(res.Status, "2"), buf
}
