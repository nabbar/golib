/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package hashicorp

import (
	"github.com/hashicorp/go-hclog"
	liblog "github.com/nabbar/golib/logger"
)

// New returns a new hclog.Logger from the given liblog.FuncLog.
// It's a convenient way to create an hclog.Logger from a logger
// that's already been set up with the liblog package.
//
// The given logger is used as the underlying logger for the
// returned hclog.Logger. This means that any log messages sent
// to the returned hclog.Logger will be forwarded to the given
// logger.
//
// The returned hclog.Logger is a fully functional hclog.Logger and
// supports all of the standard hclog.Logger methods.
func New(logger liblog.FuncLog) hclog.Logger {
	return &_hclog{
		l: logger,
	}
}

// SetDefault sets the default hclog.Logger to the given liblog.FuncLog.
// It's a convenient way to set the default hclog.Logger from a logger
// that's already been set up with the liblog package.
//
// The given logger is used as the underlying logger for the
// default hclog.Logger. This means that any log messages sent
// to the default hclog.Logger will be forwarded to the given
// logger.
//
// The default hclog.Logger is used by the hclog package whenever
// an hclog.Logger is not explicitly provided. For example, when
// creating a new hclog.Logger with the hclog.New() function, the
// default hclog.Logger is used if no other logger is provided.
func SetDefault(log liblog.FuncLog) {
	hclog.SetDefault(New(log))
}
