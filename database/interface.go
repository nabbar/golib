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
	"context"
	"sync"
	"sync/atomic"
	"time"

	libcfg "github.com/nabbar/golib/config"
	liblog "github.com/nabbar/golib/logger"
	gorlog "gorm.io/gorm/logger"

	liberr "github.com/nabbar/golib/errors"

	libsts "github.com/nabbar/golib/status"
	gormdb "gorm.io/gorm"
)

type Database interface {
	GetDB() *gormdb.DB
	SetDb(db *gormdb.DB)
	Close()

	WaitNotify(ctx context.Context, cancel context.CancelFunc)
	CheckConn() liberr.Error
	Config() *gormdb.Config

	RegisterContext(fct libcfg.FuncContext)
	RegisterLogger(fct func() liblog.Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration)
	RegisterGORMLogger(fct func() gorlog.Interface)

	StatusInfo() (name string, release string, hash string)
	StatusHealth() error
	StatusRouter(sts libsts.RouteStatus, prefix string) liberr.Error
}

func New(cfg *Config) (Database, liberr.Error) {
	if d, e := cfg.New(nil); e != nil {
		return nil, e
	} else {
		v := new(atomic.Value)
		v.Store(d)

		c := new(atomic.Value)
		c.Store(cfg)

		return &database{
			m: sync.Mutex{},
			v: v,
			c: c,
		}, nil
	}
}
