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

package gitlab

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	GitlabAPIBase    = "/api"
	GitlabAPIVersion = "/v4"
)

func getGitlbaOptions(baseUrl string, httpcli *http.Client) (opt []gitlab.ClientOptionFunc, err error) {
	var (
		u *url.URL
		e error
	)

	opt = make([]gitlab.ClientOptionFunc, 0)

	if u, e = url.Parse(baseUrl); e != nil {
		return opt, ErrorURLParse.Error(e)
	}

	if !strings.Contains(u.Path, GitlabAPIBase) {
		u.Path += GitlabAPIBase
	}

	if !strings.Contains(u.Path, GitlabAPIVersion) {
		u.Path += GitlabAPIVersion
	}

	opt = append(opt, gitlab.WithBaseURL(u.String()))

	if httpcli != nil {
		opt = append(opt, gitlab.WithHTTPClient(httpcli))
	}

	return
}

func newGitlab(ctx context.Context, c *gitlab.Client, projectId int) libart.Client {
	a := &gitlabModel{
		ClientHelper: artcli.ClientHelper{},
		c:            c,
		x:            ctx,
		p:            projectId,
	}

	a.F = a.ListReleases

	return a
}

func NewGitlabAuthUser(ctx context.Context, httpcli *http.Client, user, pass, baseUrl string, projectId int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = getGitlbaOptions(baseUrl, httpcli); err != nil {
		return
	}

	if c, e = gitlab.NewBasicAuthClient(user, pass, o...); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, projectId), err
}

func NewGitlabOAuth(ctx context.Context, httpcli *http.Client, oAuthToken, baseUrl string, projectId int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = getGitlbaOptions(baseUrl, httpcli); err != nil {
		return
	}

	if c, e = gitlab.NewOAuthClient(oAuthToken, o...); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, projectId), err
}

func NewGitlabPrivateToken(ctx context.Context, httpcli *http.Client, token, baseUrl string, projectId int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = getGitlbaOptions(baseUrl, httpcli); err != nil {
		return
	}

	if c, e = gitlab.NewClient(token, o...); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, projectId), err
}
