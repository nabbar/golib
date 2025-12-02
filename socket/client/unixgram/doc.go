//go:build linux || darwin

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

/*
Package unixgram provides a Unix domain datagram socket client implementation.

# Overview

The unixgram package implements the github.com/nabbar/golib/socket.Client interface
for Unix domain sockets in datagram mode (SOCK_DGRAM), combining characteristics of
Unix sockets and UDP:
  - Connectionless: No persistent connection like TCP
  - Message-oriented: Preserves message boundaries (datagrams)
  - Unreliable: No guaranteed delivery or ordering
  - Local-only: Uses filesystem paths, not network addresses
  - Fast: Kernel-space only, no network overhead
  - File-based security: Access controlled via filesystem permissions

# Design Philosophy

1. Connectionless Architecture: Like UDP but for local IPC
2. Filesystem-Based: Uses socket file paths instead of IP:port
3. Performance First: Minimal overhead with atomic operations
4. Observable: Lifecycle callbacks for monitoring
5. Thread-Safe: All operations safe for concurrent use

# Key Features

  - Unix domain datagram sockets (SOCK_DGRAM)
  - Connectionless datagram communication
  - Message boundary preservation
  - Thread-safe with atomic state management
  - Configurable error and info callbacks
  - Context-aware operations
  - One-shot request/response support
  - No TLS support (not applicable to Unix sockets)
  - Platform support: Linux and Darwin (macOS)

# Unix Datagram vs Unix Stream Comparison

Unix Datagram (SOCK_DGRAM) - This Package:
  - Connectionless: No persistent connection
  - Message boundaries: Each datagram is independent
  - No ordering guarantee: Datagrams may arrive out of order
  - No acknowledgment: Fire-and-forget delivery
  - Lower latency: No connection overhead
  - Best for: Event logging, notifications, metrics

Unix Stream (SOCK_STREAM) - See github.com/nabbar/golib/socket/client/unix:
  - Connection-oriented: Persistent connection
  - Byte stream: No message boundaries
  - Ordered delivery: Guaranteed order
  - Reliable: Guaranteed delivery with acknowledgment
  - Best for: Request-response, sessions, bulk data

# UDP vs Unix Datagram Comparison

Unix Datagram (SOCK_DGRAM) - This Package:
  - Local only: Same machine IPC
  - Filesystem paths: /tmp/app.sock, ./socket.sock
  - File permissions: Access control via chmod/chown
  - No network overhead: Direct kernel communication
  - Best for: Local microservices, logging daemons

UDP (SOCK_DGRAM) - See github.com/nabbar/golib/socket/client/udp:
  - Network capable: Cross-machine communication
  - IP addresses: 192.168.1.100:8080, [::1]:8080
  - Port-based: No file permissions
  - Network overhead: IP/UDP headers, routing
  - Best for: Distributed systems, multicast

# Architecture

Component Diagram:

	┌────────────────────────────────────────────┐
	│       Unix Datagram Client                 │
	├────────────────────────────────────────────┤
	│                                            │
	│  ┌──────────────┐      ┌──────────────┐    │
	│  │ Socket Path  │      │ Atomic Map   │    │
	│  │ /tmp/app.sock│      │ (state)      │    │
	│  └──────┬───────┘      └──────┬───────┘    │
	│         │                     │            │
	│         ▼                     ▼            │
	│  ┌─────────────────────────────────┐       │
	│  │     UnixConn (SOCK_DGRAM)       │       │
	│  │     Datagram Socket             │       │
	│  └─────────────┬───────────────────┘       │
	│                │                           │
	│                ▼                           │
	│    ┌──────────────────────┐                │
	│    │   User Operations    │                │
	│    │   - Write (send)     │                │
	│    │   - Read (receive)   │                │
	│    └──────────────────────┘                │
	│                                            │
	└────────────────────────────────────────────┘

Data Flow:

	Write(data) → Unix Socket → Server Socket File
	Server → Unix Socket → Read(buffer)

# Basic Usage

The primary type is ClientUnix, created using the New function:

	import (
	    "context"
	    "log"
	    "github.com/nabbar/golib/socket/client/unixgram"
	)

	// Create client with socket path
	client := unixgram.New("/tmp/app.sock")
	if client == nil {
	    log.Fatal("Invalid socket path")
	}
	defer client.Close()

	// Connect to server (creates socket)
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
	    log.Fatal(err)
	}

	// Send datagram (fire-and-forget)
	data := []byte("Event: user login")
	n, err := client.Write(data)
	if err != nil {
	    log.Fatal(err)
	}
	log.Printf("Sent %d bytes", n)

	// Note: No guarantee the datagram was received!

# Datagram Characteristics

Message Boundaries:
  - Each Write() sends one complete datagram
  - Each Read() receives one complete datagram
  - Datagrams are never merged or fragmented by the kernel

Unreliable Delivery:
  - No acknowledgment of receipt
  - No automatic retransmission
  - Datagrams may be lost or dropped
  - Application must handle missing data

Unordered Delivery:
  - Datagrams may arrive out of order
  - No sequence numbers or ordering guarantees
  - Application must handle reordering if needed

Size Limits:
  - System-dependent maximum (typically 16KB-64KB on Linux)
  - Smaller datagrams more reliable
  - Recommended: < 8KB for safety

# Common Use Cases

Event Logging:

	// Send log events to local collector
	client := unixgram.New("/var/run/logger.sock")
	defer client.Close()

	client.Connect(ctx)

	logEntry := []byte("ERROR: Database connection failed")
	client.Write(logEntry)
	// Fire-and-forget, no response expected

Metrics Collection:

	// Send metrics to local aggregator
	client := unixgram.New("/tmp/metrics.sock")
	defer client.Close()

	client.Connect(ctx)

	metric := []byte("http.requests:1|c")
	client.Write(metric)

IPC Notifications:

	// Notify other processes of events
	client := unixgram.New("/tmp/notifications.sock")
	defer client.Close()

	client.Connect(ctx)

	event := []byte("SERVICE_STARTED")
	client.Write(event)

Service Discovery:

	// Register service with local registry
	client := unixgram.New("/run/registry.sock")
	defer client.Close()

	client.Connect(ctx)

	announcement := []byte("service:api:port:8080")
	client.Write(announcement)

# Advanced Features

Error Callbacks:

Register callbacks for error notifications:

	client := unixgram.New("/tmp/app.sock")

	client.RegisterFuncError(func(errs ...error) {
	    for _, err := range errs {
	        log.Printf("Socket error: %v", err)
	    }
	})

Info Callbacks:

Monitor datagram operations:

	client.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
	    switch state {
	    case socket.ConnectionDial:
	        log.Println("Creating socket...")
	    case socket.ConnectionNew:
	        log.Println("Socket ready")
	    case socket.ConnectionWrite:
	        log.Printf("Sending to %v", remote)
	    case socket.ConnectionRead:
	        log.Printf("Receiving from %v", remote)
	    case socket.ConnectionClose:
	        log.Println("Socket closed")
	    }
	})

One-Shot Operations:

Send a request and process response in one call:

	request := bytes.NewBufferString("STATUS")

	err := client.Once(ctx, request, func(reader io.Reader) {
	    buf := make([]byte, 8192)
	    n, _ := reader.Read(buf)
	    fmt.Printf("Response: %s\n", buf[:n])
	})
	// Socket automatically closed after operation

# Performance Considerations

Datagram Size:

Choose appropriate datagram sizes for your use case:

	Small (< 1KB):     Optimal for frequent messages
	Medium (1-8KB):    Good balance for most cases
	Large (> 8KB):     May require kernel buffer tuning

Throughput:

Unix datagram sockets can handle:
  - Small messages (< 1KB): 100,000+ datagrams/second
  - Medium messages (1-8KB): 50,000+ datagrams/second
  - Large messages (> 8KB): 10,000+ datagrams/second

Actual throughput depends on system load and application processing.

Memory Usage:

	Base overhead:        ~1KB (client struct + atomics)
	Per datagram buffer:  size of allocated buffer
	Total per operation:  ~1KB + buffer size

Latency:

	Local IPC:            < 10µs typical
	System call overhead: < 1µs
	Total round-trip:     < 50µs typical

# Error Handling

The package defines specific errors:

	ErrInstance:    Operating on nil client instance
	ErrConnection:  Operating on unconnected client
	ErrAddress:     Invalid or empty socket path

Error Callbacks:

Register error callback to receive all errors:

	client.RegisterFuncError(func(errs ...error) {
	    for _, err := range errs {
	        log.Printf("Error: %v", err)
	    }
	})

Errors are reported but don't stop operations automatically.

# Context Integration

The client supports full context features:

Timeouts:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	// Connect will timeout after 5 seconds

Cancellation:

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
	    // Cancel from another goroutine
	    time.Sleep(1 * time.Second)
	    cancel()
	}()

	err := client.Connect(ctx)
	// Connect will be cancelled

Values:

Context values are available in callbacks through the info callback context parameter.

# Security Considerations

Filesystem Permissions:

Unix datagram sockets use filesystem permissions:
  - Socket file must be readable/writable by client
  - Typically 0600 (user only) or 0660 (user+group)
  - Check permissions before connecting

Socket File Location:

Choose secure locations:

	/tmp/app.sock:        Temporary, may be accessible to all users
	/var/run/app.sock:    System runtime directory (requires permissions)
	/run/user/$UID/app.sock: User-specific, more secure
	./socket.sock:        Application directory, check directory permissions

Input Validation:

Always validate received data:
  - No authentication at transport layer
  - Any process with permissions can send datagrams
  - Application must validate sender identity if needed
  - Consider message signing for critical data

# Thread Safety

All ClientUnix methods are thread-safe:
  - New() can be called concurrently
  - Connect() safe for concurrent calls (last wins)
  - IsConnected() safe for concurrent reads
  - Read() should NOT be called concurrently (undefined behavior)
  - Write() should NOT be called concurrently (undefined behavior)
  - Close() safe for concurrent calls (first succeeds)
  - RegisterFuncError(), RegisterFuncInfo() are thread-safe

Concurrent Read/Write:

Do NOT call Read() or Write() concurrently on the same client:

	// ❌ BAD: Concurrent writes (race condition)
	go client.Write([]byte("message 1"))
	go client.Write([]byte("message 2"))

	// ✅ GOOD: Sequential or mutex-protected
	mu sync.Mutex
	go func() {
	    mu.Lock()
	    defer mu.Unlock()
	    client.Write([]byte("message 1"))
	}()

# Platform Support

Supported Platforms:
  - Linux: Full support with all features
  - Darwin (macOS): Full support with all features

Unsupported Platforms:
  - Windows: Unix domain sockets not supported (see ignore.go)
  - Other Unix variants: May work but untested

For cross-platform IPC, consider:
  - TCP/IP on localhost
  - UDP on localhost
  - Named pipes (platform-specific)

# Related Packages

  - github.com/nabbar/golib/socket - Base Client interface
  - github.com/nabbar/golib/socket/client/unix - Connection-oriented Unix sockets (SOCK_STREAM)
  - github.com/nabbar/golib/socket/client/udp - UDP datagram sockets (network-capable)
  - github.com/nabbar/golib/socket/client/tcp - TCP stream sockets (network-capable)
  - github.com/nabbar/golib/socket/server/unixgram - Unix datagram server
  - github.com/nabbar/golib/atomic - Thread-safe atomic operations
  - github.com/nabbar/golib/runner - Panic recovery utilities

# Examples

See example_test.go for comprehensive usage examples:
  - Basic client creation and connection
  - Send and receive datagrams
  - Error handling patterns
  - Callback registration
  - One-shot operations
  - Context integration
*/
package unixgram
