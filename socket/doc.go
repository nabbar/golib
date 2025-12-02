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
 */

// Package socket provides a unified, production-ready framework for network socket communication
// across multiple protocols and platforms. It offers consistent interfaces for both client and
// server implementations, supporting TCP, UDP, and Unix domain sockets with optional TLS encryption.
//
// # Overview
//
// The socket package serves as the foundation for all socket-based communication in golib,
// providing platform-aware abstractions that work seamlessly across different network protocols
// and operating systems. It is designed for high-performance, concurrent applications requiring
// reliable socket communication with minimal boilerplate.
//
// This package defines the core interfaces, types, and constants that are implemented by the
// specialized sub-packages for different protocols and operation modes.
//
// # Design Philosophy
//
//  1. Unified Interface: All socket types implement common interfaces (Server, Client, Context)
//  2. Platform Awareness: Automatic protocol availability based on operating system
//  3. Type Safety: Configuration-driven construction with compile-time validation
//  4. Performance First: Zero-copy operations and minimal allocations where possible
//  5. Production Ready: Built-in error handling, logging, and monitoring capabilities
//  6. Concurrent by Design: Thread-safe operations with atomic state management
//  7. Standard Compliance: Implements io.Reader, io.Writer, io.Closer, context.Context
//
// # Architecture
//
// ## Package Structure
//
//	socket/                           # Core interfaces and types (this package)
//	├── interface.go                  # Server, Client, Context interfaces
//	├── context.go                    # Context interface definition
//	├── doc.go                        # Package documentation
//	│
//	├── config/                       # Configuration builders and validators
//	│   ├── builder_client.go         # Client configuration builder
//	│   ├── builder_server.go         # Server configuration builder
//	│   └── validator.go              # Configuration validation
//	│
//	├── client/                       # Client factory and implementations
//	│   ├── interface.go              # Factory method (New)
//	│   ├── tcp/                      # TCP client implementation
//	│   ├── udp/                      # UDP client implementation
//	│   ├── unix/                     # Unix socket client (Linux/Darwin)
//	│   └── unixgram/                 # Unix datagram client (Linux/Darwin)
//	│
//	└── server/                       # Server factory and implementations
//	    ├── interface.go              # Factory method (New)
//	    ├── tcp/                      # TCP server implementation
//	    ├── udp/                      # UDP server implementation
//	    ├── unix/                     # Unix socket server (Linux/Darwin)
//	    └── unixgram/                 # Unix datagram server (Linux/Darwin)
//
// ## Component Diagram
//
//	┌────────────────────────────────────────────────────────────────────┐
//	│                       socket Package                               │
//	│                  (Core Interfaces & Types)                         │
//	├────────────────────────────────────────────────────────────────────┤
//	│                                                                    │
//	│  ┌──────────────────────────────────────────────────────────┐      │
//	│  │             Core Interfaces                              │      │
//	│  │  • Server   - Server operations                          │      │
//	│  │  • Client   - Client operations                          │      │
//	│  │  • Context  - Connection context                         │      │
//	│  └──────────────────────────────────────────────────────────┘      │
//	│                                                                    │
//	│  ┌──────────────────────────────────────────────────────────┐      │
//	│  │             Core Types                                   │      │
//	│  │  • ConnState       - Connection state tracking           │      │
//	│  │  • HandlerFunc     - Request handler                     │      │
//	│  │  • FuncError       - Error callback                      │      │
//	│  │  • FuncInfo        - Connection info callback            │      │
//	│  └──────────────────────────────────────────────────────────┘      │
//	│                                                                    │
//	└─────┬──────────────────────────────────────────────────┬───────────┘
//	      │                                                  │
//	      ▼                                                  ▼
//	┌───────────────────────┐                   ┌───────────────────────┐
//	│   client Package      │                   │   server Package      │
//	│   (Client Factory)    │                   │   (Server Factory)    │
//	├───────────────────────┤                   ├───────────────────────┤
//	│ • TCP Client          │                   │ • TCP Server          │
//	│ • UDP Client          │                   │ • UDP Server          │
//	│ • Unix Client         │                   │ • Unix Server         │
//	│ • UnixGram Client     │                   │ • UnixGram Server     │
//	└───────────────────────┘                   └───────────────────────┘
//
// ## Data Flow
//
// ### Server Connection Flow
//
//	Listen → Accept → Handler → Read/Write → Close
//	   │        │        │          │           │
//	   │        │        │          │           └─→ ConnectionClose
//	   │        │        │          ├─→ ConnectionRead
//	   │        │        │          └─→ ConnectionWrite
//	   │        │        └─→ ConnectionHandler
//	   │        └─→ ConnectionNew
//	   └─→ Server Start
//
// ### Client Connection Flow
//
//	Connect → Read/Write → Close
//	   │          │          │
//	   │          │          └─→ ConnectionClose
//	   │          ├─→ ConnectionRead
//	   │          └─→ ConnectionWrite
//	   └─→ ConnectionDial
//
// # Key Features
//
//   - Multiple Protocol Support: TCP, UDP, Unix domain sockets, Unix datagrams
//   - TLS/SSL Encryption: Optional TLS for TCP connections
//   - Platform-Aware: Automatic Unix socket support on Linux/Darwin
//   - Unified API: Consistent interface across all protocols
//   - Configuration Builders: Type-safe configuration with validation
//   - Connection Monitoring: State tracking and event callbacks
//   - Error Handling: Comprehensive error propagation and filtering
//   - Context Integration: Full support for Go's context.Context
//   - Resource Management: Automatic cleanup and graceful shutdown
//   - High Performance: Optimized for concurrent, high-throughput scenarios
//
// # Core Interfaces
//
// ## Server Interface
//
// The Server interface provides methods for configuring, starting, and managing a socket server:
//
//	type Server interface {
//	    io.Closer
//	    RegisterFuncError(FuncError)
//	    RegisterFuncInfo(FuncInfo)
//	    RegisterFuncInfoServer(FuncInfoSrv)
//	    SetTLS(enable bool, config TLSConfig) error
//	    Listen(ctx context.Context) error
//	    Listener() (network NetworkProtocol, listener string, tls bool)
//	    Shutdown(ctx context.Context) error
//	    IsRunning() bool
//	    IsGone() bool
//	    OpenConnections() int64
//	}
//
// Implementations: tcp.ServerTcp, udp.ServerUdp, unix.ServerUnix, unixgram.ServerUnixGram
//
// ## Client Interface
//
// The Client interface provides methods for configuring and communicating with a socket server:
//
//	type Client interface {
//	    io.ReadWriteCloser
//	    SetTLS(enable bool, config TLSConfig, serverName string) error
//	    RegisterFuncError(FuncError)
//	    RegisterFuncInfo(FuncInfo)
//	    Connect(ctx context.Context) error
//	    IsConnected() bool
//	    Once(ctx context.Context, request io.Reader, fct Response) error
//	}
//
// Implementations: tcp.ClientTCP, udp.ClientUDP, unix.ClientUnix, unixgram.ClientUnix
//
// ## Context Interface
//
// The Context interface extends context.Context with I/O operations and connection state:
//
//	type Context interface {
//	    context.Context  // Deadline, Done, Err, Value
//	    io.Reader        // Read from connection
//	    io.Writer        // Write to connection
//	    io.Closer        // Close connection
//	    IsConnected() bool
//	    RemoteHost() string
//	    LocalHost() string
//	}
//
// The Context is passed to HandlerFunc and provides all necessary operations for handling a connection.
//
// # Usage Examples
//
// ## TCP Server Example
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/nabbar/golib/socket"
//	    "github.com/nabbar/golib/socket/config"
//	    "github.com/nabbar/golib/socket/server"
//	)
//
//	func main() {
//	    // Create configuration using builder
//	    cfg := config.NewServer().
//	        Network(config.NetworkTCP).
//	        Address(":8080").
//	        HandlerFunc(handleRequest).
//	        Build()
//
//	    // Create server from configuration
//	    srv, err := server.New(nil, cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer srv.Close()
//
//	    // Start listening
//	    ctx := context.Background()
//	    if err := srv.Listen(ctx); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	func handleRequest(ctx socket.Context) {
//	    // Read request
//	    buf := make([]byte, 1024)
//	    n, err := ctx.Read(buf)
//	    if err != nil {
//	        return
//	    }
//
//	    // Process and respond
//	    response := []byte("Response: " + string(buf[:n]))
//	    ctx.Write(response)
//	}
//
// ## TCP Client Example
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/nabbar/golib/socket/client"
//	    "github.com/nabbar/golib/socket/config"
//	)
//
//	func main() {
//	    // Create configuration
//	    cfg := config.NewClient().
//	        Network(config.NetworkTCP).
//	        Address("localhost:8080").
//	        Build()
//
//	    // Create client
//	    cli, err := client.New(cfg, nil)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer cli.Close()
//
//	    // Connect
//	    ctx := context.Background()
//	    if err := cli.Connect(ctx); err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Send request
//	    _, err = cli.Write([]byte("Hello, server!"))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Read response
//	    buf := make([]byte, 1024)
//	    n, err := cli.Read(buf)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    log.Printf("Response: %s", buf[:n])
//	}
//
// ## Unix Socket Server Example (Linux/Darwin)
//
//	cfg := config.NewServer().
//	    Network(config.NetworkUnix).
//	    Address("/tmp/app.sock").
//	    HandlerFunc(handleRequest).
//	    Build()
//
//	srv, err := server.New(nil, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer srv.Close()
//
//	ctx := context.Background()
//	if err := srv.Listen(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// ## UDP Server Example
//
//	cfg := config.NewServer().
//	    Network(config.NetworkUDP).
//	    Address(":9000").
//	    HandlerFunc(handleDatagram).
//	    Build()
//
//	srv, err := server.New(nil, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer srv.Close()
//
//	ctx := context.Background()
//	if err := srv.Listen(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	func handleDatagram(ctx socket.Context) {
//	    buf := make([]byte, 65536)
//	    n, err := ctx.Read(buf)
//	    if err != nil {
//	        return
//	    }
//	    log.Printf("Received datagram from %s: %s", ctx.RemoteHost(), buf[:n])
//	}
//
// # Protocol Selection Guide
//
// ## TCP (NetworkTCP, NetworkTCP4, NetworkTCP6)
//
// Use TCP for:
//   - Reliable, ordered data transmission
//   - Network communication over internet/intranet
//   - Long-lived connections
//   - TLS/SSL encryption requirements
//   - Applications requiring guaranteed delivery
//
// Characteristics:
//   - Connection-oriented
//   - Stream-based (no message boundaries)
//   - Automatic retransmission
//   - Flow control and congestion control
//   - Higher latency than UDP
//
// ## UDP (NetworkUDP, NetworkUDP4, NetworkUDP6)
//
// Use UDP for:
//   - Low-latency, connectionless communication
//   - Real-time data (video, audio, gaming)
//   - Broadcast/multicast scenarios
//   - Metrics and monitoring data
//   - Applications tolerating packet loss
//
// Characteristics:
//   - Connectionless (datagram-oriented)
//   - Message boundaries preserved
//   - No delivery guarantee
//   - No ordering guarantee
//   - Lower overhead than TCP
//
// ## Unix Domain Sockets (NetworkUnix)
//
// Use Unix sockets for:
//   - Inter-process communication on same host
//   - Microservices on same machine
//   - Database connections (PostgreSQL, MySQL)
//   - Local daemon communication
//   - Maximum performance with reliability
//
// Characteristics:
//   - Local only (same host)
//   - Highest throughput, lowest latency
//   - Connection-oriented (like TCP)
//   - File-system based addressing
//   - No network overhead
//
// ## Unix Datagram Sockets (NetworkUnixGram)
//
// Use Unix datagram sockets for:
//   - Local connectionless communication
//   - System logging (syslog)
//   - Event notifications
//   - Metrics collection
//   - Low-latency local messaging
//
// Characteristics:
//   - Local only (same host)
//   - Connectionless (like UDP)
//   - Message boundaries preserved
//   - No network overhead
//   - Higher performance than network UDP
//
// # Platform Support
//
//	┌──────────────┬───────┬───────┬─────────┬──────────┐
//	│  Protocol    │ Linux │ macOS │ Windows │ Other OS │
//	├──────────────┼───────┼───────┼─────────┼──────────┤
//	│  TCP         │   ✅   │   ✅   │    ✅    │    ✅     │
//	│  UDP         │   ✅   │   ✅   │    ✅    │    ✅     │
//	│  Unix        │   ✅   │   ✅   │    ❌    │    ❌     │
//	│  UnixGram    │   ✅   │   ✅   │    ❌    │    ❌     │
//	│  TLS (TCP)   │   ✅   │   ✅   │    ✅    │    ✅     │
//	└──────────────┴───────┴───────┴─────────┴──────────┘
//
// # Configuration
//
// Configuration is managed through the config sub-package using builder patterns:
//
//	import "github.com/nabbar/golib/socket/config"
//
//	// Server configuration
//	srvCfg := config.NewServer().
//	    Network(config.NetworkTCP).
//	    Address(":8080").
//	    HandlerFunc(myHandler).
//	    BufferSize(32 * 1024).
//	    Delimiter('\n').
//	    Build()
//
//	// Client configuration
//	cliCfg := config.NewClient().
//	    Network(config.NetworkTCP).
//	    Address("localhost:8080").
//	    BufferSize(32 * 1024).
//	    Build()
//
// See github.com/nabbar/golib/socket/config for complete configuration options.
//
// # Connection State Tracking
//
// The package provides detailed connection state tracking through the ConnState type
// and FuncInfo callback:
//
//	type ConnState uint8
//	const (
//	    ConnectionDial       // Client dialing
//	    ConnectionNew        // New connection established
//	    ConnectionRead       // Reading data
//	    ConnectionCloseRead  // Closing read side
//	    ConnectionHandler    // Handler executing
//	    ConnectionWrite      // Writing data
//	    ConnectionCloseWrite // Closing write side
//	    ConnectionClose      // Closing connection
//	)
//
// Register a callback to track state changes:
//
//	srv.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
//	    log.Printf("Connection %s -> %s: %s", remote, local, state)
//	})
//
// # Error Handling
//
// The package provides comprehensive error handling through:
//
//  1. Return Values: All operations return Go standard errors
//  2. Error Callbacks: Register FuncError for async error notification
//  3. Error Filtering: ErrorFilter() removes expected errors (closed connections)
//  4. Context Integration: Errors propagate through context cancellation
//
// Example error handling:
//
//	srv.RegisterFuncError(func(errs ...error) {
//	    for _, err := range errs {
//	        if err := socket.ErrorFilter(err); err != nil {
//	            log.Printf("Socket error: %v", err)
//	        }
//	    }
//	})
//
// # Thread Safety
//
// All interfaces and implementations are designed for concurrent use:
//
//   - Server: Listen() blocks, all other methods are thread-safe
//   - Client: Connect() can be called concurrently, Read/Write should be serialized per connection
//   - Context: All methods are safe for concurrent calls except Read/Write which follow io.Reader/Writer contracts
//
// Thread-safe operations:
//   - Server.IsRunning(), Server.OpenConnections()
//   - Client.IsConnected()
//   - Context.IsConnected(), Context.RemoteHost(), Context.LocalHost()
//   - All callback registrations
//
// # Performance Characteristics
//
//	┌────────────────────┬────────────┬──────────┬──────────────┐
//	│  Operation         │ TCP/Unix   │   UDP    │  UnixGram    │
//	├────────────────────┼────────────┼──────────┼──────────────┤
//	│  Connection Setup  │  ~1-5 ms   │  ~0 ms   │   ~0 ms      │
//	│  Read Latency      │  ~100 µs   │  ~50 µs  │   ~10 µs     │
//	│  Write Latency     │  ~100 µs   │  ~50 µs  │   ~10 µs     │
//	│  Throughput        │  GB/s      │  GB/s    │   GB/s       │
//	│  Memory/Connection │  ~32 KB    │  ~16 KB  │   ~16 KB     │
//	└────────────────────┴────────────┴──────────┴──────────────┘
//
// Performance tips:
//   - Use appropriate buffer sizes (default: 32KB)
//   - Enable TCP_NODELAY for low-latency applications
//   - Use Unix sockets for local communication
//   - Pool connections for high-throughput scenarios
//   - Use UDP/UnixGram for low-latency datagrams
//
// # Best Practices
//
// ## 1. Resource Management
//
//	srv, err := server.New(nil, cfg)
//	if err != nil {
//	    return err
//	}
//	defer srv.Close()  // Always close resources
//
// ## 2. Context Usage
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	err := srv.Listen(ctx)
//
// ## 3. Error Handling
//
//	srv.RegisterFuncError(func(errs ...error) {
//	    for _, err := range errs {
//	        if err := socket.ErrorFilter(err); err != nil {
//	            log.Printf("Error: %v", err)
//	        }
//	    }
//	})
//
// ## 4. Graceful Shutdown
//
//	sigCh := make(chan os.Signal, 1)
//	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
//	<-sigCh
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	srv.Shutdown(ctx)
//
// ## 5. Connection Monitoring
//
//	srv.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
//	    log.Printf("%s: %s -> %s", state, remote, local)
//	})
//
// # Limitations
//
//  1. Unix Sockets: Only available on Linux and Darwin (macOS)
//  2. TLS Support: Only available for TCP-based protocols
//  3. Multicast: Not directly supported (use raw sockets if needed)
//  4. Protocol Mixing: Each socket handles only one protocol type
//  5. Message Boundaries: TCP/Unix are stream-based (use delimiters or length prefixes)
//
// # Related Packages
//
//   - github.com/nabbar/golib/socket/config: Configuration builders and validators
//   - github.com/nabbar/golib/socket/client: Client factory and implementations
//   - github.com/nabbar/golib/socket/server: Server factory and implementations
//   - github.com/nabbar/golib/network/protocol: Protocol constants and utilities
//   - github.com/nabbar/golib/certificates: TLS configuration and certificate management
//   - github.com/nabbar/golib/ioutils/aggregator: Thread-safe write aggregation for socket logging
//   - github.com/nabbar/golib/ioutils/delim: Delimiter-based reading for message framing
//
// # See Also
//
// For detailed protocol-specific documentation:
//   - TCP: github.com/nabbar/golib/socket/client/tcp, github.com/nabbar/golib/socket/server/tcp
//   - UDP: github.com/nabbar/golib/socket/client/udp, github.com/nabbar/golib/socket/server/udp
//   - Unix: github.com/nabbar/golib/socket/client/unix, github.com/nabbar/golib/socket/server/unix
//   - UnixGram: github.com/nabbar/golib/socket/client/unixgram, github.com/nabbar/golib/socket/server/unixgram
//
// For configuration:
//   - github.com/nabbar/golib/socket/config
//
// For examples:
//   - See example_test.go in this package and sub-packages
package socket
