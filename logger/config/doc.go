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

// Package config provides configuration structures and validation for the logger package.
//
// # Overview
//
// The config package defines the configuration model for golib/logger, supporting multiple
// output destinations (stdout, files, syslog) with independent formatting options.
// It provides a flexible, inheritance-based configuration system with validation.
//
// Key features:
//   - Multiple output destinations: stdout/stderr, files, syslog (local or remote)
//   - Per-destination log level filtering
//   - Flexible formatting options per output
//   - Configuration inheritance from defaults
//   - JSON, YAML, TOML compatible structures
//   - Validation with detailed error reporting
//
// # Design Philosophy
//
// The package follows these design principles:
//
//  1. Separation of Concerns: Each output type (std, file, syslog) has its own
//     dedicated options structure with relevant configuration fields.
//
//  2. Inheritance: Options can inherit from a default configuration, allowing
//     centralized base settings with per-instance overrides.
//
//  3. Extensibility: The LogFileExtend and LogSyslogExtend flags allow extending
//     default configurations rather than replacing them completely.
//
//  4. Validation: Built-in validation ensures configuration consistency before
//     logger instantiation, preventing runtime errors.
//
//  5. Marshaling: All structures use standard tags (json, yaml, toml, mapstructure)
//     for seamless integration with configuration management systems.
//
// # Architecture
//
// Configuration Flow:
//
//	┌─────────────────┐
//	│  Default Config │ (optional, via RegisterDefaultFunc)
//	└────────┬────────┘
//	         │ (if InheritDefault = true)
//	         ▼
//	┌─────────────────┐
//	│     Options     │──┐
//	├─────────────────┤  │
//	│ TraceFilter     │  │
//	│ Stdout          │  │
//	│ LogFile[]       │  │ Merge logic
//	│ LogSyslog[]     │  │ (Options(), Merge())
//	└────────┬────────┘  │
//	         │           │
//	         ▼           │
//	┌─────────────────┐  │
//	│ Final Options   │◄─┘
//	└────────┬────────┘
//	         │
//	         ▼
//	   Validation (Validate())
//	         │
//	         ▼
//	   Logger Creation
//
// Component Relationships:
//
//	Options
//	  ├── Stdout: OptionsStd (single instance)
//	  ├── LogFile: OptionsFiles (slice, can extend)
//	  └── LogSyslog: OptionsSyslogs (slice, can extend)
//
// Each option type handles its own cloning, merging, and configuration logic.
//
// # Basic Usage
//
// Creating a simple stdout-only configuration:
//
//	opts := &config.Options{
//	    Stdout: &config.OptionsStd{
//	        DisableStandard:  false,  // Enable stdout
//	        DisableStack:     true,   // No goroutine ID
//	        DisableTimestamp: false,  // Show timestamps
//	        EnableTrace:      true,   // Show caller info
//	        DisableColor:     false,  // Allow colors (if TTY)
//	        EnableAccessLog:  false,  // No HTTP access logs
//	    },
//	}
//
//	// Validate before use
//	if err := opts.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// # File Logging
//
// Configuring log file output:
//
//	opts := &config.Options{
//	    LogFile: config.OptionsFiles{
//	        {
//	            LogLevel:         []string{"Error", "Fatal", "Critical"},
//	            Filepath:         "/var/log/app/errors.log",
//	            Create:           true,  // Create if not exists
//	            CreatePath:       true,  // Create directories
//	            FileMode:         0644,  // File permissions
//	            PathMode:         0755,  // Directory permissions
//	            DisableStack:     false,
//	            DisableTimestamp: false,
//	            EnableTrace:      true,
//	            EnableAccessLog:  false,
//	        },
//	    },
//	}
//
// Multiple files with different log levels:
//
//	opts := &config.Options{
//	    LogFile: config.OptionsFiles{
//	        // All logs
//	        {
//	            LogLevel: []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
//	            Filepath: "/var/log/app/all.log",
//	            Create:   true,
//	        },
//	        // Errors only
//	        {
//	            LogLevel: []string{"Error", "Fatal", "Critical"},
//	            Filepath: "/var/log/app/errors.log",
//	            Create:   true,
//	        },
//	    },
//	}
//
// # Syslog Integration
//
// Local syslog (Unix socket):
//
//	opts := &config.Options{
//	    LogSyslog: config.OptionsSyslogs{
//	        {
//	            LogLevel: []string{"Info", "Warning", "Error", "Fatal", "Critical"},
//	            Network:  "",  // Empty for local syslog
//	            Host:     "",  // Empty for local syslog
//	            Facility: "local0",
//	            Tag:      "myapp",
//	        },
//	    },
//	}
//
// Remote syslog (TCP or UDP):
//
//	opts := &config.Options{
//	    LogSyslog: config.OptionsSyslogs{
//	        {
//	            LogLevel: []string{"Error", "Fatal", "Critical"},
//	            Network:  "tcp",  // or "udp"
//	            Host:     "syslog.example.com:514",
//	            Facility: "local0",
//	            Tag:      "myapp-prod",
//	        },
//	    },
//	}
//
// # Configuration Inheritance
//
// Using default configuration with overrides:
//
//	// Define base configuration
//	defaultConfig := func() *config.Options {
//	    return &config.Options{
//	        Stdout: &config.OptionsStd{
//	            EnableTrace: true,
//	            DisableStack: true,
//	        },
//	        LogFile: config.OptionsFiles{
//	            {
//	                Filepath: "/var/log/app/default.log",
//	                Create:   true,
//	            },
//	        },
//	    }
//	}
//
//	// Create specific configuration
//	opts := &config.Options{
//	    InheritDefault: true,  // Enable inheritance
//	    TraceFilter:    "/myproject/",  // Override trace filter
//	    LogFileExtend:  true,  // Extend (not replace) default files
//	    LogFile: config.OptionsFiles{
//	        {
//	            Filepath: "/var/log/app/specific.log",
//	            Create:   true,
//	        },
//	    },
//	}
//	opts.RegisterDefaultFunc(defaultConfig)
//
//	// Get final merged configuration
//	final := opts.Options()
//	// final.LogFile will contain BOTH default.log and specific.log
//
// # Extend vs Replace
//
// The LogFileExtend and LogSyslogExtend flags control merging behavior:
//
// Replace mode (default, extend = false):
//
//	base := &config.Options{
//	    LogFile: config.OptionsFiles{
//	        {Filepath: "/tmp/base.log"},
//	    },
//	}
//
//	override := &config.Options{
//	    LogFileExtend: false,  // Replace
//	    LogFile: config.OptionsFiles{
//	        {Filepath: "/tmp/override.log"},
//	    },
//	}
//
//	base.Merge(override)
//	// base.LogFile now contains ONLY override.log
//
// Extend mode:
//
//	override := &config.Options{
//	    LogFileExtend: true,  // Extend
//	    LogFile: config.OptionsFiles{
//	        {Filepath: "/tmp/additional.log"},
//	    },
//	}
//
//	base.Merge(override)
//	// base.LogFile now contains BOTH base.log and additional.log
//
// # Configuration Sources
//
// JSON example:
//
//	{
//	  "inheritDefault": false,
//	  "traceFilter": "/myproject/",
//	  "stdout": {
//	    "disableStandard": false,
//	    "enableTrace": true
//	  },
//	  "logFile": [
//	    {
//	      "logLevel": ["Error", "Fatal"],
//	      "filepath": "/var/log/errors.log",
//	      "create": true
//	    }
//	  ]
//	}
//
// YAML example:
//
//	inheritDefault: false
//	traceFilter: /myproject/
//	stdout:
//	  disableStandard: false
//	  enableTrace: true
//	logFile:
//	  - logLevel: [Error, Fatal]
//	    filepath: /var/log/errors.log
//	    create: true
//
// # Default Configuration
//
// Retrieving the default configuration template:
//
//	// Get default config as JSON
//	defaultJSON := config.DefaultConfig("")
//
//	// Get formatted default config
//	prettyJSON := config.DefaultConfig("  ")
//
//	// Parse into Options
//	var opts config.Options
//	if err := json.Unmarshal(defaultJSON, &opts); err != nil {
//	    log.Fatal(err)
//	}
//
// Customizing defaults:
//
//	customDefault := []byte(`{
//	    "inheritDefault": false,
//	    "stdout": {
//	        "enableTrace": true
//	    }
//	}`)
//
//	config.SetDefaultConfig(customDefault)
//
// # Validation
//
// The Validate method uses go-playground/validator for struct validation:
//
//	opts := &config.Options{
//	    // ... configuration
//	}
//
//	if err := opts.Validate(); err != nil {
//	    // err is a liberr.Error containing all validation failures
//	    fmt.Println("Validation errors:", err)
//	    return
//	}
//
// Validation checks include:
//   - Required field presence (implicitly via struct tags)
//   - Field format constraints (e.g., file paths, network addresses)
//   - Cross-field dependencies
//
// # Trace Filtering
//
// The TraceFilter field allows cleaning file paths in stack traces:
//
//	opts := &config.Options{
//	    TraceFilter: "/go/src/github.com/myproject/",
//	    Stdout: &config.OptionsStd{
//	        EnableTrace: true,
//	    },
//	}
//
// With TraceFilter:
//
//	main.go:42 instead of /go/src/github.com/myproject/main.go:42
//
// This shortens log messages and removes environment-specific paths.
//
// # Log Levels
//
// Available log levels for filtering (case-sensitive):
//   - Debug: Detailed debugging information
//   - Info: General informational messages
//   - Warning: Warning messages for unusual situations
//   - Error: Error messages for failures
//   - Fatal: Critical errors causing process termination
//   - Critical: Severe errors requiring immediate attention
//
// Configuration per output:
//
//	// File receives only errors
//	file := config.OptionsFile{
//	    LogLevel: []string{"Error", "Fatal", "Critical"},
//	    Filepath: "/var/log/errors.log",
//	}
//
//	// Syslog receives everything
//	syslog := config.OptionsSyslog{
//	    LogLevel: []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
//	    Host:     "syslog.example.com:514",
//	}
//
// # Formatting Options
//
// Each output type supports formatting flags:
//
//   - DisableStack: Hide goroutine ID before each message
//   - DisableTimestamp: Hide timestamp before each message
//   - EnableTrace: Add caller file/line information
//   - EnableAccessLog: Include HTTP access logs (for web applications)
//   - DisableColor (stdout only): Disable colored output
//   - DisableStandard (stdout only): Completely disable stdout/stderr
//
// Example with different formatting:
//
//	opts := &config.Options{
//	    Stdout: &config.OptionsStd{
//	        DisableStack:     true,   // Clean output
//	        DisableTimestamp: false,  // Keep timestamps
//	        EnableTrace:      false,  // No caller info
//	        DisableColor:     false,  // Colors in terminal
//	    },
//	    LogFile: config.OptionsFiles{
//	        {
//	            DisableStack:     false,  // Include goroutine ID
//	            DisableTimestamp: false,  // Include timestamps
//	            EnableTrace:      true,   // Include caller info
//	            Filepath:         "/var/log/app.log",
//	        },
//	    },
//	}
//
// # Cloning and Merging
//
// Cloning creates independent copies:
//
//	original := &config.Options{
//	    TraceFilter: "/original/",
//	    Stdout:      &config.OptionsStd{EnableTrace: true},
//	}
//
//	clone := original.Clone()
//	clone.TraceFilter = "/modified/"  // Doesn't affect original
//
// Merging combines configurations:
//
//	base := &config.Options{
//	    Stdout: &config.OptionsStd{
//	        EnableTrace: true,
//	    },
//	}
//
//	override := &config.Options{
//	    TraceFilter: "/project/",
//	    Stdout: &config.OptionsStd{
//	        DisableColor: true,
//	    },
//	}
//
//	base.Merge(override)
//	// base.Stdout now has: EnableTrace=true, DisableColor=true
//	// base.TraceFilter is "/project/"
//
// # Error Handling
//
// The package defines two error codes:
//
//   - ErrorParamEmpty: Provided parameter is empty or nil
//   - ErrorValidatorError: Configuration validation failed
//
// Error usage:
//
//	err := config.ErrorParamEmpty.Error(nil)
//	if err != nil {
//	    log.Println("Error:", err)
//	}
//
//	// With validation
//	opts := &config.Options{}
//	if validationErr := opts.Validate(); validationErr != nil {
//	    // validationErr is of type liberr.Error
//	    // Contains all validation failures
//	    return config.ErrorValidatorError.Error(validationErr)
//	}
//
// # Best Practices
//
// 1. Always validate configuration before use:
//
//	if err := opts.Validate(); err != nil {
//	    return err
//	}
//
// 2. Use specific log levels per output:
//
//	// Stdout: everything for development
//	// File: errors only for production analysis
//	// Syslog: critical events for monitoring
//
// 3. Enable trace only when needed:
//
//	// Development: EnableTrace = true
//	// Production: EnableTrace = false (performance)
//
// 4. Use TraceFilter to clean paths:
//
//	opts.TraceFilter = os.Getenv("GOPATH") + "/src/myproject/"
//
// 5. Consider file permissions carefully:
//
//	FileMode: 0640,  // Owner read/write, group read
//	PathMode: 0750,  // Owner full, group read/execute
//
// 6. Use inheritance for DRY configuration:
//
//	// One base config, multiple specialized configs
//	opts.InheritDefault = true
//	opts.RegisterDefaultFunc(getBaseConfig)
//
// # Performance Considerations
//
// Configuration impact on performance:
//
//   - Multiple outputs: Each active output adds processing overhead
//   - Trace enabled: Adds ~10-20% CPU overhead (stack trace capture)
//   - File I/O: Disk writes can be slow; consider buffering
//   - Remote syslog: Network latency affects logging throughput
//   - Color formatting: Minimal overhead (~1-2%)
//
// Optimization tips:
//
//  1. Disable outputs not needed in production (e.g., debug files)
//  2. Use appropriate log levels to reduce volume
//  3. Disable trace in production unless troubleshooting
//  4. Consider async logging for high-volume applications
//  5. Use local syslog instead of remote when possible
//
// # Thread Safety
//
// Configuration structures are not thread-safe for modification but support:
//
//   - Safe cloning: Create independent copies
//   - Safe merging: Merge configurations before concurrent use
//   - Safe reading: Once created, Options can be read concurrently
//
// Pattern for concurrent use:
//
//	// Setup phase (single-threaded)
//	opts := &config.Options{...}
//	final := opts.Options()  // Finalize
//
//	// Runtime (multi-threaded)
//	// Read final.LogFile, final.Stdout, etc. safely
//	// Don't modify final after this point
//
// # Integration Examples
//
// See example_test.go for runnable examples including:
//   - Basic stdout logging
//   - Multi-file configuration
//   - Syslog integration (local and remote)
//   - Configuration inheritance
//   - Merge strategies
//   - JSON/YAML loading
//
// # Limitations
//
//  1. No dynamic reconfiguration: Changes require logger restart
//  2. No log rotation: Use external tools (logrotate) or file hooks
//  3. No compression: Compress rotated logs separately
//  4. No filtering by message content: Only by level
//  5. Windows event log: Requires specific syslog configuration
//
// # Compatibility
//
// Minimum Go version: 1.18 (requires generics support in dependencies)
//
// Supported platforms:
//   - Linux (all architectures)
//   - macOS (darwin/amd64, darwin/arm64)
//   - Windows (amd64, 386) - limited syslog support
//   - FreeBSD, OpenBSD, NetBSD (amd64)
//
// Configuration format compatibility:
//   - JSON (encoding/json)
//   - YAML (gopkg.in/yaml.v3)
//   - TOML (github.com/pelletier/go-toml)
//   - Viper (github.com/spf13/viper) via mapstructure tags
package config
