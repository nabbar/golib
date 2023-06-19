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

package viper

import (
	"fmt"
	"io"
	"sync/atomic"

	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"

	liblog "github.com/nabbar/golib/logger"

	libhom "github.com/mitchellh/go-homedir"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	spfvpr "github.com/spf13/viper"
)

const (
	RemoteETCD = "etcd"
)

type viperRemote struct {
	model    interface{}
	provider string
	endpoint string
	path     string
	secure   string
	fct      func()
}

type viper struct {
	v *spfvpr.Viper
	i *atomic.Uint32
	l liblog.FuncLog
	h libctx.Config[uint8]

	base string
	prfx string
	deft func() io.Reader

	remote viperRemote
}

func (v *viper) SetRemoteProvider(provider string) {
	v.remote.provider = provider
}

func (v *viper) SetRemoteEndpoint(endpoint string) {
	v.remote.endpoint = endpoint
}

func (v *viper) SetRemotePath(path string) {
	v.remote.path = path
}

func (v *viper) SetRemoteSecureKey(key string) {
	v.remote.secure = key
}

func (v *viper) SetRemoteModel(model interface{}) {
	v.remote.model = model
}

func (v *viper) SetRemoteReloadFunc(fct func()) {
	v.remote.fct = fct
}

func (v *viper) SetHomeBaseName(base string) {
	v.base = base
}

func (v *viper) SetEnvVarsPrefix(prefix string) {
	v.prfx = prefix
}

func (v *viper) SetDefaultConfig(fct func() io.Reader) {
	v.deft = fct
}

func (v *viper) SetConfigFile(fileConfig string) liberr.Error {
	if fileConfig != "" {
		v.v.SetConfigFile(fileConfig)
	} else {
		// Find home directory.
		home, err := libhom.Dir()
		if err != nil {
			return ErrorHomePathNotFound.ErrorParent(err)
		}

		// Search config in home directory with name defined in SetConfigName (without extension).
		v.v.AddConfigPath(home)

		if v.base == "" {
			return ErrorBasePathNotFound.ErrorParent(fmt.Errorf("base name of config file is empty"))
		}

		v.v.SetConfigName("." + v.base)

		if v.prfx != "" {
			v.v.SetEnvPrefix(v.prfx)
			v.v.AutomaticEnv()
		}
	}

	return nil
}

func (v *viper) Viper() *spfvpr.Viper {
	return v.v
}

func (v *viper) logEntry(lvl loglvl.Level, msg string, args ...interface{}) logent.Entry {
	return v.l().Entry(lvl, msg, args...)
}
