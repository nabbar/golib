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

package hooksyslog

import (
	"context"
	"fmt"
	"sync"
	"time"

	libsrv "github.com/nabbar/golib/runner"
)

// Run starts the background goroutine that processes buffered log entries
// and writes them to syslog. This method blocks until the context is cancelled
// or Close() is called.
//
// This method MUST be called in a separate goroutine after creating the hook:
//
//	hook, _ := New(opts, formatter)
//	go hook.Run(ctx)
//
// Behavior:
//   - Retries syslog connection every 1 second until successful
//   - Processes buffered entries from the channel (capacity: 250)
//   - Maps entries to appropriate syslog severity levels
//   - Waits for all pending writes before returning
//   - Sets running status to true after initial connection
//
// Shutdown:
//   - Context cancellation triggers graceful shutdown
//   - Done() channel is closed when Run returns
//   - All buffered entries are written before exit
//
// Error Handling:
//   - Connection errors are printed to stdout
//   - Write errors are printed to stdout (not propagated)
//   - Panics are recovered and logged
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	go hook.Run(ctx)
//	defer func() {
//		cancel()
//		hook.Close()
//		<-hook.Done()
//	}()
func (o *hks) Run(ctx context.Context) {
	var (
		s Wrapper
		w = sync.WaitGroup{}
		e error
	)

	defer func() {
		if r := recover(); r != nil {
			libsrv.RecoveryCaller("golib/logger/hooksyslog/system", r)
		}
		if s != nil {
			w.Wait()
			_ = s.Close()
		}
		o.r.Store(false)
	}()

	for {
		if s, e = o.getSyslog(); e != nil {
			fmt.Println(e.Error())
		} else {
			break
		}
		time.Sleep(time.Second)
	}

	o.prepareChan()
	//fmt.Printf("starting hook for log syslog '%s'\n", o.getSyslogInfo())
	go func() {
		time.Sleep(100 * time.Millisecond)
		o.r.Store(true)
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case <-o.Done():
			return

		case i := <-o.Data():
			w.Add(1)
			go o.writeWrapper(s, w.Done, i...)
		}
	}
}

// IsRunning returns true if the background writer goroutine is active
// and ready to process log entries.
//
// Returns:
//   - false: Before Run() connects to syslog (initial ~100ms)
//   - true: After successful syslog connection
//   - false: After context cancellation or Close()
//
// Note: There's a brief delay (~100ms) after Run() starts before
// IsRunning() returns true, allowing time for initial connection.
func (o *hks) IsRunning() bool {
	return o.r.Load()
}

// writeWrapper processes a batch of log entries and writes them to syslog.
// This is called by Run() for each batch received from the channel.
//
// Parameters:
//   - w: Syslog writer interface (platform-specific)
//   - done: Callback to signal completion (for WaitGroup)
//   - d: Slice of log entries with severity and data
//
// Behavior:
//   - Maps severity to appropriate syslog method (Panic, Fatal, Error, etc.)
//   - Writes each entry sequentially
//   - Prints errors to stdout (doesn't propagate)
//   - Recovers from panics
func (o *hks) writeWrapper(w Wrapper, done func(), d ...data) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			libsrv.RecoveryCaller("golib/logger/hooksyslog/system", r)
		}
		done()
	}()

	if w == nil {
		return
	} else if len(d) < 1 {
		return
	}

	for k := range d {
		if len(d[k].p) < 1 {
			continue
		}
		switch d[k].s {
		case SyslogSeverityAlert:
			_, err = w.Panic(d[k].p)
		case SyslogSeverityCrit:
			_, err = w.Fatal(d[k].p)
		case SyslogSeverityErr:
			_, err = w.Error(d[k].p)
		case SyslogSeverityWarning:
			_, err = w.Warning(d[k].p)
		case SyslogSeverityInfo:
			_, err = w.Info(d[k].p)
		case SyslogSeverityDebug:
			_, err = w.Debug(d[k].p)
		default:
			_, err = w.Write(d[k].p)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
