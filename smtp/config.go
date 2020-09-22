/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"

	"github.com/nabbar/golib/certificates"
	"github.com/nabbar/golib/errors"
)

type smtpConfig struct {
	DSN        string
	Host       string
	Port       int
	User       string
	Pass       string
	Net        NETMode
	TLS        TLSMode
	SkipVerify bool
	ServerName string
}

type smtpClient struct {
	con net.Conn
	cli *smtp.Client
	tls *tls.Config
	cfg *smtpConfig
}

type SMTP interface {
	Client() (*smtp.Client, errors.Error)
	Close()
	Check() errors.Error
	Clone() SMTP

	ForceHost(host string)
	ForcePort(port int)
	ForceUser(user string)
	ForcePass(pass string)
	ForceNet(mode NETMode)
	ForceTls(mode TLSMode)
	ForceTLSSkipVerify(skip bool)
	ForceTLSServerName(serverName string)

	GetDsn() string
}

type TLSMode uint8

const (
	TLS_NONE TLSMode = iota
	TLS_STARTTLS
	TLS_TLS
)

func parseTLSMode(str string) TLSMode {
	switch strings.ToLower(str) {
	case TLS_TLS.string():
		return TLS_TLS
	case TLS_STARTTLS.string():
		return TLS_STARTTLS
	}

	return TLS_NONE
}

func (tlm TLSMode) string() string {
	switch tlm {
	case TLS_TLS:
		return "tls"
	case TLS_STARTTLS:
		return "starttls"
	case TLS_NONE:
		return "none"
	}

	return TLS_NONE.string()
}

type NETMode uint8

const (
	NET_TCP NETMode = iota
	NET_TCP_4
	NET_TCP_6
	NET_UNIX
)

func parseNETMode(str string) NETMode {
	switch strings.ToLower(str) {
	case NET_TCP_4.string():
		return NET_TCP_4
	case NET_TCP_6.string():
		return NET_TCP_6
	case NET_UNIX.string():
		return NET_UNIX
	}

	return NET_TCP
}

func (n NETMode) string() string {
	switch n {
	case NET_TCP_4:
		return "tcp4"
	case NET_TCP_6:
		return "tcp6"
	case NET_UNIX:
		return "unix"
	case NET_TCP:
		return "tcp"
	}

	return NET_TCP.string()
}

// ParseDSN parses the DSN string to a Config.
// nolint: gocognit
func newSMTPConfig(dsn string) (*smtpConfig, errors.Error) {
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

// NewSMTP return a SMTP interface to operation negotiation with a SMTP server.
// the dsn parameter must be string like this '[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]".
//   - params available are : ServerName (string), SkipVerify (boolean).
//   - tls mode acceptable are :  starttls, tls, <any other value to no tls/startls>.
//   - net aceeptable are : tcp4, tcp6, unix.
func NewSMTP(dsn string, tlsConfig *tls.Config) (SMTP, errors.Error) {
	if tlsConfig == nil {
		tlsConfig = certificates.GetTLSConfig("")
	}

	if c, e := newSMTPConfig(dsn); e != nil {
		return nil, e
	} else {
		return &smtpClient{
			cfg: c,
			tls: tlsConfig,
		}, nil
	}
}

// Client Get SMTP Client interface.
func (s *smtpClient) Client() (*smtp.Client, errors.Error) {
	if s.cli == nil {
		var (
			err  error
			addr = s.cfg.Host
			tlsc = s.tls.Clone()
		)

		if s.cfg.ServerName != "" && s.cfg.Net != NET_UNIX {
			tlsc.ServerName = s.cfg.ServerName
		}

		if s.cfg.SkipVerify && s.cfg.Net != NET_UNIX {
			tlsc.InsecureSkipVerify = true
		}

		if s.cfg.Port > 0 {
			addr = fmt.Sprintf("%s:%v", s.cfg.Host, s.cfg.Port)
		}

		if s.cfg.TLS == TLS_TLS {
			s.con, err = tls.Dial(s.cfg.Net.string(), addr, tlsc)
			if err != nil {
				s.cfg.TLS = TLS_STARTTLS
			}
		}

		if s.cfg.TLS != TLS_TLS {
			s.con, err = net.Dial(s.cfg.Net.string(), addr)
			if err != nil {
				return nil, ErrorSMTPDial.ErrorParent(err)
			}
		}

		s.cli, err = smtp.NewClient(s.con, addr)
		if err != nil {
			return nil, ErrorSMTPClientInit.ErrorParent(err)
		}

		if s.cfg.TLS == TLS_STARTTLS {
			err = s.cli.StartTLS(tlsc)
			if err != nil {
				return nil, ErrorSMTPClientStartTLS.ErrorParent(err)
			}
		}

		if s.cfg.User != "" || s.cfg.Pass != "" {
			err = s.cli.Auth(smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, addr))
			if err != nil {
				return nil, ErrorSMTPClientAuth.ErrorParent(err)
			}
		}

		s.cli.Extension("8BITMIME")
	}

	return s.cli, nil
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
func (s *smtpClient) Check() errors.Error {
	defer s.Close()

	if c, e := s.Client(); e != nil {
		return e
	} else if e := c.Noop(); e != nil {
		return ErrorSMTPClientNoop.ErrorParent(e)
	}

	return nil
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

func (s *smtpClient) ForceHost(host string) {
	s.cfg.Host = host
}

func (s *smtpClient) ForcePort(port int) {
	s.cfg.Port = port
}

func (s *smtpClient) ForceUser(user string) {
	s.cfg.User = user
}

func (s *smtpClient) ForcePass(pass string) {
	s.cfg.Pass = pass
}

func (s *smtpClient) ForceNet(mode NETMode) {
	s.cfg.Net = mode
}

func (s *smtpClient) ForceTls(mode TLSMode) {
	s.cfg.TLS = mode
}

func (s *smtpClient) ForceTLSSkipVerify(skip bool) {
	s.cfg.SkipVerify = skip
}

func (s *smtpClient) ForceTLSServerName(serverName string) {
	s.cfg.ServerName = serverName
}

// GetDsn Return a correct SMTP DSN.
func (s smtpClient) GetDsn() string {
	var buf bytes.Buffer

	// [username[:password]@]
	if len(s.cfg.User) > 0 {
		_, _ = buf.WriteString(s.cfg.User)
		if len(s.cfg.Pass) > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(s.cfg.Pass)
		}
		_ = buf.WriteByte('@')
	}

	// [protocol[(address)]]
	_, _ = buf.WriteString(s.cfg.Net.string())

	// [username[:password]@]
	if len(s.cfg.Host) > 0 {
		_ = buf.WriteByte('(')
		_, _ = buf.WriteString(s.cfg.Host)
		if s.cfg.Port > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(fmt.Sprintf("%d", s.cfg.Port))
		}
		_ = buf.WriteByte(')')
	}

	// /tlsmode
	_ = buf.WriteByte('/')
	_, _ = buf.WriteString(s.cfg.TLS.string())

	// [?param1=value1&...&paramN=valueN]
	var val = &url.Values{}

	if s.cfg.ServerName != "" {
		val.Add("ServerName", s.cfg.ServerName)
	}

	if s.cfg.SkipVerify {
		val.Add("SkipVerify", "true")
	}

	params := val.Encode()

	// nolint: gomnd
	if len(params) > 2 {
		_, _ = buf.WriteString("?" + params)
	}

	return buf.String()
}
