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
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	iotagg "github.com/nabbar/golib/ioutils/aggregator"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// sysAgg manages a shared, reference-counted connection to a syslog endpoint.
// It combines a network client with a buffered aggregator to provide an
// asynchronous, non-blocking io.Writer interface.
type sysAgg struct {
	i *atomic.Int64     // i is a reference counter for the number of hooks using this aggregator.
	w libsck.Client     // w is the underlying network client for the syslog endpoint.
	l bool              // l indicates if the connection is to a local (auto-discovered) syslog.
	a iotagg.Aggregator // a is the buffered aggregator that handles asynchronous writes.
}

var (
	// agg is a global, thread-safe map that stores shared sysAgg instances.
	// The key is a unique identifier for the syslog endpoint (protocol + address),
	// and the value is the corresponding sysAgg instance. This allows multiple
	// hooks pointing to the same destination to share a single network connection.
	agg = libatm.NewMapTyped[string, *sysAgg]()
)

// init sets up a finalizer for the global aggregator map.
// This ensures that all open network connections are closed gracefully
// when the program exits, preventing resource leaks.
func init() {
	runtime.SetFinalizer(agg, func(a libatm.MapTyped[string, *sysAgg]) {
		a.Range(func(k string, v *sysAgg) bool {
			if v != nil {
				_ = v.a.Close()
				_ = v.w.Close()
			}
			return true
		})
	})
}

// ResetOpenSyslog closes all active syslog connections and clears the aggregator map.
// This is primarily useful for testing or for scenarios requiring a full reset
// of the logging infrastructure.
func ResetOpenSyslog() {
	agg.Range(func(k string, v *sysAgg) bool {
		_ = v.a.Close()
		_ = v.w.Close()
		agg.Delete(k)
		return true
	})
}

// setKey generates a unique key for a syslog endpoint based on its protocol and address.
func setKey(ptc libptc.NetworkProtocol, adr string) string {
	if adr == "" {
		ptc = libptc.NetworkEmpty
		adr = "localhost"
	}

	return fmt.Sprintf("%s-%s", ptc.Code(), adr)
}

// setAgg retrieves or creates a shared aggregator for a given syslog endpoint.
// If an aggregator for the endpoint already exists, its reference count is incremented.
// Otherwise, a new aggregator and its underlying network connection are created.
func setAgg(ptc libptc.NetworkProtocol, adr string) (io.Writer, bool, error) {
	k := setKey(ptc, adr)
	i, l := agg.Load(k)

	if l && i != nil {
		i.i.Add(1)
		agg.Store(k, i)
		return i.a, i.l, nil
	}

	var e error
	i, e = newAgg(ptc, adr)

	if e != nil {
		return nil, false, e
	}

	agg.Store(k, i)
	return i.a, i.l, nil
}

// delAgg decrements the reference count for a syslog endpoint's aggregator.
// If the reference count drops to zero, the aggregator is shut down, its network
// connection is closed, and it is removed from the global map.
func delAgg(ptc libptc.NetworkProtocol, adr string) {
	k := setKey(ptc, adr)
	i, _ := agg.Load(k)
	if i == nil {
		return
	}

	if i.i.Add(-1) > 0 {
		agg.Store(k, i)
	} else {
		agg.Delete(k)
		_ = i.a.Close()
		_ = i.w.Close()
	}
}

// newAgg creates a new sysAgg instance, including the network client and the
// buffered writer. It establishes the initial connection and starts the
// aggregator's background processing goroutine.
func newAgg(ptc libptc.NetworkProtocol, adr string) (*sysAgg, error) {
	i := &sysAgg{
		i: new(atomic.Int64),
		w: nil,
		l: false,
		a: nil,
	}

	if adr == "" {
		var err error
		ptc, adr, err = systemSyslog()
		if err != nil {
			return nil, err
		}
		i.l = true
	}

	c, e := sckclt.New(sckcfg.Client{
		Network: ptc,
		Address: adr,
		TLS:     sckcfg.TLSClient{},
	}, nil)

	if e != nil {
		return nil, e
	}

	if e = c.Connect(context.Background()); e != nil {
		_ = c.Close()
		return nil, e
	}

	// The writer function for the aggregator handles automatic reconnection.
	// If a write fails, it attempts to reconnect before retrying the write.
	a, e := iotagg.New(context.Background(), iotagg.Config{
		AsyncTimer: 0,
		AsyncMax:   0,
		AsyncFct:   nil,
		SyncTimer:  time.Second,
		SyncFct:    nil,
		BufWriter:  250, // Buffer up to 250 log entries in memory.
		FctWriter: func(p []byte) (n int, err error) {
			n, err = c.Write(p)

			if err == nil {
				return n, nil
			} else if err = c.Connect(context.Background()); err != nil {
				return 0, err
			} else {
				return c.Write(p)
			}
		},
	})

	if e != nil {
		_ = c.Close()
		return nil, e
	}

	// Route internal aggregator errors to os.Stderr.
	a.SetLoggerError(func(msg string, err ...error) {
		for _, er := range err {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", msg, er)
		}
	})

	if e = a.Start(context.Background()); e != nil {
		_ = a.Close()
		_ = c.Close()
		return nil, e
	}

	i.w = c
	i.a = a
	i.i.Store(1)

	return i, nil
}
