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

package monitor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libmon "github.com/nabbar/golib/monitor"
	montps "github.com/nabbar/golib/monitor/types"
)

func BenchmarkMonitorCreation(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = libmon.New(ctx, nfo)
	}
}

func BenchmarkMonitorSetConfig(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	cfg := montps.Config{
		Name:          "bench-monitor",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Second),
		Logger:        lo.Clone(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mon.SetConfig(ctx, cfg)
	}
}

func BenchmarkMonitorGetConfig(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	cfg := montps.Config{
		Name:          "bench-monitor",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Second),
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(ctx, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mon.GetConfig()
	}
}

func BenchmarkMonitorStatusRead(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mon.Status()
	}
}

func BenchmarkMonitorLatencyRead(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mon.Latency()
	}
}

func BenchmarkMonitorUptimeRead(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mon.Uptime()
	}
}

func BenchmarkMonitorMarshalText(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mon.MarshalText()
	}
}

func BenchmarkMonitorMarshalJSON(b *testing.B) {
	ctx := context.Background()

	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mon.MarshalJSON()
	}
}

func BenchmarkMonitorHealthCheckExecution(b *testing.B) {
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(context.Background(), nfo)

	cfg := montps.Config{
		Name:          "bench-monitor",
		CheckTimeout:  libdur.ParseDuration(1 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Microsecond), // Minimal interval for max pressure
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	var count atomic.Uint64
	mon.SetHealthCheck(func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	ctx := context.Background()
	_ = mon.Start(ctx)
	defer mon.Stop(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Just wait for some progress to be made
		for count.Load() < uint64(i) {
			// tight loop or yield
		}
	}
}

func BenchmarkMonitorConcurrentStatusReads(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mon.Status()
		}
	})
}

func BenchmarkMonitorConcurrentMetricsReads(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mon.Latency()
			_ = mon.Uptime()
			_ = mon.Downtime()
		}
	})
}

func BenchmarkMonitorClone(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	cfg := montps.Config{
		Name:          "bench-monitor",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Second),
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(ctx, cfg)

	cloneCtx, cloneCnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cloneCnl()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cloned, _ := mon.Clone(cloneCtx)
		if cloned != nil && cloned.IsRunning() {
			_ = cloned.Stop(cloneCtx)
		}
	}
}

func BenchmarkMonitorInfoOperations(b *testing.B) {
	ctx := context.Background()
	nfo, err := newInfoWithNameAlone("bench-test", func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"env":     "production",
			"region":  "us-west",
		}, nil
	})
	if err != nil {
		b.Fatal(err)
	}

	mon, _ := libmon.New(ctx, nfo)

	b.Run("InfoName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mon.InfoName()
		}
	})

	b.Run("InfoMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mon.InfoMap()
		}
	})

	b.Run("InfoGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mon.InfoGet()
		}
	})

	b.Run("InfoUpd", func(b *testing.B) {
		updatedInfo, e := newInfoWithNameAlone("updated", func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"version": "2.0.0",
			}, nil
		})
		if e != nil {
			b.Fatal(e)
		}

		for i := 0; i < b.N; i++ {
			mon.InfoUpd(updatedInfo)
		}
	})
}

func BenchmarkMonitorStartStop(b *testing.B) {
	nfo, err := newInfoWithNameAlone("bench-test", nil)
	if err != nil {
		b.Fatal(err)
	}

	cfg := montps.Config{
		Name:          "bench-monitor",
		CheckTimeout:  libdur.ParseDuration(1 * time.Second),
		IntervalCheck: libdur.ParseDuration(100 * time.Millisecond),
		Logger:        lo.Clone(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
		mon, _ := libmon.New(context.Background(), nfo)
		_ = mon.SetConfig(context.Background(), cfg)
		mon.SetHealthCheck(func(ctx context.Context) error {
			return nil
		})

		_ = mon.Start(ctx)
		_ = mon.Stop(ctx)
		cnl()
	}
}
