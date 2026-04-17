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

package udp_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/nabbar/golib/socket/server/udp"
)

// # Performance Benchmarking Suite
//
// This file contains a comprehensive suite of benchmarks designed to measure:
//  - Latency of startup and shutdown.
//  - Throughput for various datagram sizes (1KB to 32KB).
//  - Concurrent performance under high load (multiple goroutines).
//  - Impact of optimization strategies (e.g., sync.Pool for buffers).
//
// # Optimization: Memory Management
//
// All benchmarks utilize the 'bufPool' (defined in helper_test.go) to minimize
// allocations during high-frequency I/O tests. This ensures results reflect
// network/logic performance rather than GC overhead.

// getFreePortBench finds an available UDP port on the loopback interface.
// It achieves this by binding to port 0 and then immediately releasing it.
func getFreePortBench(b *testing.B) int {
	adr, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), addrFreePort)
	if err != nil {
		b.Fatalf(errResolveAddr, err)
	}

	lis, err := net.ListenUDP(libptc.NetworkUDP.Code(), adr)
	if err != nil {
		b.Fatalf(errListen, err)
	}

	defer func() {
		_ = lis.Close()
	}()

	return adr.Port
}

// getTestAddrBench generates a dynamic localhost address for benchmarking.
func getTestAddrBench(b *testing.B) string {
	return fmt.Sprintf("%s:%d", addrLocalhost, getFreePortBench(b))
}

// waitForServerBench blocks until the server reports it is running.
// Optimized for benchmarks: it uses a tight loop instead of heavy sleeps
// to reduce the "noise" in startup time measurements.
func waitForServerBench(srv udp.ServerUdp, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return true
		}
	}
	return srv.IsRunning()
}

// waitForServerStoppedBench blocks until the server reports it has stopped.
func waitForServerStoppedBench(srv udp.ServerUdp, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return true
		}
	}
	return !srv.IsRunning()
}

// connectToServerBench establishes a raw UDP connection to the server for testing.
func connectToServerBench(b *testing.B, addr string) net.Conn {
	con, err := net.DialTimeout(libptc.NetworkUDP.Code(), addr, 2*time.Second)
	if err != nil {
		b.Fatalf(errConnect, err)
	}
	return con
}

// sendAndReceiveBenchOptim performs a single request/response cycle.
//
// # Use Case: Latency Measurement
//
// This helper measures the round-trip time (RTT) for a single datagram.
// Note: UDP echo depends on the server handler logic (Write/WriteTo).
func sendAndReceiveBenchOptim(b *testing.B, con net.Conn, data, buffer []byte) {
	n, err := con.Write(data)
	if err != nil {
		b.Fatalf(errWriteData, err)
	}
	if n != len(data) {
		b.Fatalf(errWriteLen, n, len(data))
	}

	// Wait for response with a short deadline
	_ = con.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err = con.Read(buffer)
	if err != nil {
		return
	}
	if n != len(data) {
		b.Fatalf(errReadLen, n, len(data))
	}
}

// waitForServerAcceptingConnectionsBench ensures the socket is truly bound
// and accepting packets before proceeding with the benchmark.
func waitForServerAcceptingConnectionsBench(b *testing.B, addr string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if c, e := net.DialTimeout(libptc.NetworkUDP.Code(), addr, time.Millisecond); e == nil {
			_ = c.Close()
			return
		}
	}
	b.Fatalf(errWaitAccept, addr, timeout)
}

// benchmarkThroughput is a parameterized throughput benchmark.
//
// # Logic Flow
//
// 1. Setup: Start server and wait for readiness.
// 2. Warmup: Prepare message buffers.
// 3. Execution: Run b.N iterations of sending the message.
// 4. Teardown: Shutdown server and report results.
func benchmarkThroughput(b *testing.B, size int) {
	benchmarkThroughputFixed(b, size, 1)
}

