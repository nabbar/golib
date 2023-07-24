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

package config

import (
	"fmt"
	"io"

	shlcmd "github.com/nabbar/golib/shell/command"
)

func ShellCommandInfo() []shlcmd.CommandInfo {
	var res = make([]shlcmd.CommandInfo, 0)

	res = append(res, shlcmd.Info("list", "list all components"))
	res = append(res, shlcmd.Info("start", "Starting components (leave args empty to start all components)"))
	res = append(res, shlcmd.Info("stop", "Stopping components (leave args empty to start all components)"))
	res = append(res, shlcmd.Info("restart", "Restarting (stop, start) components (leave args empty to restart all components)"))

	return res
}

func (c *configModel) GetShellCommand() []shlcmd.Command {
	var res = make([]shlcmd.Command, 0)

	res = append(res, shlcmd.New("list", "list all components", func(buf io.Writer, err io.Writer, args []string) {
		for _, key := range c.ComponentDependencies() {
			if len(key) < 1 {
				continue
			} else if cpt := c.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, key)
			}
		}
	}))

	res = append(res, shlcmd.New("start", "Starting components (leave args empty to start all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = c.ComponentDependencies()
		}

		for _, key := range list {
			if len(key) < 1 {
				continue
			} else if cpt := c.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting component '%s'", key))
				e := cpt.Start()
				c.componentUpdate(key, cpt)
				if e != nil {
					_, _ = fmt.Fprintln(err, e)
				} else if !cpt.IsStarted() {
					_, _ = fmt.Fprintln(err, fmt.Errorf("component is not started"))
				}
			}
		}
	}))

	res = append(res, shlcmd.New("stop", "Stopping components (leave args empty to start all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = c.ComponentDependencies()
		}

		for i := len(list) - 1; i >= 0; i-- {
			key := list[i]

			if len(key) < 1 {
				continue
			} else if cpt := c.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping component '%s'", key))
				cpt.Stop()
			}
		}
	}))

	res = append(res, shlcmd.New("restart", "Restarting (stop, start) components (leave args empty to restart all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = c.ComponentDependencies()
		}

		for i := len(list) - 1; i >= 0; i-- {
			key := list[i]

			if len(key) < 1 {
				continue
			} else if cpt := c.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping component '%s'", key))
				cpt.Stop()
			}
		}

		_, _ = fmt.Fprintln(buf, "")

		for _, key := range list {
			if len(key) < 1 {
				continue
			} else if cpt := c.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting component '%s'", key))
				e := cpt.Start()
				c.componentUpdate(key, cpt)
				if e != nil {
					_, _ = fmt.Fprintln(err, e)
				} else if !cpt.IsStarted() {
					_, _ = fmt.Fprintln(err, fmt.Errorf("component is not started"))
				}
			}
		}
	}))

	return res
}
