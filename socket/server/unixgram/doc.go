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
Package unixgram provides a Unix domain datagram socket server implementation.

# Overview

The unixgram package implements the github.com/nabbar/golib/socket.Server interface
for Unix domain sockets in datagram mode (SOCK_DGRAM), providing a connectionless
IPC mechanism with features including:
  - Unix domain socket file creation and management
  - File permissions and group ownership control
  - Datagram handling without persistent connections
  - Single handler for all incoming datagrams
  - Callback hooks for errors and datagram events
  - Graceful shutdown support
  - Atomic state management
  - Context-aware operations

# Design Philosophy

1. Connectionless Architecture: Like UDP but for local IPC
2. Filesystem-Based: Uses socket files instead of ports
3. Security by Design: File permissions control access
4. Observable: Lifecycle callbacks for monitoring
5. Stateless: No per-client connection state

# Key Features

  - Unix domain datagram sockets (SOCK_DGRAM)
  - Filesystem-based socket files with configurable permissions
  - Group ownership control for multi-user scenarios
  - Connectionless datagram handling (similar to UDP)
  - Single handler processes all datagrams
  - Real-time monitoring via callbacks
  - Graceful shutdown with automatic socket file cleanup
  - Thread-safe atomic operations
  - Context integration for lifecycle management
  - Platform support: Linux and Darwin (macOS)

# Unix Datagram vs Unix Stream Comparison

Unix Datagram (SOCK_DGRAM) - This Package:
  - Connectionless: No per-client connections
  - Message boundaries: Each datagram is independent
  - No ordering guarantee: Datagrams may arrive out of order
  - No acknowledgment: No guaranteed delivery
  - Single handler: One handler for all datagrams
  - Best for: Fire-and-forget messages, logging, notifications
  - OpenConnections(): Returns 1 when running, 0 when stopped

Unix Stream (SOCK_STREAM) - See github.com/nabbar/golib/socket/server/unix:
  - Connection-oriented: Per-client connections
  - Byte stream: No message boundaries
  - Ordered delivery: Guaranteed order
  - Reliable: Guaranteed delivery with acknowledgment
  - Per-client handler: One handler per connection
  - Best for: Request-response, sessions, bulk data transfer
  - OpenConnections(): Returns active connection count

# UDP vs Unix Datagram Comparison

Unix Datagram (SOCK_DGRAM) - This Package:
  - Local only: Same machine IPC
  - Filesystem paths: /tmp/app.sock, ./socket.sock
  - File permissions: 0600, 0660, 0770
  - Group ownership: Control by GID
  - No network overhead: Direct kernel communication
  - Best for: Local microservices, logging daemons, IPC

UDP (SOCK_DGRAM) - See github.com/nabbar/golib/socket/server/udp:
  - Network capable: Cross-machine communication
  - IP addresses: 192.168.1.100:8080, [::1]:8080
  - Port-based: No file permissions
  - Network overhead: IP/UDP headers, routing
  - Best for: Distributed systems, multicast, discovery

# Architecture

Component Diagram:

	┌────────────────────────────────────────────────────┐
	│              Unix Datagram Server                  │
	├────────────────────────────────────────────────────┤
	│                                                    │
	│  ┌──────────────┐      ┌──────────────────┐        │
	│  │ Socket File  │      │  Context Manager │        │
	│  │ /tmp/app.sock│      │  (cancellation)  │        │
	│  └──────┬───────┘      └────────┬─────────┘        │
	│         │                       │                  │
	│         ▼                       ▼                  │
	│  ┌──────────────────────────────────────┐          │
	│  │      UnixConn Listener               │          │
	│  │      (SOCK_DGRAM)                    │          │
	│  └──────────────┬───────────────────────┘          │
	│                 │                                  │
	│                 ▼                                  │
	│    Single Handler Goroutine                        │
	│    ┌─────────────────────────┐                     │
	│    │  sCtx (I/O wrapper)     │                     │
	│    │  - ReadFrom (datagram)  │                     │
	│    │  - WriteTo (response)   │                     │
	│    │  - Sender tracking      │                     │
	│    └──────────┬──────────────┘                     │
	│               │                                    │
	│               ▼                                    │
	│    ┌─────────────────────┐                         │
	│    │   User Handler      │                         │
	│    │   (HandlerFunc)     │                         │
	│    └─────────────────────┘                         │
	│                                                    │
	└────────────────────────────────────────────────────┘

Data Flow:

	Sender → Socket File → UnixConn.ReadFrom() → Handler
	                                            ↓
	Handler → UnixConn.WriteTo() → Socket File → Sender

# Basic Usage

