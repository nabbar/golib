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
	"fmt"

	logcfg "github.com/nabbar/golib/logger/config"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

func (o *componentLog) RegisterFlag(Command *spfcbr.Command) error {
	var (
		key string
		vpr *spfvpr.Viper
		err error
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	}

	_ = Command.PersistentFlags().Bool(key+".disableStandard", false, "allow disabling to write log to standard output stdout/stderr.")
	_ = Command.PersistentFlags().Bool(key+".disableStack", false, "allow to disable the goroutine id before each message")
	_ = Command.PersistentFlags().Bool(key+".disableTimestamp", false, "allow to disable the timestamp before each message")
	_ = Command.PersistentFlags().Bool(key+".enableTrace", true, "allow to add the origin caller/file/line of each message")
	_ = Command.PersistentFlags().String(key+".traceFilter", "", "define the path to clean for trace")
	_ = Command.PersistentFlags().Bool(key+".disableColor", false, "define if color could be use or not in messages format. If the running process is not a tty, no color will be used.")

	if err = vpr.BindPFlag(key+".disableStandard", Command.PersistentFlags().Lookup(key+".disableStandard")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disableStack", Command.PersistentFlags().Lookup(key+".disableStack")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disableTimestamp", Command.PersistentFlags().Lookup(key+".disableTimestamp")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".enableTrace", Command.PersistentFlags().Lookup(key+".enableTrace")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".traceFilter", Command.PersistentFlags().Lookup(key+".traceFilter")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disableColor", Command.PersistentFlags().Lookup(key+".disableColor")); err != nil {
		return err
	}

	return nil
}

func (o *componentLog) _getConfig() (*logcfg.Options, error) {
	var (
		key string
		cfg logcfg.Options
		vpr libvpr.Viper
		err error
	)

	if vpr = o._getViper(); vpr == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else if !vpr.Viper().IsSet(key) {
		return nil, ErrorParamInvalid.Error(fmt.Errorf("missing config key '%s'", key))
	} else if e := vpr.UnmarshalKey(key, &cfg); e != nil {
		return nil, ErrorParamInvalid.Error(e)
	}

	if cfg.Stdout == nil {
		cfg.Stdout = &logcfg.OptionsStd{}
	}

	if val := vpr.GetBool(key + "disableStandard"); val {
		cfg.Stdout.DisableStandard = true
	}

	if val := vpr.GetBool(key + "disableStack"); val {
		cfg.Stdout.DisableStack = true
	}

	if val := vpr.GetBool(key + "disableTimestamp"); val {
		cfg.Stdout.DisableTimestamp = true
	}

	if val := vpr.GetBool(key + "enableTrace"); val {
		cfg.Stdout.EnableTrace = true
	}

	if val := vpr.GetString(key + "traceFilter"); val != "" {
		cfg.TraceFilter = val
	}

	if val := vpr.GetBool(key + "disableColor"); val {
		cfg.Stdout.DisableColor = true
	}

	if err = cfg.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	}

	return &cfg, nil
}
