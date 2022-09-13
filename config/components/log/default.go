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

package log

import (
	"bytes"
	"encoding/json"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`
{
   "disableStandard":false,
   "disableStack":false,
   "disableTimestamp":false,
   "enableTrace":true,
   "traceFilter":"",
   "disableColor":false,
   "logFile":[
      {
         "logLevel":[
            "Debug",
            "Info",
            "Warning",
            "Error",
            "Fatal",
            "Critical"
         ],
         "filepath":"",
         "create":false,
         "createPath":false,
         "fileMode":"0644",
         "pathMode":"0755",
         "disableStack":false,
         "disableTimestamp":false,
         "enableTrace":true
      }
   ],
   "logSyslog":[
      {
         "logLevel":[
            "Debug",
            "Info",
            "Warning",
            "Error",
            "Fatal",
            "Critical"
         ],
         "network":"tcp",
         "host":"",
         "severity":"Error",
         "facility":"local0",
         "tag":"",
         "disableStack":false,
         "disableTimestamp":false,
         "enableTrace":true
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

func (c *componentLog) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentLog) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	_ = Command.PersistentFlags().Bool(c.key+".disableStandard", false, "allow disabling to write log to standard output stdout/stderr.")
	_ = Command.PersistentFlags().Bool(c.key+".disableStack", false, "allow to disable the goroutine id before each message")
	_ = Command.PersistentFlags().Bool(c.key+".disableTimestamp", false, "allow to disable the timestamp before each message")
	_ = Command.PersistentFlags().Bool(c.key+".enableTrace", true, "allow to add the origin caller/file/line of each message")
	_ = Command.PersistentFlags().String(c.key+".traceFilter", "", "define the path to clean for trace")
	_ = Command.PersistentFlags().Bool(c.key+".disableColor", false, "define if color could be use or not in messages format. If the running process is not a tty, no color will be used.")

	if err := Viper.BindPFlag(c.key+".disableStandard", Command.PersistentFlags().Lookup(c.key+".disableStandard")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disableStack", Command.PersistentFlags().Lookup(c.key+".disableStack")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disableTimestamp", Command.PersistentFlags().Lookup(c.key+".disableTimestamp")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".enableTrace", Command.PersistentFlags().Lookup(c.key+".enableTrace")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".traceFilter", Command.PersistentFlags().Lookup(c.key+".traceFilter")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disableColor", Command.PersistentFlags().Lookup(c.key+".disableColor")); err != nil {
		return err
	}

	return nil
}

func (c *componentLog) _GetOptions(getCfg libcfg.FuncComponentConfigGet) (*liblog.Options, liberr.Error) {
	var (
		err liberr.Error
		cfg = liblog.Options{}
		vpr = c.vpr()
	)

	if err = getCfg(c.key, &cfg); err != nil {
		return nil, ErrorParamInvalid.Error(err)
	}

	if val := vpr.GetBool(c.key + "disableStandard"); val {
		cfg.DisableStandard = true
	}
	if val := vpr.GetBool(c.key + "disableStack"); val {
		cfg.DisableStack = true
	}
	if val := vpr.GetBool(c.key + "disableTimestamp"); val {
		cfg.DisableTimestamp = true
	}
	if val := vpr.GetBool(c.key + "enableTrace"); val {
		cfg.EnableTrace = true
	}
	if val := vpr.GetString(c.key + "traceFilter"); val != "" {
		cfg.TraceFilter = val
	}
	if val := vpr.GetBool(c.key + "disableColor"); val {
		cfg.DisableColor = true
	}

	if err = cfg.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	}

	return &cfg, nil
}
