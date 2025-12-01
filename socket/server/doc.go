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
 */

// Package server provides a unified, platform-aware factory for creating socket servers
// across different network protocols. It serves as a convenient entry point that
// automatically selects the appropriate protocol-specific implementation based on the
// network type specified in the configuration.
//
// # Overview
//
// This package acts as a factory that abstracts protocol-specific server implementations
// and provides a consistent API through the github.com/nabbar/golib/socket.Server interface.
// The factory automatically delegates to the appropriate sub-package based on the protocol:
//
//   - TCP, TCP4, TCP6: github.com/nabbar/golib/socket/server/tcp
//   - UDP, UDP4, UDP6: github.com/nabbar/golib/socket/server/udp
//   - Unix (Linux/Darwin): github.com/nabbar/golib/socket/server/unix
//   - UnixGram (Linux/Darwin): github.com/nabbar/golib/socket/server/unixgram
//
// # Architecture
//
// ## Factory Pattern
//
// The server package implements the Factory Method pattern, providing a single entry
// point (New) that instantiates the appropriate server implementation:
//
//	┌─────────────────────────────────────────────────────┐
//	│                     server.New()                    │
//	│                   (Factory Function)                │
//	└───────────────────────────┬─────────────────────────┘
//	                            │
//	        ┌─────────────┬─────┴───────┬───────────┐
//	        │             │             │           │
//	        ▼             ▼             ▼           ▼
//	 ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
//	 │   TCP    │  │   UDP    │  │   Unix   │  │ UnixGram │
//	 │  Server  │  │  Server  │  │  Server  │  │  Server  │
//	 └──────────┘  └──────────┘  └──────────┘  └──────────┘
//	      │             │             │             │
//	      └─────────────┴──────┬──────┴─────────────┘
//	                           │
//	                 ┌─────────▼─────────┐
//	                 │  socket.Server    │
//	                 │    (Interface)    │
//	                 └───────────────────┘
//
// ## Platform-Specific Implementations
//
// The package uses build constraints to provide platform-specific implementations:
//
//   - interface_linux.go (//go:build linux): Full protocol support including Unix sockets
//   - interface_darwin.go (//go:build darwin): Full protocol support including Unix sockets
//   - interface_other.go (//go:build !linux && !darwin): TCP and UDP only
//
// This ensures that Unix domain sockets are only available on platforms that support them.
//
// ## Component Diagram
//
//	┌───────────────────────────────────────────────────────────┐
//	│                    server Package                         │
//	├───────────────────────────────────────────────────────────┤
//	│                                                           │
//	│  ┌─────────────────────────────────────────────────┐      │
//	│  │          New(upd, handler, cfg)                 │      │
//	│  │                                                 │      │
//	│  │  1. Validate config.Network                     │      │
//	│  │  2. Switch on protocol type                     │      │
//	│  │  3. Delegate to appropriate package             │      │
//	│  │  4. Return socket.Server implementation         │      │
//	│  └─────────────────────────┬───────────────────────┘      │
//	│                            │                              │
//	│                            ▼                              │
//	│     ┌───────────────┬──────┴────┬──────────────┐          │
//	│     │               │           │              │          │
//	│     ▼               ▼           ▼              ▼          │
//	│  tcp.New()       udp.New()  unix.New() unixgram.New()     │
//	│                                                           │
//	└───────────────────────────────────────────────────────────┘
//
// # Design Philosophy
//
//  1. Simplicity First: Single entry point (New) for all protocol types
//  2. Platform Awareness: Automatic protocol availability based on OS
//  3. Type Safety: Configuration-based server creation with validation
//  4. Consistent API: All servers implement socket.Server interface
//  5. Zero Overhead: Factory only adds a single switch statement
//
// # Key Features
//
//   - Unified API: Single New() function for all protocols
//   - Platform-aware: Automatic Unix socket support detection
//   - Type-safe configuration: Uses config.Server struct
//   - Protocol validation: Returns error for unsupported protocols
//   - Zero dependencies: Only delegates to sub-packages
//   - Minimal overhead: Direct delegation without wrapping
//
// # Usage Examples
//
// ## TCP Server
//
//	import (
//	    "context"
//	    "github.com/nabbar/golib/network/protocol"
//	    "github.com/nabbar/golib/socket"
//	    "github.com/nabbar/golib/socket/config"
//	    "github.com/nabbar/golib/socket/server"
//	)
//
//	func main() {
//	    handler := func(c socket.Context) {
//	        defer c.Close()
//	        // Handle connection...
//	    }
//
//	    cfg := config.Server{
//	        Network: protocol.NetworkTCP,
//	        Address: ":8080",
//	    }
//
//	    srv, err := server.New(nil, handler, cfg)
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    if err := srv.Listen(context.Background()); err != nil {
//	        panic(err)
//	    }
//	}
//
// ## UDP Server
//
//	cfg := config.Server{
//	    Network: protocol.NetworkUDP,
//	    Address: ":9000",
//	}
//
//	srv, err := server.New(nil, handler, cfg)
//	if err != nil {
//	    panic(err)
//	}
//
//	if err := srv.Listen(context.Background()); err != nil {
//	    panic(err)
//	}
//
// ## Unix Socket Server (Linux/Darwin only)
//
//	import "github.com/nabbar/golib/file/perm"
//
//	cfg := config.Server{
//	    Network:   protocol.NetworkUnix,
//	    Address:   "/tmp/app.sock",
//	    PermFile:  perm.Perm(0660),
//	    GroupPerm: -1,
//	}
//
//	srv, err := server.New(nil, handler, cfg)
//	if err != nil {
//	    panic(err)
//	}
//
//	if err := srv.Listen(context.Background()); err != nil {
//	    panic(err)
//	}
//
// # Protocol Selection
//
// The New() function selects the appropriate server implementation based on
// cfg.Network value:
//
//	┌─────────────────────┬──────────────────┬─────────────────────┐
//	│  Protocol Value     │  Platform        │  Delegates To       │
//	├─────────────────────┼──────────────────┼─────────────────────┤
//	│  NetworkTCP         │  All             │  tcp.New()          │
//	│  NetworkTCP4        │  All             │  tcp.New()          │
//	│  NetworkTCP6        │  All             │  tcp.New()          │
//	│  NetworkUDP         │  All             │  udp.New()          │
//	│  NetworkUDP4        │  All             │  udp.New()          │
//	│  NetworkUDP6        │  All             │  udp.New()          │
//	│  NetworkUnix        │  Linux/Darwin    │  unix.New()         │
//	│  NetworkUnixGram    │  Linux/Darwin    │  unixgram.New()     │
//	│  Other values       │  All             │  ErrInvalidProtocol │
//	└─────────────────────┴──────────────────┴─────────────────────┘
//
// # Configuration
//
// The New() function accepts a config.Server struct containing:
//
//   - Network: Protocol type (NetworkTCP, NetworkUDP, NetworkUnix, etc.)
//   - Address: Protocol-specific address string
//   - TCP/UDP: "[host]:port" format
//   - Unix: filesystem path
//   - PermFile: File permissions for Unix sockets (ignored for TCP/UDP)
//   - GroupPerm: Group ID for Unix socket ownership (ignored for TCP/UDP)
//   - ConIdleTimeout: Idle timeout for connections (applies to all protocols)
//   - TLS: TLS configuration (TCP only)
//
// See github.com/nabbar/golib/socket/config.Server for complete configuration options.
//
// # Error Handling
//
// The New() function returns an error if:
//
//  1. Protocol is not supported on the current platform
//     Example: NetworkUnix on Windows returns ErrInvalidProtocol
//
//  2. Protocol value is invalid or unrecognized
//     Example: Undefined protocol constant returns ErrInvalidProtocol
//
//  3. Sub-package constructor fails
//     Example: tcp.New() fails due to invalid configuration
//
// All errors are propagated from the underlying protocol implementation.
//
// # Platform Support
//
// ## Linux (//go:build linux)
//
//   - Supported: TCP, TCP4, TCP6, UDP, UDP4, UDP6, Unix, UnixGram
//   - Unix sockets: Full support with file permissions and group ownership
//   - Special features: SCM_CREDENTIALS for process authentication
//
// ## Darwin/macOS (//go:build darwin)
//
//   - Supported: TCP, TCP4, TCP6, UDP, UDP4, UDP6, Unix, UnixGram
//   - Unix sockets: Full support with file permissions and group ownership
//   - Special features: Standard Unix socket features
//
// ## Other Platforms (//go:build !linux && !darwin)
//
//   - Supported: TCP, TCP4, TCP6, UDP, UDP4, UDP6
//   - Unix sockets: Not supported (returns ErrInvalidProtocol)
//   - Note: Includes Windows, BSD, Solaris, etc.
//
// # Thread Safety
//
// The New() function is thread-safe and can be called concurrently from multiple
// goroutines. Each call creates a new, independent server instance with its own
// state and resources.
//
// The returned socket.Server implementations are also thread-safe for concurrent
// method calls, but each connection is handled in its own goroutine.
//
// # Performance Considerations
//
// ## Factory Overhead
//
// The server factory adds minimal overhead:
//
//   - One switch statement on protocol type
//   - One function call to the appropriate constructor
//   - No additional memory allocations
//   - No runtime reflection
//
// Typical overhead: <1 microsecond per New() call.
//
// ## Protocol Performance Characteristics
//
//	┌──────────────┬──────────────┬─────────────┬──────────────────┐
//	│  Protocol    │  Throughput  │  Latency    │  Best Use Case   │
//	├──────────────┼──────────────┼─────────────┼──────────────────┤
//	│  TCP         │  High        │  Low        │  Network IPC     │
//	│  UDP         │  Very High   │  Very Low   │  Datagrams       │
//	│  Unix        │  Highest     │  Lowest     │  Local IPC       │
//	│  UnixGram    │  Highest     │  Lowest     │  Local datagrams │
//	└──────────────┴──────────────┴─────────────┴──────────────────┘
//
// # Limitations
//
//  1. Factory Only: This package provides no direct functionality, only delegation
//  2. Platform-Specific: Unix socket support depends on OS (Linux/Darwin only)
//  3. No Protocol Mixing: Each server handles only one protocol type
//  4. No Auto-Detection: Protocol must be explicitly specified in configuration
//  5. No Fallback: If a protocol is unsupported, an error is returned (no fallback)
//
// # Best Practices
//
// ## 1. Use Appropriate Protocol for Use Case
//
//	// Local IPC between processes on same host
//	cfg.Network = protocol.NetworkUnix  // Lowest latency
//
//	// Network communication
//	cfg.Network = protocol.NetworkTCP   // Reliable, ordered
//
//	// Real-time datagrams (logging, metrics)
//	cfg.Network = protocol.NetworkUDP   // Fastest, connectionless
//
// ## 2. Handle Platform-Specific Errors
//
//	srv, err := server.New(nil, handler, cfg)
//	if err == config.ErrInvalidProtocol {
//	    // Protocol not supported on this platform
//	    // Fall back to TCP
//	    cfg.Network = protocol.NetworkTCP
//	    srv, err = server.New(nil, handler, cfg)
//	}
//
// ## 3. Configure Protocol-Specific Options
//
//	// TCP-specific: Enable TLS
//	if cfg.Network.IsTCP() {
//	    cfg.TLS.Enable = true
//	}
//
//	// Unix-specific: Set file permissions
//	if cfg.Network == protocol.NetworkUnix {
//	    cfg.PermFile = perm.Perm(0660)
//	}
//
// ## 4. Resource Management
//
//	srv, err := server.New(nil, handler, cfg)
//	if err != nil {
//	    return err
//	}
//	defer srv.Close()  // Always clean up
//
//	if err := srv.Listen(ctx); err != nil {
//	    return err
//	}
//
// ## 5. Graceful Shutdown
//
//	srv, err := server.New(nil, handler, cfg)
//	if err != nil {
//	    return err
//	}
//
//	// Handle signals
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
//
//	go func() {
//	    <-sigChan
//	    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	    defer cancel()
//	    srv.Shutdown(ctx)
//	}()
//
//	srv.Listen(context.Background())
//
// # Comparison with Direct Protocol Packages
//
// ## Using Factory (server.New)
//
// Advantages:
//   - Single import for all protocols
//   - Consistent API across protocols
//   - Configuration-based server selection
//   - Easier to switch protocols
//
// Disadvantages:
//   - Slight indirection overhead
//   - Less explicit about protocol used
//
// ## Using Direct Protocol Packages
//
// Advantages:
//   - More explicit (tcp.New vs server.New)
//   - Direct access to protocol-specific features
//   - Slightly better IDE autocomplete
//
// Disadvantages:
//   - Multiple imports needed
//   - More code changes when switching protocols
//   - Repetitive error handling
//
// Recommendation: Use server.New for most cases. Use protocol-specific packages
// when you need direct access to protocol-specific features or when protocol
// selection is static and won't change.
//
// # Related Packages
//
//   - github.com/nabbar/golib/socket: Base interfaces and types
//   - github.com/nabbar/golib/socket/config: Server configuration structures
//   - github.com/nabbar/golib/socket/server/tcp: TCP server implementation
//   - github.com/nabbar/golib/socket/server/udp: UDP server implementation
//   - github.com/nabbar/golib/socket/server/unix: Unix socket server (Linux/Darwin)
//   - github.com/nabbar/golib/socket/server/unixgram: Unix datagram server (Linux/Darwin)
//   - github.com/nabbar/golib/network/protocol: Protocol constants and utilities
//
// # See Also
//
// For detailed documentation on individual protocol implementations, refer to:
//   - TCP servers: github.com/nabbar/golib/socket/server/tcp/doc.go
//   - UDP servers: github.com/nabbar/golib/socket/server/udp/doc.go
//   - Unix servers: github.com/nabbar/golib/socket/server/unix/doc.go
//   - UnixGram servers: github.com/nabbar/golib/socket/server/unixgram/doc.go
//
// For examples, see example_test.go in this package.
package server
