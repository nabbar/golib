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

package oauth

import (
	"context"
	"net/http"

	"github.com/nabbar/golib/errors"
	"golang.org/x/oauth2"
)

func NewConfigOAuth(clientID, clientSecret, endpointToken, endpointAuth, redirectUrl string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpointAuth,
			TokenURL: endpointToken,
		},
		RedirectURL: redirectUrl,
		Scopes:      scopes,
	}
}

// ConfigGetAuthCodeUrl returns a URL to OAuth 2.0 provider's consent page
// that asks for permissions for the required scopes explicitly.
//
// State is a token to protect the user from CSRF attacks. You must
// always provide a non-empty string and validate that it matches the
// the state query parameter on your redirect callback.
// See http://tools.ietf.org/html/rfc6749#section-10.12 for more info.
//
// online may include true for AccessTypeOnline or false for AccessTypeOffline, as well
// as ApprovalForce.
func ConfigGetAuthCodeUrl(oa *oauth2.Config, state string, online bool) string {
	if online {
		return oa.AuthCodeURL(state, oauth2.AccessTypeOnline)
	}

	return oa.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func ConfigExchangeCode(oa *oauth2.Config, ctx context.Context, httpcli *http.Client, code string) (*http.Client, errors.Error) {
	if httpcli != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpcli)
	}

	if tok, err := oa.Exchange(ctx, code); err != nil {
		return nil, ErrorOAuthExchange.Error(err)
	} else {
		return oa.Client(ctx, tok), nil
	}
}

func NewClientFromToken(ctx context.Context, httpcli *http.Client, tokenOAuth string) *http.Client {
	if httpcli != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpcli)
	}

	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenOAuth},
	))
}

func NewClientFromTokenSource(ctx context.Context, httpcli *http.Client, token oauth2.TokenSource) *http.Client {
	if httpcli != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpcli)
	}

	return oauth2.NewClient(ctx, token)
}
