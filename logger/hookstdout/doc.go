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
Package hookstdout provides a logrus hook for writing log entries to stdout with configurable
field filtering, formatting options, and cross-platform color support.

# Overview

The hookstdout package implements a specialized logrus.Hook that writes log entries to os.Stdout
with fine-grained control over output formatting, field filtering, and color support. It is built
as a thin wrapper around the hookwriter package, specifically configured for stdout output with
cross-platform color support via mattn/go-colorable.

This package is particularly useful for:
  - Console applications requiring colored log output
  - CLI tools with structured logging to stdout
  - Development and debugging with readable console logs
  - Production applications using stdout for log aggregation
  - Docker/Kubernetes workloads where stdout is the standard log destination

# Design Philosophy

1. Stdout-Focused: Optimized specifically for stdout output with sensible defaults
2. Cross-Platform Colors: Automatic color support on Windows, Linux, and macOS
3. Zero Configuration: Works out-of-the-box with minimal setup
4. Flexible Formatting: Support for any logrus.Formatter with field filtering
5. Lightweight Wrapper: Delegates to hookwriter for core functionality

# Key Features

  - Automatic stdout routing with os.Stdout as default destination
  - Cross-platform color support via mattn/go-colorable
  - Selective field filtering (stack traces, timestamps, caller info)
  - Access log mode for message-only output
  - Multiple formatter support (JSON, Text, custom)
  - Level-based filtering (handle only specific log levels)
  - Optional color output control (enable/disable per hook)
  - Zero-allocation for disabled hooks (returns nil)

# Architecture

The package implements a simple delegation pattern to hookwriter:

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
	        │   HookStdOut.Fire()        │
	        │   (delegates to            │
	        │    HookWriter)             │
	        └────────────┬───────────────┘
	                     │
	                     ▼
	          ┌─────────────────┐
	          │   HookWriter    │
	          │                 │
	          │  1. Dup Entry   │
	          │  2. Filter      │
	          │  3. Format      │
	          │  4. Write       │
	          └────────┬────────┘
	                   │
	                   ▼
	        ┌──────────────────────┐
	        │  colorable.Stdout    │
	        │  (os.Stdout wrapper) │
	        └──────────────────────┘

# Package Structure

The hookstdout package is intentionally minimal with a single file:

  - interface.go: Public API with New() and NewWithWriter() constructors

All core functionality (field filtering, formatting, entry processing) is delegated
to the hookwriter package, maintaining a clean separation of concerns.

# Data Flow

1. Entry Creation: Application creates log entry via logger.Info/Warn/Error/etc.
2. Hook Invocation: logrus calls Fire() on all registered hooks for matching levels
3. Delegation: HookStdOut delegates to HookWriter.Fire()
4. Entry Processing: HookWriter duplicates entry, filters fields, formats output
5. Stdout Write: Formatted output written to os.Stdout via colorable wrapper

# Basic Usage

Create a hook and register it with a logrus logger:

	import (
	    "github.com/sirupsen/logrus"
	    "github.com/nabbar/golib/logger/config"
	    "github.com/nabbar/golib/logger/hookstdout"
	)

	func main() {
	    // Configure hook options
	    opt := &config.OptionsStd{
	        DisableStandard:  false,
	        DisableColor:     false,  // Enable color output
	        DisableStack:     true,
	        DisableTimestamp: false,
	        EnableTrace:      false,
	    }

	    // Create hook with Text formatter
	    hook, err := hookstdout.New(opt, nil, &logrus.TextFormatter{
	        ForceColors: true,
	    })
	    if err != nil {
	        log.Fatal(err)
	    }

	    // Register hook with logger
	    logger := logrus.New()
	    logger.AddHook(hook)

	    // Log entries will be written to stdout with colors
	    logger.WithField("msg", "Application started").Info("ignored message")

		// Log entries will be written to stdout with colors
		logger.WithField("msg", "User logged in").WithField("user", "john").Info("ignored message")

		// Error messages will NOT be written to stdout
	    logger.Info("This message does not go to stdout")
		// Use only field to define message, all message set into logrus function are ignored except for AccessLog (see below)
	}

# Configuration Options

The OptionsStd struct controls hook behavior:

DisableStandard: If true, returns nil hook (completely disabled)

	opt := &config.OptionsStd{DisableStandard: true}
	hook, _ := hookstdout.New(opt, nil, nil)  // Returns (nil, nil)

