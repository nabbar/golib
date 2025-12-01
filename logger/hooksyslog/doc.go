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

// Package hooksyslog provides a logrus hook for writing log entries to syslog.
//
// # Overview
//
// The hooksyslog package implements a logrus hook that sends log messages to
// system syslog. It supports both Unix/Linux syslog (via log/syslog) and Windows
// Event Log (via golang.org/x/sys/windows/svc/eventlog). The hook provides
// asynchronous, buffered logging with automatic reconnection and graceful shutdown.
//
// # Design Philosophy
//
//   - Cross-platform: Transparent support for Unix/Linux syslog and Windows Event Log
//   - Asynchronous: Non-blocking writes using buffered channels (capacity: 250)
//   - Flexible: Configurable levels, formatters, and syslog facilities
//   - Robust: Automatic reconnection on connection failures
//   - Standard: Full compliance with RFC 5424 syslog severity levels
//
// # Architecture
//
// The package consists of several key components:
//
//	┌─────────────────────────────────────────────┐
//	│          HookSyslog Interface               │
//	│  (logrus.Hook + Done + WriteSev)            │
//	└──────────────┬──────────────────────────────┘
//	               │
//	     ┌─────────▼─────────┐
//	     │   hks struct      │
//	     │  - channels       │
//	     │  - options        │
//	     │  - running state  │
//	     └─────────┬─────────┘
//	               │
//	     ┌─────────▼─────────┐
//	     │   Run() goroutine │
//	     │  - reads channel  │
//	     │  - writes syslog  │
//	     │  - reconnects     │
//	     └─────────┬─────────┘
//	               │
//	       ┌───────┴───────┐
//	       │               │
//	┌──────▼──────┐  ┌─────▼──────┐
//	│  _Syslog    │  │  _WinLog   │
//	│  (Unix)     │  │ (Windows)  │
//	└─────────────┘  └────────────┘
//
// # Platform-Specific Implementation
//
// Unix/Linux (sys_syslog.go):
//   - Uses log/syslog package
//   - Supports network protocols: tcp, udp, unixgram
//   - Full RFC 5424 severity and facility support
//   - Build tags: linux || darwin
//
// Windows (sys_winlog.go):
//   - Uses golang.org/x/sys/windows/svc/eventlog
//   - Maps severity to Windows event types (Error, Warning, Info)
//   - Automatic Windows Event Source registration
//   - Build tags: windows
//
// # Key Features
//
// Asynchronous Buffered Writing:
//   - Buffered channel capacity: 250 entries
//   - Non-blocking Fire() method (unless buffer full)
//   - Background goroutine handles actual syslog writes
//   - Graceful shutdown with Done() channel
//
// Automatic Reconnection:
//   - Retries connection on startup (1-second intervals)
//   - Continues logging even if syslog temporarily unavailable
//   - Connection errors printed to stdout (not propagated to logrus)
//
// Level Mapping:
//   - logrus.PanicLevel  → SyslogSeverityAlert
//   - logrus.FatalLevel  → SyslogSeverityCrit
//   - logrus.ErrorLevel  → SyslogSeverityErr
//   - logrus.WarnLevel   → SyslogSeverityWarning
//   - logrus.InfoLevel   → SyslogSeverityInfo
//   - logrus.DebugLevel  → SyslogSeverityDebug
//
// Field Filtering:
//   - DisableStack: Remove "stack" field
//   - DisableTimestamp: Remove "time" field
//   - EnableTrace: Include "caller", "file", "line" fields
//
// Access Log Mode:
//   - EnableAccessLog: true → writes entry.Message, ignores fields
//   - EnableAccessLog: false → writes formatted fields, ignores Message
//
// # Advantages
//
//   - Non-blocking: Logrus calls don't wait for syslog I/O
//   - Reliable: Buffering prevents log loss during temporary unavailability
//   - Flexible: Works with any logrus.Formatter (JSON, Text, custom)
//   - Cross-platform: Single API for Unix and Windows
//   - Standard: RFC 5424 compliant severity and facility
//
// # Disadvantages
//
//   - Memory overhead: Buffered channel holds up to 250 entries (~250KB typical)
//   - Latency: Async writes introduce small delay before syslog write
//   - Errors silent: Syslog write errors printed to stdout, not returned to caller
//   - Requires Run(): Must call Run(ctx) in goroutine, not automatic
//   - No batching: Each log entry results in separate syslog write
//
// # Limitations
//
//   - Buffer overflow: If buffer fills (>250 entries), Fire() blocks
//   - Platform-specific: Windows has limited severity mapping (3 types)
//   - Network only: Windows remote logging requires network protocol
//   - No TLS: Unix syslog over network doesn't support TLS
//   - Tag immutable: Syslog tag set at creation, can't change per-entry
//
// # Typical Use Cases
//
//   - System service logging to local syslog
//   - Application logging to remote syslog server
//   - Windows service logging to Event Log
//   - Structured logging with JSON formatter
//   - HTTP access logs (EnableAccessLog mode)
//   - Multi-destination logging (syslog + file + console)
//
// # Performance Considerations
//
//   - Channel capacity: 250 entries is default, suitable for most use cases
//   - Write throughput: Limited by syslog server/daemon, not this package
//   - Memory usage: ~1KB per buffered entry (depends on entry size)
//   - Goroutine overhead: Single goroutine per hook instance
//   - Reconnection: 1-second retry interval on connection failures
//
// # Example Usage
//
// See the example_test.go file for complete examples, including:
//   - Basic setup with local syslog
//   - Remote syslog via TCP/UDP
//   - Windows Event Log
//   - Custom formatter with JSON
//   - Access log mode
//   - Graceful shutdown pattern
package hooksyslog
