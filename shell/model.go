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

package shell

import (
	"fmt"
	"io"

	liberr "github.com/nabbar/golib/errors"
	shlcmd "github.com/nabbar/golib/shell/command"
)

type shell struct {
	c map[string]shlcmd.Command
}

func (s *shell) Run(buf io.Writer, err io.Writer, args []string) {
	if len(args) == 0 {
		return
	}

	if cmd, ok := s.c[args[0]]; ok {
		if cmd == nil {
			_, _ = fmt.Fprintf(err, "Command not runable...\n")
			panic(nil)
		}

		cmd.Run(buf, err, args[1:])
	} else {
		_, _ = fmt.Fprintf(err, "Invalid command\n")
	}
}

func (s *shell) Walk(fct func(name string, item shlcmd.Command) (shlcmd.Command, liberr.Error)) liberr.Error {
	if fct == nil {
		return nil
	}

	for k, c := range s.c {
		if s.c[k] == nil {
			continue
		}

		if r, e := fct(k, c); e != nil {
			return e
		} else if r == nil {
			continue
		} else {
			s.c[k] = r
		}
	}

	return nil
}

func (s *shell) Add(prefix string, cmd ...shlcmd.Command) {
	if len(s.c) == 0 {
		s.c = make(map[string]shlcmd.Command)
	}

	for i := 0; i < len(cmd); i++ {
		n := cmd[i].Name()
		c := cmd[i]

		if c == nil {
			continue
		}

		if len(prefix) > 0 {
			n = prefix + n
		}

		s.c[n] = c
	}
}

func (s *shell) Get(cmd string) []shlcmd.Command {
	var res = make([]shlcmd.Command, 0)

	_ = s.Walk(func(name string, item shlcmd.Command) (shlcmd.Command, liberr.Error) {
		if len(cmd) == 0 || name == cmd {
			res = append(res, item)
		}

		return nil, nil
	})

	return res
}

func (s *shell) Desc(cmd string) map[string]string {
	var res = make(map[string]string)

	_ = s.Walk(func(name string, item shlcmd.Command) (shlcmd.Command, liberr.Error) {
		if len(cmd) == 0 || name == cmd {
			res[name] = item.Describe()
		}

		return nil, nil
	})

	return res
}
