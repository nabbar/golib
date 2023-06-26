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

package hookfile

import (
	"os"

	"github.com/sirupsen/logrus"
)

func (o *hkf) getFormatter() logrus.Formatter {
	return o.o.format
}

func (o *hkf) getFlags() int {
	return o.o.flags
}

func (o *hkf) getLevel() []logrus.Level {
	return o.o.levels
}

func (o *hkf) getDisableStack() bool {
	return o.o.disableStack
}

func (o *hkf) getDisableTimestamp() bool {
	return o.o.disableTimestamp
}

func (o *hkf) getEnableTrace() bool {
	return o.o.enableTrace
}

func (o *hkf) getEnableAccessLog() bool {
	return o.o.enableAccessLog
}

func (o *hkf) getCreatePath() bool {
	return o.o.createPath
}

func (o *hkf) getFilepath() string {
	return o.o.filepath
}

func (o *hkf) getFileMode() os.FileMode {
	return o.o.fileMode
}

func (o *hkf) getPathMode() os.FileMode {
	return o.o.pathMode
}
