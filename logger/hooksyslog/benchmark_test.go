/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package hooksyslog_test

import (
	"runtime"
	"testing"

	logcfg "github.com/nabbar/golib/logger/config"
	logsys "github.com/nabbar/golib/logger/hooksyslog"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

func BenchmarkNewHook(b *testing.B) {
	opt := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUnix.Code(),
		LogLevel: []string{"info"},
	}
	defer logsys.ResetOpenSyslog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook, err := logsys.New(opt, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = hook.Close()
	}
}

func BenchmarkFire(b *testing.B) {
	opt := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUnix.Code(),
		LogLevel: []string{"info"},
	}

	hook, err := logsys.New(opt, nil)
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
		return
	}
	defer func() {
		_ = hook.Close()
		logsys.ResetOpenSyslog()
	}()

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Data:    logrus.Fields{"key": "value", "foo": "bar"},
		Level:   logrus.InfoLevel,
		Message: "test message",
	}

	b.Run("Single", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := hook.Fire(entry); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := hook.Fire(entry); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkFireJSON(b *testing.B) {
	opt := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUnix.Code(),
		LogLevel: []string{"info"},
	}

	hook, err := logsys.New(opt, &logrus.JSONFormatter{})
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
		return
	}
	defer func() {
		_ = hook.Close()
		logsys.ResetOpenSyslog()
	}()

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Data:    logrus.Fields{"key": "value", "foo": "bar"},
		Level:   logrus.InfoLevel,
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hook.Fire(entry); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGoroutineLeakCheck(b *testing.B) {
	opt := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUnix.Code(),
		LogLevel: []string{"info"},
	}

	initialGoroutines := runtime.NumGoroutine()

	for i := 0; i < b.N; i++ {
		hook, err := logsys.New(opt, nil)
		if err == nil {
			_ = hook.Fire(&logrus.Entry{Level: logrus.InfoLevel, Message: "test"})
			_ = hook.Close()
		}
		logsys.ResetOpenSyslog()
	}

	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines+10 { // Allow some slack for runtime goroutines
		b.Errorf("Possible goroutine leak: started with %d, ended with %d", initialGoroutines, finalGoroutines)
	}
}
