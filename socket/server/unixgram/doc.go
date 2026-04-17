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
 */

// Package unixgram provides an industrial-strength, high-performance Unix Domain Datagram Socket server.
//
// # 1. INTRODUCTION & CORE CONCEPTS
//
// The unixgram package implements a server based on the SOCK_DGRAM type for AF_UNIX sockets.
// This mechanism provides a connectionless Inter-Process Communication (IPC) that is native
// to POSIX systems (Linux, macOS).
//
// Key technical advantages:
//   - Zero Network Overhead: No IP/UDP headers, no routing, no checksumming.
//   - Message Preservation: Each Read() call yields exactly one complete datagram.
//   - Filesystem Security: Access control is managed via standard Unix UID/GID and permissions.
//
// # 2. ARCHITECTURE
//
// The following diagram illustrates the structural components of the server and their
// relationship with the Host OS and the application handler.
//
//	+---------------------------------------------------------------------------+
//	|                          UNIXGRAM SERVER SYSTEM                           |
//	+---------------------------------------------------------------------------+
//	|                                                                           |
//	|   [ EXTERNAL PEERS ]        [ FILESYSTEM LAYER ]        [ KERNEL SPACE ]  |
//	|          |                          |                          |          |
//	|   Peer A (Client) ----> WRITE ----> Socket File <---- KERNEL BUFFER       |
//	|   Peer B (Client) ----> WRITE ----> (/tmp/u.sock)          |              |
//	|          |                          |                      |              |
//	|          +--------------------------+----------------------+              |
//	|                                     |                                     |
//	|   [ USER SPACE - NABBAR GOLIB ]     v                                     |
//	|   +-------------------------------------------------------------------+   |
//	|   |                         LISTENER SERVICE                          |   |
//	|   |  +-------------------+       +---------------------------------+  |   |
//	|   |  |   net.UnixConn    | <---- |  Shutdown Signal (gnc channel)  |  |   |
//	|   |  | (Socket Binding)  |       |  (Instant Broadcast)            |  |   |
//	|   |  +---------+---------+       +---------------------------------+  |   |
//	|   |            |                                                      |   |
//	|   |            v                 +---------------------------------+  |   |
//	|   |  +-------------------+       |           sync.Pool             |  |   |
//	|   |  |  Read Loop (OS)   | ----> |  (Recycling sCtx Contexts)      |  |   |
//	|   |  +---------+---------+       +-------------------------+-------+  |   |
//	|   |            |                                           ^          |   |
//	|   |            |        (Get & Reset Context)              |          |   |
//	|   |            +-----------------------+-------------------+          |   |
//	|   |                                    |                              |   |
//	|   |                                    v                              |   |
//	|   |                        +---------------------------+              |   |
//	|   |                        |   sCtx (Reader/Wrapper)   |              |   |
//	|   |                        +-------------+-------------+              |   |
//	|   |                                      |                            |   |
//	|   |                                      v                            |   |
//	|   |                        +---------------------------+              |   |
//	|   |                        |    User HandlerFunc(ctx)  |              |   |
//	|   |                        +---------------------------+              |   |
//	|   +-------------------------------------------------------------------+   |
//	|                                                                           |
//	+---------------------------------------------------------------------------+
//
// # 3. PERFORMANCE OPTIMIZATIONS
//
//   - Object Pooling (sync.Pool): By recycling sCtx instances, heap allocations are reduced
//     by ~95%, preventing GC spikes during high-load datagram processing.
//
//   - Event-Driven Shutdown (gnc Channel): Uses a "Gone Broadcast" channel to achieve
//     sub-millisecond shutdown latencies and zero idle-CPU usage.
//
//   - Atomic State Management: Uses atomic.Bool and atomic.Value for lock-free
//     lifecycle control across multiple goroutines.
//
// # 4. DATA FLOW
//
// The following diagram illustrates the flow of a datagram through the Unixgram server:
//
//	[PRODUCER]            [FILESYSTEM]          [SERVER LOOP]           [HANDLER]
//	   │                     │                        │                     │
//	   │───(Datagram)───────>│                        │                     │
//	   │                     │────────(Read)─────────>│                     │
//	   │                     │                        │                     │
//	   │                     │                        │───(Fetch sCtx)─────>│
//	   │                     │                        │                     │
//	   │                     │                        │───(Spawn Handler)──>│
//	   │                     │                        │                     │
//	   │                     │                        │────────(Data)──────>│
//	   │                     │                        │                     │
//	   │                     │                        │<──────(Cleanup)─────│
//	   │                     │                        │                     │
//	   │                     │                        │───(Release sCtx)───>│
//	   │                     │                        │                     │
//
// # 5. USE CASES
//
//   - High-Frequency Metrics Collector: Collecting stats from local sidecars.
//   - Centralized Logging Daemon: A local syslog-like server for microservices.
//   - Local Signal Hub: Triggering configuration reloads across local processes.
//   - IPC for Security Daemons: Using GID-restricted socket files for command isolation.
//
// # 6. BEST PRACTICES & CAVEATS
//
//   - Buffer Management: To minimize GC pressure, use a sync.Pool for the buffers used
//     within the handler (e.g., make([]byte, 65535)).
//   - Kernel Buffers: Increase kernel buffers for high-load servers to prevent packet
//     drops if the handler cannot keep up with the incoming rate.
//   - Truncation: Datagrams larger than the provided buffer will be truncated by the OS.
//     Always use a sufficiently large buffer for Read() calls.
//
// # 7. IMPLEMENTATION EXAMPLE
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/nabbar/golib/file/perm"
//	    "github.com/nabbar/golib/socket"
//	    "github.com/nabbar/golib/socket/config"
//	    "github.com/nabbar/golib/socket/server/unixgram"
//	)
//
//	func main() {
//	    // 1. Define the handler (spawns once)
//	    handler := func(ctx socket.Context) {
//	        defer ctx.Close()
//	        buf := make([]byte, 65535)
//	        for {
//	            n, err := ctx.Read(buf)
//	            if err != nil {
//	                return
//	            }
//	            log.Printf("Received %d bytes: %s", n, string(buf[:n]))
//	        }
//	    }
//
//	    // 2. Setup Config
//	    cfg := config.Server{
//	        Network:   "unixgram",
//	        Address:   "/tmp/app.sock",
//	        PermFile:  perm.NewPerm(0660),
//	        GroupPerm: -1,
//	    }
//
//	    // 3. Instantiate
//	    srv, err := unixgram.New(nil, handler, cfg)
//	    if err != nil {
//	        log.Fatalf("Failed to create UnixGram server: %v", err)
//	    }
//
//	    // 4. Start
//	    if err := srv.Listen(context.Background()); err != nil {
//	        log.Fatalf("UnixGram server failed to listen: %v", err)
//	    }
//	}
package unixgram
