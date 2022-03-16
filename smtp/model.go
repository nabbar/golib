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
	"runtime"
	"strings"
	"sync"
	"time"

	libsts "github.com/nabbar/golib/status"

	"github.com/nabbar/golib/errors"
)

type smtpClient struct {
	mut sync.Mutex
	con net.Conn
	cli *smtp.Client
	tls *tls.Config
	cfg Config
}

// Clone is used to clone current smtp pointer to a new one with same config.
func (s *smtpClient) Clone() SMTP {
	return &smtpClient{
		mut: sync.Mutex{},
		con: nil,
		cli: nil,
		tls: s.tls,
		cfg: s.cfg,
	}
}

// Close Terminate SMTP negotiation client and close connection.
func (s *smtpClient) Close() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s._close()
}

// Client Get SMTP Client interface.
func (s *smtpClient) Client(ctx context.Context) (*smtp.Client, errors.Error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s._client(ctx)
}

// Check Try to initiate SMTP dial and negotiation and try to close connection.
func (s *smtpClient) Check(ctx context.Context) errors.Error {
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
func (s *smtpClient) Send(ctx context.Context, from string, to []string, data io.WriterTo) errors.Error {
	//from smtp.SendMail()

	var (
		e error
		c *smtp.Client
		w io.WriteCloser

		err errors.Error
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

// StatusInfo is used to return the information part for the router status process
func (s *smtpClient) StatusInfo() (name string, release string, hash string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	hash = ""
	release = strings.TrimLeft(strings.ToLower(runtime.Version()), "go")
	name = fmt.Sprintf("SMTP %s:%d", s.cfg.GetHost(), s.cfg.GetPort())

	return name, release, hash
}

// StatusHealth is used to return the status of the SMTP connection to the server (with a timeout of 5 sec).
func (s *smtpClient) StatusHealth() error {
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return s.Check(ctx)
}

// StatusRouter is used to initiate a router status component and register it to the router status interface
func (s *smtpClient) StatusRouter(sts libsts.RouteStatus, prefix string) {
	if prefix != "" {
		prefix = fmt.Sprintf("%s SMTP %s:%d", prefix, s.cfg.GetHost(), s.cfg.GetPort())
	} else {
		prefix = fmt.Sprintf("SMTP %s:%d", s.cfg.GetHost(), s.cfg.GetPort())
	}

	cfg := s.cfg.GetStatusConfig()
	cfg.RegisterStatus(sts, prefix, s.StatusInfo, s.StatusHealth)
}

func (s *smtpClient) _ValidateLine(line string) errors.Error {
	if strings.ContainsAny(line, "\n\r") {
		return ErrorSMTPLineCRLF.Error(nil)
	}

	return nil
}

func (s *smtpClient) _client(ctx context.Context) (*smtp.Client, errors.Error) {
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
	} else if err = s.auth(cli, addr); err != nil {
		return nil, err
	} else {
		s.con = con
		s.cli = cli
	}

	return s.cli, nil
}

func (s *smtpClient) _close() {
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
