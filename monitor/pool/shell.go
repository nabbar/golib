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

package pool

import (
	"context"
	"fmt"
	"io"

	monsts "github.com/nabbar/golib/monitor/status"
	shlcmd "github.com/nabbar/golib/shell/command"
)

// ShellCommandInfo returns a list of shell command descriptions available for the pool.
// These commands can be used to interact with the pool through a shell interface.
func ShellCommandInfo() []shlcmd.CommandInfo {
	var res = make([]shlcmd.CommandInfo, 0)

	res = append(res, shlcmd.Info("list", "Print the monitors' List"))
	res = append(res, shlcmd.Info("info", "Print information about monitors (leave args empty to print info for all monitors)"))
	res = append(res, shlcmd.Info("start", "Starting monitor (leave args empty to start all monitors)"))
	res = append(res, shlcmd.Info("stop", "Stopping monitor (leave args empty to stop all monitors)"))
	res = append(res, shlcmd.Info("restart", "Restarting monitor (leave args empty to restart all monitors)"))
	res = append(res, shlcmd.Info("status", "Print status & message for monitor (leave args empty to print status of all monitors)"))

	return res
}

// GetShellCommand returns a list of executable shell commands for interacting with the pool.
// The commands include: list, info, status, start, stop, and restart.
func (p *pool) GetShellCommand(ctx context.Context) []shlcmd.Command {
	return []shlcmd.Command{
		p.shlCmdList(ctx),
		p.shlCmdInfo(ctx),
		p.shlCmdStatus(ctx),
		p.shlCmdStart(ctx),
		p.shlCmdStop(ctx),
		p.shlCmdRestart(ctx),
	}
}

func (p *pool) shlCmdList(_ context.Context) shlcmd.Command {
	return shlcmd.New("list", "Print the monitors' List", func(buf io.Writer, err io.Writer, args []string) {
		var list = p.MonitorList()

		for i := 0; i < len(list); i++ {
			_, _ = fmt.Fprintln(buf, list[i])
		}
	})
}

func (p *pool) shlCmdInfo(_ context.Context) shlcmd.Command {
	return shlcmd.New("info", "Print information about monitors (leave args empty to print info for all monitors)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = p.MonitorList()
		}

		for i := 0; i < len(list); i++ {
			if len(list[i]) < 1 {
				continue
			}

			m := p.MonitorGet(list[i])

			if m == nil {
				continue
			}

			inf := m.InfoGet()
			_, _ = fmt.Fprintln(buf, inf.Name()) // nolint
			for k, v := range inf.Info() {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("\t%s: %s", k, v)) // nolint
			}
			_, _ = fmt.Fprintln(buf, "") // nolint
		}
	})
}

func (p *pool) shlCmdStart(ctx context.Context) shlcmd.Command {
	return shlcmd.New("start", "Starting monitor (leave args empty to start all monitors)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = p.MonitorList()
		}

		for i := 0; i < len(list); i++ {
			if len(list[i]) < 1 {
				continue
			}

			m := p.MonitorGet(list[i])

			if m == nil {
				continue
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting monitor '%s'", list[i])) // nolint
			if e := m.Start(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i])) // nolint
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}
		}
	})
}

func (p *pool) shlCmdStop(ctx context.Context) shlcmd.Command {
	return shlcmd.New("stop", "Stopping monitor (leave args empty to stop all monitors)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = p.MonitorList()
		}

		for i := 0; i < len(list); i++ {
			if len(list[i]) < 1 {
				continue
			}

			m := p.MonitorGet(list[i])

			if m == nil {
				continue
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping monitor '%s'", list[i])) // nolint
			if e := m.Stop(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i])) // nolint
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}
		}
	})
}

func (p *pool) shlCmdRestart(ctx context.Context) shlcmd.Command {
	return shlcmd.New("restart", "Restarting monitor (leave args empty to restart all monitors)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = p.MonitorList()
		}

		for i := 0; i < len(list); i++ {
			if len(list[i]) < 1 {
				continue
			}

			m := p.MonitorGet(list[i])

			if m == nil {
				continue
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping monitor '%s'", list[i])) // nolint
			if e := m.Stop(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting monitor '%s'", list[i])) // nolint
			if e := m.Start(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i])) // nolint
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e) // nolint
			}
		}
	})
}

func (p *pool) shlCmdStatus(_ context.Context) shlcmd.Command {
	return shlcmd.New("status", "Print status & message for monitor (leave args empty to print status of all monitors)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = p.MonitorList()
		}

		for i := 0; i < len(list); i++ {
			if len(list[i]) < 1 {
				continue
			}

			m := p.MonitorGet(list[i])

			if m == nil {
				continue
			}

			s := m.Status()

			if s == monsts.OK {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("%s - %s", s.String(), list[i])) // nolint
			} else {
				_, _ = fmt.Fprintln(err, fmt.Sprintf("%s - %s: %s", s.String(), list[i], m.Message())) // nolint
			}
		}
	})
}
