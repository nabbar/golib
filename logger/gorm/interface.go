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

package gorm

import (
	"time"

	liblog "github.com/nabbar/golib/logger"
	gorlog "gorm.io/gorm/logger"
)

// New returns a gorlog.Interface that logs queries using the given logger and parameters.
//
//   - fct is a function that returns a golib Logger. This function is called whenever a new logger is needed.
//   - ignoreRecordNotFoundError determines whether gorm should ignore record not found errors.
//   - slowThreshold is a time.Duration that determines what constitutes a slow query.
//     If the query takes longer than this threshold, a warning will be logged.
//
// The returned gorlog.Interface is safe for concurrent use.
func New(fct func() liblog.Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration) gorlog.Interface {
	return &logGorm{
		i: ignoreRecordNotFoundError,
		s: slowThreshold,
		l: fct,
	}
}
