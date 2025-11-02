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

// ShellCommandInfo returns metadata for all available shell commands.
// This is used by shell integration packages to discover available commands
// without creating the full command implementations.
//
// Returns:
//   - Slice of CommandInfo containing name and description for each command
//
// Available commands:
//   - list: Display all registered components
//   - start: Start components (optionally specify component keys as arguments)
//   - stop: Stop components (optionally specify component keys as arguments)
//   - restart: Restart components by stopping then starting them
//
// Example usage in shell integration:
//
//	for _, info := range config.ShellCommandInfo() {
//	    fmt.Printf("%s: %s\n", info.Name(), info.Description())
//	}
func ShellCommandInfo() []shlcmd.CommandInfo {
	var res = make([]shlcmd.CommandInfo, 0)

	res = append(res, shlcmd.Info("list", "list all components"))
	res = append(res, shlcmd.Info("start", "Starting components (leave args empty to start all components)"))
	res = append(res, shlcmd.Info("stop", "Stopping components (leave args empty to start all components)"))
	res = append(res, shlcmd.Info("restart", "Restarting (stop, start) components (leave args empty to restart all components)"))

	return res
}

// GetShellCommand returns executable shell commands for runtime component management.
// These commands can be integrated into CLI applications or interactive shells
// for manual component control during development or operations.
//
// Returns:
//   - Slice of Command objects that can be executed with Run(stdout, stderr, args)
//
// Available commands:
//
// 1. list:
//   - Lists all registered components in dependency order
//   - Arguments: none
//   - Output: One component key per line to stdout
//
// 2. start:
//   - Starts specified components or all if no arguments provided
//   - Arguments: component keys (optional)
//   - Output: Progress messages to stdout, errors to stderr
//   - Starts components in dependency order
//
// 3. stop:
//   - Stops specified components or all if no arguments provided
//   - Arguments: component keys (optional)
//   - Output: Progress messages to stdout
//   - Stops components in reverse dependency order
//
// 4. restart:
//   - Restarts components by stopping then starting them
//   - Arguments: component keys (optional)
//   - Output: Progress messages to stdout, errors to stderr
//   - Executes stop (reverse order) then start (dependency order)
//
// Example usage:
//
//	commands := cfg.GetShellCommand()
//	for _, cmd := range commands {
//	    if cmd.Name() == "list" {
//	        cmd.Run(os.Stdout, os.Stderr, nil)
//	    }
//	}
//
// Thread Safety:
// These commands access the component registry which is thread-safe.
// However, concurrent execution of start/stop/restart commands may lead
// to unpredictable results and should be avoided.
func (o *model) GetShellCommand() []shlcmd.Command {
	return []shlcmd.Command{
		o.shellCmdList(),
		o.shellCmdStart(),
		o.shellCmdStop(),
		o.shellCmdRestart(),
	}
}

func (o *model) shellCmdList() shlcmd.Command {
	return shlcmd.New("list", "list all components", func(buf io.Writer, err io.Writer, args []string) {
		for _, key := range o.ComponentDependencies() {
			if len(key) < 1 {
				continue
			} else if cpt := o.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, key) // nolint
			}
		}
	})
}

func (o *model) shellCmdStart() shlcmd.Command {
	return shlcmd.New("start", "Starting components (leave args empty to start all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = o.ComponentDependencies()
		}

		for _, key := range list {
			if len(key) < 1 {
				continue
			} else if cpt := o.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting component '%s'", key)) // nolint
				e := cpt.Start()
				o.componentUpdate(key, cpt)
				if e != nil {
					_, _ = fmt.Fprintln(err, e) // nolint
				} else if !cpt.IsStarted() {
					_, _ = fmt.Fprintln(err, fmt.Errorf("component is not started")) // nolint
				}
			}
		}
	})
}

func (o *model) shellCmdStop() shlcmd.Command {
	return shlcmd.New("stop", "Stopping components (leave args empty to start all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = o.ComponentDependencies()
		}

		for i := len(list) - 1; i >= 0; i-- {
			key := list[i]

			if len(key) < 1 {
				continue
			} else if cpt := o.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping component '%s'", key)) // nolint
				cpt.Stop()
			}
		}
	})
}

func (o *model) shellCmdRestart() shlcmd.Command {
	return shlcmd.New("restart", "Restarting (stop, start) components (leave args empty to restart all components)", func(buf io.Writer, err io.Writer, args []string) {
		var list []string
		if len(args) > 0 {
			list = args
		} else {
			list = o.ComponentDependencies()
		}

		for i := len(list) - 1; i >= 0; i-- {
			key := list[i]

			if len(key) < 1 {
				continue
			} else if cpt := o.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Stopping component '%s'", key)) // nolint
				cpt.Stop()
			}
		}

		_, _ = fmt.Fprintln(buf, "") // nolint

		for _, key := range list {
			if len(key) < 1 {
				continue
			} else if cpt := o.ComponentGet(key); cpt == nil {
				continue
			} else {
				_, _ = fmt.Fprintln(buf, fmt.Sprintf("Starting component '%s'", key)) // nolint
				e := cpt.Start()
				o.componentUpdate(key, cpt)
				if e != nil {
					_, _ = fmt.Fprintln(err, e) // nolint
				} else if !cpt.IsStarted() {
					_, _ = fmt.Fprintln(err, fmt.Errorf("component is not started")) // nolint
				}
			}
		}
	})
}
