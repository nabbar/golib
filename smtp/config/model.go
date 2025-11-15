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

package config

import (
	"bytes"
	"fmt"
	"net/url"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	smtptp "github.com/nabbar/golib/smtp/tlsmode"
)

type smtpConfig struct {
	DSN        string
	Host       string
	Port       int
	User       string
	Pass       string
	Net        libptc.NetworkProtocol
	TLS        smtptp.TLSMode
	SkipVerify bool
	ServerName string

	TLSCfg libtls.Config
}

func (c *smtpConfig) SetHost(host string) {
	c.Host = host
}

func (c *smtpConfig) GetHost() string {
	return c.Host
}

func (c *smtpConfig) SetPort(port int) {
	c.Port = port
}

func (c *smtpConfig) GetPort() int {
	return c.Port
}

func (c *smtpConfig) SetUser(user string) {
	c.User = user
}

func (c *smtpConfig) GetUser() string {
	return c.User
}

func (c *smtpConfig) SetPass(pass string) {
	c.Pass = pass
}

func (c *smtpConfig) GetPass() string {
	return c.Pass
}

func (c *smtpConfig) SetNet(mode libptc.NetworkProtocol) {
	c.Net = mode
}

func (c *smtpConfig) GetNet() libptc.NetworkProtocol {
	return c.Net
}

func (c *smtpConfig) SetTlsMode(mode smtptp.TLSMode) {
	c.TLS = mode
}

func (c *smtpConfig) GetTlsMode() smtptp.TLSMode {
	return c.TLS
}

func (c *smtpConfig) SetTls(tls libtls.Config) {
	c.TLSCfg = tls
}

func (c *smtpConfig) GetTls() libtls.Config {
	return c.TLSCfg
}

func (c *smtpConfig) ForceTLSSkipVerify(skip bool) {
	c.SkipVerify = skip
}

func (c *smtpConfig) IsTLSSkipVerify() bool {
	return c.SkipVerify
}

func (c *smtpConfig) SetTLSServerName(serverName string) {
	c.ServerName = serverName
}

func (c *smtpConfig) GetTlSServerName() string {
	return c.ServerName
}

// GetDsn Return a correct SMTP DSN.
func (c *smtpConfig) GetDsn() string {
	var buf bytes.Buffer

	// [username[:password]@]
	if len(c.User) > 0 {
		_, _ = buf.WriteString(c.User)
		if len(c.Pass) > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(c.Pass)
		}
		_ = buf.WriteByte('@')
	}

	// [protocol[(address)]]
	_, _ = buf.WriteString(c.Net.String())

	// [username[:password]@]
	if len(c.Host) > 0 {
		_ = buf.WriteByte('(')
		_, _ = buf.WriteString(c.Host)
		if c.Port > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(fmt.Sprintf("%d", c.Port))
		}
		_ = buf.WriteByte(')')
	}

	// /tlsmode
	_ = buf.WriteByte('/')
	_, _ = buf.WriteString(c.TLS.String())

	// [?param1=value1&...&paramN=valueN]
	var val = &url.Values{}

	if c.ServerName != "" {
		val.Add("ServerName", c.ServerName)
	}

	if c.SkipVerify {
		val.Add("SkipVerify", "true")
	}

	params := val.Encode()

	// nolint: gomnd
	if len(params) > 2 {
		_, _ = buf.WriteString("?" + params)
	}

	return buf.String()
}
