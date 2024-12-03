/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package auth

import (
	"crypto/tls"
	"strings"
)

const (
	strict  = "strict"
	require = "require"
	verify  = "verify"
	request = "request"
	none    = "none"
)

type ClientAuth tls.ClientAuthType

const (
	NoClientCert               = ClientAuth(tls.NoClientCert)
	RequestClientCert          = ClientAuth(tls.RequestClientCert)
	RequireAnyClientCert       = ClientAuth(tls.RequireAnyClientCert)
	VerifyClientCertIfGiven    = ClientAuth(tls.VerifyClientCertIfGiven)
	RequireAndVerifyClientCert = ClientAuth(tls.RequireAndVerifyClientCert)
)

func List() []ClientAuth {
	return []ClientAuth{
		NoClientCert,
		RequestClientCert,
		RequireAnyClientCert,
		VerifyClientCertIfGiven,
		RequireAndVerifyClientCert,
	}
}

func Parse(s string) ClientAuth {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "'", "", -1)
	s = strings.TrimSpace(s)

	switch {
	case strings.Contains(s, strict) || (strings.Contains(s, require) && strings.Contains(s, verify)):
		return RequireAndVerifyClientCert
	case strings.Contains(s, verify):
		return VerifyClientCertIfGiven
	case strings.Contains(s, require) && !strings.Contains(s, verify):
		return RequireAnyClientCert
	case strings.Contains(s, request):
		return RequestClientCert
	default:
		return NoClientCert
	}
}

func ParseInt(d int) ClientAuth {
	switch tls.ClientAuthType(d) {
	case tls.RequireAndVerifyClientCert:
		return RequireAndVerifyClientCert
	case tls.VerifyClientCertIfGiven:
		return VerifyClientCertIfGiven
	case tls.RequireAnyClientCert:
		return RequireAnyClientCert
	case tls.RequestClientCert:
		return RequestClientCert
	default:
		return NoClientCert
	}
}

func parseBytes(p []byte) ClientAuth {
	return Parse(string(p))
}
