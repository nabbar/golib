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
	"fmt"

	libmon "github.com/nabbar/golib/monitor/types"
	smtpcf "github.com/nabbar/golib/smtp/config"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

func (o *componentSmtp) RegisterFlag(Command *spfcbr.Command) error {
	var (
		key string
		vpr *spfvpr.Viper
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	}

	_ = Command.PersistentFlags().String(key+".dsn", "", "A DSN like string to describe the smtp connection. Format allowed is [user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN] ")

	if err := vpr.BindPFlag(key+".dsn", Command.PersistentFlags().Lookup(key+".dsn")); err != nil {
		return err
	}

	return nil
}

func (o *componentSmtp) _getConfig() (smtpcf.Config, *libmon.Config, error) {
	var (
		key string
		cfg smtpcf.ConfigModel
		vpr libvpr.Viper
		err error
	)

	if vpr = o._getViper(); vpr == nil {
		return nil, nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, nil, ErrorComponentNotInitialized.Error(nil)
	} else if !vpr.Viper().IsSet(key) {
		return nil, nil, ErrorParamInvalid.Error(fmt.Errorf("missing config key '%s'", key))
	} else if e := vpr.UnmarshalKey(key, &cfg); e != nil {
		return nil, nil, ErrorParamInvalid.Error(e)
	}

	if val := vpr.GetString(key + "dsn"); val != "" {
		cfg.DSN = val
	}

	if err = cfg.Validate(); err != nil {
		return nil, nil, ErrorConfigInvalid.Error(err)
	}

	if c, e := cfg.Config(); e != nil {
		return nil, nil, e
	} else {
		return c, &cfg.Monitor, nil
	}
}
