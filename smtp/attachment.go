/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package smtp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/nabbar/golib/errors"
)

type attachment struct {
	name string
	buff *bytes.Buffer
}

func (a *attachment) Write(p []byte) (n int, e error) {
	return a.buff.Write(p)
}

func (a *attachment) WriteString(s string) (n int, e error) {
	return a.buff.WriteString(s)
}

func (a *attachment) WriteRune(r rune) (n int, e error) {
	return a.buff.WriteRune(r)
}

func (a attachment) GetContentType() string {
	return http.DetectContentType(a.buff.Bytes())
}

func (a attachment) GetName() string {
	return a.name
}

func (a attachment) GetBuffer() *bytes.Buffer {
	return a.buff
}

func (a attachment) Clone() Attachment {
	return &attachment{
		name: a.name,
		buff: bytes.NewBuffer(a.buff.Bytes()),
	}
}

type Attachment interface {
	Clone() Attachment

	GetContentType() string
	GetName() string
	GetBuffer() *bytes.Buffer

	Write(p []byte) (n int, e error)
	WriteString(s string) (n int, e error)
	WriteRune(r rune) (n int, e error)
}

func NewAttachment(name string) Attachment {
	return &attachment{
		name: name,
		buff: bytes.NewBuffer([]byte{}),
	}
}

func NewAttachmentFile(name string, filepath string) (Attachment, Error) {
	var b = bytes.NewBuffer([]byte{})

	if _, e := os.Stat(filepath); e != nil {
		return nil, ErrorFileStat.ErrorParent(e)
	}

	// #nosec
	if bb, e := ioutil.ReadFile(filepath); e != nil {
		return nil, ErrorFileRead.ErrorParent(e)
	} else {
		b.Write(bb)
	}

	return &attachment{
		name: name,
		buff: b,
	}, nil
}
