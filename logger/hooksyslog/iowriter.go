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

package hooksyslog

import (
	"fmt"
	"time"
)

func (o *hks) Write(p []byte) (n int, err error) {
	c := o.d.Load()

	if c != nil {
		if c.(chan data) != closeByte {
			c.(chan data) <- newData(0, p)
			return len(p), nil
		}
	}

	return 0, errStreamClosed

}

func (o *hks) WriteSev(s SyslogSeverity, p []byte) (n int, err error) {
	c := o.d.Load()

	if c != nil {
		if c.(chan data) != closeByte {
			c.(chan data) <- newData(s, p)
			return len(p), nil
		}
	}

	return 0, fmt.Errorf("%v, path: %s", errStreamClosed, o.getSyslogInfo())

}

func (o *hks) Close() error {
	//fmt.Printf("closing hook for log syslog '%s'\n", o.getSyslogInfo())

	o.d.Store(closeByte)
	time.Sleep(10 * time.Millisecond)

	o.s.Store(closeStruct)
	return nil
}
