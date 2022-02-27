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
	"bytes"
	"encoding/json"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libsmtp "github.com/nabbar/golib/smtp"
	libsts "github.com/nabbar/golib/status"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
  "dsn": "",
  "tls": ` + string(cpttls.DefaultConfig(libcfg.JSONIndent)) + `,
  "status": ` + string(libsts.DefaultConfig(libcfg.JSONIndent)) + `
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, libcfg.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (c *componentSmtp) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentSmtp) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	_ = Command.PersistentFlags().String(c.key+".dsn", "", "A DSN like string to describe the smtp connection. Format allowed is [user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN] ")

	if err := Viper.BindPFlag(c.key+".dsn", Command.PersistentFlags().Lookup(c.key+".dsn")); err != nil {
		return err
	}

	return nil
}

func (c *componentSmtp) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libsmtp.ConfigModel, liberr.Error) {
	var (
		cfg = libsmtp.ConfigModel{}
		vpr = c.vpr()
		err liberr.Error
	)

	if e := getCfg(c.key, &cfg); e != nil {
		return cfg, ErrorParamsInvalid.Error(e)
	}

	if val := vpr.GetString(c.key + "dsn"); val != "" {
		cfg.DSN = val
	}

	if err = cfg.Validate(); err != nil {
		return cfg, ErrorConfigInvalid.Error(err)
	}

	cfg.RegisterDefaultTLS(func() libtls.TLSConfig {
		var (
			t libtls.TLSConfig
			e liberr.Error
		)

		if t, e = c._GetTLS(); e != nil {
			return t
		} else {
			return nil
		}
	})

	return cfg, nil
}
