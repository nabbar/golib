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
Package hookstderr provides a logrus hook for writing log entries to standard error (stderr)
with configurable field filtering and formatting options.

# Overview

The hookstderr package is a specialized wrapper around the hookwriter package, specifically
configured to write log entries to standard error (stderr). It implements a logrus.Hook that
intercepts log entries and writes them to stderr (or a custom writer) with fine-grained control
over which fields are included, how they're formatted, and whether special modes like access
logging are enabled.

This package is particularly useful for:
  - Separating error/diagnostic output from normal application output
  - Writing structured error logs to stderr while keeping stdout clean
  - Filtering sensitive or verbose fields from error output
  - Creating specialized error log formats distinct from stdout logs
  - Implementing standard Unix conventions (errors to stderr, output to stdout)

# Design Philosophy

1. **Standard Error Focus**: Dedicated to stderr output following Unix conventions
2. **Wrapper Simplicity**: Thin wrapper over hookwriter providing stderr-specific defaults
3. **Configuration Consistency**: Uses same OptionsStd as other logger hooks for uniformity
4. **Color Support**: Automatic color handling via mattn/go-colorable for cross-platform compatibility
5. **Flexible Writing**: Supports custom writers for testing and advanced use cases

# Key Features

  - Dedicated stderr output for error and diagnostic messages
  - Automatic color support detection and stripping when needed
  - Selective field filtering (stack traces, timestamps, caller info)
  - Access log mode for message-only output
  - Multiple formatter support (JSON, Text, custom)
  - Level-based filtering (handle only specific log levels)
  - Custom writer support for testing and advanced scenarios
  - Zero-allocation for disabled hooks (returns nil)

# Architecture

The package architecture is intentionally minimal, delegating most functionality to hookwriter:

	┌──────────────────────────────────────────────┐
	│             logrus.Logger                    │
	│                                              │
	│  ┌────────────────────────────────────┐      │
	│  │  logger.Error("error message")     │      │
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
	        │   HookStdErr.Fire()        │
	        │   (delegates to hookwriter)│
	        └────────────┬───────────────┘
	                     │
	                     ▼
	              ┌──────────────┐
	              │  os.Stderr   │
	              │  (or custom) │
	              └──────────────┘

# Comparison with hookstdout

HookStdOut (stdout):
  - Intended for normal application output and informational messages
  - Typically used for Info, Debug, and Trace level logs
  - Follows Unix convention of stdout for program output

HookStdErr (stderr):
  - Intended for error messages and diagnostic information
  - Typically used for Error, Fatal, Panic, and Warning level logs
  - Follows Unix convention of stderr for error/diagnostic output

Both packages:
  - Use identical configuration structure (OptionsStd)
  - Share the same field filtering capabilities
  - Delegate to hookwriter for actual implementation
  - Support custom writers for testing

# Basic Usage

Create a stderr hook and register it with a logrus logger:

	import (
	    "github.com/sirupsen/logrus"
	    "github.com/nabbar/golib/logger/config"
	    "github.com/nabbar/golib/logger/hookstderr"
	)

	func main() {
	    // Configure hook options
	    opt := &config.OptionsStd{
	        DisableStandard:  false,
	        DisableColor:     false,  // Enable color on stderr
	        DisableStack:     true,   // Filter stack traces
	        DisableTimestamp: false,
	        EnableTrace:      false,  // Filter caller info
	    }

	    // Create stderr hook
	    hook, err := hookstderr.New(opt, nil, &logrus.TextFormatter{})
	    if err != nil {
	        log.Fatal(err)
	    }

	    // Register hook with logger
	    logger := logrus.New()
	    logger.AddHook(hook)

	    // Error messages will be written to stderr
	    logger.WithField("msg", "This error goes to stderr").Error("ignored message")
	    logger.WithField("msg", "This error goes to stderr").WithField("err", err).Error("ignored message")

		// Error messages will NOT be written to stderr
	    logger.Error("This error does not go to stderr")
		// Use only field to define message, all message set into logrus function are ignored except for AccessLog (see below)
	}

# Configuration Options

The OptionsStd struct controls hook behavior:

DisableStandard: Completely disables the hook

	opt := &config.OptionsStd{DisableStandard: true}
	hook, _ := hookstderr.New(opt, nil, nil)  // Returns (nil, nil)

DisableColor: Strips ANSI color escape sequences from output

	opt := &config.OptionsStd{DisableColor: true}
	// Color codes will be removed from stderr output

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
	logger.WithField("status", 500).Error("Internal Server Error")
	// Output: "Internal Server Error\n" (fields ignored)

# Common Use Cases

Separate Error and Info Logs:

	// Errors to stderr
	stderrOpt := &config.OptionsStd{DisableStandard: false}
	errHook, _ := hookstderr.New(stderrOpt, []logrus.Level{
	    logrus.ErrorLevel,
	    logrus.FatalLevel,
	    logrus.PanicLevel,
	}, nil)

	// Info to stdout
	stdoutOpt := &config.OptionsStd{DisableStandard: false}
	infoHook, _ := hookstdout.New(stdoutOpt, []logrus.Level{
	    logrus.InfoLevel,
	    logrus.DebugLevel,
	}, nil)

	logger.AddHook(errHook)
	logger.AddHook(infoHook)
	// Errors and info now properly separated

Structured Error Logging with JSON:

	opt := &config.OptionsStd{
	    DisableStandard: false,
	    DisableColor:    true,  // No color in JSON
	}
	hook, _ := hookstderr.New(opt, nil, &logrus.JSONFormatter{})
	logger.AddHook(hook)

	logger.WithFields(logrus.Fields{
	    "error_code": "E500",
	    "request_id": "abc123",
	}).Error("Database connection failed")
	// JSON error message written to stderr

