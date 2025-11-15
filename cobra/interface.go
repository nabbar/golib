/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

// Package cobra provides a comprehensive, production-ready wrapper around spf13/cobra
// for building professional CLI applications in Go.
//
// This package simplifies CLI application development by providing:
//   - Zero-boilerplate setup for complete CLI applications
//   - Thread-safe, instance-based design (no global state)
//   - Type-safe flags for 20+ Go types with compile-time safety
//   - Auto-generated shell completion (bash, zsh, fish, PowerShell)
//   - Auto-generated configuration file support
//   - Integrated logging and version management
//
// Design Philosophy:
//
// 1. Instance-Based Architecture: Unlike raw cobra which often relies on global variables,
// this package uses instance-based design ensuring thread-safety and testability.
//
// 2. Progressive Enhancement: Start with a basic CLI, then add features (completion,
// config generation) with single function calls.
//
// 3. Type Safety: Comprehensive flag support with Go's type system preventing runtime errors.
//
// Quick Example:
//
//	package main
//
//	import (
//	    libcbr "github.com/nabbar/golib/cobra"
//	    libver "github.com/nabbar/golib/version"
//	)
//
//	func main() {
//	    app := libcbr.New()
//	    app.SetVersion(version)
//	    app.Init()
//
//	    var config string
//	    app.SetFlagConfig(true, &config)
//	    app.AddCommandCompletion()
//
//	    app.Execute()
//	}
//
// Built-in Features:
//   - Shell completion generation (--completion bash|zsh|fish|powershell)
//   - Configuration file generation (--configure json|yaml|toml)
//   - Error code printing (--print-error-code)
//   - Version information (--version)
//
// For more examples and detailed documentation, see README.md.
package cobra

