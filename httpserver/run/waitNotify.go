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
	"os"
	"os/signal"
	"syscall"
)

func (o *sRun) StartWaitNotify(ctx context.Context) {
	if !o.IsRunning() {
		return
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	o.initChan()
	select {
	case <-quit:
		_ = o.Stop(ctx)
		return
	case <-o.getContext().Done():
		if o.IsRunning() {
			_ = o.Stop(ctx)
		}
		return
	case <-o.getChan():
		return
	}
}

func (o *sRun) StopWaitNotify() {
	o.m.Lock()
	defer o.m.Unlock()
	o.chn <- struct{}{}
}

func (o *sRun) initChan() {
	o.m.Lock()
	defer o.m.Unlock()
	o.chn = make(chan struct{})
}

func (o *sRun) getChan() <-chan struct{} {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.chn
}
