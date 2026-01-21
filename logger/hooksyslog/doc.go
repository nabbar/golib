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

// Package hooksyslog provides a robust and flexible logrus hook for writing log entries to syslog endpoints.
// It leverages a custom socket implementation for cross-platform network communication (TCP, UDP, Unix sockets),
// ensuring that syslog messages can be sent from any supported OS, including Windows.
// The package is designed for high-performance, asynchronous, and buffered logging with features like
// connection aggregation, automatic reconnection, and RFC 5424 compliance.
//
// # Overview
//
// The `hooksyslog` package integrates with `logrus` to redirect log messages to local or remote syslog servers.
// By abstracting network communication, it provides a unified logging mechanism across different operating systems.
// The implementation ensures non-blocking log operations by delegating buffering and asynchronous writing to an
// underlying aggregator component, which processes log entries in a dedicated background goroutine.
//
// # Design Philosophy
//
// The core principles guiding the design of this package are:
//
//   - True Cross-Platform Support: A single, consistent API for sending syslog messages over the network from any OS,
//     including Windows, without relying on platform-specific logging APIs like Windows Event Log.
//   - Asynchronous Operation: Log calls from the application are non-blocking. The hook delegates the I/O operations
//     to an internal aggregator that uses a buffered channel and a background worker, preventing performance bottlenecks.
//   - Connection Efficiency: A global connection aggregator manages and shares network connections to syslog endpoints.
//     This reduces the overhead of creating new connections for each hook instance and minimizes resource usage.
//   - Robustness and Reliability: The underlying socket client implements automatic reconnection logic, ensuring that
//     logging can resume seamlessly after transient network or service disruptions.
//   - Standard Compliance: Adheres to RFC 5424 for syslog message formatting, including priority, facility, and severity,
//     enabling seamless integration with standard syslog collectors and analyzers.
//
// # Architecture
//
// The package's architecture is designed for modularity, connection sharing, and asynchronous processing.
//
//	┌───────────────────────────────────────────────────────────────────────────┐
//	│                           Application (logrus)                            │
//	│                                                                           │
//	│  ┌─────────────────────────────────────────────────────────────────────┐  │
//	│  │                          HookSyslog Interface                       │  │
//	│  │                        (logrus.Hook + Closer)                       │  │
//	│  └───────────────────────────────▲─────────────────────────────────────┘  │
//	│                                   │                                       │
//	│                                   │ Fire(entry *logrus.Entry)             │
//	│                                   │ (non-blocking, writes to aggregator)  │
//	│                                   │                                       │
//	│  ┌────────────────────────────────▼────────────────────────────────────┐  │
//	│  │                           hks struct                                │  │
//	│  │  - Formatter and log level configuration                            │  │
//	│  │  - Reference to a shared, global connection aggregator (io.Writer)  │  │
//	│  └─────────────────────────────────▲───────────────────────────────────┘  │
//	│                                    │                                      │
//	│                                    │ (Shared across multiple hooks)       │
//	│                                    │                                      │
//	│  ┌─────────────────────────────────▼───────────────────────────────────┐  │
//	│  │                    Global Connection Aggregator Map                 │  │
//	│  │           (Manages one aggregator per unique syslog endpoint)       │  │
//	│  │                                                                     │  │
//	│  │  ┌───────────────────────────────────────────────────────────────┐  │  │
//	│  │  │                 sysAgg (Connection Aggregator)                │  │  │
//	│  │  │  - Reference counting for active hooks                        │  │  │
//	│  │  │  - Wraps an `iotagg.Aggregator` for buffering & async writes  │  │  │
//	│  │  │  - Holds the underlying `libsck.Client` socket connection     │  │  │
//	│  │  └──────────────────────────────▲────────────────────────────────┘  │  │
//	│  └─────────────────────────────────│───────────────────────────────────┘  │
//	│                                    │                                      │
//	│                                    │ (Delegated Async Processing)         │
//	│                                    │                                      │
//	│  ┌─────────────────────────────────▼───────────────────────────────────┐  │
//	│  │               iotagg.Aggregator (from ioutils package)              │  │
//	│  │  - Internal buffer (250 entries) and background goroutine           │  │
//	│  │  - Pulls formatted messages and writes them to the socket           │  │
//	│  │  - Manages write synchronization and error handling                 │  │
//	│  └─────────────────────────────────▲───────────────────────────────────┘  │
//	│                                    │                                      │
//	│                                    │ Write to network socket              │
//	│                                    │                                      │
//	│  ┌─────────────────────────────────▼───────────────────────────────────┐  │
//	│  │                libsck.Client (from socket package)                  │  │
//	│  │  - Cross-platform TCP, UDP, and Unix socket client                  │  │
//	│  │  - Handles connection establishment and automatic reconnection      │  │
//	│  └─────────────────────────────────────────────────────────────────────┘  │
//	│                                                                           │
//	└───────────────────────────────────────────────────────────────────────────┘
//
// # Platform-Specific Behavior
//
// The hook now provides consistent behavior across all platforms by using a network-first approach.
//
//   - Unix/Linux (`sys_syslog.go`): If no network address is specified, the hook attempts to auto-discover the local
//     syslog service by checking for standard Unix domain sockets (`/dev/log`, `/var/run/syslog`).
//
//   - Windows (`sys_winlog.go`): Local syslog auto-discovery is not supported. An explicit network address (e.g., "udp", "localhost:514")
//     must be provided in the configuration. The hook will send standard syslog messages to this endpoint.
//
// # Key Features
//
//   - Connection Aggregation: A global, thread-safe map manages a pool of aggregators, one for each unique syslog endpoint
//     (protocol/address pair). This allows multiple `HookSyslog` instances to share a single underlying network connection,
//     significantly reducing resource consumption. Reference counting ensures connections are closed only when no hooks are using them.
//
//   - Delegated Asynchronous Writing: The hook itself is lightweight. The heavy lifting of buffering, asynchronous processing,
//     and I/O is delegated to the `iotagg.Aggregator` component. This component runs a background goroutine to process a buffer
//     of 250 log entries, ensuring `Fire()` calls are non-blocking.
//
//   - Automatic Reconnection: The underlying `libsck.Client` automatically handles connection failures. If a write fails,
//     it will attempt to reconnect to the syslog server before retrying the write, providing resilience against transient network issues.
//
//   - RFC 5424 Formatting: The hook manually constructs RFC 5424-compliant syslog messages, including the priority value
//     (calculated from facility and severity), timestamp, hostname, tag, and the log message itself.
//
// # Advantages
//
//   - True Cross-Platform Operation: Eliminates dependencies on platform-specific APIs like `log/syslog` or Windows Event Log,
//     providing a consistent syslog-over-network implementation everywhere.
//   - Resource Efficiency: Connection aggregation prevents the proliferation of sockets when multiple hooks point to the same destination.
//   - Non-blocking: Logrus calls return immediately, as I/O is handled by a background worker in the shared aggregator.
//   - Reliable: The combination of buffering and automatic reconnection minimizes log loss during temporary service unavailability.
//
// # Disadvantages
//
//   - Latency: Asynchronous writes introduce a small delay between the log call and the actual transmission of the syslog message.
//   - Error Handling: Network errors are handled internally by the aggregator and printed to `os.Stderr`. They are not propagated
//     back to the application's logging call.
//   - Global State: The use of a global map for connection aggregation introduces a shared state within the application, which
//     can have implications for testing and resource management in complex scenarios.
//
// # Limitations
//
//   - Buffer Overflow: If the rate of log production consistently exceeds the network write speed, the internal buffer
//     (250 entries) can fill up, causing subsequent `Fire()` calls to block.
//   - No TLS: The current socket implementation does not support TLS-encrypted syslog connections.
//
// # Example Usage
//
// Refer to the `example_test.go` file for comprehensive examples, including:
//   - Configuration for remote syslog via TCP/UDP on any platform.
//   - Auto-discovery of local syslog on Unix-like systems.
//   - Graceful shutdown patterns using the `Close()` method on the hook.
package hooksyslog
