//go:build !386 && !arm && !mips && !mipsle
// +build !386,!arm,!mips,!mipsle

/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nutsdb

import (
	"context"
	"sync/atomic"
	"time"

	libclu "github.com/nabbar/golib/cluster"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	shlcmd "github.com/nabbar/golib/shell/command"
	libver "github.com/nabbar/golib/version"
)

const LogLib = "NutsDB"

type NutsDB interface {
	Listen() liberr.Error
	Restart() liberr.Error
	Shutdown() liberr.Error

	ForceRestart()
	ForceShutdown()

	IsRunning() bool
	IsReady(ctx context.Context) bool
	IsReadyTimeout(parent context.Context, dur time.Duration) bool
	WaitReady(ctx context.Context, tick time.Duration)

	GetLogger() liblog.Logger
	SetLogger(l liblog.FuncLog)

	Monitor(ctx libctx.FuncContext, vrs libver.Version) (montps.Monitor, error)

	Cluster() libclu.Cluster
	Client(ctx context.Context, tickSync time.Duration) Client
	ShellCommand(ctx func() context.Context, tickSync time.Duration) []shlcmd.Command
}

func New(c Config) NutsDB {
	return &ndb{
		c: c,
		t: new(atomic.Value),
		r: new(atomic.Value),
	}
}
