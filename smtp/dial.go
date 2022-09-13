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
	"net"
	"net/smtp"

	liberr "github.com/nabbar/golib/errors"
)

func (s *smtpClient) dialTLS(ctx context.Context, addr string, tlsConfig *tls.Config) (con net.Conn, err liberr.Error) {
	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	if con, err = s.dial(ctx, addr); err != nil {
		return
	}

	if con = tls.Client(con, tlsConfig); con == nil {
		err = ErrorSMTPClientStartTLS.Error(nil)
	}

	return
}

func (s *smtpClient) dial(ctx context.Context, addr string) (con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	d := net.Dialer{}

	if con, e = d.DialContext(ctx, s.cfg.GetNet().string(), addr); e != nil {
		return con, ErrorSMTPDial.ErrorParent(e)
	}

	return
}

func (s *smtpClient) client(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil {
			if cli != nil {
				_ = cli.Close()
			}
			if con != nil {
				_ = con.Close()
			}
		}
	}()

	if s.cfg.GetTlsMode() == TLS_STARTTLS && tlsConfig == nil {
		err = ErrorParamEmpty.Error(nil)
		return
	} else if s.cfg.GetTlsMode() == TLS_TLS && tlsConfig == nil {
		err = ErrorParamEmpty.Error(nil)
		return
	}

	if s.cfg.GetTlsMode() == TLS_TLS && tlsConfig != nil {
		if con, err = s.dialTLS(ctx, addr, tlsConfig); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.ErrorParent(e)
			return
		}
	} else {
		if con, err = s.dial(ctx, addr); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.ErrorParent(e)
			return
		}

		try := s.checkExtension(cli, "STARTTLS")

		if s.cfg.GetTlsMode() == TLS_STARTTLS || try {
			if e = cli.StartTLS(tlsConfig); e != nil && !try {
				err = ErrorSMTPClientStartTLS.ErrorParent(e)
				return
			} else if e == nil && try {
				s.cfg.SetTlsMode(TLS_STARTTLS)
			}
		}
	}

	return
}

func (s *smtpClient) tryClient(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	cli, con, err = s.client(ctx, addr, tlsConfig)

	if err == nil {
		return
	}

	switch s.cfg.GetTlsMode() {
	case TLS_TLS:
		s.cfg.SetTlsMode(TLS_STARTTLS)
		return s.tryClient(ctx, addr, tlsConfig)
	case TLS_STARTTLS, TLS_NONE:
		return
	}

	return
}

func (s *smtpClient) auth(cli *smtp.Client, addr string) liberr.Error {
	usr := s.cfg.GetUser()
	pwd := s.cfg.GetPass()
	err := ErrorSMTPClientAuth.Error(nil)

	if usr == "" && pwd == "" {
		return nil
	}
	/*
		if e := cli.Auth(smtp.CRAMMD5Auth(usr, pwd)); e != nil {
			err.AddParent(e)
		} else {
			return nil
		}
	*/
	if e := cli.Auth(smtp.PlainAuth("", usr, pwd, addr)); e != nil {
		err.AddParent(e)
	} else {
		return nil
	}

	return err
}

func (s *smtpClient) checkExtension(cli *smtp.Client, ext string) bool {
	ok, _ := cli.Extension(ext)
	return ok
}
