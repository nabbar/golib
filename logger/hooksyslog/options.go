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

	"github.com/sirupsen/logrus"
)

func (o *hks) getFormatter() logrus.Formatter {
	return o.o.format
}

func (o *hks) getLevel() []logrus.Level {
	return o.o.levels
}

func (o *hks) getDisableStack() bool {
	return o.o.disableStack
}

func (o *hks) getDisableTimestamp() bool {
	return o.o.disableTimestamp
}

func (o *hks) getEnableTrace() bool {
	return o.o.enableTrace
}

func (o *hks) getEnableAccessLog() bool {
	return o.o.enableAccessLog
}

func (o *hks) getSyslog() (Wrapper, error) {
	return newSyslog(o.o.network, o.o.endpoint, o.o.tag, o.o.fac)
}

func (o *hks) getSyslogInfo() string {
	return fmt.Sprintf("syslog to '%s %s' with tag '%s'", o.o.network.Code(), o.o.endpoint, o.o.tag)
}