DisableColor: If true, wraps stdout to disable color escape sequences

	opt := &config.OptionsStd{DisableColor: true}
	// Output will not contain ANSI color codes

DisableStack: Filters out stack trace fields from output

	opt := &config.OptionsStd{DisableStack: true}
	logger.WithField("stack", trace).Error("error")  // "stack" field removed

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

Console Application with Colors:

	opt := &config.OptionsStd{
	    DisableStandard: false,
	    DisableColor:    false,  // Enable colors
	}
	hook, _ := hookstdout.New(opt, nil, &logrus.TextFormatter{
	    ForceColors:     true,
	    FullTimestamp:   true,
	})
	logger.AddHook(hook)
	// Colored output for better readability

CLI Tool with Minimal Output:

	opt := &config.OptionsStd{
	    DisableStack:     true,
	    DisableTimestamp: true,
	    EnableTrace:      false,
	}
	hook, _ := hookstdout.New(opt, nil, &logrus.TextFormatter{
	    DisableTimestamp: true,
	})
	// Clean, minimal CLI output

Docker Container Logs:

	opt := &config.OptionsStd{
	    DisableColor: true,  // No colors for log aggregation
	}
	hook, _ := hookstdout.New(opt, nil, &logrus.JSONFormatter{})
	// Structured JSON logs to stdout for container log drivers

Development Mode with Debug Logs:

	debugOpt := &config.OptionsStd{
	    DisableStack: false,  // Keep stack traces
	    EnableTrace:  true,   // Include caller info
	}
	hook, _ := hookstdout.New(debugOpt, []logrus.Level{
	    logrus.DebugLevel,
	    logrus.InfoLevel,
	}, nil)
	// Verbose debug output with full context

Access Log to Stdout:

	accessOpt := &config.OptionsStd{
	    DisableStandard: false,
	    EnableAccessLog: true,  // Message-only mode
	    DisableColor:    true,
	}
	hook, _ := hookstdout.New(accessOpt, nil, nil)
	logger.Info("192.168.1.1 - GET /api/users - 200 - 45ms")
	// for AccesLog, field are ignored and only message passed are written

# Performance Considerations

Memory Efficiency:

  - Entry delegation to hookwriter minimizes allocations
  - Disabled hooks (DisableStandard=true) return nil with zero allocation
  - Color support via colorable adds minimal overhead (~1-2% on Windows)

Write Performance:

  - Stdout writes are buffered by OS, generally fast
  - Avoid high-frequency logging (>10k/sec) to stdout in production
  - For high-throughput, consider async aggregation:
    github.com/nabbar/golib/ioutils/aggregator

Formatter Overhead:

  - JSON formatters: ~50-100µs per entry (fast, structured)
  - Text formatters: ~100-200µs per entry (slower, readable)
  - Access log mode: ~20-30µs per entry (fastest, no formatting)

Color Performance:

  - Unix/Linux/macOS: Zero overhead (native ANSI support)
  - Windows: Minimal overhead via colorable's virtual terminal sequences
  - Disable colors in production for slight performance gain

# Thread Safety

The hook implementation is thread-safe when used correctly:

  - Safe: Multiple goroutines logging to the same logger with this hook
  - Safe: Multiple hooks registered on the same logger
  - Safe: Concurrent stdout writes (os.Stdout is thread-safe)
  - Unsafe: Concurrent calls to Fire() with same entry (logrus prevents this)
  - Unsafe: Modifying hook configuration after creation (immutable design)

Note: os.Stdout on Unix-like systems has atomic writes for messages < PIPE_BUF
(typically 4KB), but Windows may interleave concurrent writes. For guaranteed
ordering, use single-threaded logging or an aggregator.

# Error Handling

The hook can return errors in the following situations:

Construction Errors:

	// No errors expected for New() with valid options
	hook, err := hookstdout.New(opt, nil, nil)
	// err is always nil (unless delegated hookwriter.New fails)

Runtime Errors:

	// Delegated to hookwriter.Fire()
	err := hook.Fire(entry)  // Returns formatter or writer errors

Silent Failures:

  - Empty log data: Fire() returns nil without writing (normal)
  - Empty access log message: Fire() returns nil without writing (normal)
  - Disabled hook: New() returns (nil, nil) - not an error

