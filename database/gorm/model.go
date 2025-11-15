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

package gorm

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	gormdb "gorm.io/gorm"
	gorlog "gorm.io/gorm/logger"
)

type database struct {
	m sync.Mutex
	v *atomic.Value
	c *atomic.Value
}

func (d *database) getConfig() *Config {
	d.m.Lock()
	defer d.m.Unlock()

	if d.v == nil {
		return nil
	} else if i := d.c.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Config); !ok {
		return nil
	} else {
		return o
	}
}

func (d *database) setConfig(cfg *Config) {
	d.m.Lock()
	defer d.m.Unlock()

	if d.v == nil {
		d.v = new(atomic.Value)
	}

	d.c.Store(cfg)
}

func (d *database) GetDB() *gormdb.DB {
	d.m.Lock()
	defer d.m.Unlock()

	if d.v == nil {
		return nil
	} else if i := d.v.Load(); i == nil {
		return nil
	} else if o, ok := i.(*gormdb.DB); !ok {
		return nil
	} else {
		return o
	}
}

func (d *database) SetDb(db *gormdb.DB) {
	d.m.Lock()
	defer d.m.Unlock()

	if d.v == nil {
		d.v = new(atomic.Value)
	}

	d.v.Store(db)
}

func (d *database) Close() {
	if o := d.GetDB(); o != nil {
		if i, e := o.DB(); e != nil {
			return
		} else {
			_ = i.Close()
		}
	}
}

func (d *database) WaitNotify(ctx context.Context, cancel context.CancelFunc) {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	select {
	case <-quit:
		d.Close()
		if cancel != nil {
			cancel()
		}
	case <-ctx.Done():
		d.Close()
		if cancel != nil {
			cancel()
		}
	}
}

func (d *database) CheckConn() liberr.Error {
	var o *gormdb.DB

	if o = d.GetDB(); o == nil {
		return ErrorDatabaseNotInitialized.Error(nil)
	}

	if v, e := o.DB(); e != nil {
		return ErrorDatabaseCannotSQLDB.Error(e)
	} else if e = v.Ping(); e != nil {
		return ErrorDatabasePing.Error(e)
	}

	return nil
}

func (d *database) Config() *gormdb.Config {
	cfg := d.getConfig()
	if cfg == nil {
		return nil
	}

	return cfg.Config()
}

func (d *database) RegisterContext(fct context.Context) {
	cfg := d.getConfig()
	if cfg == nil {
		return
	}

	cfg.RegisterContext(fct)
	d.setConfig(cfg)
}

func (d *database) RegisterLogger(fct func() liblog.Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration) {
	cfg := d.getConfig()
	if cfg == nil {
		return
	}

	cfg.RegisterLogger(fct, ignoreRecordNotFoundError, slowThreshold)
	d.setConfig(cfg)
}

func (d *database) RegisterGORMLogger(fct func() gorlog.Interface) {
	cfg := d.getConfig()
	if cfg == nil {
		return
	}

	cfg.RegisterGORMLogger(fct)
	d.setConfig(cfg)
}
