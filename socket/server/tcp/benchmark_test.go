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
 *
 *
 */

// Package tcp_test contains performance benchmarks for the TCP server.
//
// # Benchmark Strategy
//
// The 'benchmark_test.go' file is used to measure the impact of recent 
// optimizations (sync.Pool, idlemgr, etc.) on throughput and latency.
//
// # Measured Metrics
//
//   - Startup/Shutdown Latency: Measures the overhead of initializing resources like the idle manager.
//   - Connection Establishment: Measures the cost of Accept() and sCtx allocation.
//   - Echo Latency: Round-trip time for small packets (checks TCP_NODELAY effect).
//   - Throughput (1KB-32KB): Measures total bandwidth and syscall efficiency.
//   - Concurrency (C1-C25): Measures lock contention and context-switching overhead.
//   - TLS Overhead: Compares plain TCP vs. TLS connection costs.
//
// # Execution
//
// Run benchmarks using:
//
//	go test -bench=. -benchmem ./socket/server/tcp/...
package tcp_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
	libptc "github.com/nabbar/golib/network/protocol"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
)

var benchTLSCfg libtls.Config

// init initializes the TLS configuration for benchmarks.
func init() {
	pub, key, err := genCertPair()
	if err != nil {
		panic(fmt.Errorf(errGenCert, err))
	}

	crt, err := tlscrt.ParsePair(key, pub)
	if err != nil {
		panic(fmt.Errorf(errParsePair, err))
	}

	benchTLSCfg = libtls.Config{
		CurveList:  tlscrv.List(),
		CipherList: tlscpr.List(),
		Certs:      []tlscrt.Certif{crt.Model()},
		VersionMin: tlsvrs.VersionTLS12,
		VersionMax: tlsvrs.VersionTLS13,
	}
}

// createTLSConfigBench creates a server config with TLS enabled for benchmarking.
func createTLSConfigBench(b *testing.B, addr string) sckcfg.Server {
	cfg := createDefaultConfig(addr)
	cfg.TLS.Enabled = true
	cfg.TLS.Config = benchTLSCfg
	return cfg
}

// getFreePortBench finds a free port for a benchmark.
func getFreePortBench(b *testing.B) int {
	adr, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), addrFreePort)
	if err != nil {
		b.Fatalf(errResolveAddr, err)
	}

	lis, err := net.ListenTCP(libptc.NetworkTCP.Code(), adr)
	if err != nil {
		b.Fatalf(errListen, err)
	}

	defer func() {
		_ = lis.Close()
	}()

	return lis.Addr().(*net.TCPAddr).Port
}

// getTestAddrBench returns a loopback address for benchmarking.
func getTestAddrBench(b *testing.B) string {
	return fmt.Sprintf("%s:%d", addrLocalhost, getFreePortBench(b))
}

