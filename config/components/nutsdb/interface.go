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

package nutsdb

import (
	"sync"
	"time"

	libsts "github.com/nabbar/golib/status"

	cptlog "github.com/nabbar/golib/config/components/log"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libndb "github.com/nabbar/golib/nutsdb"
)

const (
	ComponentType = "nutsdb"
)

type ComponentNutsDB interface {
	libcfg.Component

	SetLogger(key string)
	GetServer() (libndb.NutsDB, liberr.Error)
	GetClient(tickSync time.Duration) (libndb.Client, liberr.Error)
	SetStatusRouter(sts libsts.RouteStatus, prefix string)
}

func New(logKey string) ComponentNutsDB {
	if logKey == "" {
		logKey = cptlog.ComponentType
	}

	return &componentNutsDB{
		ctx: nil,
		get: nil,
		fsa: nil,
		fsb: nil,
		fra: nil,
		frb: nil,
		m:   sync.Mutex{},
		l:   logKey,
		n:   nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentNutsDB) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key, logKey string) {
	cfg.ComponentSet(key, New(logKey))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentNutsDB {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentNutsDB); !ok {
		return nil
	} else {
		return h
	}
}