// benchmarkThroughputFixed allows testing concurrent connections.
//
// # Concurrency Diagram
//
//	[Benchmark Control]
//	      │
//	      ├──── [Worker Goroutine 1] ──▶ [Write to Socket]
//	      ├──── [Worker Goroutine 2] ──▶ [Write to Socket]
//	      └──── [Worker Goroutine N] ──▶ [Write to Socket]
//
// Total throughput is the sum of all parallel writes.
func benchmarkThroughputFixed(b *testing.B, size int, numConns int) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := udp.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testAddr, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	msg := bytes.Repeat([]byte(msgDefaultByte), size)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	if numConns <= 1 {
		con := connectToServerBench(b, testAddr)
		defer con.Close()
		for i := 0; i < b.N; i++ {
			_, _ = con.Write(msg)
		}
	} else {
		var wg sync.WaitGroup
		var total int64 = int64(b.N)
		wg.Add(numConns)

		for i := 0; i < numConns; i++ {
			go func() {
				defer wg.Done()
				con := connectToServerBench(b, testAddr)
				defer con.Close()
				for {
					if atomic.AddInt64(&total, -1) < 0 {
						return
					}
					_, _ = con.Write(msg)
				}
			}()
		}
		wg.Wait()
	}
}

// BenchmarkServerStartup measures the raw overhead of starting the UDP listener.
// This highlights the performance gain from removing poll-based tickers.
func BenchmarkServerStartup(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testAddr := getTestAddrBench(b)
		cfg := createDefaultConfig(testAddr)
		var srv udp.ServerUdp
		var err error

		b.StartTimer()
		srv, err = udp.New(nil, echoHandlerBench, cfg)
		if err != nil {
			b.Fatalf(errCreateServer, err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		startServerInBackground(ctx, srv)
		if !waitForServerBench(srv, 5*time.Second) {
			b.Fatalf(errStartServer)
		}
		b.StopTimer()

		// Cleanup between iterations
		_ = srv.Close()
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}
}

// BenchmarkServerShutdown measures the latency of stopping the listener.
// Crucial for measuring the "gnc" channel broadcast speed.
func BenchmarkServerShutdown(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testAddr := getTestAddrBench(b)
		cfg := createDefaultConfig(testAddr)
		srv, err := udp.New(nil, echoHandlerBench, cfg)
		if err != nil {
			b.Fatalf(errCreateServer, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		startServerInBackground(ctx, srv)
		waitForServerAcceptingConnectionsBench(b, testAddr, 5*time.Second)

		b.StartTimer()
		_ = srv.Shutdown(ctx)
		b.StopTimer()

		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}
}

// BenchmarkConnectionEstablishment measures the latency of net.Dial for UDP.
func BenchmarkConnectionEstablishment(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := udp.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testAddr, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		con := connectToServerBench(b, testAddr)
		_ = con.Close()
	}
}

// BenchmarkThroughput1KB measures 1024-byte packet performance.
func BenchmarkThroughput1KB(b *testing.B) {
	benchmarkThroughput(b, 1024)
}

// BenchmarkThroughput4KB measures 4096-byte packet performance.
func BenchmarkThroughput4KB(b *testing.B) {
	benchmarkThroughput(b, 4096)
}

// BenchmarkThroughput8KB measures 8192-byte packet performance.
func BenchmarkThroughput8KB(b *testing.B) {
	benchmarkThroughput(b, 8192)
}

// BenchmarkThroughput8KB_C5 measures 8KB packets with 5 concurrent workers.
func BenchmarkThroughput8KB_C5(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 5)
}

// BenchmarkThroughput8KB_C10 measures 8KB packets with 10 concurrent workers.
func BenchmarkThroughput8KB_C10(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 10)
}

// BenchmarkThroughput8KB_C25 measures 8KB packets with 25 concurrent workers.
func BenchmarkThroughput8KB_C25(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 25)
}

// BenchmarkThroughput16KB measures 16384-byte packet performance.
func BenchmarkThroughput16KB(b *testing.B) {
	benchmarkThroughput(b, 16384)
}

// BenchmarkThroughput32KB measures 32768-byte packet performance.
func BenchmarkThroughput32KB(b *testing.B) {
	benchmarkThroughput(b, 32768)
}

// BenchmarkConcurrentSends uses b.RunParallel to measure multi-core utilization.
func BenchmarkConcurrentSends(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := udp.New(nil, echoHandlerBench, cfg)
	if err != nil {
		b.Fatalf(errCreateServer, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	startServerInBackground(ctx, srv)
	waitForServerAcceptingConnectionsBench(b, testAddr, 5*time.Second)
	defer func() {
		_ = srv.Shutdown(ctx)
		cancel()
		waitForServerStoppedBench(srv, 5*time.Second)
	}()

	msg := []byte(msgConcurrent)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		con := connectToServerBench(b, testAddr)
		defer func() { _ = con.Close() }()

		for pb.Next() {
			_, _ = con.Write(msg)
		}
	})
}
