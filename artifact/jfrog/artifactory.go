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
	"net/url"
	"runtime"

	"github.com/nabbar/golib/logger"

	"github.com/jfrog/jfrog-client-go/artifactory"
	aauth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	jauth "github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/artifact/client"
	"github.com/nabbar/golib/errors"
)

func NewArtifactory(ctx context.Context, repos string, authDetail jauth.ServiceDetails, opts Options, tlsInsecure bool, certPath string, threadNumber int, dryRun bool) (artifact.Client, errors.Error) {
	s := config.NewConfigBuilder()
	s.SetServiceDetails(authDetail)

	if certPath != "" {
		s.SetCertificatesPath(certPath)
	}

	if tlsInsecure {
		s.SetInsecureTls(true)
	} else {
		s.SetInsecureTls(false)
	}

	if threadNumber > 0 {
		s.SetThreads(threadNumber)
	} else {
		s.SetThreads(runtime.GOMAXPROCS(0))
	}

	if dryRun {
		s.SetDryRun(true)
	} else {
		s.SetDryRun(false)
	}

	log.SetLogger(
		log.NewLogger(
			log.DEBUG,
			logger.GetIOWriter(logger.DebugLevel, "[Artifactory]", ""),
		),
	)

	if cfg, err := s.Build(); err != nil {
		return nil, ErrorClientInit.ErrorParent(err)
	} else if art, err := artifactory.New(&authDetail, cfg); err != nil {
		return nil, ErrorClientInit.ErrorParent(err)
	} else {
		a := &artifactoryModel{
			ClientHelper: client.ClientHelper{},
			c:            art,
			a:            authDetail,
			o:            opts,
			x:            ctx,
			r:            repos,
		}

		a.ClientHelper.F = a.ListReleases

		return a, nil
	}
}

func NewAuthUserPass(artifactoryUrl, user, pass string) jauth.ServiceDetails {
	a := aauth.NewArtifactoryDetails()
	a.SetUrl(artifactoryUrl)

	if user != "" {
		a.SetUser(user)
	}

	if pass != "" {
		a.SetPassword(pass)
	}

	return a
}

func NewAuthSSH(artifactoryUrl, sshUrl, sshKeyPath, sshPassPhrase string, sshHeader url.Values) jauth.ServiceDetails {
	a := aauth.NewArtifactoryDetails()
	a.SetUrl(artifactoryUrl)

	if sshUrl != "" {
		a.SetSshUrl(sshUrl)
	}

	if sshKeyPath != "" {
		a.SetSshKeyPath(sshKeyPath)
	}

	if sshPassPhrase != "" {
		a.SetSshPassphrase(sshPassPhrase)
	}

	if len(sshHeader) > 0 {
		var h = make(map[string]string)

		for k, v := range sshHeader {
			h[k] = v[0]
		}

		a.SetSshAuthHeaders(h)
	}

	return a
}

func NewAuthCert(artifactoryUrl, clientCertKeyPath, clientCertCrtPath string) jauth.ServiceDetails {
	a := aauth.NewArtifactoryDetails()
	a.SetUrl(artifactoryUrl)

	if clientCertKeyPath == "" {
		a.SetClientCertKeyPath(clientCertKeyPath)
	}
	if clientCertCrtPath == "" {
		a.SetClientCertPath(clientCertCrtPath)
	}

	return a
}

func NewAuthToken(artifactoryUrl, apiKey, accessToken string) jauth.ServiceDetails {
	a := aauth.NewArtifactoryDetails()
	a.SetUrl(artifactoryUrl)

	if apiKey == "" {
		a.SetApiKey(apiKey)
	}
	if accessToken == "" {
		a.SetAccessToken(accessToken)
	}

	return a
}

func NewOptionSearch(recursive bool, pattern, excludePatterns string, props, excludeProps url.Values) Options {
	if props == nil {
		props = make(url.Values)
	}

	if excludeProps == nil {
		excludeProps = make(url.Values)
	}

	return &artifactoryOptions{
		a: recursive,
		r: pattern,
		e: excludePatterns,
		p: props,
		x: excludeProps,
	}
}
