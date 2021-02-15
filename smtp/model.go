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
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"

	"github.com/nabbar/golib/errors"
)

type smtpClient struct {
	con net.Conn
	cli *smtp.Client
	tls *tls.Config
	cfg Config
}

// Check Try to initiate SMTP dial and negotiation and try to close connection.
func (s smtpClient) Clone() SMTP {
	return &smtpClient{
		con: nil,
		cli: nil,
		tls: s.tls,
		cfg: s.cfg,
	}
}

// Close Terminate SMTP negotiation client and close connection.
func (s *smtpClient) Close() {
	if s.cli != nil {
		if e := s.cli.Quit(); e != nil {
			_ = s.cli.Close()
		}
		s.cli = nil
	}

	if s.con != nil {
		_ = s.con.Close()
		s.con = nil
	}
}

// Check Try to initiate SMTP dial and negotiation and try to close connection.
func (s *smtpClient) Check(ctx context.Context) errors.Error {
	defer s.Close()

	if s.cli == nil {
		if _, e := s.Client(ctx); e != nil {
			return e
		}
	}

	if e := s.cli.Noop(); e != nil {
		return ErrorSMTPClientNoop.ErrorParent(e)
	}

	return nil
}

// Client Get SMTP Client interface.
func (s *smtpClient) Client(ctx context.Context) (*smtp.Client, errors.Error) {
	if s.cli == nil {
		var (
			addr = s.cfg.GetHost()
			tlsc = s.tls.Clone()
		)

		if s.cfg.GetTlSServerName() != "" && s.cfg.GetNet() != NET_UNIX {
			tlsc.ServerName = s.cfg.GetTlSServerName()
		}

		if s.cfg.IsTLSSkipVerify() && s.cfg.GetNet() != NET_UNIX {
			tlsc.InsecureSkipVerify = true
		}

		if s.cfg.GetPort() > 0 {
			addr = fmt.Sprintf("%s:%v", s.cfg.GetHost(), s.cfg.GetPort())
		}

		if cli, con, err := s.tryClient(ctx, addr, tlsc); err != nil {
			return nil, err
		} else if err := s.auth(cli, addr); err != nil {
			return nil, err
		} else {
			s.Close()
			s.con = con
			s.cli = cli
		}
	}

	return s.cli, nil
}

// validateLine checks to see if a line has CR or LF as per RFC 5321
func (s smtpClient) validateLine(line string) errors.Error {
	if strings.ContainsAny(line, "\n\r") {
		return ErrorSMTPLineCRLF.Error(nil)
	}

	return nil
}

func (s *smtpClient) Send(ctx context.Context, from string, to []string, data io.WriterTo) errors.Error {
	//from smtp.SendMail()

	var (
		e error
		w io.WriteCloser
	)

	defer func() {
		if w != nil {
			_ = w.Close()
		}

		s.Close()
	}()

	if err := s.validateLine(from); err != nil {
		return err
	}

	for _, recp := range to {
		if err := s.validateLine(recp); err != nil {
			return err
		}
	}

	if err := s.Check(ctx); err != nil {
		return err
	}

	if _, e = s.Client(ctx); e != nil {
		return ErrorSMTPClientInit.ErrorParent(e)
	}

	if e = s.cli.Mail(from); e != nil {
		return ErrorSMTPClientMail.ErrorParent(e)
	}

	for _, addr := range to {
		if e = s.cli.Rcpt(addr); e != nil {
			return ErrorSMTPClientRcpt.ErrorParent(e)
		}
	}

	if w, e = s.cli.Data(); e != nil {
		return ErrorSMTPClientData.ErrorParent(e)
	}

	if _, e = data.WriteTo(w); e != nil {
		return ErrorSMTPClientWrite.ErrorParent(e)
	}

	return nil
}
