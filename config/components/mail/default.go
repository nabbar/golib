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

package mail

import (
	"bytes"
	"encoding/json"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libmail "github.com/nabbar/golib/mail"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
  "charset": "",
  "subject": "",
  "encoding": "",
  "priority": "",
  "headers": {
    "": ""
  },
  "from": "",
  "sender": "",
  "replyTo": "",
  "returnPath": "",
  "to": [
    ""
  ],
  "cc": [
    ""
  ],
  "bcc": [
    ""
  ],
  "attach": [
    {
      "name": "",
      "mime": "",
      "path": ""
    }
  ],
  "inline": [
    {
      "name": "",
      "mime": "",
      "path": ""
    }
  ]
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

func (c *componentMail) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentMail) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	_ = Command.PersistentFlags().String(c.key+".charset", "", "define the charset to use into mail header")
	_ = Command.PersistentFlags().String(c.key+".encoding", "", "define the encoding mode for contents of mail")
	_ = Command.PersistentFlags().String(c.key+".priority", "", "define the priority of the mail")
	_ = Command.PersistentFlags().String(c.key+".subject", "", "define the subject of the mail")
	_ = Command.PersistentFlags().String(c.key+".from", "", "define the email use for sending the mail")
	_ = Command.PersistentFlags().String(c.key+".sender", "", "define the email show as sender")
	_ = Command.PersistentFlags().String(c.key+".replyTo", "", "define the email to use for reply")

	if err := Viper.BindPFlag(c.key+".charset", Command.PersistentFlags().Lookup(c.key+".charset")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".encoding", Command.PersistentFlags().Lookup(c.key+".encoding")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".priority", Command.PersistentFlags().Lookup(c.key+".priority")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".subject", Command.PersistentFlags().Lookup(c.key+".subject")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".from", Command.PersistentFlags().Lookup(c.key+".from")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".sender", Command.PersistentFlags().Lookup(c.key+".sender")); err != nil {
		return err
	}
	if err := Viper.BindPFlag(c.key+".replyTo", Command.PersistentFlags().Lookup(c.key+".replyTo")); err != nil {
		return err
	}

	return nil
}

func (c *componentMail) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libmail.Config, liberr.Error) {
	var (
		cfg = libmail.Config{}
		vpr = c.vpr()
		err liberr.Error
	)

	if len(cfg.Headers) < 1 {
		cfg.Headers = make(map[string]string, 0)
	}

	if len(cfg.To) < 1 {
		cfg.To = make([]string, 0)
	}

	if len(cfg.Cc) < 1 {
		cfg.Cc = make([]string, 0)
	}

	if len(cfg.Bcc) < 1 {
		cfg.Bcc = make([]string, 0)
	}

	if len(cfg.Attach) < 1 {
		cfg.Attach = make([]libmail.ConfigFile, 0)
	}

	if len(cfg.Inline) < 1 {
		cfg.Inline = make([]libmail.ConfigFile, 0)
	}

	if e := getCfg(c.key, &cfg); e != nil {
		return cfg, ErrorParamInvalid.Error(e)
	}

	if val := vpr.GetString(c.key + ".charset"); val != "" {
		cfg.Charset = val
	}
	if val := vpr.GetString(c.key + ".encoding"); val != "" {
		cfg.Encoding = val
	}
	if val := vpr.GetString(c.key + ".priority"); val != "" {
		cfg.Priority = val
	}
	if val := vpr.GetString(c.key + ".subject"); val != "" {
		cfg.Subject = val
	}
	if val := vpr.GetString(c.key + ".from"); val != "" {
		cfg.From = val
	}
	if val := vpr.GetString(c.key + ".sender"); val != "" {
		cfg.Sender = val
	}
	if val := vpr.GetString(c.key + ".replyTo"); val != "" {
		cfg.ReplyTo = val
	}

	if err = cfg.Validate(); err != nil {
		return cfg, ErrorConfigInvalid.Error(err)
	}

	return cfg, nil
}
