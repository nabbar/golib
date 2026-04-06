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

// Package tcp provides a high-performance, production-ready TCP server implementation
// with support for TLS, connection pooling, and centralized idle management.
//
// # 1. ARCHITECTURE
//
// The TCP server is engineered to handle massive concurrency while maintaining a
// predictable and low memory footprint. It achieves this by combining Go's
// standard library net.TCPListener with several advanced architectural patterns.
//
//	┌─────────────────────────────────────────────────────────────────────────────┐
//	│                              TCP SERVER SYSTEM                              │
//	├─────────────────────────────────────────────────────────────────────────────┤
//	│                                                                             │
//	│   [ NETWORK LAYER ]         [ KERNEL SPACE ]          [ USER SPACE ]        │
//	│          │                         │                        │               │
//	│   TCP SYN Package ----> [ TCP Accept Queue ] ----> [ Accept Loop (srv) ]    │
//	│                                                             │               │
//	│                                                             ▼               │
//	│   [ RESOURCE MANAGEMENT ]                           [ CONNECTION HANDLER ]  │
//	│          │                                                  │               │
//	│   ┌──────────────┐          ┌────────────────┐      ┌───────────────┐       │
//	│   │  sync.Pool   │ <─────── │ Context Reset  │ <─── │  net.Conn     │       │
//	│   │ (sCtx items) │          └────────────────┘      │  (Raw Socket) │       │
//	│   └──────────────┘                                  └───────┬───────┘       │
//	│          ^                                                  │               │
//	│          │                                                  v               │
//	│   ┌──────────────┐          ┌────────────────┐      ┌───────────────┐       │
//	│   │ Idle Manager │ <─────── │ Registration   │ <─── │ User Handler  │       │
//	│   │ (sckidl.Mgr) │          └────────────────┘      │ (Goroutine)   │       │
//	│   └──────────────┘                                  └───────────────┘       │
//	│                                                             │               │
//	│   [ SHUTDOWN CONTROL ]                                      v               │
//	│   ┌──────────────────┐                              ┌───────────────┐       │
//	│   │   Gone Channel   │ ───────(Broadcast)─────────> │   Cleanup     │       │
//	│   │      (gnc)       │                              │ (Pool Return) │       │
//	│   └──────────────────┘                              └───────────────┘       │
//	└─────────────────────────────────────────────────────────────────────────────┘
//
// # 2. PERFORMANCE OPTIMIZATIONS
//
//   - Zero-Allocation Connection Path: The server uses a sync.Pool to manage connection
//     contexts (sCtx). Instead of allocating a new context for every client, the server
//     fetches one from the pool, resets its internal state (atomic counters, context
//     propagation), and returns it once the connection is closed. This reduces GC
//     pause times by up to 90% in high-load scenarios.
//
//   - Synchronous Acceptance Loop: The listener loop blocks directly on net.Listener.Accept().
//     Recent performance profiling showed that using intermediate channels for connection
//     distribution introduced unnecessary scheduling overhead.
//
//   - Centralized Idle Scanning: Replaced individual per-connection tickers with
//     integration into the nabbar/golib/socket/idlemgr. A single background scanner
//     handles thousands of timeouts by checking atomic activity counters, reducing
//     the overall "timer" overhead in the Go runtime.
//
//   - Event-Driven Shutdown (Gone Channel): The 'gnc' channel acts as a broadcast
//     mechanism. When the server starts shutting down, this channel is closed. All
//     active connection goroutines select on this channel, allowing them to terminate
//     gracefully and instantly.
//
//   - Systematic NoDelay: TCP_NODELAY is enabled by default to ensure minimal latency
//     for small packets, bypassing Nagle's algorithm.
//
// # 3. DATA FLOW
//
// The following diagram illustrates the lifecycle of a connection within the server:
//
//	[CLIENT]          [SERVER LISTENER]          [IDLE MGR]          [HANDLER]
//	   │                     │                       │                   │
//	   │───(TCP Connect)────>│                       │                   │
//	   │                     │───(Fetch sCtx)───────>│                   │
//	   │                     │                       │                   │
//	   │                     │───(Register)─────────>│                   │
//	   │                     │                       │                   │
//	   │                     │───(Spawn Handler)────────────────────────>│
//	   │                     │                       │                   │
//	   │<───(I/O Activity)───┼───────────────────────────────────────────│
//	   │                     │                       │                   │
//	   │                     │                       │<──(Atomic Reset)──│
//	   │                     │                       │                   │
//	   │───(TCP Close)──────>│                       │                   │
//	   │                     │───(Unregister)───────>│                   │
//	   │                     │                       │                   │
//	   │                     │───(Release sCtx)─────>│                   │
//	   │                     │                       │                   │
//
// # 4. SECURITY & TLS (RFC 8446, RFC 5246)
//
// The server provides first-class support for TLS 1.2 and 1.3 through the SetTLS method.
// It integrates seamlessly with the nabbar/golib/certificates package.
//
// ## TLS over Pure Sockets: Limitations and Constraints
//
// When using TLS over pure sockets (without a higher-level protocol like HTTP/1.1 or
// HTTP/2 which provide explicit host headers), several security and validation
// constraints must be acknowledged:
//
//  1. Trust Chain Relativity: Since the TLS handshake often happens on a direct IP
//     dial, the verification of the certificate's Common Name (CN) or Subject
//     Alternative Name (SAN) is relative to the IP or the provided ServerName.
//
//  2. SNI Constraints (RFC 6066): Server Name Indication (SNI) is used to select the
//     appropriate certificate. If the client does not provide SNI, the server
//     falls back to its default certificate. In pure socket environments, clients
//     must be explicitly configured to send SNI for proper virtual hosting support.
//
//  3. RFC 8446 (TLS 1.3): The server prioritizes TLS 1.3, which eliminates
//     vulnerable cryptographic primitives and provides "0-RTT" (not enabled by
//     default for security reasons).
//
//  4. Trust Chain Validation: Chain of trust verification is performed by the
//     underlying crypto/tls package. However, in pure socket scenarios, the
//     lack of standardized application-level verification (like HTTP's HSTS)
//     means the security of the initial handshake is paramount.
//
// # 5. BEST PRACTICES & USE CASES
//
//   - Use Case: Building high-performance microservices communicating over TCP.
//   - Use Case: Implementing custom protocols (e.g., database drivers, legacy IPC).
//   - Best Practice: Always call defer ctx.Close() within your HandlerFunc to ensure
//     resource return even if a panic occurs.
//   - Best Practice: Use the UpdateConn callback for OS-level tuning like SO_RCVBUF
//     and SO_SNDBUF for high-bandwidth applications.
//
// # 6. IMPLEMENTATION EXAMPLE
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "github.com/nabbar/golib/socket"
//	    "github.com/nabbar/golib/socket/config"
//	    "github.com/nabbar/golib/socket/server/tcp"
//	)
//
//	func main() {
//	    // 1. Define the handler
//	    handler := func(ctx socket.Context) {
//	        defer ctx.Close()
//	        buf := make([]byte, 1024)
//	        n, _ := ctx.Read(buf)
//	        fmt.Printf("Received: %s\n", string(buf[:n]))
//	        ctx.Write([]byte("ACK"))
//	    }
//
//	    // 2. Setup Config
//	    cfg := config.Server{
//	        Network: "tcp",
//	        Address: ":9090",
//	    }
//
//	    // 3. Instantiate
//	    srv, _ := tcp.New(nil, handler, cfg)
//
//	    // 4. Start
//	    srv.Listen(context.Background())
//	}
package tcp
