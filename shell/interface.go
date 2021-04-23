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
	"io"

	"github.com/c-bata/go-prompt"

	liberr "github.com/nabbar/golib/errors"
)

type Command interface {
	Name() string
	Describe() string
	Run(buf io.Writer, err io.Writer, args []string)
}

type Shell interface {
	Run(buf io.Writer, err io.Writer, args []string)
	Add(prefix string, cmd ...Command)
	Get(cmd string) []Command
	Desc(cmd string) map[string]string
	Walk(fct func(name string, item Command) (Command, liberr.Error)) liberr.Error

	//go prompt
	RunPrompt(out, err io.Writer, opt ...prompt.Option)
}

func New() Shell {
	return &shell{
		c: make(map[string]Command),
	}
}
