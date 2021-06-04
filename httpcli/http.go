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

package httpcli

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type httpClient struct {
	url *url.URL
	cli *http.Client
	ctx context.Context
}

type HTTP interface {
	SetContext(ctx context.Context)
	Check() liberr.Error
	Call(file *bytes.Buffer) (bool, *bytes.Buffer, liberr.Error)
}

func NewClient(uri string) (HTTP, liberr.Error) {
	var (
		pUri *url.URL
		err  error
		host string
	)

	if uri != "" {
		pUri, err = url.Parse(uri)

		if err != nil {
			return nil, URL_PARSE.ErrorParent(err)
		}

		host = pUri.Host
	} else {
		pUri = nil
		host = ""
	}

	c, e := GetClientError(host)

	if e != nil {
		return nil, HTTP_CLIENT.Error(e)
	}

	return &httpClient{
		url: pUri,
		cli: c,
		ctx: context.Background(),
	}, nil
}

func (obj *httpClient) SetContext(ctx context.Context) {
	if ctx != nil {
		obj.ctx = ctx
	}
}

func (obj *httpClient) Check() liberr.Error {
	req, e := obj.newRequest(http.MethodHead, nil)

	if e != nil {
		return e
	}

	var r *http.Response

	r, e = obj.doRequest(req)

	if r != nil && r.Body != nil {
		_ = r.Body.Close()
	}

	return e
}

func (obj *httpClient) Call(body *bytes.Buffer) (bool, *bytes.Buffer, liberr.Error) {
	req, e := obj.newRequest(http.MethodPost, body)

	if e != nil {
		return false, nil, e
	}

	res, e := obj.doRequest(req)

	if e != nil {
		return false, nil, e
	}

	return obj.checkResponse(res)
}

func (obj *httpClient) newRequest(method string, body *bytes.Buffer) (*http.Request, liberr.Error) {
	var reader *bytes.Reader

	if body != nil && body.Len() > 0 {
		reader = bytes.NewReader(body.Bytes())
	}

	req, e := http.NewRequestWithContext(obj.ctx, method, obj.url.String(), reader)
	if e != nil {
		return req, HTTP_REQUEST.ErrorParent(e)
	}

	return req, nil
}

func (obj *httpClient) doRequest(req *http.Request) (*http.Response, liberr.Error) {
	res, e := obj.cli.Do(req)

	if e != nil {
		return res, HTTP_DO.ErrorParent(e)
	}

	return res, nil
}

func (obj *httpClient) checkResponse(res *http.Response) (bool, *bytes.Buffer, liberr.Error) {
	var buf *bytes.Buffer

	if res.Body != nil {
		bdy, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return false, nil, IO_READ.ErrorParent(err)
		}

		_, err = buf.Write(bdy)

		if err != nil {
			return false, nil, BUFFER_WRITE.ErrorParent(err)
		}

		liblog.GetDefault().Entry(liblog.DebugLevel, "").ErrorAdd(true, err).FieldAdd("remote.uri", res.Request.URL.String()).FieldAdd("remote.method", res.Request.Method).Log()
	}

	return strings.HasPrefix(res.Status, "2"), buf, nil
}
