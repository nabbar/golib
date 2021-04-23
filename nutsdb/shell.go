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
	"encoding/json"
	"fmt"
	"io"

	"github.com/xujiajun/nutsdb/ds/zset"

	libsh "github.com/nabbar/golib/shell"
	"github.com/xujiajun/nutsdb"
)

type shellCommand struct {
	n CmdCode
	c func() Client
}

func newShellCommand(code CmdCode, cli func() Client) libsh.Command {
	if code == CmdUnknown {
		return nil
	}

	return &shellCommand{
		c: cli,
		n: code,
	}
}

func (s *shellCommand) Name() string {
	return s.n.Name()
}

func (s *shellCommand) Describe() string {
	return s.n.Desc()
}

func (s *shellCommand) Run(buf io.Writer, err io.Writer, args []string) {
	if s.n == CmdUnknown {
		_, _ = fmt.Fprintf(err, "error: %v\n", ErrorClientCommandInvalid.Error(nil).GetError())
	}

	var cli Client

	if s.c == nil {
		_, _ = fmt.Fprintf(err, "error: %v\n", ErrorClientCommandCommit.Error(nil).GetError())
	} else if cli = s.c(); cli == nil {
		_, _ = fmt.Fprintf(err, "error: %v\n", ErrorClientCommandCommit.Error(nil).GetError())
	} else if r, e := cli.Run(s.n, args); e != nil {
		_, _ = fmt.Fprintf(err, "error: %v\n", e)
	} else if len(r.Value) < 1 {
		_, _ = fmt.Fprintf(buf, "No result.\n")
	} else {
		for i, val := range r.Value {
			if val == nil {
				continue
			}
			s.parse(buf, err, r.Value[i])
		}
	}
}

func (s *shellCommand) json(buf, err io.Writer, val interface{}) {
	if p, e := json.MarshalIndent(val, "", "  "); e != nil {
		_, _ = fmt.Fprintf(err, "error: %v\n", e)
	} else {
		_, _ = buf.Write(p)
	}
}

func (s *shellCommand) parse(buf, err io.Writer, val interface{}) {
	if values, ok := val.(*nutsdb.Entries); ok {
		for _, v := range *values {
			_, _ = fmt.Fprintf(buf, "Key: %s\n", string(v.Key))
			_, _ = fmt.Fprintf(buf, "Val: %s\n", string(v.Value))
			_, _ = fmt.Fprintf(buf, "\n")
		}
		return
	}

	if values, ok := val.(nutsdb.Entries); ok {
		for _, v := range values {
			_, _ = fmt.Fprintf(buf, "Key: %s\n", string(v.Key))
			_, _ = fmt.Fprintf(buf, "Val: %s\n", string(v.Value))
			_, _ = fmt.Fprintf(buf, "\n")
		}
		return
	}

	if values, ok := val.(*nutsdb.Entry); ok {
		_, _ = fmt.Fprintf(buf, "Key: %s\n", string(values.Key))
		_, _ = fmt.Fprintf(buf, "Val: %s\n", string(values.Value))
		_, _ = fmt.Fprintf(buf, "\n")
		return
	}

	if values, ok := val.(nutsdb.Entry); ok {
		_, _ = fmt.Fprintf(buf, "Key: %s\n", string(values.Key))
		_, _ = fmt.Fprintf(buf, "Val: %s\n", string(values.Value))
		_, _ = fmt.Fprintf(buf, "\n")
		return
	}

	if values, ok := val.(map[string]*zset.SortedSetNode); ok {
		for _, v := range values {
			_, _ = fmt.Fprintf(buf, "Key: %s\n", v.Key())
			_, _ = fmt.Fprintf(buf, "Val: %s\n", string(v.Value))
			_, _ = fmt.Fprintf(buf, "Score: %v\n", v.Score())
			_, _ = fmt.Fprintf(buf, "\n")
		}
		return
	}

	if values, ok := val.([]*zset.SortedSetNode); ok {
		for _, v := range values {
			_, _ = fmt.Fprintf(buf, "Key: %s\n", v.Key())
			_, _ = fmt.Fprintf(buf, "Val: %s\n", string(v.Value))
			_, _ = fmt.Fprintf(buf, "Score: %v\n", v.Score())
			_, _ = fmt.Fprintf(buf, "\n")
		}
		return
	}

	s.json(buf, err, val)
	return
}
