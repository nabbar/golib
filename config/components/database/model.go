/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package database

import (
	"sync"
	"time"

	libdbs "github.com/nabbar/golib/database/gorm"

	libctx "github.com/nabbar/golib/context"
	montps "github.com/nabbar/golib/monitor/types"
)

type componentDatabase struct {
	m  sync.RWMutex
	x  libctx.Config[uint8]
	li bool
	ls time.Duration
	d  libdbs.Database
	p  montps.FuncPool
}

func (o *componentDatabase) SetLogOptions(ignoreRecordNotFoundError bool, slowThreshold time.Duration) {
	o.m.Lock()
	defer o.m.Unlock()

	o.li = ignoreRecordNotFoundError
	o.ls = slowThreshold
}

func (o *componentDatabase) SetDatabase(db libdbs.Database) {
	o.m.Lock()
	defer o.m.Unlock()

	o.d = db
}

func (o *componentDatabase) GetDatabase() libdbs.Database {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.d
}
