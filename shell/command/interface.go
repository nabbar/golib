/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

// Package command provides a simple interface for creating and managing shell commands.
// It supports both executable commands and informational command metadata.
//
// The package is designed to be thread-safe and can be used in concurrent environments.
// All command instances are immutable after creation, making them safe for concurrent access.
//
// # Usage
//
// Create a new executable command:
//
//	cmd := command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
//	    if len(args) > 0 {
//	        fmt.Fprintf(out, "Hello, %s!\n", args[0])
//	    } else {
//	        fmt.Fprintln(out, "Hello, World!")
//	    }
//	})
//
// Create a command info for documentation purposes:
//
//	info := command.Info("exit", "Exit the shell")
//
// Execute a command:
//
//	cmd.Run(os.Stdout, os.Stderr, []string{"Alice"})
//
// # Integration
//
// This package is typically used with github.com/nabbar/golib/shell to create
// interactive command-line interfaces. Commands created with this package can be
// registered with a shell instance for execution.
//
// See github.com/nabbar/golib/shell for the main Shell interface that uses this package.
package command

import "io"

// FuncRun is a function type that executes a command.
// It receives output and error writers for standard output and error streams,
// and a slice of string arguments.
//
// Parameters:
//   - buf: Writer for standard output. Can be nil; implementations should check before writing.
//   - err: Writer for error output. Can be nil; implementations should check before writing.
//   - args: Command arguments. Can be nil or empty. The first element (if present) is typically
//     the first argument after the command name.
//
// The function should:
//   - Write standard output to buf (e.g., results, success messages)
//   - Write errors and diagnostics to err (e.g., error messages, warnings)
//   - Parse and validate args as needed for the command's logic
//   - Handle nil writers gracefully (check before writing)
//   - Be safe for concurrent execution if used in a concurrent context
//
// Example implementations:
//
//	// Simple command that always succeeds
//	func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, "Command executed")
//	}
//
//	// Command with argument validation
//	func(out, err io.Writer, args []string) {
//	    if len(args) == 0 {
//	        fmt.Fprintln(err, "Error: missing required argument")
//	        return
//	    }
//	    fmt.Fprintf(out, "Processing: %s\n", args[0])
//	}
//
//	// Command that handles nil writers
//	func(out, err io.Writer, args []string) {
//	    result := doSomething(args)
//	    if out != nil {
//	        fmt.Fprintln(out, result)
//	    }
//	}
type FuncRun func(buf io.Writer, err io.Writer, args []string)

// CommandInfo provides read-only access to command metadata.
// It allows querying a command's name and description without executing it.
//
// This interface is useful for:
//   - Building help systems that list available commands
//   - Creating command documentation without execution capability
//   - Storing command metadata in registries or catalogs
//
// Example usage:
//
//	func displayHelp(commands []CommandInfo) {
//	    fmt.Println("Available commands:")
//	    for _, cmd := range commands {
//	        fmt.Printf("  %-15s %s\n", cmd.Name(), cmd.Describe())
//	    }
//	}
type CommandInfo interface {
	// Name returns the command name.
	// The name is typically used as the identifier to invoke the command.
	// Returns an empty string if the command has no name.
	Name() string

	// Describe returns a human-readable description of the command.
	// The description explains what the command does and is typically shown in help text.
	// Returns an empty string if no description was provided.
	Describe() string
}

// Command represents an executable command with metadata.
// It extends CommandInfo with the ability to execute the command.
//
// Commands are immutable and thread-safe after creation. Multiple goroutines
// can safely call any method on a Command instance concurrently.
//
// Usage example:
//
//	cmd := command.New("greet", "Greet a user", func(out, err io.Writer, args []string) {
//	    name := "World"
//	    if len(args) > 0 {
//	        name = args[0]
//	    }
//	    fmt.Fprintf(out, "Hello, %s!\n", name)
//	})
//
//	// Query metadata
//	fmt.Println(cmd.Name())        // "greet"
//	fmt.Println(cmd.Describe())    // "Greet a user"
//
//	// Execute the command
//	cmd.Run(os.Stdout, os.Stderr, []string{"Alice"})  // Output: Hello, Alice!
type Command interface {
	CommandInfo

	// Run executes the command with the provided writers and arguments.
	//
	// Parameters:
	//   - buf: Writer for standard output (can be nil)
	//   - err: Writer for error output (can be nil)
	//   - args: Slice of string arguments for the command (can be nil or empty)
	//
	// Behavior:
	//   - If the command's function is nil, Run returns immediately without doing anything
	//   - The method is safe for concurrent calls
	//   - Arguments are passed as-is to the underlying FuncRun implementation
	//   - The method does not return errors; implementations should write errors to the err writer
	//
	// Example:
	//
	//	cmd.Run(os.Stdout, os.Stderr, []string{"arg1", "arg2"})
	Run(buf io.Writer, err io.Writer, args []string)
}

// New creates a new executable Command with the given name, description, and function.
//
// Parameters:
//   - name: The command name (can be empty)
//   - desc: A human-readable description (can be empty)
//   - fct: The function to execute when Run is called (can be nil for no-op commands)
//
// Returns a Command that is immediately usable and thread-safe.
//
// Example:
//
//	// Create a simple echo command
//	echo := command.New("echo", "Echo arguments", func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, strings.Join(args, " "))
//	})
//
//	// Create a command with error handling
//	divide := command.New("divide", "Divide two numbers", func(out, err io.Writer, args []string) {
//	    if len(args) < 2 {
//	        fmt.Fprintln(err, "Error: requires two arguments")
//	        return
//	    }
//	    // ... division logic
//	})
//
//	// Create a no-op placeholder command
//	placeholder := command.New("future-feature", "Coming soon", nil)
func New(name, desc string, fct FuncRun) Command {
	return &model{
		n: name,
		d: desc,
		r: fct,
	}
}

// Info creates a new CommandInfo with the given name and description.
// This is useful for creating command metadata without an executable function.
//
// Parameters:
//   - name: The command name (can be empty)
//   - desc: A human-readable description (can be empty)
//
// Returns a CommandInfo that is immediately usable and thread-safe.
// The returned CommandInfo can also be used as a Command, but calling Run will be a no-op.
//
// Example:
//
//	// Create command metadata for documentation
//	exitInfo := command.Info("exit", "Exit the shell")
//	helpInfo := command.Info("help", "Show available commands")
//
//	// Use in a help system
//	commands := []command.CommandInfo{
//	    command.Info("ls", "List files"),
//	    command.Info("cd", "Change directory"),
//	    command.Info("pwd", "Print working directory"),
//	}
//
//	// Display command list
//	for _, cmd := range commands {
//	    fmt.Printf("%-10s %s\n", cmd.Name(), cmd.Describe())
//	}
//
// Note: The returned value implements both CommandInfo and Command interfaces,
// but calling Run() on it will do nothing since no function is attached.
func Info(name, desc string) CommandInfo {
	return &model{
		n: name,
		d: desc,
		r: nil,
	}
}
