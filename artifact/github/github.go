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

package github

import (
	"context"
	"net/http"
	"strings"

	github "github.com/google/go-github/v33/github"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
)

func getOrgProjectFromRepos(repos string) (owner string, project string) {
	if strings.HasPrefix(repos, "/") || strings.HasSuffix(repos, "/") {
		repos = strings.Trim(repos, "/")
	}

	lst := strings.SplitN(repos, "/", 2)
	return lst[0], lst[1]
}

func NewGithub(ctx context.Context, httpcli *http.Client, repos string) (cli libart.Client, err error) {
	o, p := getOrgProjectFromRepos(repos)

	a := &githubModel{
		ClientHelper: artcli.ClientHelper{},
		c:            github.NewClient(httpcli),
		x:            ctx,
		o:            o,
		p:            p,
	}

	a.F = a.ListReleases

	return a, err
}

func NewGithubWithTokenOAuth(ctx context.Context, repos string, oauth2client *http.Client) (cli libart.Client, err error) {
	o, p := getOrgProjectFromRepos(repos)

	a := &githubModel{
		ClientHelper: artcli.ClientHelper{},
		c:            github.NewClient(oauth2client),
		x:            ctx,
		o:            o,
		p:            p,
	}

	a.F = a.ListReleases

	return a, err
}
