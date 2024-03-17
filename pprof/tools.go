/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package pprof

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	srvtck "github.com/nabbar/golib/server/runner/ticker"
)

var (
	c        *os.File
	m        string
	ctx, cnl = context.WithCancel(context.Background())
	s        = srvtck.New(5*time.Minute, ProfilingMemRun)
)

func StartProfiling() {
	ProfilingCPUStart()
	ProfilingMemStart()
}

func StopProfiling() {
	ProfilingMemDefer()
	ProfilingCPUDefer()
}

func getPath(basename string) (*os.File, error) {
	var (
		h *os.File
		p string
		e error
	)

	p, e = os.Executable()
	if e != nil {
		return nil, e
	}

	p = filepath.Join(filepath.Dir(p), basename)

	if _, e = os.Stat(p); e != nil && !errors.Is(e, os.ErrNotExist) {
		return nil, e
	} else if e != nil {
		h, e = os.Create(p)
	} else {
		h, e = os.Open(p)
	}

	if e != nil {
		return nil, e
	}

	if e = h.Truncate(0); e != nil {
		_ = h.Close()
		return nil, e
	}

	return h, nil
}

func ProfilingCPUStart() {
	var e error
	if c, e = getPath("cpu.prof"); e != nil {
		panic(e)
	} else if e = pprof.StartCPUProfile(c); e != nil {
		panic(e)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Starting pprof for CPU to file '%s'\n", c.Name())
}

func ProfilingCPUDefer() {
	_, _ = fmt.Fprintf(os.Stdout, "Stopping pprof for CPU to file '%s'\n", c.Name())
	pprof.StopCPUProfile()
	_ = c.Close()
}

func ProfilingMemStart() {
	if h, e := getPath("mem.prof"); e != nil {
		panic(e)
	} else {
		m = h.Name()
	}

	if e := s.Start(ctx); e != nil {
		panic(e)
	}
}

func ProfilingMemRun(ctx context.Context, tck *time.Ticker) error {
	if ctx.Err() != nil {
		return nil
	} else if len(m) < 1 {
		return nil
	} else if h, e := os.OpenFile(m, os.O_RDWR|os.O_EXCL|os.O_SYNC, 0644); e != nil {
		return e
	} else {
		defer func() {
			_ = h.Close()
		}()

		runtime.GC()

		if e = pprof.WriteHeapProfile(h); e != nil {
			return e
		}

		return nil
	}
}

func ProfilingMemDefer() {
	if cnl != nil {
		cnl()
	}

	x, l := context.WithTimeout(context.Background(), 15*time.Second)
	defer l()

	_ = s.Stop(x)
}
