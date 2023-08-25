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
	libmail "github.com/nabbar/golib/mail"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

func (o *componentMail) RegisterFlag(Command *spfcbr.Command) error {
	var (
		key string
		vpr *spfvpr.Viper
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	}

	_ = Command.PersistentFlags().String(key+".charset", "", "define the charset to use into mail header")
	_ = Command.PersistentFlags().String(key+".encoding", "", "define the encoding mode for contents of mail")
	_ = Command.PersistentFlags().String(key+".priority", "", "define the priority of the mail")
	_ = Command.PersistentFlags().String(key+".subject", "", "define the subject of the mail")
	_ = Command.PersistentFlags().String(key+".from", "", "define the email use for sending the mail")
	_ = Command.PersistentFlags().String(key+".sender", "", "define the email show as sender")
	_ = Command.PersistentFlags().String(key+".replyTo", "", "define the email to use for reply")

	if err := vpr.BindPFlag(key+".charset", Command.PersistentFlags().Lookup(key+".charset")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".encoding", Command.PersistentFlags().Lookup(key+".encoding")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".priority", Command.PersistentFlags().Lookup(key+".priority")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".subject", Command.PersistentFlags().Lookup(key+".subject")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".from", Command.PersistentFlags().Lookup(key+".from")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".sender", Command.PersistentFlags().Lookup(key+".sender")); err != nil {
		return err
	}
	if err := vpr.BindPFlag(key+".replyTo", Command.PersistentFlags().Lookup(key+".replyTo")); err != nil {
		return err
	}

	return nil
}

func (o *componentMail) _getConfig() (*libmail.Config, error) {
	var (
		key string
		cfg libmail.Config
		vpr *spfvpr.Viper
		err error
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if e := vpr.UnmarshalKey(key, &cfg); e != nil {
		return nil, ErrorParamInvalid.Error(e)
	}

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

	if val := vpr.GetString(key + ".charset"); val != "" {
		cfg.Charset = val
	}
	if val := vpr.GetString(key + ".encoding"); val != "" {
		cfg.Encoding = val
	}
	if val := vpr.GetString(key + ".priority"); val != "" {
		cfg.Priority = val
	}
	if val := vpr.GetString(key + ".subject"); val != "" {
		cfg.Subject = val
	}
	if val := vpr.GetString(key + ".from"); val != "" {
		cfg.From = val
	}
	if val := vpr.GetString(key + ".sender"); val != "" {
		cfg.Sender = val
	}
	if val := vpr.GetString(key + ".replyTo"); val != "" {
		cfg.ReplyTo = val
	}

	if err = cfg.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	}

	return &cfg, nil
}
