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

// Package config provides declarative configuration structures for socket clients and servers.
//
// # Overview
//
// This package implements a configuration-first approach to socket programming,
// allowing you to define connection parameters before instantiation. It's designed
// for scenarios where socket settings are loaded from external sources (configuration
// files, environment variables, databases) and need validation before use.
//
// The package supports all common socket types through a unified interface:
//   - TCP: Reliable, connection-oriented network sockets
//   - UDP: Fast, connectionless network sockets
//   - Unix: Connection-oriented inter-process communication via filesystem sockets
//   - Unixgram: Connectionless inter-process communication via filesystem sockets
//
// # Design Philosophy
//
// The package follows these design principles:
//
// 1. Configuration Separation: Socket parameters are defined independently from
// socket instances, enabling validation before connection attempts. This separation
// allows for early detection of configuration errors without allocating network
// resources.
//
// 2. Declarative API: Configuration uses simple struct field assignments rather
// than complex builder patterns or method chaining. This makes configurations
// easy to serialize, deserialize, and manipulate programmatically.
//
// 3. Fail-Fast Validation: The Validate() methods catch configuration errors
// early, before network operations are attempted. Validation includes DNS resolution,
// address format checking, and platform compatibility verification.
//
// 4. Platform Awareness: Built-in checks for platform-specific limitations
// (e.g., Unix sockets on Windows). The package provides clear error messages
// when attempting to use unsupported features on incompatible platforms.
//
// 5. Security by Default: TLS/SSL configuration is validated to ensure proper
// certificate handling and protocol restrictions. The package enforces that TLS
// is only used with appropriate protocols and that all required parameters are present.
//
// # Key Features
//
// Protocol Flexibility: Single configuration API supports TCP, UDP, Unix, and
// Unixgram sockets through the NetworkProtocol interface. This unification
// simplifies code that needs to support multiple transport types.
//
// TLS Support: First-class TLS/SSL configuration for TCP connections with
// certificate validation and server name verification. Supports both client
// and server TLS with customizable cipher suites and protocol versions.
//
// Unix Socket Permissions: Fine-grained control over Unix socket file permissions
// and group ownership for security-sensitive IPC. Allows specifying both file
// mode bits and group ownership to implement proper access control.
//
// Connection Management: Configurable idle timeouts for connection-oriented
// protocols to prevent resource exhaustion from stale connections.
//
// Validation Guarantees: Comprehensive validation catches address format errors,
// protocol mismatches, platform incompatibilities, and TLS configuration issues
// before any network operations are attempted.
//
// # Architecture
//
// The package consists of two main configuration structures:
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                     socket/config                            │
//	├─────────────────────────────────────────────────────────────┤
//	│                                                               │
//	│  ┌──────────────┐              ┌──────────────┐            │
//	│  │   Client     │              │   Server     │            │
//	│  ├──────────────┤              ├──────────────┤            │
//	│  │ Network      │              │ Network      │            │
//	│  │ Address      │              │ Address      │            │
//	│  │ TLS          │              │ PermFile     │            │
//	│  │              │              │ GroupPerm    │            │
//	│  │ + Validate() │              │ ConIdleTimeout│           │
//	│  └──────────────┘              │ TLS          │            │
//	│                                 │              │            │
//	│                                 │ + Validate() │            │
//	│                                 │ + DefaultTLS()│           │
//	│                                 │ + GetTLS()   │            │
//	│                                 └──────────────┘            │
//	│                                                               │
//	│  ┌────────────────────────────────────────────┐            │
//	│  │              Error Types                    │            │
//	│  ├────────────────────────────────────────────┤            │
//	│  │ ErrInvalidProtocol                         │            │
//	│  │ ErrInvalidTLSConfig                        │            │
//	│  │ ErrInvalidGroup                            │            │
//	│  └────────────────────────────────────────────┘            │
//	└─────────────────────────────────────────────────────────────┘
//	           │                              │
//	           │                              │
//	           ▼                              ▼
//	┌────────────────────┐         ┌────────────────────┐
//	│ socket/client      │         │ socket/server      │
//	│ implementations    │         │ implementations    │
//	└────────────────────┘         └────────────────────┘
//
// Configuration flows from external sources → config structs → validation →
// socket implementations. This separation ensures that invalid configurations
// are caught before any network resources are allocated.
//
// # Key Features
//
// Protocol Flexibility: Single configuration API supports TCP, UDP, Unix, and
// Unixgram sockets through the NetworkProtocol interface.
//
// TLS Support: First-class TLS/SSL configuration for TCP connections with
// certificate validation and server name verification.
//
// Unix Socket Permissions: Fine-grained control over Unix socket file permissions
// and group ownership for security-sensitive IPC.
//
// Connection Management: Configurable idle timeouts for connection-oriented
// protocols to prevent resource exhaustion.
//
// Validation Guarantees: Comprehensive validation catches address format errors,
// protocol mismatches, platform incompatibilities, and TLS configuration issues.
//
// # Limitations
//
// 1. Platform-Specific Features: Unix domain sockets are not available on Windows.
// Configuration validation will return ErrInvalidProtocol on unsupported platforms.
//
// 2. TLS Protocol Restrictions: TLS/SSL is only supported for TCP-based protocols.
// UDP and Unix sockets cannot use TLS through this package.
//
// 3. No Dynamic Reconfiguration: Once a socket is created from a configuration,
// changing the configuration does not affect the existing socket. You must create
// a new socket instance.
//
// 4. Group Permission Limits: Unix socket group IDs are limited to MaxGID (32767)
// for portability across Unix-like systems.
//
// 5. No IPv6 Scope IDs: While IPv6 addresses are supported, zone/scope IDs in
// link-local addresses may have platform-specific behavior.
//
// # Performance Considerations
//
// The configuration structures are designed for infrequent creation (e.g., application
// startup) rather than high-frequency operations. Validation involves DNS resolution
// and filesystem checks, which may block.
//
// Performance characteristics:
//
//	Structure Creation:  < 100µs (simple struct initialization)
//	TCP Validation:      < 1ms average (includes DNS resolution)
//	UDP Validation:      < 1ms average (includes DNS resolution)
//	Unix Validation:     < 100µs (no DNS, just address checks)
//	Structure Copy:      < 10µs (small memory footprint)
//
// For hot-path operations:
//   - Cache validated configurations rather than re-validating on each use
//   - Create socket instances once at startup and reuse them throughout the application lifecycle
//   - Avoid calling Validate() in request handling loops or performance-critical paths
//   - Validation may block for DNS resolution, so perform it asynchronously if needed
//
// The structs are small (< 100 bytes) and safe to copy by value. However, they contain
// interface fields (TLS.Config) that may reference larger objects. Copying the struct
// creates a shallow copy that shares these referenced objects.
//
// # Use Cases
//
// Configuration File Loading:
//
// Load socket parameters from YAML, JSON, or TOML files and validate them
// before starting services. This catches configuration errors at startup
// rather than during operation.
//
// Environment-Based Configuration:
//
// Read socket settings from environment variables (12-factor app pattern) and
// create properly validated socket instances for different deployment environments.
//
// Dynamic Service Discovery:
//
// Receive socket addresses from service discovery systems (Consul, etcd) and
// validate them before establishing connections.
//
// Multi-Protocol Services:
//
// Configure services to listen on multiple socket types simultaneously (TCP for
// network access, Unix sockets for local IPC) using a unified configuration format.
//
// Secure Service Communication:
//
// Define TLS/SSL parameters for encrypted client-server communication with proper
// certificate validation and server name verification.
//
// # Error Handling
//
// All validation methods return typed errors that can be checked:
//
//	if err := cfg.Validate(); err != nil {
//	    switch {
//	    case errors.Is(err, config.ErrInvalidProtocol):
//	        // Handle unsupported protocol
//	    case errors.Is(err, config.ErrInvalidTLSConfig):
//	        // Handle TLS configuration error
//	    case errors.Is(err, config.ErrInvalidGroup):
//	        // Handle group permission error
//	    default:
//	        // Handle address resolution or other errors
//	    }
//	}
//
// Address resolution errors come from the standard net package and preserve
// the original error information for debugging.
//
// # Thread Safety
//
// Configuration structures are safe to read from multiple goroutines after
// creation, but must not be modified concurrently. If you need to share
// configurations across goroutines, either:
//
//   - Create separate copies for each goroutine (structs are small)
//   - Protect shared instances with sync.RWMutex
//   - Use immutable patterns (create new configs instead of modifying)
//
// The Server.DefaultTLS() and Server.GetTLS() methods are not thread-safe and
// should not be called concurrently with other operations on the same instance.
//
// # Examples
//
// See the package examples for detailed usage patterns:
//   - Basic TCP client and server configuration
//   - Unix socket configuration with permissions
//   - TLS-enabled secure connections
//   - Configuration loading from external sources
//   - Error handling and validation
package config