import (
	"io"
	"time"

	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

// FuncInit is a function type for initialization callbacks.
// It is called during cobra initialization to perform custom setup operations.
type FuncInit func()

// FuncLogger is a function type that returns a logger instance.
// This pattern allows for lazy initialization of the logger and dependency injection.
type FuncLogger func() liblog.Logger

// FuncViper is a function type that returns a viper configuration instance.
// This pattern allows for lazy initialization and dependency injection of configuration.
type FuncViper func() libvpr.Viper

// FuncPrintErrorCode is a function type for printing error code information.
// It receives an item identifier and its corresponding value for display.
type FuncPrintErrorCode func(item, value string)

// Cobra is the main interface for building CLI applications with enhanced features.
// It wraps spf13/cobra with additional functionality for logging, configuration,
// and version management while maintaining thread-safety through instance-based design.
type Cobra interface {
	// SetVersion sets the version of the Cobra application. The version is used
	// to construct the help message and the version message.
	//
	// The SetVersion function does not return any value. It is designed to be used as a
	// setter function, which means that the function will not return until the Cobra
	// application's logic has been set.
	//
	// If an error occurs while setting the version, the function will panic.
	SetVersion(v libver.Version)

	// SetFuncInit sets the initialization function. The initialization function is called
	// when the Cobra application is initialized. The initialization function takes
	// no arguments and returns no value.
	//
	// The SetFuncInit function does not return any value. It is designed to be used as a
	// setter function, which means that the function will not return until the Cobra
	// application's logic has been set.
	//
	// If an error occurs while setting the initialization function, the error is returned.
	//
	// The SetFuncInit function is safe to call concurrently.
	SetFuncInit(fct FuncInit)
	// SetViper sets the viper function. The viper function is called when the Cobra
	// application needs to access the viper configuration. The viper function takes
	// no arguments and returns a viper configuration.
	//
	// The SetViper function does not return any value. It is designed to be used as a
	// setter function, which means that the function will not return until the Cobra
	// application's logic has been set.
	//
	// If an error occurs while setting the viper, the error is returned.
	SetViper(fct FuncViper)
	// SetLogger sets the logger function. The logger function is called when the Cobra
	// application needs to log information. The logger function takes no arguments and
	// returns a logger.
	//
	// The SetLogger function does not return any value. It is designed to be used as a
	// setter function, which means that the function will not return until the Cobra
	// application's logic has been set.
	//
	// If an error occurs while setting the logger, the error is returned.
	//
	// The SetLogger function is safe to call concurrently.
	SetLogger(fct FuncLogger)
	// SetForceNoInfo sets the ForceNoInfo flag. If the flag is set to true, the
	// Cobra application will not print any information related to the
	// package path. If the flag is set to false, the Cobra application will
	// print information related to the package path.
	//
	// The SetForceNoInfo function does not return any value. It is designed to be
	// used as a setter function, which means that the function will not return
	// until the Cobra application's logic has been set.
	SetForceNoInfo(flag bool)

	// Init initializes the Cobra application. It sets the logger, the version,
	// and the force no info flag. The function must be called before executing
	// the Cobra application.
	//
	// If an error occurs while initializing the Cobra application, the
	// error is returned.
	//
	// The Init function does not return any value. It is designed to be used as
	// a setter function, which means that the function will not return until the
	// Cobra application's logic has been set.
	Init()

	// SetFlagConfig adds a string flag to the Cobra application that can be used
	// to configure the application.
	//
	// The SetFlagConfig function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a string
	//
	// If an error occurs while adding the flag, the error is returned.
	SetFlagConfig(persistent bool, flagVar *string) error
	// SetFlagVerbose adds a verbose flag to the Cobra application
	//
	// The SetFlagVerbose function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int
	//
	// The verbose flag is used to enable verbose mode (multi allowed v, vv, vvv)
	//
	// Example:
	// cobra.SetFlagVerbose(true, &verbose)
	SetFlagVerbose(persistent bool, flagVar *int)

	// AddFlagString adds a string flag to the Cobra application
	//
	// The AddFlagString function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a string
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagString(persistent bool, p *string, name, shorthand string, value string, usage string)
	// AddFlagCount adds a count flag to the Cobra application
	//
	// The AddFlagCount function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - usage: a description of the flag
	AddFlagCount(persistent bool, p *int, name, shorthand string, usage string)
	// AddFlagBool adds a bool flag to the Cobra application
	//
	// The AddFlagBool function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a bool
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagBool(persistent bool, p *bool, name, shorthand string, value bool, usage string)
	// AddFlagDuration adds a time.Duration flag to the Cobra application
	//
	// The AddFlagDuration function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a time.Duration
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagDuration(persistent bool, p *time.Duration, name, shorthand string, value time.Duration, usage string)
	// AddFlagFloat32 adds a float32 flag to the Cobra application
	//
	// The AddFlagFloat32 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a float32
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagFloat32(persistent bool, p *float32, name, shorthand string, value float32, usage string)
	// AddFlagFloat64 adds a float64 flag to the Cobra application
	//
	// The AddFlagFloat64 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a float64
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagFloat64(persistent bool, p *float64, name, shorthand string, value float64, usage string)
	// AddFlagInt adds an int flag to the Cobra application
	//
	// The AddFlagInt function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt(persistent bool, p *int, name, shorthand string, value int, usage string)
	// AddFlagInt8 adds an int8 flag to the Cobra application
	//
	// The AddFlagInt8 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int8
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt8(persistent bool, p *int8, name, shorthand string, value int8, usage string)
	// AddFlagInt16 adds an int16 flag to the Cobra application
	//
	// The AddFlagInt16 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int16
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt16(persistent bool, p *int16, name, shorthand string, value int16, usage string)
	// AddFlagInt32 adds an int32 flag to the Cobra application
	//
	// The AddFlagInt32 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int32
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt32(persistent bool, p *int32, name, shorthand string, value int32, usage string)
	// AddFlagInt32Slice adds an int32 slice flag to the Cobra application
	//
	// The AddFlagInt32Slice function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int32 slice
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt32Slice(persistent bool, p *[]int32, name, shorthand string, value []int32, usage string)
	// AddFlagInt64 adds an int64 flag to the Cobra application
	//
	// The AddFlagInt64 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int64
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt64(persistent bool, p *int64, name, shorthand string, value int64, usage string)
	// AddFlagInt64Slice adds an int64 slice flag to the Cobra application
	//
	// The AddFlagInt64Slice function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to an int64 slice
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagInt64Slice(persistent bool, p *[]int64, name, shorthand string, value []int64, usage string)
	// AddFlagUint adds a uint flag to the Cobra application
	//
	// The AddFlagUint function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagUint(persistent bool, p *uint, name, shorthand string, value uint, usage string)
	// AddFlagUintSlice adds a uint slice flag to the Cobra application
	//
	// The AddFlagUintSlice function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint slice
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagUintSlice(persistent bool, p *[]uint, name, shorthand string, value []uint, usage string)
	// AddFlagUint8 adds a uint8 flag to the Cobra application
	//
	// The AddFlagUint8 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint8
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagUint8(persistent bool, p *uint8, name, shorthand string, value uint8, usage string)
	// AddFlagUint16 adds a uint16 flag to the Cobra application
	//
	// The AddFlagUint16 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint16
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagUint16(persistent bool, p *uint16, name, shorthand string, value uint16, usage string)
	// AddFlagUint32 adds a uint32 flag to the Cobra application
	//
	// The AddFlagUint32 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint32
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	//
	// The AddFlagUint32 function does not return any value. It is used
	// to add a uint32 flag to the Cobra application.
	//
	// The AddFlagUint32 function allocates memory for the returned
	// flag. The caller is responsible for freeing this memory when it is no
	// longer needed.
	AddFlagUint32(persistent bool, p *uint32, name, shorthand string, value uint32, usage string)
	// AddFlagUint64 adds a uint64 flag to the Cobra application
	//
	// The AddFlagUint64 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a uint64
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	//
	// The AddFlagUint64 function does not return any value. It is used
	// to add a uint64 flag to the Cobra application.
	//
	// The AddFlagUint64 function allocates memory for the returned
	// flag. The caller is responsible for freeing this memory when it is no
	// longer needed.
	AddFlagUint64(persistent bool, p *uint64, name, shorthand string, value uint64, usage string)
	// AddFlagStringArray adds a string array flag to the Cobra application
	//
	// The AddFlagStringArray function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a string array
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	//
	// The AddFlagStringArray function does not return any value. It is used
	// to add a string array flag to the Cobra application.
	//
	// The AddFlagStringArray function allocates memory for the returned
	// flag. The caller is responsible for freeing this memory when it is no
	// longer needed.
	AddFlagStringArray(persistent bool, p *[]string, name, shorthand string, value []string, usage string)
	// AddFlagStringToInt adds a string flag to the Cobra application
	// that maps string keys to int values.
	//
	// The AddFlagStringToInt function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a map of string keys to int values
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	AddFlagStringToInt(persistent bool, p *map[string]int, name, shorthand string, value map[string]int, usage string)
	// AddFlagStringToInt64 adds a string flag to the Cobra application
	// that maps string keys to int64 values.
	//
	// The AddFlagStringToInt64 function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a map of string keys to int64 values
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	//
	// The AddFlagStringToInt64 function does not return any value. It is used
	// to add a string flag to the Cobra application.
	//
	// The AddFlagStringToInt64 function allocates memory for the returned
	// flag. The caller is responsible for freeing this memory when it is no
	// longer needed.
	AddFlagStringToInt64(persistent bool, p *map[string]int64, name, shorthand string, value map[string]int64, usage string)
	// AddFlagStringToString adds a string flag to the Cobra application
	// that maps string keys to string values.
	//
	// The AddFlagStringToString function takes the following arguments:
	//
	// - persistent: a boolean indicating whether the flag is persistent
	//   or not
	// - p: a pointer to a map of string keys to string values
	// - name: the name of the flag
	// - shorthand: a shorthand string for the flag
	// - value: the default value of the flag
	// - usage: a description of the flag
	//
	// The AddFlagStringToString function does not return any value. It is used
	// to add a string flag to the Cobra application.
	//
	// The AddFlagStringToString function allocates memory for the returned
	// flag. The caller is responsible for freeing this memory when it is no
	// longer needed.
	AddFlagStringToString(persistent bool, p *map[string]string, name, shorthand string, value map[string]string, usage string)

	// NewCommand returns a new Cobra command with the given command name,
	// short description, long description, usage without command, and
	// example usage without command.
	//
	// The NewCommand function takes the following arguments:
	//
	// - cmd: the name of the command
	// - short: a short description of the command
	// - long: a long description of the command
	// - useWithoutCmd: the usage string without the command name
	// - exampleWithoutCmd: an example usage string without the command name
	//
	// The NewCommand function returns a pointer to a spfcbr.Command object.
	//
	// The NewCommand function does not modify the state of the Cobra application.
	//
	// The NewCommand function allocates memory for the returned Cobra command.
	// The caller is responsible for freeing this memory when it is no longer needed.
	NewCommand(cmd, short, long, useWithoutCmd, exampleWithoutCmd string) *spfcbr.Command
	// AddCommand adds one or more subcommands to the Cobra application.
	//
	// The AddCommand function takes one or more arguments. Each argument is a
	// pointer to a spfcbr.Command object.
	//
	// The AddCommand function does not return any value. It is used to add
	// subcommands to the Cobra application.
	AddCommand(subCmd ...*spfcbr.Command)

	// AddCommandCompletion adds a subcommand to the Cobra application
	// that prints the possible completions for the current command.
	//
	// The AddCommandCompletion function does not return any value. It is
	// used to add a completion subcommand to the Cobra application.
	//
	// The AddCommandCompletion function takes no arguments.
	AddCommandCompletion()
	// AddCommandConfigure adds a subcommand to the Cobra application
	// that configures the Cobra application.
	//
	// The AddCommandConfigure function takes three arguments:
	// - alias: an alias for the configure subcommand
	// - basename: the filename of the configuration file
	// - defaultConfig: a function that returns a reader with the default
	//   configuration
	//
	// The AddCommandConfigure function does not return any value. It is
	// designed to be used as a setter function, which means that the
	// function will not return until the Cobra application's logic has
	// been set.
	AddCommandConfigure(alias, basename string, defaultConfig func() io.Reader)
	// AddCommandPrintErrorCode adds a subcommand to the Cobra application
	// that prints the error code with package path related.
	//
	// The AddCommandPrintErrorCode function takes a single argument of type
	// FuncPrintErrorCode, which is a function that takes two arguments of
	// type string and prints the error code and the related package path.
	//
	// The AddCommandPrintErrorCode function does not return any value. It is
	// designed to be used as a setter function, which means that the
	// function will not return until the Cobra application's logic has
	// been set.
	AddCommandPrintErrorCode(fct FuncPrintErrorCode)

	// Execute executes the Cobra application's logic.
	//
	// The Execute function will execute the application's logic by
	// calling the Cobra application's root command's Execute function.
	//
	// The Execute function will return an error if the application's
	// logic encounters an error during execution.
	//
	// The Execute function does not return any other value. It is
	// designed to be used as a terminal function, which means that
	// the function will not return until the application's logic has
	// been executed or an error has been encountered.
	Execute() error

	// Cobra returns a new Cobra object.
	//
	// The Cobra object is returned with all its fields set to nil.
	// The caller is responsible for initializing the fields of the Cobra object.
	//
	// The Cobra object returned by this function is a root command of the Cobra application.
	// It is the top-level command of the application, and it is responsible for
	// executing the application's logic.
	//
	// The Cobra object returned by this function is not configured by default.
	// The caller is responsible for configuring the Cobra object before executing
	// the application's logic.
	Cobra() *spfcbr.Command

	// ConfigureCheckArgs checks if the arguments passed to the Configure command are valid.
	//
	// It returns an error if the arguments are invalid or if there is a problem with the
	// configuration file generation.
	//
	// basename is the base name of the file to be generated. If basename is empty, the
	// package name will be used as the base name.
	//
	// args is the list of arguments passed to the Configure command.
	//
	// The function returns an error if anything goes wrong.
	ConfigureCheckArgs(basename string, args []string) error
	// ConfigureWriteConfig generates a configuration file based on giving existing config flag
	// override by passed flag in command line and completed with default for non existing values.
	//
	// basename is the base name of the file to be generated. If basename is empty, the
	// package name will be used as the base name.
	//
	// defaultConfig is a function to read the default configuration.
	//
	// printMsg is a function to print the success message after generating the configuration file.
	// If printMsg is nil, the function will not be called.
	//
	// The function returns an error if anything goes wrong.
	ConfigureWriteConfig(basename string, defaultConfig func() io.Reader, printMsg func(pkg, file string)) error
}

// New returns a new Cobra object.
//
// The Cobra object is returned with all its fields set to nil.
// The caller is responsible for initializing the fields of the Cobra object.
func New() Cobra {
	return &cobra{
		c: nil,
		s: nil,
		v: nil,
		i: nil,
		l: nil,
		f: "",
	}
}