// waitForServerBench waits until the server's IsRunning flag is true.
func waitForServerBench(srv scksrt.ServerTcp, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// waitForServerStoppedBench waits until the server's IsRunning flag is false.
func waitForServerStoppedBench(srv scksrt.ServerTcp, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// connectToServerBench dials the server and fails the benchmark on error.
func connectToServerBench(b *testing.B, addr string) net.Conn {
	con, err := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 2*time.Second)
	if err != nil {
		b.Fatalf(errConnect, err)
	}
	return con
}

// sendAndReceiveBenchOptim performs a single I/O roundtrip using pre-allocated buffers.
func sendAndReceiveBenchOptim(b *testing.B, con net.Conn, data, buffer []byte) {
	n, err := con.Write(data)
	if err != nil {
		b.Fatalf(errWriteData, err)
	}
	if n != len(data) {
		b.Fatalf(errWriteLen, n, len(data))
	}

	n, err = io.ReadFull(con, buffer)
	if err != nil {
		b.Fatalf(errReadData, err)
	}
	if n != len(data) {
		b.Fatalf(errReadLen, n, len(data))
	}
}

// waitForServerAcceptingConnectionsBench verifies that the server is actually 
// ready to accept TCP connections before starting measurements.
func waitForServerAcceptingConnectionsBench(b *testing.B, addr string, timeout time.Duration) {
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	tck := time.NewTicker(50 * time.Millisecond)
	defer tck.Stop()

	for {
		select {
		case <-tmr.C:
			b.Fatalf(errWaitAccept, addr, timeout)
			return
		case <-tck.C:
			if c, e := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}

// benchmarkThroughput runs a throughput test with a single connection and fixed message size.
func benchmarkThroughput(b *testing.B, size int) {
	benchmarkThroughputFixed(b, size, 1)
}

// benchmarkThroughputFixed runs a throughput test with multiple connections and fixed message size.
func benchmarkThroughputFixed(b *testing.B, size int, numConns int) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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

	msg := bytes.Repeat([]byte("a"), size)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	if numConns <= 1 {
		con := connectToServerBench(b, testAddr)
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
				con := connectToServerBench(b, testAddr)
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

// BenchmarkServerStartup measures the latency of creating and starting a new server.
func BenchmarkServerStartup(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testAddr := getTestAddrBench(b)
		cfg := createDefaultConfig(testAddr)
		var srv scksrt.ServerTcp
		var err error

		b.StartTimer()
		srv, err = scksrt.New(nil, echoHandlerBench, cfg)
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
	}
}

// BenchmarkServerShutdown measures the latency of the graceful shutdown process.
func BenchmarkServerShutdown(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testAddr := getTestAddrBench(b)
		cfg := createDefaultConfig(testAddr)
		srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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

// BenchmarkConnectionEstablishment measures the overhead of Accept() and internal state setup.
func BenchmarkConnectionEstablishment(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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

// BenchmarkEchoLatency measures roundtrip latency for a minimal message.
func BenchmarkEchoLatency(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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

	con := connectToServerBench(b, testAddr)
	defer func() { _ = con.Close() }()

	msg := []byte(msgEchoLatency)
	buf := make([]byte, len(msg))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sendAndReceiveBenchOptim(b, con, msg, buf)
	}
}

// BenchmarkThroughput1KB measures bandwidth for 1KB payloads.
func BenchmarkThroughput1KB(b *testing.B) {
	benchmarkThroughput(b, 1024)
}

// BenchmarkThroughput4KB measures bandwidth for 4KB payloads.
func BenchmarkThroughput4KB(b *testing.B) {
	benchmarkThroughput(b, 4096)
}

// BenchmarkThroughput8KB measures bandwidth for 8KB payloads.
func BenchmarkThroughput8KB(b *testing.B) {
	benchmarkThroughput(b, 8192)
}

// BenchmarkThroughput8KB_C5 measures bandwidth for 8KB payloads with 5 concurrent clients.
func BenchmarkThroughput8KB_C5(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 5)
}

// BenchmarkThroughput8KB_C10 measures bandwidth for 8KB payloads with 10 concurrent clients.
func BenchmarkThroughput8KB_C10(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 10)
}

// BenchmarkThroughput8KB_C25 measures bandwidth for 8KB payloads with 25 concurrent clients.
func BenchmarkThroughput8KB_C25(b *testing.B) {
	benchmarkThroughputFixed(b, 8192, 25)
}

// BenchmarkThroughput16KB measures bandwidth for 16KB payloads.
func BenchmarkThroughput16KB(b *testing.B) {
	benchmarkThroughput(b, 16384)
}

// BenchmarkThroughput32KB measures bandwidth for 32KB payloads.
func BenchmarkThroughput32KB(b *testing.B) {
	benchmarkThroughput(b, 32768)
}

// BenchmarkConcurrentConnections stresses the server with parallel client requests.
func BenchmarkConcurrentConnections(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createDefaultConfig(testAddr)
	srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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
		buf := make([]byte, len(msg))

		for pb.Next() {
			sendAndReceiveBenchOptim(b, con, msg, buf)
		}
	})
}

// BenchmarkTLSConnectionEstablishment measures the overhead of the TLS Handshake.
func BenchmarkTLSConnectionEstablishment(b *testing.B) {
	testAddr := getTestAddrBench(b)
	cfg := createTLSConfigBench(b, testAddr)
	srv, err := scksrt.New(nil, echoHandlerBench, cfg)
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

	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		con, err := tls.Dial(libptc.NetworkTCP.Code(), testAddr, tlsCfg)
		if err != nil {
			b.Fatalf(errConnectTLS, err)
		}
		_ = con.Close()
	}
}
