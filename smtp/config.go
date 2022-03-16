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
	"fmt"
	"net/url"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	libsts "github.com/nabbar/golib/status"
)

type ConfigModel struct {
	DSN    string              `json:"dsn" yaml:"dsn" toml:"dsn" mapstructure:"dsn"`
	TLS    libtls.Config       `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty" mapstructure:"tls,omitempty"`
	Status libsts.ConfigStatus `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty" mapstructure:"status,omitempty"`

	_tls func() libtls.TLSConfig
}

func (c ConfigModel) Validate() liberr.Error {
	err := ErrorConfigValidator.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if c.DSN != "" {
		if _, er := NewConfig(c); er != nil {
			err.AddParent(er)
		}
	} else {
		err.AddParent(ErrorConfigInvalidDSN.Error(nil))
	}

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (c ConfigModel) RegisterDefaultTLS(fct func() libtls.TLSConfig) {
	c._tls = fct
}

func (c ConfigModel) GetSMTP() (SMTP, liberr.Error) {
	if c._tls == nil {
		return c.SMTP(nil)
	}

	return c.SMTP(c._tls())
}

func (c ConfigModel) SMTP(tlsDefault libtls.TLSConfig) (SMTP, liberr.Error) {
	var (
		err liberr.Error
		cfg Config
		tls libtls.TLSConfig
	)

	cfg, err = NewConfig(c)
	if err != nil {
		return nil, err
	}

	tls, err = c.TLS.NewFrom(tlsDefault)
	if err != nil {
		return nil, err
	}

	return NewSMTP(cfg, tls.TlsConfig(""))
}

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

	TLSCfg libtls.Config
	Status libsts.ConfigStatus
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

func (c *smtpConfig) SetNet(mode NETMode) {
	c.Net = mode
}

func (c *smtpConfig) GetNet() NETMode {
	return c.Net
}

func (c *smtpConfig) SetTlsMode(mode TLSMode) {
	c.TLS = mode
}

func (c *smtpConfig) GetTlsMode() TLSMode {
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
	_, _ = buf.WriteString(c.Net.string())

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
	_, _ = buf.WriteString(c.TLS.string())

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

func (c *smtpConfig) SetStatusConfig(sts libsts.ConfigStatus) {
	c.Status = sts
}

func (c *smtpConfig) GetStatusConfig() libsts.ConfigStatus {
	return c.Status
}
