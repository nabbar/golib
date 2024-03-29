/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package jfrog

import (
	"context"
	"net/http"
	"net/url"

	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
)

func NewArtifactory(ctx context.Context, Do func(req *http.Request) (*http.Response, error), uri, releaseRegex string, releaseGroup int, reposPath ...string) (libart.Client, error) {
	if u, e := url.Parse(uri); e != nil {
		return nil, ErrorURLParse.Error(e)
	} else {
		a := &artifactoryModel{
			ClientHelper: artcli.ClientHelper{},
			Do:           Do,
			ctx:          ctx,
			endpoint:     u,
			path:         reposPath,
			group:        releaseGroup,
			regex:        releaseRegex,
		}

		a.ClientHelper.F = a.ListReleases

		return a, nil
	}
}
