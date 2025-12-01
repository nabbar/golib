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
Package hookwriter provides a logrus hook for writing log entries to custom io.Writer instances
with configurable field filtering and formatting options.

# Overview

The hookwriter package implements a logrus.Hook that intercepts log entries and writes them to
any io.Writer with fine-grained control over which fields are included, how they're formatted,
and whether special modes like access logging are enabled. This is particularly useful for:

  - Writing logs to multiple destinations (files, network, buffers)
  - Filtering sensitive or verbose fields from output
  - Creating specialized log formats for different outputs
  - Implementing access log patterns separate from application logs

# Design Philosophy

1. Flexible Output: Write to any io.Writer (files, buffers, network sockets, custom writers)
2. Non-invasive Filtering: Filter fields without modifying the original entry
3. Format Agnostic: Support any logrus.Formatter or use default serialization
4. Simple Integration: Single-function creation with clear configuration options
5. Stateless Operation: No background goroutines or complex lifecycle management

# Key Features

  - Custom io.Writer support for any output destination
  - Selective field filtering (stack traces, timestamps, caller info)
  - Access log mode for message-only output
  - Multiple formatter support (JSON, Text, custom)
  - Level-based filtering (handle only specific log levels)
  - Optional color output via mattn/go-colorable
  - Zero-allocation for disabled hooks (returns nil)

# Architecture

The package consists of a simple architecture with minimal components:

	┌──────────────────────────────────────────────┐
	│             logrus.Logger                    │
	│                                              │
	│  ┌────────────────────────────────────┐      │
	│  │  logger.Info("message")            │      │
	│  └────────────────┬───────────────────┘      │
	│                   │                          │
	│                   ▼                          │
	│         ┌──────────────────┐                 │
	│         │  logrus.Entry    │                 │
	│         └──────────┬───────┘                 │
	│                    │                         │
	└────────────────────┼─────────────────────────┘
	                     │
	                     ▼
	        ┌────────────────────────────┐
	        │     HookWriter.Fire()      │
	        │                            │
	        │  1. Duplicate Entry        │
	        │  2. Filter Fields          │
	        │     - Stack (opt)          │
	        │     - Time (opt)           │
	        │     - Caller/File (opt)    │
	        │  3. Format Entry           │
	        │     - Formatter            │
	        │     - Access Log Mode      │
	        │  4. Write to io.Writer     │
	        └────────────┬───────────────┘
	                     │
	                     ▼
	              ┌──────────────┐
	              │  io.Writer   │
	              │  (file, net) │
	              └──────────────┘

# Data Flow

1. Entry Creation: Application code creates log entry via logger.Info/Warn/Error/etc.
2. Hook Invocation: logrus calls Fire() on all registered hooks for matching log levels
3. Entry Duplication: Hook duplicates entry to avoid modifying original
4. Field Filtering: Removes configured fields (stack, time, caller, file, line)
5. Formatting: Applies formatter or access log mode to serialize entry
6. Write: Outputs formatted bytes to configured io.Writer

# Basic Usage

Create a hook and register it with a logrus logger:

	import (
	    "os"
	    "github.com/sirupsen/logrus"
	    "github.com/nabbar/golib/logger/config"
	    "github.com/nabbar/golib/logger/hookwriter"
	)

	func main() {
	    // Create file writer
	    file, _ := os.Create("app.log")
	    defer file.Close()

	    // Configure hook options
	    opt := &config.OptionsStd{
	        DisableStandard:  false,
	        DisableColor:     true,
	        DisableStack:     true,
	        DisableTimestamp: false,
	        EnableTrace:      false,
	    }

	    // Create hook with JSON formatter
	    hook, err := hookwriter.New(file, opt, nil, &logrus.JSONFormatter{})
	    if err != nil {
	        log.Fatal(err)
	    }

	    // Register hook with logger
	    logger := logrus.New()
	    logger.AddHook(hook)

	    // Log entries will be written to file
	    logger.Info("Application started")
	}

# Configuration Options

The OptionsStd struct controls hook behavior:

