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

package pool

import (
	"context"
	"time"

	libhtp "github.com/nabbar/golib/httpserver"
)

func (o *pool) Start(ctx context.Context) error {
	var err = ErrorPoolStart.Error(nil)

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if e := srv.Start(ctx); e != nil {
			err.Add(e)
		} else {
			o.Store(srv)
		}

		return true
	})

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (o *pool) Stop(ctx context.Context) error {
	var err = ErrorPoolStop.Error(nil)

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if e := srv.Stop(ctx); e != nil {
			err.Add(e)
		} else {
			o.Store(srv)
		}

		return true
	})

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (o *pool) Restart(ctx context.Context) error {
	var err = ErrorPoolRestart.Error(nil)

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if e := srv.Restart(ctx); e != nil {
			err.Add(e)
		} else {
			o.Store(srv)
		}

		return true
	})

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (o *pool) IsRunning() bool {
	var run = false

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if srv.IsRunning() {
			run = true
			return false
		}

		return true
	})

	return run
}

func (o *pool) Uptime() time.Duration {
	var res time.Duration

	o.Walk(func(name string, val libhtp.Server) bool {
		if dur := val.Uptime(); res < dur {
			res = dur
		}

		return true
	})

	return res
}