The primary type is ServerUnixGram, created using the New function:

	import (
	    libprm "github.com/nabbar/golib/file/perm"
	    libsck "github.com/nabbar/golib/socket"
	    sckcfg "github.com/nabbar/golib/socket/config"
	    "github.com/nabbar/golib/socket/server/unixgram"
	)

	// Define handler for incoming datagrams
	handler := func(ctx libsck.Context) {
	    defer ctx.Close()

	    buf := make([]byte, 65507) // Max datagram size
	    for {
	        n, err := ctx.Read(buf)
	        if err != nil {
	            break
	        }
	        // Process datagram
	        log.Printf("Received: %s", buf[:n])
	    }
	}

	// Create server configuration
	cfg := sckcfg.Server{
	    Network:   libptc.NetworkUnixGram,
	    Address:   "/tmp/app.sock",
	    PermFile:  libprm.NewPerm(0660), // rw-rw----
	    GroupPerm: -1, // Use process group
	}

	// Create server
	srv, err := unixgram.New(nil, handler, cfg)
	if err != nil {
	    log.Fatal(err)
	}

	// Start server
	ctx := context.Background()
	if err := srv.Listen(ctx); err != nil {
	    log.Fatal(err)
	}

# Socket File Management

Socket File Creation:

The socket file is created when Listen() is called:
 1. Validates the socket path
 2. Removes existing file if present
 3. Creates Unix datagram socket (SOCK_DGRAM)
 4. Applies file permissions via chmod
 5. Changes group ownership via chown
 6. Starts accepting datagrams

Socket File Permissions:

	0600 (rw-------): Only owner can send datagrams
	0660 (rw-rw----): Owner and group can send datagrams
	0666 (rw-rw-rw-): Anyone can send datagrams (use with caution)
	0770 (rwxrwx---): Owner and group (executable not meaningful for sockets)

Group Ownership:

	-1: Use process's default group (typically user's primary group)
	 0: Root group (requires elevated privileges)
	>0: Specific GID (must be valid on system, max 32767)

Socket File Cleanup:

The socket file is automatically removed during shutdown:
  - Shutdown() method removes the file
  - Close() method removes the file
  - Context cancellation triggers cleanup
  - Failed Listen() cleans up on error

# Datagram Handling

Unlike connection-oriented Unix sockets, Unix datagram servers:

No Persistent Connections:
  - Each datagram is independent
  - No connection setup or teardown
  - No per-sender state maintained

Single Handler:
  - One handler goroutine processes all datagrams
  - Handler runs for the server's lifetime
  - Handler receives datagrams from all senders

Message Boundaries:
  - Each Read() receives one complete datagram
  - Datagrams are never merged or split
  - Maximum datagram size: ~65507 bytes (system dependent)

No Ordering Guarantee:
  - Datagrams may arrive out of order
  - No acknowledgment or retry mechanism
  - Application must handle duplicate or missing datagrams

# Common Use Cases

Local Logging Daemon:

	// Log collector receiving from multiple processes
	handler := func(ctx libsck.Context) {
	    buf := make([]byte, 8192)
	    for {
	        n, err := ctx.Read(buf)
	        if err != nil {
	            break
	        }
	        logFile.Write(buf[:n])
	    }
	}

	srv, _ := unixgram.New(nil, handler, cfg)
	srv.Listen(context.Background())

Metrics Collection:

	// Metrics aggregator for local services
	handler := func(ctx libsck.Context) {
	    buf := make([]byte, 1024)
	    for {
	        n, err := ctx.Read(buf)
	        if err != nil {
	            break
	        }
	        metric := parseMetric(buf[:n])
	        metricsDB.Record(metric)
	    }
	}

IPC Notification System:

	// Notification server for inter-process events
	handler := func(ctx libsck.Context) {
	    buf := make([]byte, 4096)
	    for {
	        n, err := ctx.Read(buf)
	        if err != nil {
	            break
	        }
	        event := parseEvent(buf[:n])
	        eventBus.Publish(event)
	    }
	}

Service Discovery:

	// Local service registry for microservices
	handler := func(ctx libsck.Context) {
	    buf := make([]byte, 2048)
	    for {
	        n, err := ctx.Read(buf)
	        if err != nil {
	            break
	        }
	        service := parseServiceAnnouncement(buf[:n])
	        registry.Register(service)
	    }
	}

# Performance Considerations

Datagram Size:

Unix datagram sockets have size limits (typically ~65507 bytes on Linux):

	Small datagrams (< 1KB):  Optimal for frequent messages
	Medium datagrams (1-8KB): Good balance
	Large datagrams (> 8KB):  May require fragmentation

Buffer sizing:

	buf := make([]byte, 65507) // Max size, safe for all datagrams
	buf := make([]byte, 8192)  // 8KB, good for most use cases
	buf := make([]byte, 1024)  // 1KB, minimal overhead

Memory Usage:

	Base overhead:        ~2KB (server struct + atomics)
	Per handler:          ~8KB (goroutine stack)
	Per datagram buffer:  size of allocated buffer
	Total:                ~10KB + buffer size

Throughput:

Unix datagram sockets can handle:
  - Small messages (< 1KB): 100,000+ datagrams/second
  - Medium messages (1-8KB): 50,000+ datagrams/second
  - Large messages (> 8KB): 10,000+ datagrams/second

Actual throughput depends on handler processing speed.

# Error Handling

