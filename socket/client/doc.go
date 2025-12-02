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

// Package client provides a unified, platform-aware factory for creating socket clients
// across different network protocols. It serves as a convenient entry point that
// automatically selects the appropriate protocol-specific implementation based on the
// network type specified in the configuration.
//
// # Overview
//
// This package acts as a factory that abstracts protocol-specific client implementations
// and provides a consistent API through the github.com/nabbar/golib/socket.Client interface.
// The factory automatically delegates to the appropriate sub-package based on the protocol:
//
//   - TCP, TCP4, TCP6: github.com/nabbar/golib/socket/client/tcp
//   - UDP, UDP4, UDP6: github.com/nabbar/golib/socket/client/udp
//   - Unix (Linux/Darwin): github.com/nabbar/golib/socket/client/unix
//   - UnixGram (Linux/Darwin): github.com/nabbar/golib/socket/client/unixgram
//
// # Architecture
//
// ## Factory Pattern
//
// The client package implements the Factory Method pattern, providing a single entry
// point (New) that instantiates the appropriate client implementation:
//
//	┌─────────────────────────────────────────────────────┐
//	│                     client.New()                    │
//	│                   (Factory Function)                │
//	└───────────────────────────┬─────────────────────────┘
//	                            │
//	        ┌─────────────┬─────┴───────┬───────────┐
//	        │             │             │           │
//	        ▼             ▼             ▼           ▼
//	 ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
//	 │   TCP    │  │   UDP    │  │   Unix   │  │ UnixGram │
//	 │  Client  │  │  Client  │  │  Client  │  │  Client  │
//	 └──────────┘  └──────────┘  └──────────┘  └──────────┘
//	      │             │             │             │
//	      └─────────────┴──────┬──────┴─────────────┘
//	                           │
//	                 ┌─────────▼─────────┐
//	                 │  socket.Client    │
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
//	│                    client Package                         │
//	├───────────────────────────────────────────────────────────┤
//	│                                                           │
//	│  ┌─────────────────────────────────────────────────┐      │
//	│  │             New(cfg, def)                       │      │
//	│  │                                                 │      │
//	│  │  1. Validate cfg.Network                        │      │
//	│  │  2. Switch on protocol type                     │      │
//	│  │  3. Delegate to appropriate package             │      │
//	│  │  4. Return socket.Client implementation         │      │
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
//  3. Type Safety: Configuration-based client creation with validation
//  4. Consistent API: All clients implement socket.Client interface
//  5. Zero Overhead: Factory only adds a single switch statement
//
// # Key Features
//
//   - Unified API: Single New() function for all protocols
//   - Platform-aware: Automatic Unix socket support detection
//   - Type-safe configuration: Uses config.Client struct
//   - Protocol validation: Returns error for unsupported protocols
//   - TLS support: Transparent TLS configuration for TCP clients
//   - Zero dependencies: Only delegates to sub-packages
//   - Minimal overhead: Direct delegation without wrapping
//   - Panic recovery: Automatic recovery with detailed logging
//
// # Usage Examples
//
// ## TCP Client
//
//	import (
//	    "github.com/nabbar/golib/network/protocol"
//	    "github.com/nabbar/golib/socket/client"
//	    "github.com/nabbar/golib/socket/config"
//	)
//
//	func main() {
//	    cfg := config.Client{
//	        Network: protocol.NetworkTCP,
//	        Address: "localhost:8080",
//	    }
//
//	    cli, err := client.New(cfg, nil)
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer cli.Close()
//
//	    // Use client...
//	}
//
// ## TCP Client with TLS
//
//	import (
//	    "github.com/nabbar/golib/certificates"
//	    "github.com/nabbar/golib/socket/config"
//	)
//
//	// Create TLS config
//	tlsCfg := certificates.TLSConfig{
//	    // Configure TLS...
//	}
//
//	cfg := config.Client{
//	    Network: protocol.NetworkTCP,
//	    Address: "secure.example.com:443",
//	    TLS: config.ClientTLS{
//	        Enabled:    true,
//	        ServerName: "secure.example.com",
//	    },
//	}
//
//	cli, err := client.New(cfg, tlsCfg)
//	if err != nil {
//	    panic(err)
//	}
//	defer cli.Close()
//
// ## UDP Client
//
//	cfg := config.Client{
//	    Network: protocol.NetworkUDP,
//	    Address: "localhost:9000",
//	}
//
//	cli, err := client.New(cfg, nil)
//	if err != nil {
//	    panic(err)
//	}
//	defer cli.Close()
//
// ## Unix Socket Client (Linux/Darwin only)
//
//	cfg := config.Client{
//	    Network: protocol.NetworkUnix,
//	    Address: "/tmp/app.sock",
//	}
//
//	cli, err := client.New(cfg, nil)
//	if err != nil {
//	    panic(err)
//	}
//	defer cli.Close()
//
// # Protocol Selection
//
// The New() function selects the appropriate client implementation based on
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
// The New() function accepts a config.Client struct containing:
//
//   - Network: Protocol type (NetworkTCP, NetworkUDP, NetworkUnix, etc.)
//   - Address: Protocol-specific address string
//   - TCP/UDP: "host:port" format
//   - Unix: filesystem path
//   - TLS: TLS configuration (TCP only)
//   - Enabled: Enable TLS
//   - Config: TLS certificate configuration
//   - ServerName: Server name for TLS verification
//
// See github.com/nabbar/golib/socket/config.Client for complete configuration options.
//
// # Error Handling
//
// The New() function returns an error if:
//
//  1. Configuration validation fails
//     Example: Empty address or invalid network value
//
//  2. Protocol is not supported on the current platform
//     Example: NetworkUnix on Windows returns ErrInvalidProtocol
//
//  3. Protocol value is invalid or unrecognized
//     Example: Undefined protocol constant returns ErrInvalidProtocol
//
//  4. Sub-package constructor fails
//     Example: tcp.New() fails due to invalid address format
//
//  5. TLS configuration fails (TCP only)
//     Example: Invalid TLS certificate or configuration
//
// All errors are propagated from the underlying protocol implementation or configuration
// validation. The factory uses panic recovery to catch and log unexpected errors.
//
// # Platform Support
//
// ## Linux (//go:build linux)
//
//   - Supported: TCP, TCP4, TCP6, UDP, UDP4, UDP6, Unix, UnixGram
//   - Unix sockets: Full support with abstract socket namespace
//   - Special features: Abstract sockets (addresses starting with @)
//
// ## Darwin/macOS (//go:build darwin)
//
//   - Supported: TCP, TCP4, TCP6, UDP, UDP4, UDP6, Unix, UnixGram
//   - Unix sockets: Full support with filesystem sockets
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
// goroutines. Each call creates a new, independent client instance with its own
// state and resources.
//
// The returned socket.Client implementations are also thread-safe for concurrent
// method calls, though each connection should typically be used by a single goroutine
// for reading and another for writing.
//
// # Performance Considerations
//
// ## Factory Overhead
//
// The client factory adds minimal overhead:
//
//   - One configuration validation
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
//  3. No Protocol Mixing: Each client handles only one protocol type
//  4. No Auto-Detection: Protocol must be explicitly specified in configuration
//  5. No Fallback: If a protocol is unsupported, an error is returned (no fallback)
//  6. TLS for TCP Only: TLS configuration only applies to TCP-based protocols
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
//	// Real-time datagrams (metrics, logging)
//	cfg.Network = protocol.NetworkUDP   // Fastest, connectionless
//
// ## 2. Handle Platform-Specific Errors
//
//	cli, err := client.New(cfg, nil)
//	if err == config.ErrInvalidProtocol {
//	    // Protocol not supported on this platform
//	    // Fall back to TCP
//	    cfg.Network = protocol.NetworkTCP
//	    cli, err = client.New(cfg, nil)
//	}
//
// ## 3. Configure Protocol-Specific Options
//
//	// TCP-specific: Enable TLS
//	if cfg.Network.IsTCP() {
//	    cfg.TLS.Enabled = true
//	    cfg.TLS.ServerName = "example.com"
//	}
//
// ## 4. Resource Management
//
//	cli, err := client.New(cfg, nil)
//	if err != nil {
//	    return err
//	}
//	defer cli.Close()  // Always clean up
//
//	// Use client...
//
// ## 5. Error Handling
//
//	cli, err := client.New(cfg, nil)
//	if err != nil {
//	    if err == config.ErrInvalidProtocol {
//	        // Handle unsupported protocol
//	    } else {
//	        // Handle other errors
//	    }
//	    return err
//	}
//
// # Comparison with Direct Protocol Packages
//
// ## Using Factory (client.New)
//
// Advantages:
//   - Single import for all protocols
//   - Consistent API across protocols
//   - Configuration-based client selection
//   - Easier to switch protocols
//   - Centralized error handling
//
// Disadvantages:
//   - Slight indirection overhead
//   - Less explicit about protocol used
//
// ## Using Direct Protocol Packages
//
// Advantages:
//   - More explicit (tcp.New vs client.New)
//   - Direct access to protocol-specific features
//   - Slightly better IDE autocomplete
//
// Disadvantages:
//   - Multiple imports needed
//   - More code changes when switching protocols
//   - Repetitive error handling
//
// Recommendation: Use client.New for most cases. Use protocol-specific packages
// when you need direct access to protocol-specific features or when protocol
// selection is static and won't change.
//
// # Related Packages
//
//   - github.com/nabbar/golib/socket: Base interfaces and types
//   - github.com/nabbar/golib/socket/config: Client configuration structures
//   - github.com/nabbar/golib/socket/client/tcp: TCP client implementation
//   - github.com/nabbar/golib/socket/client/udp: UDP client implementation
//   - github.com/nabbar/golib/socket/client/unix: Unix socket client (Linux/Darwin)
//   - github.com/nabbar/golib/socket/client/unixgram: Unix datagram client (Linux/Darwin)
//   - github.com/nabbar/golib/network/protocol: Protocol constants and utilities
//   - github.com/nabbar/golib/certificates: TLS configuration and utilities
//
// # See Also
//
// For detailed documentation on individual protocol implementations, refer to:
//   - TCP clients: github.com/nabbar/golib/socket/client/tcp/doc.go
//   - UDP clients: github.com/nabbar/golib/socket/client/udp/doc.go
//   - Unix clients: github.com/nabbar/golib/socket/client/unix/doc.go
//   - UnixGram clients: github.com/nabbar/golib/socket/client/unixgram/doc.go
//
// For examples, see example_test.go in this package.
package client
