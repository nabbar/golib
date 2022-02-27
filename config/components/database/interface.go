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

	libcfg "github.com/nabbar/golib/config"
	libdbs "github.com/nabbar/golib/database"
)

const (
	ComponentType = "database"
)

type ComponentDatabase interface {
	libcfg.Component

	SetLOGKey(logKey string)
	SetLogOptions(ignoreRecordNotFoundError bool, slowThreshold time.Duration)
	GetDatabase() libdbs.Database
	SetDatabase(db libdbs.Database)
}

func New(logKey string) ComponentDatabase {
	return &componentDatabase{
		m: sync.Mutex{},
		l: logKey,
		d: nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentDatabase) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key, logKey string) {
	cfg.ComponentSet(key, New(logKey))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentDatabase {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentDatabase); !ok {
		return nil
	} else {
		return h
	}
}
