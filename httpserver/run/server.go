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

package run

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
)

func (o *sRun) Start(ctx context.Context) error {
	o.m.Lock()
	defer o.m.Unlock()

	if o == nil || o.srv == nil {
		return fmt.Errorf("invalid instance")
	}

	o.ctx, o.cnl = context.WithCancel(ctx)
	go o.runServer()
	return nil
}

func (o *sRun) runServer() {
	defer func() {
		cnl := o.getCancel()
		if cnl != nil {
			cnl()
		}

		o.setRunning(false)
		o.logger().Entry(liblog.InfoLevel, "server stopped").Log()
	}()

	var err error
	o.logger().Entry(liblog.InfoLevel, "Server is starting").Log()
	o.setError(nil)
	o.setRunning(true)

	if o.isTLS() {
		err = o.getServer().ListenAndServeTLS("", "")
	} else {
		err = o.getServer().ListenAndServe()
	}

	if err != nil {
		x := o.getContext()
		if x != nil && x.Err() != nil && errors.Is(err, x.Err()) {
			err = nil
		} else if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}

	o.setError(err)
	o.logger().Entry(liblog.InfoLevel, "Server return an error").ErrorAdd(true, err).Check(liblog.NilLevel)
}

func (o *sRun) Stop(ctx context.Context) error {
	var cnl context.CancelFunc

	if ctx.Err() != nil {
		ctx = context.Background()
	}

	if t, ok := ctx.Deadline(); !ok || t.IsZero() {
		ctx, cnl = context.WithTimeout(ctx, srvtps.TimeoutWaitingStop)
	}

	defer func() {
		if cnl != nil {
			cnl()
		}
		//o.delServer()
	}()

	var err error
	if o.IsRunning() {
		o.StopWaitNotify()
		if srv := o.getServer(); srv != nil {
			err = srv.Shutdown(ctx)
		}
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	o.logger().Entry(liblog.ErrorLevel, "Shutting down server").ErrorAdd(true, err).Check(liblog.NilLevel)
	return err
}

func (o *sRun) Restart(ctx context.Context) error {
	_ = o.Stop(ctx)
	return o.Start(ctx)
}

func (o *sRun) IsRunning() bool {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.run
}

func (o *sRun) setRunning(flag bool) {
	o.m.Lock()
	defer o.m.Unlock()
	o.run = flag
}
