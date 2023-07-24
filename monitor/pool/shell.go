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

func (p *pool) GetShellCommand(ctx context.Context) []shlcmd.Command {
	var res = make([]shlcmd.Command, 0)

	res = append(res, shlcmd.New("list", "Print the monitors' List", func(buf io.Writer, err io.Writer, args []string) {
		var list = p.MonitorList()

		for i := 0; i < len(list); i++ {
			_, _ = fmt.Fprintln(buf, list[i])
		}
	}))

	res = append(res, shlcmd.New("info", "Print information about monitors (leave args empty to print info for all monitors)", func(buf io.Writer, err io.Writer, args []string) {
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
			_, _ = fmt.Fprintln(buf, inf.Name())
			for k, v := range inf.Info() {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("\t%s: %s", k, v))
			}
			_, _ = fmt.Fprintln(buf, "")
		}
	}))

	res = append(res, shlcmd.New("start", "Starting monitor (leave args empty to start all monitors)", func(buf io.Writer, err io.Writer, args []string) {
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

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting monitor '%s'", list[i]))
			if e := m.Start(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i]))
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}
		}
	}))

	res = append(res, shlcmd.New("stop", "Stopping monitor (leave args empty to stop all monitors)", func(buf io.Writer, err io.Writer, args []string) {
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

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping monitor '%s'", list[i]))
			if e := m.Stop(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i]))
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}
		}
	}))

	res = append(res, shlcmd.New("restart", "Restarting monitor (leave args empty to restart all monitors)", func(buf io.Writer, err io.Writer, args []string) {
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

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping monitor '%s'", list[i]))
			if e := m.Stop(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting monitor '%s'", list[i]))
			if e := m.Start(ctx); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}

			_, _ = fmt.Fprintln(buf, fmt.Sprintf("Updating monitor '%s' on pool", list[i]))
			if e := p.MonitorSet(m); e != nil {
				_, _ = fmt.Fprintln(err, e)
			}
		}
	}))

	res = append(res, shlcmd.New("status", "Print status & message for monitor (leave args empty to print status of all monitors)", func(buf io.Writer, err io.Writer, args []string) {
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
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("%s - %s", s.String(), list[i]))
			} else {
				_, _ = fmt.Fprintln(err, fmt.Sprintf("%s - %s: %s", s.String(), list[i], m.Message()))
			}
		}
	}))

	return res
}
