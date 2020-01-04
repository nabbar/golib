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

package njs_logger

import (
	"fmt"
	"io"
)

// IOWriter is struct redirected all entry to the current logger
type IOWriter struct {
	lvl Level
	prf string
}

// GetIOWriter return a io.Writer instance to Write on logger with a specified log level
/*
	level specify the log level to use to redirect all entry to current logger
	msgPrefixPattern is a string pattern to prefix all entry
	msgPrefixArgs is a list of args to apply on the msgPrefixPattern pattern to prefix all entry
*/
func GetIOWriter(level Level, msgPrefixPattern string, msgPrefixArgs ...interface{}) io.Writer {
	return &IOWriter{
		lvl: level,
		prf: fmt.Sprintf(msgPrefixPattern, msgPrefixArgs...),
	}
}

// Write implement the Write function of the io.Writer interface and redirect all entry to current logger
//
// the return n will always return the len on the p parameter and err will always be nil
/*
	p the entry to be redirect to current logger
*/
func (iow IOWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	iow.lvl.Log(iow.prf + " " + string(p))
	return
}
