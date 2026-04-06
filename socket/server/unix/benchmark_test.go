//go:build linux || darwin

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

// Package unix_test provides high-performance benchmarks to measure the Unix domain socket server's
// throughput, latency, and resource utilization under extreme load.
//
// # benchmark_test.go: Performance and Scalability Metrics
//
// These benchmarks are crucial for validating the recent performance optimizations, such as the
// introduction of `sync.Pool` for connection contexts and the centralized Idle Manager.
//
// # Optimization Highlights in Benchmarks:
//
// ## 1. Handling "Resource Temporarily Unavailable" (EAGAIN)
// Under extreme load (e.g., 25 concurrent connections with large payloads), the OS's Unix socket
// backlog and buffers can saturate. The benchmark helper `connectToServerBench` and the I/O
// operations implement a retry loop (up to 10 attempts with 1ms delay) when `syscall.EAGAIN`
// is encountered. This ensures the benchmark measures the server's peak performance rather
// than system congestion limits.
//
// ## 2. Zero-Allocation Strategy
// The benchmarks utilize a `bufPool` (defined in helper_test.go) to avoid allocations during
// the Read/Write loop. This allows the benchmarks to isolate the server's internal overhead.
//
// # Benchmarks Included:
//
// ## Server Lifecycle
//   - BenchmarkServerStartup: Measures the time to initialize the server, create the socket, and start the listener.
//   - BenchmarkServerShutdown: Measures the time to stop the listener and clean up the socket file.
//
// ## Connection & Latency
//   - BenchmarkConnectionEstablishment: Measures the overhead of a full net.Dial + Accept cycle.
//   - BenchmarkEchoLatency: Measures the round-trip time for a small message, reflecting the minimal per-request overhead.
//
// ## Throughput (Bandwidth)
//   - BenchmarkThroughput[1KB-32KB]: Measures data transfer rate for various payload sizes on a single connection.
//   - BenchmarkThroughput8KB_C[5-25]: Measures aggregate throughput across multiple concurrent connections to test scaling.
//
// ## Parallelism
//   - BenchmarkConcurrentConnections: Uses `b.RunParallel` to simulate a real-world scenario where multiple goroutines
//     simultaneously interact with the server.
//
// # Data Flow in Benchmarks:
//
//	[ Benchmark Loop (b.N) ]
//	        |
//	        +---> [ Client net.Conn ] --- (Write data) ---> [ Server sCtx (from pool) ]
//	                                                             |
//	                                                      [ echoHandlerBench ]
//	                                                             |
//	        [ Client net.Conn ] <--- (Read data) <--- [ Server sCtx (from pool) ]
//
// # How to Run:
//
//	go test -v -bench=. -benchmem ./socket/server/unix
//
// # Interpretation of Results:
//   - ns/op: Lower is better. Indicates the overhead of a single operation.
//   - B/op: Should be near zero for steady-state I/O, indicating successful pooling.
//   - allocs/op: Should be minimal, isolating the server's efficiency.
package unix_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	scksru "github.com/nabbar/golib/socket/server/unix"
)

