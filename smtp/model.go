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
	"crypto/tls"
	"net"
	"net/smtp"
	"strings"
	"sync"

	liberr "github.com/nabbar/golib/errors"
	smtpcf "github.com/nabbar/golib/smtp/config"
)

type smtpClient struct {
	mut sync.Mutex
	con net.Conn
	cli *smtp.Client
	tls *tls.Config
	cfg smtpcf.SMTP
}

func (s *smtpClient) _ValidateLine(line string) liberr.Error {
	if strings.ContainsAny(line, "\n\r") {
		return ErrorSMTPLineCRLF.Error(nil)
	}

	return nil
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

// UpdConfig is used to update the config & TLS Config for an instance.
func (s *smtpClient) UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config) {
	s._close()

	s.mut.Lock()
	defer s.mut.Unlock()

	s.cfg = cfg
	s.tls = tslConfig
}
