/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package smtp

import (
	"context"
	"fmt"
	"io"
	"net/smtp"

	liberr "github.com/nabbar/golib/errors"
	smtpnt "github.com/nabbar/golib/smtp/network"
)

func (s *smtpClient) _client(ctx context.Context) (*smtp.Client, liberr.Error) {
	if s.cli != nil && s.con != nil {
		return s.cli, nil
	}

	if s.con == nil && s.cli != nil {
		if e := s.cli.Quit(); e != nil {
			_ = s.cli.Close()
		}
	} else if s.con != nil && s.cli == nil {
		_ = s.con.Close()
	}

	s.cli = nil
	s.con = nil

	var (
		addr = s.cfg.GetHost()
		tlsc = s.tls.Clone()
	)

	if s.cfg.GetTlSServerName() != "" && s.cfg.GetNet() != smtpnt.NetworkUnixSocket {
		tlsc.ServerName = s.cfg.GetTlSServerName()
	}

	if s.cfg.IsTLSSkipVerify() && s.cfg.GetNet() != smtpnt.NetworkUnixSocket {
		tlsc.InsecureSkipVerify = true
	}

	if s.cfg.GetPort() > 0 {
		addr = fmt.Sprintf("%s:%v", s.cfg.GetHost(), s.cfg.GetPort())
	}

	if cli, con, err := s.tryClient(ctx, addr, tlsc); err != nil {
		return nil, err
	} else if err = s.auth(cli, addr); err != nil {
		return nil, err
	} else {
		s.con = con
		s.cli = cli
	}

	return s.cli, nil
}

// Client Get SMTP Client interface.
func (s *smtpClient) Client(ctx context.Context) (*smtp.Client, liberr.Error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s._client(ctx)
}

// Check Try to initiate SMTP dial and negotiation and try to close connection.
func (s *smtpClient) Check(ctx context.Context) liberr.Error {
	s.mut.Lock()
	defer func() {
		s._close()
		s.mut.Unlock()
	}()

	if c, err := s._client(ctx); err != nil {
		return err
	} else if e := c.Noop(); e != nil {
		return ErrorSMTPClientNoop.ErrorParent(e)
	}

	return nil
}

// Send is used to initiate the smtp connection with the client and send a mail before closing the connection.
// This function is based on smtp.SendMail function.
func (s *smtpClient) Send(ctx context.Context, from string, to []string, data io.WriterTo) liberr.Error {
	//from smtp.SendMail()

	var (
		e error
		c *smtp.Client
		w io.WriteCloser

		err liberr.Error
	)

	s.mut.Lock()

	defer func() {
		if w != nil {
			_ = w.Close()
		}

		//mandatory for SMTP protocol
		s._close()

		s.mut.Unlock()
	}()

	if err = s._ValidateLine(from); err != nil {
		return err
	}

	for _, recp := range to {
		if err = s._ValidateLine(recp); err != nil {
			return err
		}
	}

	if c, err = s._client(ctx); err != nil {
		return ErrorSMTPClientInit.Error(err)
	}

	if e = c.Noop(); e != nil {
		return ErrorSMTPClientNoop.ErrorParent(e)
	}

	if e = c.Mail(from); e != nil {
		return ErrorSMTPClientMail.ErrorParent(e)
	}

	for _, addr := range to {
		if e = c.Rcpt(addr); e != nil {
			return ErrorSMTPClientRcpt.ErrorParent(e)
		}
	}

	if w, e = c.Data(); e != nil {
		return ErrorSMTPClientData.ErrorParent(e)
	}

	if _, e = data.WriteTo(w); e != nil {
		return ErrorSMTPClientWrite.ErrorParent(e)
	}

	return nil
}
