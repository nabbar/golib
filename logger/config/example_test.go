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

package config_test

import (
	"encoding/json"
	"fmt"

	libprm "github.com/nabbar/golib/file/perm"
	"github.com/nabbar/golib/logger/config"
)

// ExampleOptions_basic demonstrates basic stdout logging configuration.
// This is the simplest setup for logging to standard output only.
func ExampleOptions_basic() {
	opts := &config.Options{
		Stdout: &config.OptionsStd{
			DisableStandard:  false, // Enable stdout
			DisableStack:     true,  // Hide goroutine ID
			DisableTimestamp: false, // Show timestamps
			EnableTrace:      false, // No caller info (faster)
			DisableColor:     false, // Allow colors (if TTY)
		},
	}

	// Validate configuration
	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Println("Basic stdout configuration created successfully")
	// Output: Basic stdout configuration created successfully
}

// ExampleOptions_fileLogging demonstrates logging to files with different levels.
// Shows how to separate logs by severity into different files.
func ExampleOptions_fileLogging() {
	// Create file mode permissions
	fileMode, _ := libprm.Parse("0644")
	pathMode, _ := libprm.Parse("0755")

	opts := &config.Options{
		LogFile: config.OptionsFiles{
			// All logs to main file
			{
				LogLevel:         []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
				Filepath:         "/var/log/app/all.log",
				Create:           true,
				CreatePath:       true,
				FileMode:         fileMode,
				PathMode:         pathMode,
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
			},
			// Errors only to separate file
			{
				LogLevel:         []string{"Error", "Fatal", "Critical"},
				Filepath:         "/var/log/app/errors.log",
				Create:           true,
				CreatePath:       true,
				FileMode:         fileMode,
				PathMode:         pathMode,
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
			},
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("File logging configured with %d outputs\n", len(opts.LogFile))
	// Output: File logging configured with 2 outputs
}

// ExampleOptions_syslogLocal demonstrates local syslog configuration.
// Shows how to send logs to the local system syslog daemon.
func ExampleOptions_syslogLocal() {
	opts := &config.Options{
		LogSyslog: config.OptionsSyslogs{
			{
				LogLevel: []string{"Info", "Warning", "Error", "Fatal", "Critical"},
				Network:  "", // Empty for local syslog
				Host:     "", // Empty for local syslog
				Facility: "local0",
				Tag:      "myapp",
			},
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Println("Local syslog configuration created")
	// Output: Local syslog configuration created
}

// ExampleOptions_syslogRemote demonstrates remote syslog configuration.
// Shows how to send logs to a remote syslog server over TCP or UDP.
func ExampleOptions_syslogRemote() {
	opts := &config.Options{
		LogSyslog: config.OptionsSyslogs{
			{
				LogLevel: []string{"Error", "Fatal", "Critical"},
				Network:  "tcp",
				Host:     "syslog.example.com:514",
				Facility: "local0",
				Tag:      "myapp-prod",
			},
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Println("Remote syslog configuration created")
	// Output: Remote syslog configuration created
}

// ExampleOptions_multiOutput demonstrates multiple output destinations.
// Shows combining stdout, files, and syslog in one configuration.
func ExampleOptions_multiOutput() {
	fileMode, _ := libprm.Parse("0644")
	pathMode, _ := libprm.Parse("0755")

	opts := &config.Options{
		TraceFilter: "/myproject/",
		Stdout: &config.OptionsStd{
			DisableStandard: false,
			EnableTrace:     true,
		},
		LogFile: config.OptionsFiles{
			{
				LogLevel: []string{"Error", "Fatal", "Critical"},
				Filepath: "/var/log/app/errors.log",
				Create:   true,
				FileMode: fileMode,
				PathMode: pathMode,
			},
		},
		LogSyslog: config.OptionsSyslogs{
			{
				LogLevel: []string{"Fatal", "Critical"},
				Facility: "local0",
				Tag:      "myapp",
			},
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("Multi-output: stdout + %d files + %d syslog\n",
		len(opts.LogFile), len(opts.LogSyslog))
	// Output: Multi-output: stdout + 1 files + 1 syslog
}

// ExampleOptions_inheritance demonstrates configuration inheritance.
// Shows how to use default configurations with overrides.
func ExampleOptions_inheritance() {
	// Define default configuration
	defaultConfig := func() *config.Options {
		return &config.Options{
			Stdout: &config.OptionsStd{
				EnableTrace:  true,
				DisableStack: true,
			},
		}
	}

	// Create specific configuration with inheritance
	opts := &config.Options{
		InheritDefault: true,
		TraceFilter:    "/myproject/",
		Stdout: &config.OptionsStd{
			DisableColor: true, // Override default
		},
	}
	opts.RegisterDefaultFunc(defaultConfig)

	// Get final merged configuration
	final := opts.Options()

	fmt.Printf("Inherited - EnableTrace: %v, DisableColor: %v\n",
		final.Stdout.EnableTrace, final.Stdout.DisableColor)
	// Output: Inherited - EnableTrace: true, DisableColor: true
}

// ExampleOptions_Merge demonstrates merging configurations.
// Shows how to combine two configuration objects.
func ExampleOptions_Merge() {
	base := &config.Options{
		Stdout: &config.OptionsStd{
			EnableTrace: true,
		},
	}

	override := &config.Options{
		TraceFilter: "/project/",
		Stdout: &config.OptionsStd{
			DisableColor: true,
		},
	}

	base.Merge(override)

	fmt.Printf("Merged - TraceFilter: %s, EnableTrace: %v, DisableColor: %v\n",
		base.TraceFilter, base.Stdout.EnableTrace, base.Stdout.DisableColor)
	// Output: Merged - TraceFilter: /project/, EnableTrace: true, DisableColor: true
}

// ExampleOptions_Clone demonstrates cloning configurations.
// Shows how to create independent copies of configurations.
func ExampleOptions_Clone() {
	original := &config.Options{
		TraceFilter: "/original/",
		Stdout: &config.OptionsStd{
			EnableTrace: true,
		},
	}

	clone := original.Clone()
	clone.TraceFilter = "/modified/"

	fmt.Printf("Original: %s, Clone: %s\n", original.TraceFilter, clone.TraceFilter)
	// Output: Original: /original/, Clone: /modified/
}

// ExampleOptionsFile_Clone demonstrates cloning file options.
// Shows how to create independent copies of file configurations.
func ExampleOptionsFile_Clone() {
	fileMode, _ := libprm.Parse("0644")

	original := config.OptionsFile{
		LogLevel: []string{"Error", "Fatal"},
		Filepath: "/var/log/original.log",
		Create:   true,
		FileMode: fileMode,
	}

	clone := original.Clone()
	clone.Filepath = "/var/log/clone.log"

	fmt.Printf("Original: %s, Clone: %s\n", original.Filepath, clone.Filepath)
	// Output: Original: /var/log/original.log, Clone: /var/log/clone.log
}

// ExampleOptionsSyslog_Clone demonstrates cloning syslog options.
// Shows how to create independent copies of syslog configurations.
func ExampleOptionsSyslog_Clone() {
	original := config.OptionsSyslog{
		LogLevel: []string{"Error", "Fatal"},
		Network:  "tcp",
		Host:     "syslog1.example.com:514",
		Tag:      "original",
	}

	clone := original.Clone()
	clone.Host = "syslog2.example.com:514"
	clone.Tag = "clone"

	fmt.Printf("Original: %s, Clone: %s\n", original.Tag, clone.Tag)
	// Output: Original: original, Clone: clone
}

// ExampleDefaultConfig demonstrates using default configuration template.
// Shows how to retrieve and use the built-in default configuration.
func ExampleDefaultConfig() {
	// Get default configuration as JSON
	defaultJSON := config.DefaultConfig("")

	// Parse into Options structure
	var opts config.Options
	if err := json.Unmarshal(defaultJSON, &opts); err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Default has stdout: %v, %d files, %d syslogs\n",
		opts.Stdout != nil, len(opts.LogFile), len(opts.LogSyslog))
	// Output: Default has stdout: true, 1 files, 1 syslogs
}

// ExampleSetDefaultConfig demonstrates customizing default configuration.
// Shows how to replace the built-in default configuration template.
func ExampleSetDefaultConfig() {
	// Save original
	original := config.DefaultConfig("")

	// Set custom default
	custom := []byte(`{
		"inheritDefault": false,
		"stdout": {
			"enableTrace": true,
			"disableColor": true
		}
	}`)
	config.SetDefaultConfig(custom)

	// Use custom default
	newDefault := config.DefaultConfig("")

	// Restore original
	config.SetDefaultConfig(original)

	fmt.Println("Custom default configuration set successfully")
	// Output: Custom default configuration set successfully
	_ = newDefault
}

// ExampleOptions_extendMode demonstrates extend vs replace mode.
// Shows the difference between extending and replacing configurations.
func ExampleOptions_extendMode() {
	fileMode, _ := libprm.Parse("0644")

	base := &config.Options{
		LogFile: config.OptionsFiles{
			{
				Filepath: "/var/log/base.log",
				Create:   true,
				FileMode: fileMode,
			},
		},
	}

	// Replace mode (default)
	replaceOpts := &config.Options{
		LogFileExtend: false,
		LogFile: config.OptionsFiles{
			{
				Filepath: "/var/log/replace.log",
				Create:   true,
				FileMode: fileMode,
			},
		},
	}

	baseClone := base.Clone()
	baseClone.Merge(replaceOpts)
	fmt.Printf("Replace mode: %d files\n", len(baseClone.LogFile))

	// Extend mode
	extendOpts := &config.Options{
		LogFileExtend: true,
		LogFile: config.OptionsFiles{
			{
				Filepath: "/var/log/extend.log",
				Create:   true,
				FileMode: fileMode,
			},
		},
	}

	base.Merge(extendOpts)
	fmt.Printf("Extend mode: %d files\n", len(base.LogFile))
	// Output:
	// Replace mode: 1 files
	// Extend mode: 2 files
}

// ExampleOptions_jsonConfiguration demonstrates JSON configuration.
// Shows how to load configuration from JSON format.
func ExampleOptions_jsonConfiguration() {
	jsonConfig := []byte(`{
		"inheritDefault": false,
		"traceFilter": "/myproject/",
		"stdout": {
			"disableStandard": false,
			"enableTrace": true,
			"disableColor": false
		},
		"logFile": [
			{
				"logLevel": ["Error", "Fatal", "Critical"],
				"filepath": "/var/log/errors.log",
				"create": true,
				"fileMode": "0644",
				"pathMode": "0755"
			}
		]
	}`)

	var opts config.Options
	if err := json.Unmarshal(jsonConfig, &opts); err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("JSON config loaded: TraceFilter=%s, Files=%d\n",
		opts.TraceFilter, len(opts.LogFile))
	// Output: JSON config loaded: TraceFilter=/myproject/, Files=1
}

// ExampleOptions_developmentSetup demonstrates development configuration.
// Shows a typical setup for development with verbose logging.
func ExampleOptions_developmentSetup() {
	opts := &config.Options{
		TraceFilter: "/myproject/",
		Stdout: &config.OptionsStd{
			DisableStandard:  false,
			DisableStack:     false, // Show goroutine IDs
			DisableTimestamp: false,
			EnableTrace:      true, // Show caller info
			DisableColor:     false,
			EnableAccessLog:  false,
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Println("Development configuration: verbose, colored, with traces")
	// Output: Development configuration: verbose, colored, with traces
}

// ExampleOptions_productionSetup demonstrates production configuration.
// Shows a typical setup for production with optimized logging.
func ExampleOptions_productionSetup() {
	fileMode, _ := libprm.Parse("0640")
	pathMode, _ := libprm.Parse("0750")

	opts := &config.Options{
		Stdout: &config.OptionsStd{
			DisableStandard: true, // No stdout in production
		},
		LogFile: config.OptionsFiles{
			// Application logs
			{
				LogLevel:         []string{"Info", "Warning", "Error", "Fatal", "Critical"},
				Filepath:         "/var/log/app/application.log",
				Create:           true,
				CreatePath:       true,
				FileMode:         fileMode,
				PathMode:         pathMode,
				DisableStack:     true,  // Clean logs
				DisableTimestamp: false, // Keep timestamps
				EnableTrace:      false, // No trace for performance
			},
			// Critical errors only
			{
				LogLevel:         []string{"Error", "Fatal", "Critical"},
				Filepath:         "/var/log/app/errors.log",
				Create:           true,
				CreatePath:       true,
				FileMode:         fileMode,
				PathMode:         pathMode,
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true, // Trace for errors
			},
		},
		LogSyslog: config.OptionsSyslogs{
			{
				LogLevel: []string{"Fatal", "Critical"},
				Facility: "local0",
				Tag:      "myapp-prod",
			},
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("Production setup: %d files + %d syslog\n",
		len(opts.LogFile), len(opts.LogSyslog))
	// Output: Production setup: 2 files + 1 syslog
}

// ExampleOptions_Validate demonstrates configuration validation.
// Shows how validation catches configuration errors.
func ExampleOptions_Validate() {
	opts := &config.Options{
		Stdout: &config.OptionsStd{
			DisableStandard: false,
			EnableTrace:     true,
		},
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Configuration is valid")
	// Output: Configuration is valid
}