DisableStandard: If true, returns nil hook (completely disabled)

	opt := &config.OptionsStd{DisableStandard: true}
	hook, _ := hookwriter.New(writer, opt, nil, nil)  // Returns (nil, nil)

DisableColor: If true, wraps writer with colorable.NewNonColorable() to disable color output

	opt := &config.OptionsStd{DisableColor: true}
	// Disables color escape sequences in output

DisableStack: Filters out stack trace fields from output

	opt := &config.OptionsStd{DisableStack: true}
	logger.WithField("stack", trace).Error("error")  // "stack" field removed from output

DisableTimestamp: Filters out timestamp fields from output

	opt := &config.OptionsStd{DisableTimestamp: true}
	// "time" field removed from all entries

EnableTrace: Controls caller/file/line field inclusion

	opt := &config.OptionsStd{EnableTrace: false}
	// Removes "caller", "file", "line" fields from output

EnableAccessLog: Enables message-only mode (ignores fields and formatters)

	opt := &config.OptionsStd{EnableAccessLog: true}
	logger.WithField("status", 200).Info("GET /api/users")
	// Output: "GET /api/users\n" (fields ignored)

# Common Use Cases

Multiple Output Destinations:

	fileHook, _ := hookwriter.New(logFile, fileOpt, nil, &logrus.JSONFormatter{})
	netHook, _ := hookwriter.New(networkConn, netOpt, nil, &logrus.TextFormatter{})
	logger.AddHook(fileHook)
	logger.AddHook(netHook)
	// Logs written to both file and network

Level-Specific Hooks:

	errorHook, _ := hookwriter.New(errorFile, opt, []logrus.Level{
	    logrus.ErrorLevel,
	    logrus.FatalLevel,
	    logrus.PanicLevel,
	}, nil)
	// Only errors written to error file

Access Log Pattern:

	accessOpt := &config.OptionsStd{
	    DisableStandard: false,
	    EnableAccessLog: true,
	}
	accessHook, _ := hookwriter.New(accessLog, accessOpt, nil, nil)
	logger.AddHook(accessHook)
	logger.Info("GET /api/users - 200 OK")  // Clean access log format

Filtered Debug Output:

	debugOpt := &config.OptionsStd{
	    DisableStack:     true,
	    DisableTimestamp: true,
	    EnableTrace:      false,
	}
	debugHook, _ := hookwriter.New(os.Stdout, debugOpt, []logrus.Level{logrus.DebugLevel}, nil)
	// Minimal debug output without clutter

# Performance Considerations

Memory Efficiency:

  - Entry duplication uses entry.Dup() which shares data structures where possible
  - Field filtering modifies the duplicated entry's Data map without allocating new maps
  - Disabled hooks (DisableStandard=true) return nil with zero allocation

Write Performance:

  - Write performance depends entirely on the underlying io.Writer
  - Buffered writers (bufio.Writer) recommended for high-frequency logging
  - Network writers should have reasonable timeouts to avoid blocking
  - File writers benefit from OS-level buffering

Formatter Overhead:

  - JSON formatters are faster but produce larger output
  - Text formatters are slower but more human-readable
  - Access log mode bypasses formatting entirely (fastest)

