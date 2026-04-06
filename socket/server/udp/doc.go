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

// Package udp provides a high-performance, stateless UDP server implementation
// designed for robustness and low-latency datagram processing.
//
// # 1. ARCHITECTURE
//
// This package implements a production-grade UDP server that minimizes the 
// overhead of connection management. It is designed to scale horizontally 
// by leveraging asynchronous datagram processing.
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                    UDP SERVER ARCHITECTURE                  │
//	├─────────────────────────────────────────────────────────────┤
//	│                                                             │
//	│   [ NETWORK INTERFACE ]     [ KERNEL SPACE ]      [ USER SPACE ]
//	│          │                         │                    │
//	│   UDP Datagram 1 --------> [ Receive Buffer ] <--- [ Read() ]
//	│   UDP Datagram 2 --------> [ Receive Buffer ] <--- [ Read() ]
//	│          │                         │                    │
//	│          v                         v                    v
//	│   [ EVENT DISPATCH ]        [ CONTEXT WRAPPER ]   [ USER HANDLER ]
//	│          │                         │                    │
//	│   ┌──────────────┐          ┌────────────────┐    ┌───────────────┐
//	│   │  Done()      │ <─────── │ context.Context│ <──│ HandlerFunc   │
//	│   │  (Cancel)    │          └────────────────┘    │ (Goroutine)   │
//	│   └──────────────┘                                └───────┬───────┘
//	│          ^                                                │
//	│          │          ┌─────────────────────────────────────┘
//	│   ┌──────────────┐  │       ┌───────────────────┐
//	│   │ Gone Channel │ ─┴─────> │  Shutdown Signal  │
//	│   │ (gnc Broadcast)         │  (Instant Exit)   │
//	│   └──────────────┘          └───────────────────┘
//	└─────────────────────────────────────────────────────────────┘
//
// # 2. KEY FEATURES & OPTIMIZATIONS
//
//   - Stateless Operation: Unlike TCP, the UDP server maintains no session state 
//     by default. This allows the server to handle millions of datagrams with 
//     near-zero memory footprint for per-client management.
//
//   - Event-Driven Shutdown (Gone Channel): Traditional UDP servers often rely 
//     on periodic polling to check for shutdown. This package uses the "gnc" 
//     broadcast channel. When the server is closed, the 'gnc' channel is closed 
//     instantly, notifying the main listener loop to exit without any wait.
//
//   - Atomic State Control: All lifecycle flags (IsRunning, IsGone) are managed 
//     via atomic operations (sync/atomic), ensuring lock-free thread safety across 
//     monitoring goroutines.
//
//   - Hook-Based Tuning: The UpdateConn callback provides a hook to call 
//     SetReadBuffer and SetWriteBuffer directly on the underlying net.UDPConn, 
//     which is critical for preventing packet drops under high-bandwidth loads.
//
// # 3. DATA FLOW
//
// The following diagram illustrates the flow of a datagram through the server:
//
//	  [PEER]            [KERNEL BUFFER]            [SERVER LOOP]           [HANDLER]
//	     │                     │                         │                     │
//	     │──(UDP Datagram)────>│                         │                     │
//	     │                     │                         │                     │
//	     │                     │<────(Blocking Read)─────│                     │
//	     │                     │                         │                     │
//	     │                     │────────(Payload)───────>│                     │
//	     │                     │                         │                     │
//	     │                     │                         │────────(Data)──────>│
//	     │                     │                         │                     │
//	     │                     │                         │<──────(Processing)──│
//	     │                     │                         │                     │
//	     │<───(UDP Response)───┼─────────────────────────┼────────(WriteTo)────│
//	     │                     │                         │                     │
//
// # 4. UDP HANDLING SEMANTICS & CAVEATS (RFC 768)
//
// UDP is inherently connectionless, which has specific implications for this server:
//
//  1. Shared Socket: There is only one listener socket for all incoming data. 
//     This means only one HandlerFunc is spawned per Listen() call. 
//     The handler is responsible for managing its own concurrency if needed.
//
//  2. Reliability (RFC 768): This package does NOT implement retries, ACKs, or message
//     ordering. If your application requires these, you must implement them 
//     within your HandlerFunc or use a protocol like TCP.
//
//  3. Max Payload: Datagrams exceeding the MTU (typically 1500 bytes) may be 
//     fragmented by the network stack. It's recommended to keep datagram sizes 
//     under 1472 bytes for IPv4 or 1280 bytes for IPv6 for maximum reliability.
//
// # 5. BEST PRACTICES & PERFORMANCE TUNING
//
//   - High-Throughput Tuning: Always increase kernel buffers for high-load UDP 
//     servers to prevent "ICMP Destination Unreachable" or packet drops:
//
//	updateFn := func(conn net.Conn) {
//	    if udp, ok := conn.(*net.UDPConn); ok {
//	        _ = udp.SetReadBuffer(2 * 1024 * 1024)  // 2MB Read Buffer
//	        _ = udp.SetWriteBuffer(2 * 1024 * 1024) // 2MB Write Buffer
//	    }
//	}
//
//   - Buffer Management: To minimize Garbage Collector (GC) pressure, use a 
//     sync.Pool for the buffers used within the handler:
//
//	var bufPool = sync.Pool{
//	    New: func() any { return make([]byte, 65535) },
//	}
//	// Inside handler...
//	buf := bufPool.Get().([]byte)
//	defer bufPool.Put(buf)
//	n, remoteAddr, _ := ctx.Read(buf)
//
// # 6. USE CASES
//
//   - Case 1: High-Performance Metrics Gathering (StatsD-like systems).
//   - Case 2: Real-time Streaming (Audio/Video datagrams).
//   - Case 3: Network Probing & Monitoring Tools.
//   - Case 4: DNS-like request/response protocols.
//
// # 7. CONCRETE IMPLEMENTATION EXAMPLE
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/nabbar/golib/socket"
//	    "github.com/nabbar/golib/socket/config"
//	    "github.com/nabbar/golib/socket/server/udp"
//	)
//
//	func main() {
//	    // 1. Define the handler (spawns once)
//	    handler := func(ctx socket.Context) {
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
//	        Network: "udp",
//	        Address: ":1234",
//	    }
//
//	    // 3. Instantiate and Start
//	    srv, _ := udp.New(nil, handler, cfg)
//	    srv.Listen(context.Background())
//	}
package udp