func waitForServerBench(srv scksru.ServerUnix, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func waitForServerStoppedBench(srv scksru.ServerUnix, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func connectToServerBench(b *testing.B, socketPath string) net.Conn {
	var (
		con net.Conn
		err error
	)

	// Retry loop for "resource temporarily unavailable" (EAGAIN)
	// This happens when the listen backlog is full during intense benchmarks.
	for i := 0; i < 10; i++ {
		con, err = net.DialTimeout(libptc.NetworkUnix.Code(), socketPath, 2*time.Second)
		if err == nil {
			return con
		}

		if errors.Is(err, syscall.EAGAIN) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		b.Fatalf(errConnect, err)
	}

	b.Fatalf(errConnect, err)
	return nil
}

func sendAndReceiveBenchOptim(b *testing.B, con net.Conn, data, buffer []byte) {
	var (
		n   int
		err error
	)

	// Write with retry for "resource temporarily unavailable" (EAGAIN)
	for i := 0; i < 10; i++ {
		n, err = con.Write(data)
		if err == nil {
			break
		}
		if errors.Is(err, syscall.EAGAIN) {
			time.Sleep(time.Millisecond)
			continue
		}
		b.Fatalf(errWriteData, err)
	}

	if n != len(data) {
		b.Fatalf(errWriteLen, n, len(data))
	}

	// Read with retry for "resource temporarily unavailable" (EAGAIN)
	for i := 0; i < 10; i++ {
		n, err = io.ReadFull(con, buffer)
		if err == nil {
			break
		}
		if errors.Is(err, syscall.EAGAIN) {
			time.Sleep(time.Millisecond)
			continue
		}
		b.Fatalf(errReadData, err)
	}

	if n != len(data) {
		b.Fatalf(errReadLen, n, len(data))
	}
}

func waitForServerAcceptingConnectionsBench(b *testing.B, socketPath string, timeout time.Duration) {
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	tck := time.NewTicker(50 * time.Millisecond)
	defer tck.Stop()

	for {
		select {
		case <-tmr.C:
			b.Fatalf(errWaitAccept, socketPath, timeout)
			return
		case <-tck.C:
			if c, e := net.DialTimeout(libptc.NetworkUnix.Code(), socketPath, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}

func benchmarkThroughput(b *testing.B, size int) {
	benchmarkThroughputFixed(b, size, 1)
}

func benchmarkThroughputFixed(b *testing.B, size int, numConns int) {
	testSocket := getTestSocketPath()
	defer cleanupSocketFile(testSocket)

	cfg := createDefaultConfig(testSocket)
	srv, err := scksru.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testSocket, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	msg := bytes.Repeat([]byte("a"), size)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	if numConns <= 1 {
		con := connectToServerBench(b, testSocket)
		defer con.Close()
		buf := make([]byte, size)
		for i := 0; i < b.N; i++ {
			sendAndReceiveBenchOptim(b, con, msg, buf)
		}
	} else {
		var wg sync.WaitGroup
		var total int64 = int64(b.N)
		wg.Add(numConns)

		for i := 0; i < numConns; i++ {
			go func() {
				defer wg.Done()
				con := connectToServerBench(b, testSocket)
				defer con.Close()
				buf := make([]byte, size)
				for {
					if atomic.AddInt64(&total, -1) < 0 {
						return
					}
					sendAndReceiveBenchOptim(b, con, msg, buf)
				}
			}()
		}
		wg.Wait()
	}
}

// BenchmarkServerStartup measures the time it takes to start the Unix server.
func BenchmarkServerStartup(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testSocket := getTestSocketPath()
		cfg := createDefaultConfig(testSocket)
		var srv scksru.ServerUnix
		var err error

		b.StartTimer()
		srv, err = scksru.New(nil, echoHandlerBench, cfg)
		if err != nil {
			b.Fatalf(errCreateServer, err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		startServerInBackground(ctx, srv)
		if !waitForServerBench(srv, 5*time.Second) {
			b.Fatalf(errStartServer)
		}
		b.StopTimer()

		// Cleanup
		_ = srv.Close()
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
		cleanupSocketFile(testSocket)
	}
}

// BenchmarkServerShutdown measures the time it takes to shut down the Unix server.
func BenchmarkServerShutdown(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testSocket := getTestSocketPath()
		cfg := createDefaultConfig(testSocket)
		srv, err := scksru.New(nil, echoHandlerBench, cfg)
		if err != nil {
			b.Fatalf(errCreateServer, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		startServerInBackground(ctx, srv)
		waitForServerAcceptingConnectionsBench(b, testSocket, 5*time.Second)

		b.StartTimer()
		_ = srv.Shutdown(ctx)
		b.StopTimer()

		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
		cleanupSocketFile(testSocket)
	}
}

// BenchmarkConnectionEstablishment measures the time it takes to establish a new Unix connection.
func BenchmarkConnectionEstablishment(b *testing.B) {
	testSocket := getTestSocketPath()
	defer cleanupSocketFile(testSocket)

	cfg := createDefaultConfig(testSocket)
	srv, err := scksru.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testSocket, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		con := connectToServerBench(b, testSocket)
		_ = con.Close()
	}
}

// BenchmarkEchoLatency measures the round-trip time for a message echoed by the server.
func BenchmarkEchoLatency(b *testing.B) {
	testSocket := getTestSocketPath()
	defer cleanupSocketFile(testSocket)

	cfg := createDefaultConfig(testSocket)
	srv, err := scksru.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testSocket, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	con := connectToServerBench(b, testSocket)
	defer func() { _ = con.Close() }()

	msg := []byte(msgEchoLatency)
	buf := make([]byte, len(msg))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sendAndReceiveBenchOptim(b, con, msg, buf)
	}
}

// BenchmarkThroughput1KB measures the rate for 1KB messages on 1 connection.
func BenchmarkThroughput1KB(b *testing.B) {
	benchmarkThroughput(b, 1024)
}

// BenchmarkThroughput4KB measures the rate for 4KB messages on 1 connection.
func BenchmarkThroughput4KB(b *testing.B) {
	benchmarkThroughput(b, 4096)
}

// BenchmarkThroughput8KB measures the rate for 8KB messages on 1 connection.
func BenchmarkThroughput8KB(b *testing.B) {
	benchmarkThroughput(b, 8192)
}

// BenchmarkThroughput8KB_C5 measures the rate for 8KB messages on 5 concurrent connections.
func BenchmarkThroughput8KB_C5(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 5)
}

// BenchmarkThroughput8KB_C10 measures the rate for 8KB messages on 10 concurrent connections.
func BenchmarkThroughput8KB_C10(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 10)
}

// BenchmarkThroughput8KB_C25 measures the rate for 8KB messages on 25 concurrent connections.
func BenchmarkThroughput8KB_C25(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 25)
}

// BenchmarkThroughput16KB measures the rate for 16KB messages on 1 connection.
func BenchmarkThroughput16KB(b *testing.B) {
	benchmarkThroughput(b, 16384)
}

// BenchmarkThroughput32KB measures the rate for 32KB messages on 1 connection.
func BenchmarkThroughput32KB(b *testing.B) {
	benchmarkThroughput(b, 32768)
}

// BenchmarkConcurrentConnections measures the performance of handling multiple concurrent connections.
func BenchmarkConcurrentConnections(b *testing.B) {
	testSocket := getTestSocketPath()
	defer cleanupSocketFile(testSocket)

	cfg := createDefaultConfig(testSocket)
	srv, err := scksru.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testSocket, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	msg := []byte(msgConcurrent)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		con := connectToServerBench(b, testSocket)
		defer func() { _ = con.Close() }()
		buf := make([]byte, len(msg))

		for pb.Next() {
			sendAndReceiveBenchOptim(b, con, msg, buf)
		}
	})
}
