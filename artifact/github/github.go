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

	github "github.com/google/go-github/v76/github"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
)

// GetOrgProjectFromRepos parses a GitHub repository path and extracts the owner and project name.
// The repos parameter should be in the format "owner/project" (with or without leading/trailing slashes).
//
// Example:
//
//	owner, project := GetOrgProjectFromRepos("google/go-github")
//	// Returns: owner="google", project="go-github"
func GetOrgProjectFromRepos(repos string) (owner string, project string) {
	if strings.HasPrefix(repos, "/") || strings.HasSuffix(repos, "/") {
		repos = strings.Trim(repos, "/")
	}

	lst := strings.SplitN(repos, "/", 2)
	return lst[0], lst[1]
}

// NewGithub returns a new Github client.
//
// The context is used to set the http client.
// The httpcli parameter is used to set the http client.
// The repos parameter is used to get the owner and project from the repository path.
//
// The returned client can be used to list releases and download artifacts.
func NewGithub(ctx context.Context, httpcli *http.Client, repos string) (cli libart.Client, err error) {
	o, p := GetOrgProjectFromRepos(repos)

	a := &githubModel{
		Helper: artcli.Helper{},
		c:      github.NewClient(httpcli),
		x:      ctx,
		o:      o,
		p:      p,
	}

	a.F = a.ListReleases

	return a, err
}

// NewGithubWithTokenOAuth returns a new Github client with OAuth2 authentication.
//
// The context is used to set the http client.
// The repos parameter is used to get the owner and project from the repository path.
// The oauth2client parameter is used to set the OAuth2 authenticated http client.
//
// The returned client can be used to list releases and download artifacts.
func NewGithubWithTokenOAuth(ctx context.Context, repos string, oauth2client *http.Client) (cli libart.Client, err error) {
	o, p := GetOrgProjectFromRepos(repos)

	a := &githubModel{
		Helper: artcli.Helper{},
		c:      github.NewClient(oauth2client),
		x:      ctx,
		o:      o,
		p:      p,
	}

	a.F = a.ListReleases

	return a, err
}
