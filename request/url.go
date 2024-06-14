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

package request

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

func (r *request) SetEndpoint(u string) error {
	if uri, err := url.Parse(u); err != nil {
		return err
	} else {
		r.mux.Lock()
		defer r.mux.Unlock()

		r.uri = uri
		return nil
	}
}

func (r *request) GetEndpoint() string {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.uri.String()
}

func (r *request) SetPath(raw bool, path string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if raw {
		r.uri.RawPath = path
	} else {
		r.uri.Path = path
	}
}

func (r *request) AddPath(raw bool, pathPart ...string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.uri == nil {
		return
	}

	var str string
	if raw {
		str = path.Clean(r.uri.RawPath)
	} else {
		str = path.Clean(r.uri.Path)
	}

	for i := range pathPart {
		if strings.HasSuffix(str, "/") && strings.HasPrefix(pathPart[i], "/") {
			pathPart[i] = strings.TrimPrefix(pathPart[i], "/")
		}

		if strings.HasSuffix(pathPart[i], "/") {
			pathPart[i] = strings.TrimSuffix(pathPart[i], "/")
		}

		str = path.Join(str, pathPart[i])
	}

	if raw {
		r.uri.RawPath = path.Clean(str)
	} else {
		r.uri.Path = path.Clean(str)
	}
}

func (r *request) SetMethod(method string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	switch strings.ToUpper(method) {
	case http.MethodGet:
		r.mth = http.MethodGet
	case http.MethodHead:
		r.mth = http.MethodHead
	case http.MethodPost:
		r.mth = http.MethodPost
	case http.MethodPut:
		r.mth = http.MethodPut
	case http.MethodPatch:
		r.mth = http.MethodPatch
	case http.MethodDelete:
		r.mth = http.MethodDelete
	case http.MethodConnect:
		r.mth = http.MethodConnect
	case http.MethodOptions:
		r.mth = http.MethodOptions
	case http.MethodTrace:
		r.mth = http.MethodTrace
	default:
		r.mth = strings.ToUpper(method)
	}

	if r.mth == "" {
		r.mth = http.MethodGet
	}
}

func (r *request) GetMethod() string {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.mth
}