Scalability:

  - Hooks are called synchronously by logrus for each entry
  - Multiple hooks add cumulative overhead (each hook's Fire() is called)
  - For high-throughput scenarios, use aggregation packages:
    github.com/nabbar/golib/ioutils/aggregator for async writes

# Thread Safety

The hook implementation is thread-safe when used correctly:

  - Safe: Multiple goroutines logging to the same logger with this hook
  - Safe: Multiple hooks registered on the same logger
  - Unsafe: Concurrent calls to Fire() with the same entry instance (logrus prevents this)
  - Unsafe: Modifying hook configuration after creation (immutable design)

The underlying io.Writer must be thread-safe for concurrent writes. Most standard
writers (os.File, bufio.Writer, bytes.Buffer) are not inherently thread-safe for
concurrent writes. Use synchronization or the aggregator package for concurrent scenarios.

# Error Handling

The hook can return errors in the following situations:

Construction Errors:

	hook, err := hookwriter.New(nil, opt, nil, nil)
	// err: "hook writer is nil"

Runtime Errors:

	// Formatter error during Fire()
	err := hook.Fire(entry)  // Returns formatter.Format() error

	// Writer error during Fire()
	err := hook.Fire(entry)  // Returns writer.Write() error

Silent Failures:

  - Empty log data: Fire() returns nil without writing (normal behavior)
  - Empty access log message: Fire() returns nil without writing (normal behavior)
  - Disabled hook: New() returns (nil, nil) - not an error

# Comparison with Standard Output

Standard logrus output (logger.SetOutput):

  - Single output destination
  - No field filtering
  - Applied to all log levels
  - Direct write (no hook overhead)

HookWriter advantages:

  - Multiple simultaneous outputs
  - Per-hook field filtering
  - Per-hook level filtering
  - Per-hook formatting
  - Doesn't replace SetOutput (additive)

Use standard output for simple cases, hooks for advanced routing.

# Integration with golib Packages

Logger Package:

	import "github.com/nabbar/golib/logger"
	// Main logger package that uses this hook internally

Logger Config:

	import "github.com/nabbar/golib/logger/config"
	// Provides OptionsStd configuration structure

Logger Types:

	import "github.com/nabbar/golib/logger/types"
	// Defines Hook interface and field constants

IOUtils Aggregator:

	import "github.com/nabbar/golib/ioutils/aggregator"
	// For async high-performance log aggregation

# Limitations

1. Synchronous Writes: Hook writes are synchronous with log calls. Slow writers block logging.
   Mitigation: Use aggregator package for async writes or buffered writers.

2. No Write Retries: Failed writes return errors but don't retry or queue.
   Mitigation: Use reliable writers or add retry logic in custom writers.

3. No Buffer Management: Hook doesn't buffer or flush data.
   Mitigation: Use bufio.Writer and call Flush() explicitly when needed.

4. No Compression: No built-in log compression or rotation.
   Mitigation: Use external log rotation tools (logrotate) or writer wrappers.

5. Writer Lifecycle: Hook doesn't manage writer Close().
   Mitigation: Caller must close writers when done. Not an issue - proper design.

# Best Practices

DO:
  - Use bufio.Writer for high-frequency logging to amortize I/O costs
  - Set reasonable timeouts on network writers to prevent blocking
  - Close writers explicitly when shutting down
  - Use level filtering to send different levels to different destinations
  - Enable access log mode for HTTP access logs or similar patterns
  - Check for nil when DisableStandard is conditionally true

DON'T:
  - Use unbuffered network writers in performance-critical paths
  - Ignore errors from New() (check for nil writer error)
  - Share non-thread-safe writers across multiple hooks without synchronization
  - Modify opt struct after passing to New() (not effective, options are copied)
  - Use this for extremely high-throughput logging (>100k/sec) without aggregation

# Testing

The package includes comprehensive tests covering:

  - Hook creation with various configurations
  - Field filtering (stack, time, caller, file, line)
  - Access log mode with empty messages
  - Formatter integration (JSON, Text)
  - Integration with logrus.Logger
  - Level filtering behavior
  - Multiple hooks on single logger
  - Error paths (nil writer, write failures)

Run tests:

	go test -v github.com/nabbar/golib/logger/hookwriter

Check coverage:

	go test -cover github.com/nabbar/golib/logger/hookwriter

Current coverage: 90.2% (exceeds 80% target)

# Examples

See example_test.go for runnable examples demonstrating:
  - Basic hook creation and usage
  - File writing with JSON formatter
  - Access log mode for HTTP logs
  - Multiple hooks for different outputs
  - Level-specific filtering
  - Field filtering configurations

# Related Packages

  - github.com/sirupsen/logrus - Underlying logging framework
  - github.com/mattn/go-colorable - Color support on Windows
  - github.com/nabbar/golib/logger - Main logger package
  - github.com/nabbar/golib/logger/config - Configuration types
  - github.com/nabbar/golib/logger/types - Hook interface and constants
  - github.com/nabbar/golib/ioutils/aggregator - Async write aggregation

# License

MIT License - See LICENSE file for details.

Copyright (c) 2025 Nicolas JUHEL
*/
package hookwriter