Error Messages Without Clutter:

	opt := &config.OptionsStd{
	    DisableStandard:  false,
	    DisableStack:     true,  // No stack traces
	    DisableTimestamp: true,  // No timestamps
	    EnableTrace:      false, // No caller info
	}
	hook, _ := hookstderr.New(opt, nil, &logrus.TextFormatter{})
	logger.AddHook(hook)

	logger.Error("Configuration file not found")
	// Clean error message without extra fields

Testing with Buffer:

	var buf bytes.Buffer
	opt := &config.OptionsStd{DisableStandard: false}
	hook, _ := hookstderr.NewWithWriter(&buf, opt, nil, nil)
	logger.AddHook(hook)

	logger.Error("test error")
	// Error written to buffer for assertion
	assert.Contains(t, buf.String(), "test error")

# Performance Considerations

Memory Efficiency:
  - Entry duplication uses entry.Dup() which shares data structures
  - Field filtering modifies duplicated entry without new allocations
  - Disabled hooks (DisableStandard=true) have zero allocation cost

Write Performance:
  - Direct writes to stderr (unbuffered by default)
  - Consider using bufio.Writer wrapper for high-frequency logging
  - Network writers should have reasonable timeouts

Formatter Overhead:
  - JSON formatters are faster but produce larger output
  - Text formatters are slower but more human-readable
  - Access log mode bypasses formatting (fastest)

Scalability:
  - Hooks are called synchronously for each entry
  - Multiple hooks add cumulative overhead
  - For high-throughput scenarios, use aggregation:
    github.com/nabbar/golib/ioutils/aggregator

# Thread Safety

The hook implementation is thread-safe when used correctly:

  - Safe: Multiple goroutines logging to the same logger with this hook
  - Safe: Multiple hooks registered on the same logger
  - Unsafe: Concurrent calls to Fire() with the same entry (logrus prevents this)
  - Unsafe: Modifying configuration after hook creation (immutable design)

The underlying writer (os.Stderr) is thread-safe for writes in most operating systems,
but custom writers must ensure thread safety if used concurrently.

# Error Handling

The hook can return errors in the following situations:

Construction Errors:

	// None - hookstderr.New never returns errors directly
	// Errors propagate from underlying hookwriter.New (e.g., if extended in future)

Runtime Errors:

	// Formatter error during Fire()
	err := hook.Fire(entry)  // Returns formatter.Format() error

	// Writer error during Fire()
	err := hook.Fire(entry)  // Returns writer.Write() error

Silent Behaviors:

  - Empty log data: Fire() returns nil without writing (normal)
  - Empty access log message: Fire() returns nil without writing (normal)
  - Disabled hook: New() returns (nil, nil) - not an error

# Limitations

 1. **Synchronous Writes**: Hook writes are synchronous with log calls. Slow stderr blocks logging.
    Mitigation: Use aggregator package for async writes or buffered writers.

 2. **No Write Retries**: Failed writes return errors but don't retry or queue.
    Mitigation: Use reliable writers or add retry logic in custom writers.

 3. **No Buffer Management**: Hook doesn't buffer or flush data.
    Mitigation: Use bufio.Writer and call Flush() explicitly when needed.

 4. **Writer Lifecycle**: Hook doesn't manage writer Close().
    Mitigation: Caller must close custom writers when done. Not an issue - proper design.

# Best Practices

DO:
  - Use stderr hook for error, warning, fatal, and panic level logs
  - Separate stdout and stderr hooks for clean separation of concerns
  - Enable color for terminal output, disable for file/pipe redirection
  - Use level filtering to route different severities appropriately
  - Test with custom writers (buffers) to avoid stderr pollution in tests
  - Check for nil when DisableStandard is conditionally true

DON'T:
  - Mix error and info output on the same stream without good reason
  - Use unbuffered stderr in extremely high-frequency error scenarios
  - Ignore errors from New() (though rare, check for future compatibility)
  - Share non-thread-safe custom writers across multiple hooks
  - Modify opt struct after passing to New() (not effective, options copied)

# Testing

The package includes comprehensive tests covering:

  - Hook creation with various configurations
  - Field filtering (stack, time, caller, file, line)
  - Access log mode with empty messages
  - Formatter integration (JSON, Text)
  - Integration with logrus.Logger
  - Level filtering behavior
  - Multiple hooks on single logger
  - Writer interface compliance
  - RegisterHook and Run methods

Run tests:

	go test -v github.com/nabbar/golib/logger/hookstderr

Check coverage:

	go test -cover github.com/nabbar/golib/logger/hookstderr

Run with race detector:

	CGO_ENABLED=1 go test -race github.com/nabbar/golib/logger/hookstderr

# Examples

See example_test.go for runnable examples demonstrating:
  - Basic stderr hook creation and usage
  - Error-specific logging with level filtering
  - Field filtering configurations
  - JSON formatted error logs
  - Testing with custom writers
  - Access log mode for clean error messages
  - Integration with standard logrus workflows

# Related Packages

  - github.com/sirupsen/logrus - Underlying logging framework
  - github.com/mattn/go-colorable - Cross-platform color support
  - github.com/nabbar/golib/logger - Main logger package
  - github.com/nabbar/golib/logger/config - Configuration types
  - github.com/nabbar/golib/logger/types - Hook interface and constants
  - github.com/nabbar/golib/logger/hookwriter - Core hook implementation
  - github.com/nabbar/golib/logger/hookstdout - Stdout equivalent package
  - github.com/nabbar/golib/ioutils/aggregator - Async write aggregation

# License

MIT License - See LICENSE file for details.

Copyright (c) 2025 Nicolas JUHEL
*/
package hookstderr
