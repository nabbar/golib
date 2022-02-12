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
	"io"

	spfvpr "github.com/spf13/viper"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type Viper interface {
	SetRemoteProvider(provider string)
	SetRemoteEndpoint(endpoint string)
	SetRemotePath(path string)
	SetRemoteSecureKey(key string)
	SetRemoteModel(model interface{})
	SetRemoteReloadFunc(fct func())

	SetHomeBaseName(base string)
	SetEnvVarsPrefix(prefix string)
	SetDefaultConfig(fct func() io.Reader)
	SetConfigFile(fileConfig string) liberr.Error

	Config(logLevelRemoteKO, logLevelRemoteOK liblog.Level) liberr.Error
	Viper() *spfvpr.Viper
	WatchFS(logLevelFSInfo liblog.Level)
	Unset(key ...string) error
}

func New() Viper {
	return &viper{
		v: spfvpr.New(),
	}
}
