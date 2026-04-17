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

package idlemgr_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	durbig "github.com/nabbar/golib/duration/big"
	idlemgr "github.com/nabbar/golib/socket/idlemgr"
)

// BenchmarkNew measures the performance of creating a new manager.
func BenchmarkNew(b *testing.B) {
	ctx := context.Background()
	idle := durbig.Seconds(300)
	tick := durbig.Seconds(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = idlemgr.New(ctx, idle, tick)
	}
}

// BenchmarkRegister measures the performance of registering a single client.
func BenchmarkRegister(b *testing.B) {
	ctx := context.Background()
	idle := durbig.Seconds(300)
	tick := durbig.Seconds(1)
	manager, _ := idlemgr.New(ctx, idle, tick)
	_ = manager.Start(ctx)
	defer manager.Close()

	// Pre-create clients to avoid fmt.Sprintf and StopTimer/StartTimer in the loop
	clients := make([]*mockClient, b.N)
	for i := 0; i < b.N; i++ {
		clients[i] = &mockClient{id: fmt.Sprintf("client-%d", i)}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Register(clients[i])
	}
}

// BenchmarkRegisterUnregister measures the overhead of adding and removing clients.
func BenchmarkRegisterUnregister(b *testing.B) {
	ctx := context.Background()
	idle := durbig.Seconds(300)
	tick := durbig.Seconds(1)
	manager, _ := idlemgr.New(ctx, idle, tick)
	_ = manager.Start(ctx)
	defer manager.Close()

	// Pre-create clients
	clients := make([]*mockClient, b.N)
	for i := 0; i < b.N; i++ {
		clients[i] = &mockClient{id: fmt.Sprintf("client-%d", i)}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = manager.Register(clients[i])
		_ = manager.Unregister(clients[i])
	}
}

// BenchmarkNotifyActivity measures the performance of resetting client counters.
// This is the most frequent operation in a real-world scenario (Read/Write).
func BenchmarkNotifyActivity(b *testing.B) {
	ctx := context.Background()
	idle := durbig.Seconds(300)
	tick := durbig.Seconds(1)
	manager, _ := idlemgr.New(ctx, idle, tick)
	_ = manager.Start(ctx)
	defer manager.Close()

	numClients := 1000
	clients := make([]*mockClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &mockClient{id: fmt.Sprintf("client-%d", i)}
		_ = manager.Register(clients[i])
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var i uint32
		for pb.Next() {
			idx := atomic.AddUint32(&i, 1) % uint32(numClients)
			clients[idx].Reset()
		}
	})
}

// BenchmarkManagerLoop measures the CPU impact of the main ticker loop with many clients.
func BenchmarkManagerLoop(b *testing.B) {
	ctx := context.Background()
	idle := durbig.Seconds(300)
	tick := durbig.Seconds(1)
	manager, _ := idlemgr.New(ctx, idle, tick)
	_ = manager.Start(ctx)
	defer manager.Close()

	// Register a large number of clients to stress the loop
	numClients := 10000
	for i := 0; i < numClients; i++ {
		_ = manager.Register(&mockClient{id: fmt.Sprintf("client-%d", i)})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		time.Sleep(10 * time.Microsecond)
	}
}

// BenchmarkMassiveTimeout measures performance when many clients timeout simultaneously.
func BenchmarkMassiveTimeout(b *testing.B) {
	numClients := 5000
	ctx := context.Background()
	idle := durbig.Seconds(1)
	tick := durbig.Seconds(1)

	// Since this benchmark depends on time.Sleep, we limit b.N to avoid excessive duration.
	// Measuring one full cleanup cycle is enough to see the performance.
	if b.N > 1 {
		b.N = 1
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		manager, _ := idlemgr.New(ctx, idle, tick)

		for j := 0; j < numClients; j++ {
			c := &mockClient{id: fmt.Sprintf("timeout-%d", j)}
			_ = manager.Register(c)
		}

		b.StartTimer()
		_ = manager.Start(ctx)
		time.Sleep(1200 * time.Millisecond)
		b.StopTimer()
		_ = manager.Close()
	}
	b.ReportAllocs()
}
