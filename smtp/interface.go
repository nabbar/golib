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
	"net/smtp"
	"sync"

	libctx "github.com/nabbar/golib/context"
	montps "github.com/nabbar/golib/monitor/types"
	smtpcf "github.com/nabbar/golib/smtp/config"
	libver "github.com/nabbar/golib/version"
)

type SMTP interface {
	Clone() SMTP
	Close()

	UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config)

	Client(ctx context.Context) (*smtp.Client, error)
	Check(ctx context.Context) error
	Send(ctx context.Context, from string, to []string, data io.WriterTo) error

	Monitor(ctx libctx.FuncContext, vrs libver.Version) (montps.Monitor, error)
}

// New return a SMTP interface to operation negotiation with a SMTP server.
// the dsn parameter must be string like this '[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]".
//   - params available are : ServerName (string), SkipVerify (boolean).
//   - tls mode acceptable are :  starttls, tls, <any other value to no tls/startls>.
//   - net acceptable are : tcp4, tcp6, unix.
func New(cfg smtpcf.SMTP, tlsConfig *tls.Config) (SMTP, error) {
	if tlsConfig == nil {
		/* #nosec */
		//nolint #nosec
		tlsConfig = &tls.Config{}
	}

	if cfg == nil {
		return nil, ErrorParamEmpty.Error(nil)
	} else {
		return &smtpClient{
			mut: sync.Mutex{},
			cfg: cfg,
			tls: tlsConfig,
		}, nil
	}
}
