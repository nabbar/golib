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
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"
)

const (
	GitlabAPIBase    = "/api" // GitLab API base path
	GitlabAPIVersion = "/v4"  // GitLab API version path
)

// GetGitlabOptions constructs GitLab client options with proper API path configuration.
// It ensures the base URL includes the GitLab API base path (/api) and version (/v4).
//
// Parameters:
//   - baseUrl: GitLab instance URL (e.g., "https://gitlab.com" or "https://gitlab.example.com")
//   - httpcli: Optional HTTP client for custom transport/timeout configuration
//
// Returns client options that can be passed to gitlab.NewClient or similar constructors.
//
// Example:
//
//	opts, err := GetGitlabOptions("https://gitlab.example.com", nil)
//	// Results in baseUrl: "https://gitlab.example.com/api/v4"
func GetGitlabOptions(baseUrl string, httpcli *http.Client) (opt []gitlab.ClientOptionFunc, err error) {
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

// NewGitlabAuthUser creates a GitLab artifact client using basic authentication (username/password).
//
// Deprecated: Basic authentication is being phased out by GitLab. Use NewGitlabPrivateToken instead.
//
// Parameters:
//   - ctx: Request context
//   - htc: Optional HTTP client
//   - usr: GitLab username
//   - pwd: GitLab password
//   - uri: GitLab instance URL
//   - pid: Project ID (numeric)
func NewGitlabAuthUser(ctx context.Context, htc *http.Client, usr, pwd, uri string, pid int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = GetGitlabOptions(uri, htc); err != nil {
		return
	}

	if c, e = gitlab.NewBasicAuthClient(usr, pwd, o...); e != nil { // nolint
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, pid), err
}

// NewGitlabOAuth creates a GitLab artifact client using OAuth token authentication.
//
// Deprecated: Use NewGitlabAuthSource for OAuth2 token source authentication.
//
// Parameters:
//   - ctx: Request context
//   - htc: Optional HTTP client
//   - tkn: OAuth token
//   - uri: GitLab instance URL
//   - pid: Project ID (numeric)
func NewGitlabOAuth(ctx context.Context, htc *http.Client, tkn, uri string, pid int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = GetGitlabOptions(uri, htc); err != nil {
		return
	}

	if c, e = gitlab.NewOAuthClient(tkn, o...); e != nil { // nolint
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, pid), err
}

// NewGitlabAuthSource creates a GitLab artifact client using an OAuth2 token source.
// This method supports automatic token refresh and is recommended for OAuth2 authentication.
//
// Parameters:
//   - ctx: Request context
//   - htc: Optional HTTP client for custom transport/timeout
//   - uri: GitLab instance URL (e.g., "https://gitlab.com")
//   - pid: Project ID (numeric, found in project settings)
//   - tkn: OAuth2 token source (handles token refresh automatically)
//
// Returns a client implementing the artifact.Client interface.
func NewGitlabAuthSource(ctx context.Context, htc *http.Client, uri string, pid int, tkn oauth2.TokenSource) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = GetGitlabOptions(uri, htc); err != nil {
		return
	}

	if c, e = gitlab.NewAuthSourceClient(gitlab.OAuthTokenSource{TokenSource: tkn}, o...); e != nil { // nolint
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, pid), err
}

// NewGitlabPrivateToken creates a GitLab artifact client using a private token (recommended).
// This is the most common and recommended authentication method for GitLab.
//
// Parameters:
//   - ctx: Request context for API calls
//   - httpcli: Optional HTTP client for custom transport/timeout configuration
//   - token: GitLab private token (create at User Settings > Access Tokens)
//   - baseUrl: GitLab instance URL (e.g., "https://gitlab.com" or "https://gitlab.example.com")
//   - projectId: Project ID (numeric, found in project settings under "Project ID")
//
// Returns a client implementing the artifact.Client interface for:
//   - Listing releases with version filtering
//   - Retrieving artifact download URLs
//   - Streaming artifact downloads
//
// Example:
//
//	ctx := context.Background()
//	client, err := NewGitlabPrivateToken(ctx, nil, "glpat-xxxxx", "https://gitlab.com", 12345)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	versions, _ := client.ListReleases()
func NewGitlabPrivateToken(ctx context.Context, httpcli *http.Client, token, baseUrl string, projectId int) (cli libart.Client, err error) {
	var (
		o []gitlab.ClientOptionFunc
		c *gitlab.Client
		e error
	)

	if o, err = GetGitlabOptions(baseUrl, httpcli); err != nil {
		return
	}

	if c, e = gitlab.NewClient(token, o...); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	return newGitlab(ctx, c, projectId), err
}

// newGitlab is an internal constructor that wraps a GitLab API client
// into the artifact.Client interface with version management capabilities.
//
// It initializes the Helper with a reference to ListReleases for version organization.
func newGitlab(ctx context.Context, c *gitlab.Client, projectId int) libart.Client {
	a := &gitlabModel{
		Helper: artcli.Helper{},
		c:      c,
		x:      ctx,
		p:      projectId,
	}

	a.F = a.ListReleases

	return a
}