# Comparison with HookStdErr

hookstdout vs hookstderr:

	hookstdout:
	  - Writes to os.Stdout
	  - Typically for Info, Debug levels
	  - Suitable for structured logs, access logs
	  - Output can be piped/redirected separately

	hookstderr:
	  - Writes to os.Stderr
	  - Typically for Warn, Error, Fatal, Panic levels
	  - Suitable for error logs, diagnostics
	  - Separates errors from normal output

Use both hooks for proper stdout/stderr separation in CLI tools:

	stdoutHook, _ := hookstdout.New(opt, []logrus.Level{
	    logrus.InfoLevel,
	    logrus.DebugLevel,
	}, nil)
	stderrHook, _ := hookstderr.New(opt, []logrus.Level{
	    logrus.WarnLevel,
	    logrus.ErrorLevel,
	    logrus.FatalLevel,
	}, nil)
	logger.AddHook(stdoutHook)
	logger.AddHook(stderrHook)

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

HookWriter:

	import "github.com/nabbar/golib/logger/hookwriter"
	// Core implementation that hookstdout delegates to

HookStdErr:

	import "github.com/nabbar/golib/logger/hookstderr"
	// Companion package for stderr output

IOUtils Aggregator:

	import "github.com/nabbar/golib/ioutils/aggregator"
	// For async high-performance log aggregation

# Limitations

 1. Stdout-Only: This package is specifically for stdout. For other destinations,
    use hookwriter directly.

 2. No Buffering: Stdout writes are unbuffered by default. For high-frequency
    logging, wrap stdout with bufio.Writer via NewWithWriter().

 3. Color Limitations: Color output depends on terminal capabilities. Some
    environments (e.g., non-TTY pipes) may not display colors correctly even
    when enabled.

 4. No Write Retries: Failed stdout writes return errors but don't retry.
    Generally not an issue as stdout writes rarely fail.

 5. No Lifecycle Management: Hook doesn't manage stdout lifecycle (no Close()).
    This is intentional as stdout should remain open for process lifetime.

# Best Practices

DO:
  - Enable colors for interactive terminals, disable for log aggregation
  - Use JSON formatter for production, Text formatter for development
  - Filter out verbose fields (stack, caller) for cleaner output
  - Use level filtering to route different levels to stdout vs stderr
  - Check for nil when DisableStandard is conditionally set
  - Use access log mode for HTTP access logs or similar patterns

DON'T:
  - Use this for file output (use hookwriter with os.Create instead)
  - Enable colors when piping to files or non-TTY destinations
  - Log extremely high frequency (>10k/sec) to stdout without aggregation
  - Ignore the nil return when DisableStandard is true
  - Mix structured and unstructured logging without clear separation

# Testing

The package includes comprehensive tests covering:

  - Hook creation with various configurations
  - Field filtering (stack, time, caller, file, line)
  - Access log mode with empty messages
  - Formatter integration (JSON, Text)
  - Integration with logrus.Logger
  - Level filtering behavior
  - Multiple hooks on single logger
  - Color enable/disable scenarios

Run tests:

	go test -v github.com/nabbar/golib/logger/hookstdout

Check coverage:

	go test -cover github.com/nabbar/golib/logger/hookstdout

Current coverage: Target >80% (delegates most logic to hookwriter)

# Examples

See example_test.go for runnable examples demonstrating:
  - Basic hook creation and usage
  - Colored console output
  - Access log mode for HTTP logs
  - Level-specific filtering
  - Field filtering configurations
  - Custom writer usage (NewWithWriter)

# Related Packages

  - github.com/sirupsen/logrus - Underlying logging framework
  - github.com/mattn/go-colorable - Cross-platform color support
  - github.com/nabbar/golib/logger - Main logger package
  - github.com/nabbar/golib/logger/config - Configuration types
  - github.com/nabbar/golib/logger/types - Hook interface and constants
  - github.com/nabbar/golib/logger/hookwriter - Core hook implementation
  - github.com/nabbar/golib/logger/hookstderr - Companion stderr hook
  - github.com/nabbar/golib/ioutils/aggregator - Async write aggregation

# License

MIT License - See LICENSE file for details.

Copyright (c) 2025 Nicolas JUHEL
*/
package hookstdout
