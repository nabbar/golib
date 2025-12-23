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

package server_test

import (
	"context"
	"fmt"
	"io"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"
)

// Example demonstrates creating a basic TCP server using the factory.
// This is the simplest way to create a socket server.
func Example() {
	// Create a simple echo handler
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, _ = c.Write(buf[:n])
			}
		}
	}

	// Create server configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":8080",
	}

	// Create server using factory
	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		panic(err)
	}

	// Start server
	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown after demonstration
	_ = srv.Shutdown(ctx)
	// Output:
}

// ExampleNew demonstrates creating a TCP server using the factory.
func ExampleNew() {
	// Define connection handler
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		_, _ = io.Copy(c, c) // Echo
	}

	// Create configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9000",
	}

	// Create server using factory
	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	fmt.Printf("Server created successfully\n")
	_ = srv
	// Output: Server created successfully
}

// ExampleNew_tcp demonstrates creating a TCP server.
func ExampleNew_tcp() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9001",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("TCP server created")
	_ = srv
	// Output: TCP server created
}

// ExampleNew_tcp4 demonstrates creating a TCP4 server (IPv4 only).
func ExampleNew_tcp4() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP4,
		Address: ":9002",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("TCP4 server created")
	_ = srv
	// Output: TCP4 server created
}

// ExampleNew_tcp6 demonstrates creating a TCP6 server (IPv6 only).
func ExampleNew_tcp6() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP6,
		Address: ":9003",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("TCP6 server created")
	_ = srv
	// Output: TCP6 server created
}

// ExampleNew_udp demonstrates creating a UDP server.
func ExampleNew_udp() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9004",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("UDP server created")
	_ = srv
	// Output: UDP server created
}

// ExampleNew_udp4 demonstrates creating a UDP4 server (IPv4 only).
func ExampleNew_udp4() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP4,
		Address: ":9005",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("UDP4 server created")
	_ = srv
	// Output: UDP4 server created
}

// ExampleNew_udp6 demonstrates creating a UDP6 server (IPv6 only).
func ExampleNew_udp6() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP6,
		Address: ":9006",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("UDP6 server created")
	_ = srv
	// Output: UDP6 server created
}

// ExampleNew_invalidProtocol demonstrates error handling for invalid protocols.
func ExampleNew_invalidProtocol() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	// Use an invalid protocol value
	cfg := sckcfg.Server{
		Network: 255, // Invalid protocol
		Address: ":9007",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Println("Error: invalid protocol")
	}
	_ = srv
	// Output: Error: invalid protocol
}

// ExampleNew_withIdleTimeout demonstrates configuring idle timeout.
func ExampleNew_withIdleTimeout() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":9008",
		ConIdleTimeout: libdur.Minutes(5),
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Server with idle timeout created")
	_ = srv
	// Output: Server with idle timeout created
}

// Example_multipleServers demonstrates running multiple servers with different protocols.
func Example_multipleServers() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	// Create TCP server
	tcpCfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9100",
	}
	tcpSrv, err := scksrv.New(nil, handler, tcpCfg)
	if err != nil {
		fmt.Printf("TCP Error: %v\n", err)
		return
	}

	// Create UDP server
	udpCfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9101",
	}
	udpSrv, err := scksrv.New(nil, handler, udpCfg)
	if err != nil {
		fmt.Printf("UDP Error: %v\n", err)
		return
	}

	fmt.Println("Multiple servers created successfully")

	// Start both servers
	ctx := context.Background()
	go func() {
		_ = tcpSrv.Listen(ctx)
	}()
	go func() {
		_ = udpSrv.Listen(ctx)
	}()

	// Wait and shutdown
	time.Sleep(50 * time.Millisecond)
	_ = tcpSrv.Shutdown(ctx)
	_ = udpSrv.Shutdown(ctx)

	// Output: Multiple servers created successfully
}

// Example_gracefulShutdown demonstrates graceful server shutdown.
func Example_gracefulShutdown() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		time.Sleep(100 * time.Millisecond)
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9200",
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Start server
	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err == nil {
		fmt.Println("Server shut down gracefully")
	}

	// Output: Server shut down gracefully
}
