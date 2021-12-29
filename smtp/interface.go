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

package smtp

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/nabbar/golib/errors"
)

type Config interface {
	SetHost(host string)
	GetHost() string

	SetPort(port int)
	GetPort() int

	SetUser(user string)
	GetUser() string

	SetPass(pass string)
	GetPass() string

	SetNet(mode NETMode)
	GetNet() NETMode

	SetTls(mode TLSMode)
	GetTls() TLSMode

	ForceTLSSkipVerify(skip bool)
	IsTLSSkipVerify() bool

	SetTLSServerName(serverName string)
	GetTlSServerName() string

	GetDsn() string
}

type SMTP interface {
	Client(ctx context.Context) (*smtp.Client, errors.Error)
	Close()
	Check(ctx context.Context) errors.Error
	Clone() SMTP
	Send(ctx context.Context, from string, to []string, data io.WriterTo) errors.Error
}

// NewSMTP return a SMTP interface to operation negotiation with a SMTP server.
// the dsn parameter must be string like this '[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]".
//   - params available are : ServerName (string), SkipVerify (boolean).
//   - tls mode acceptable are :  starttls, tls, <any other value to no tls/startls>.
//   - net aceeptable are : tcp4, tcp6, unix.
func NewSMTP(cfg Config, tlsConfig *tls.Config) (SMTP, errors.Error) {
	if tlsConfig == nil {
		/* #nosec */
		//nolint #nosec
		tlsConfig = &tls.Config{}
	}

	if cfg == nil {
		return nil, ErrorParamsEmpty.Error(nil)
	} else {
		return &smtpClient{
			mut: sync.Mutex{},
			cfg: cfg,
			tls: tlsConfig,
		}, nil
	}
}

// NewConfig parses the DSN string to a Config.
// nolint: gocognit
func NewConfig(dsn string) (Config, errors.Error) {
	var (
		smtpcnf = &smtpConfig{
			DSN: dsn,
		}
	)

	// [user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	if !strings.ContainsRune(dsn, '?') && !strings.ContainsRune(dsn, '/') {
		dsn += "/"
	} else if strings.ContainsRune(dsn, '?') && !strings.ContainsRune(dsn, '/') {
		v := strings.Split(dsn, "?")
		v[len(v)-2] += "/"
		dsn = strings.Join(v, "?")
	}

	foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			foundSlash = true
			var j, k int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								smtpcnf.Pass = dsn[k+1 : j]
								break
							}
						}
						smtpcnf.User = dsn[:k]

						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return nil, ErrorConfigInvalidDSN.Error(nil)
							}
							return nil, ErrorConfigInvalidNetwork.Error(nil)
						}

						if strings.ContainsRune(dsn[k+1:i-1], ':') {
							h, p, e := net.SplitHostPort(dsn[k+1 : i-1])
							if e == nil && p != "" {
								pint, er := strconv.ParseInt(p, 10, 64)
								if er == nil {
									smtpcnf.Port = int(pint)
								}
								smtpcnf.Host = h
							}
						}

						if smtpcnf.Host == "" || smtpcnf.Port == 0 {
							smtpcnf.Host = dsn[k+1 : i-1]
						}

						break
					}
				}

				smtpcnf.Net = parseNETMode(dsn[j+1 : k])

			}

			// [?param1=value1&...&paramN=valueN]
			// Find the first '?' in dsn[i+1:]
			for j = i + 1; j < len(dsn); j++ {
				if dsn[j] == '?' {

					if val, err := url.ParseQuery(dsn[j+1:]); err != nil {
						return nil, ErrorConfigInvalidParams.ErrorParent(err)
					} else {

						if val.Get("ServerName") != "" {
							smtpcnf.ServerName = val.Get("ServerName")
						}

						if val.Get("SkipVerify") != "" {
							vi, e := strconv.ParseBool(val.Get("SkipVerify"))
							if e == nil {
								smtpcnf.SkipVerify = vi
							}
						}
					}

					break
				}
			}

			smtpcnf.TLS = parseTLSMode(dsn[i+1 : j])
			break
		}
	}

	if !foundSlash && len(dsn) > 0 {
		return nil, ErrorConfigInvalidHost.Error(nil)
	}

	return smtpcnf, nil
}