The package defines specific errors:

	ErrInvalidUnixFile:   Socket file path is invalid or empty
	ErrInvalidGroup:      GID exceeds MaxGID (32767)
	ErrInvalidHandler:    Handler function is nil
	ErrShutdownTimeout:   Shutdown exceeded context timeout
	ErrInvalidInstance:   Operating on nil server instance

Error Callbacks:

Register error callback to receive all errors:

	srv.RegisterFuncError(func(errs ...error) {
	    for _, err := range errs {
	        log.Printf("Server error: %v", err)
	    }
	})

Errors are reported but don't stop the server automatically.

# Monitoring and Callbacks

Connection Info Callback:

Monitor datagram events:

	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
	    switch state {
	    case libsck.ConnectionNew:
	        log.Printf("Handler started")
	    case libsck.ConnectionRead:
	        log.Printf("Datagram from %s", remote)
	    case libsck.ConnectionWrite:
	        log.Printf("Response to %s", remote)
	    case libsck.ConnectionClose:
	        log.Printf("Handler stopped")
	    }
	})

Server Info Callback:

Receive server lifecycle messages:

	srv.RegisterFuncInfoServer(func(msg string) {
	    log.Printf("Server: %s", msg)
	})

State Monitoring:

	running := srv.IsRunning()  // true when accepting datagrams
	gone := srv.IsGone()        // true when stopped
	conns := srv.OpenConnections() // 1 if running, 0 if stopped

# Graceful Shutdown

Shutdown via Context:

	ctx, cancel := context.WithCancel(context.Background())
	go srv.Listen(ctx)

	// Later, trigger shutdown
	cancel()

Shutdown via Method:

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
	    log.Printf("Shutdown error: %v", err)
	}

Close Immediately:

	srv.Close() // Equivalent to Shutdown(context.Background())

Shutdown process:
 1. Sets server to "gone" state
 2. Stops accepting new datagrams
 3. Waits for handler to exit
 4. Closes socket connection
 5. Removes socket file

# Context Integration

The server implements full context support:

Deadlines:

	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	// Server will auto-shutdown after 10 seconds
	srv.Listen(ctx)

Values:

	type ctxKey string
	ctx := context.WithValue(parent, ctxKey("requestID"), "12345")

	// Handler can access context values
	handler := func(ctx libsck.Context) {
	    if reqID := ctx.Value(ctxKey("requestID")); reqID != nil {
	        log.Printf("Request ID: %v", reqID)
	    }
	}

Cancellation:

	ctx, cancel := context.WithCancel(parent)

	// Any cancellation propagates to the server
	go srv.Listen(ctx)

	// Trigger shutdown from any goroutine
	cancel()

# Security Considerations

File Permissions:

Choose appropriate permissions for your security model:

	0600: Maximum security - only socket owner can communicate
	0660: Group access - owner and group members can communicate
	0666: Open access - any local user can communicate (risky)

Group Ownership:

Use group ownership for multi-user scenarios:

	-1: Process's default group (safest)
	gid: Specific group (e.g., "www-data" group for web services)

Socket File Location:

	/tmp/app.sock:        Temporary, world-readable directory
	/var/run/app.sock:    System runtime directory (requires permissions)
	./socket.sock:        Current directory (application-specific)
	/home/user/app.sock:  User's home directory (user-specific)

Input Validation:

Always validate datagram contents:
  - Datagrams can come from any local process (within permissions)
  - No authentication at transport layer
  - Application must validate sender identity if needed
  - Consider message signing for critical data

# Thread Safety

All ServerUnixGram methods are thread-safe:
  - Listen() can be called from any goroutine
  - Shutdown() can be called from any goroutine
  - IsRunning(), IsGone(), OpenConnections() are safe for concurrent reads
  - RegisterFuncError(), RegisterFuncInfo(), RegisterFuncInfoServer() are thread-safe

Internal synchronization uses atomic operations for lock-free state management.

# Platform Support

Supported Platforms:
  - Linux: Full support with all features
  - Darwin (macOS): Full support with all features

Unsupported Platforms:
  - Windows: Unix domain sockets not supported (see ignore.go)
  - Other Unix variants: May work but untested

For Windows IPC, consider:
  - Named pipes (Windows native)
  - TCP/IP on localhost
  - UDP on localhost

# Related Packages

  - github.com/nabbar/golib/socket - Base Server interface
  - github.com/nabbar/golib/socket/server/unix - Connection-oriented Unix sockets (SOCK_STREAM)
  - github.com/nabbar/golib/socket/server/udp - UDP datagram sockets (network-capable)
  - github.com/nabbar/golib/socket/server/tcp - TCP stream sockets (network-capable)
  - github.com/nabbar/golib/socket/config - Server configuration types
  - github.com/nabbar/golib/file/perm - File permission handling
  - github.com/nabbar/golib/network/protocol - Network protocol constants

# Examples

See example_test.go for comprehensive usage examples:
  - Basic datagram server
  - Server with callbacks
  - Custom socket configuration
  - Graceful shutdown patterns
  - Error handling
  - State monitoring
*/
package unixgram
