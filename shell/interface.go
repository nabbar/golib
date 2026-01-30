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

// Package shell provides an interactive command-line shell interface with command registration,
// execution, and terminal interaction capabilities.
//
// The package supports both direct command execution and interactive prompt mode using go-prompt.
// It provides a thread-safe command registry with prefix support, allowing organization of commands
// into namespaces (e.g., "sys:", "user:", etc.).
//
// # Key Features
//
//   - Thread-safe command registration and execution using atomic operations
//   - Command prefix support for namespacing and organization
//   - Interactive prompt mode with autocomplete and suggestions
//   - Terminal state management and restoration via tty subpackage
//   - Command walking and inspection capabilities
//   - Built-in quit/exit commands for interactive mode
//   - Signal handling for graceful shutdown
//
// # Architecture
//
// The package consists of three main components:
//
//  1. Shell Interface: Public API for command management
//  2. Command Registry: Thread-safe storage using github.com/nabbar/golib/atomic.MapTyped
//  3. Terminal Management: TTY state preservation via github.com/nabbar/golib/shell/tty
//
// Commands are stored with their full name (prefix + command name) as the key, enabling
// efficient lookup and namespace isolation. The atomic map ensures all operations are
// thread-safe without explicit locking in user code.
//
// # Basic Usage
//
// Create a shell and register commands:
//
//	sh := shell.New(nil) // nil TTYSaver for non-interactive mode
//	sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, "Hello, World!")
//	}))
//	sh.Run(os.Stdout, os.Stderr, []string{"hello"})
//
// # Interactive Mode
//
// Start an interactive prompt with autocomplete:
//
//	// Create TTYSaver for terminal state management
//	ttySaver, _ := tty.New(nil, true) // Enable signal handling
//	sh := shell.New(ttySaver)
//
//	// Register your commands...
//	sh.Add("sys:", command.New("info", "System info", infoFunc))
//	sh.Add("user:", command.New("list", "List users", listFunc))
//
//	// Start interactive prompt (blocks until user exits)
//	sh.RunPrompt(os.Stdout, os.Stderr)
//
// # Namespacing with Prefixes
//
// Organize commands into logical groups:
//
//	sh := shell.New(nil)
//
//	// System commands
//	sh.Add("sys:", sysInfo, sysStatus, sysRestart)
//
//	// User management commands
//	sh.Add("user:", userAdd, userDel, userList)
//
//	// Database commands
//	sh.Add("db:", dbConnect, dbQuery, dbBackup)
//
//	// Commands accessible as: sys:info, user:add, db:connect, etc.
//
// # Command Inspection
//
// Walk through all registered commands:
//
//	sh := shell.New(nil)
//	// ... register commands ...
//
//	count := 0
//	sh.Walk(func(name string, item command.Command) bool {
//	    fmt.Printf("%s: %s\n", name, item.Describe())
//	    count++
//	    return true // continue walking
//	})
//	fmt.Printf("Total: %d commands\n", count)
//
// # Error Handling
//
// The shell handles various error conditions gracefully:
//   - Missing commands: Writes "Invalid command" to error writer
//   - Nil commands: Writes "Command not runable..." to error writer
//   - Empty arguments: Returns immediately without action
//   - Terminal not available: Interactive mode fails with tty.ErrorNotTTY
//
// # Thread Safety
//
// All Shell methods are safe for concurrent use. The internal command registry uses
// github.com/nabbar/golib/atomic.MapTyped which provides lock-free thread-safe operations.
// Multiple goroutines can safely:
//   - Add commands concurrently
//   - Execute different commands simultaneously
//   - Walk the registry while commands are being added
//
// # Dependencies
//
// This package depends on:
//   - github.com/nabbar/golib/shell/command: Command interface and creation
//   - github.com/nabbar/golib/shell/tty: Terminal state management
//   - github.com/nabbar/golib/atomic: Thread-safe map implementation
//   - github.com/c-bata/go-prompt: Interactive prompt library
//
// # Subpackages
//
//   - command: Command interface and creation utilities
//   - tty: Terminal state save/restore for interactive mode
//
// See also:
//   - github.com/nabbar/golib/shell/command for creating commands
//   - github.com/nabbar/golib/shell/tty for terminal state management
//   - github.com/c-bata/go-prompt for prompt customization options
package shell

import (
	"io"

	libshl "github.com/c-bata/go-prompt"
	libatm "github.com/nabbar/golib/atomic"
	shlcmd "github.com/nabbar/golib/shell/command"
	"github.com/nabbar/golib/shell/tty"
)

// Shell provides an interface for managing and executing shell commands.
// All methods are safe for concurrent use by multiple goroutines.
//
// The Shell maintains an internal command registry that can be modified via Add()
// and queried via Get(), Desc(), and Walk(). Commands can be executed directly
// via Run() or interactively via RunPrompt().
//
// Command Organization:
//   - Commands can be registered with optional prefixes for namespacing
//   - The same command name can exist in different namespaces
//   - Commands are retrieved by their full name (prefix + name)
//
// Example:
//
//	sh := shell.New()
//	sh.Add("sys:", command.New("info", "System info", func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, "System information...")
//	}))
//	sh.Run(os.Stdout, os.Stderr, []string{"sys:info"})
type Shell interface {
	// Run executes a command with the provided arguments.
	// The first element of args should be the command name (including any prefix).
	// Subsequent elements are passed as arguments to the command's function.
	//
	// Parameters:
	//   - buf: Writer for standard output (can be nil)
	//   - err: Writer for error output (can be nil)
	//   - args: Command name and arguments. If empty or nil, the method returns immediately.
	//
	// Behavior:
	//   - If the command is not found, writes "Invalid command" to err writer
	//   - If the command is nil, writes "Command not runable..." to err writer
	//   - Otherwise, executes the command's function with the remaining arguments
	//
	// Thread-safety: Safe for concurrent use
	//
	// Example:
	//
	//	sh.Run(os.Stdout, os.Stderr, []string{"hello", "Alice"})
	Run(buf io.Writer, err io.Writer, args []string)

	// Add registers one or more commands with an optional prefix.
	// If a command with the same name already exists, it is replaced.
	//
	// Parameters:
	//   - prefix: Optional prefix to prepend to each command name (e.g., "sys:", "user:")
	//   - cmd: One or more commands to register. Nil commands are skipped.
	//
	// The full command name is formed by concatenating the prefix and the command's Name().
	// Commands can be registered multiple times with different prefixes.
	//
	// Thread-safety: Safe for concurrent use
	//
	// Example:
	//
	//	sh.Add("", command.New("help", "Show help", helpFunc))
	//	sh.Add("sys:", command.New("info", "System info", infoFunc))
	Add(prefix string, cmd ...shlcmd.Command)

	// Get retrieves a command by its full name (including prefix).
	//
	// The method performs an exact match lookup in the command registry.
	// Command names are case-sensitive and must include any prefix used during registration.
	//
	// Parameters:
	//   - cmd: Full command name to search for (e.g., "help" or "sys:info")
	//
	// Returns:
	//   - command: The Command instance if found
	//   - found: true if the command exists, false otherwise
	//
	// Behavior:
	//   - Empty string will only match if a command was registered with empty name
	//   - Prefix is part of the lookup key: "info" != "sys:info"
	//   - Returns (nil, false) if command not found
	//
	// Thread-safety: Safe for concurrent use via atomic map operations
	//
	// Example:
	//
	//	cmd, found := sh.Get("help")
	//	if found {
	//	    fmt.Println("Description:", cmd.Describe())
	//	}
	//
	//	sysCmd, found := sh.Get("sys:info")
	//	if !found {
	//	    fmt.Println("Command not found")
	//	}
	Get(cmd string) (shlcmd.Command, bool)

	// Desc retrieves the description of a command by its full name.
	//
	// The method looks up the command and returns its description string.
	// This is a convenience method that combines Get() and Command.Describe().
	//
	// Parameters:
	//   - cmd: Full command name to search for (including prefix if any)
	//
	// Returns:
	//   - string: The command's description, or empty string if command not found
	//
	// Behavior:
	//   - Returns empty string if command doesn't exist
	//   - Returns empty string if command is nil
	//   - Command names are case-sensitive
	//   - Prefix must be included: "info" != "sys:info"
	//
	// Thread-safety: Safe for concurrent use via atomic map operations
	//
	// Example:
	//
	//	desc := sh.Desc("help")
	//	fmt.Println(desc) // "Show help"
	//
	//	desc = sh.Desc("sys:info")
	//	if desc == "" {
	//	    fmt.Println("Command not found")
	//	}
	Desc(cmd string) string

	// Walk iterates over all registered commands, allowing inspection and enumeration.
	// The provided function is called once for each command in the registry.
	//
	// The iteration order is non-deterministic as it depends on the internal map structure.
	// If you need ordered iteration, collect commands and sort them externally.
	//
	// Parameters:
	//   - fct: Function called for each command. If nil, Walk returns immediately.
	//
	// The function receives:
	//   - name: Full command name (including prefix)
	//   - item: The command instance (may be nil if cleanup detected)
	//
	// The function should return:
	//   - true: Continue walking to the next command
	//   - false: Stop walking immediately
	//
	// Use Cases:
	//   - Counting commands: Count total or by prefix
	//   - Generating help: Collect all command names and descriptions
	//   - Command validation: Check all commands meet certain criteria
	//   - Statistics: Gather metrics about registered commands
	//
	// Implementation Notes:
	// The method uses atomic.MapTyped.Range() which provides a consistent snapshot
	// of the registry during iteration. Nil commands are automatically removed
	// during walking (cleanup phase).
	//
	// Thread-safety: Safe for concurrent use. The registry can be modified by other
	// goroutines during walking without causing data races.
	//
	// Example - Count Commands:
	//
	//	count := 0
	//	sh.Walk(func(name string, item command.Command) bool {
	//	    count++
	//	    return true
	//	})
	//	fmt.Printf("Total commands: %d\n", count)
	//
	// Example - List by Prefix:
	//
	//	var sysCommands []string
	//	sh.Walk(func(name string, item command.Command) bool {
	//	    if strings.HasPrefix(name, "sys:") {
	//	        sysCommands = append(sysCommands, name)
	//	    }
	//	    return true
	//	})
	//
	// Example - Early Exit:
	//
	//	found := false
	//	sh.Walk(func(name string, item command.Command) bool {
	//	    if name == "special" {
	//	        found = true
	//	        return false // stop walking
	//	    }
	//	    return true
	//	})
	Walk(fct func(name string, item shlcmd.Command) bool)

	// RunPrompt starts an interactive shell prompt using the go-prompt library.
	// This method blocks until the user exits the shell (via "quit" or "exit" commands).
	//
	// The prompt provides a rich interactive experience with autocomplete, suggestions,
	// and terminal state management. It's designed for building interactive CLI tools
	// and administrative shells.
	//
	// Prerequisites:
	// The Shell must be created with a valid TTYSaver (via New()) for terminal state management.
	// Signal handling should be enabled in the TTYSaver (sig=true in tty.New()) for graceful
	// shutdown on Ctrl+C and other termination signals.
	//
	// Parameters:
	//   - out: Writer for command output (uses os.Stdout if nil)
	//   - err: Writer for error output (uses os.Stderr if nil)
	//   - opt: Optional go-prompt configuration options (see github.com/c-bata/go-prompt)
	//
	// Interactive Features:
	//   - Auto-completion: Tab completion of registered command names
	//   - Suggestions: Live dropdown with command descriptions
	//   - History: Arrow keys for command history navigation
	//   - Built-in commands: "quit" and "exit" to terminate the prompt
	//   - Prefix support: Autocompletes with namespace awareness
	//
	// Terminal Management:
	// The method uses the TTYSaver provided during Shell creation to manage terminal state:
	//   1. Terminal state is already saved by the TTYSaver
	//   2. go-prompt enables raw mode for character-by-character input
	//   3. Terminal state is restored on exit via defer
	//   4. Signal handlers (if enabled in TTYSaver) ensure cleanup on interruption
	//
	// Signal Handling:
	// If the Shell was created with a TTYSaver that has signal handling enabled
	// (sig=true in tty.New()), the following signals are handled:
	//   - SIGINT (Ctrl+C): Restores terminal and exits gracefully
	//   - SIGTERM: Graceful shutdown from systemd/docker
	//   - SIGQUIT (Ctrl+\): Quit with terminal restoration
	//   - SIGHUP: Terminal hangup handling
	//
	// The signal handler goroutine is automatically started when RunPrompt begins.
	//
	// Blocking Behavior:
	// This method blocks the calling goroutine until the user explicitly exits.
	// To run the prompt in the background, start it in a separate goroutine:
	//
	//	go sh.RunPrompt(os.Stdout, os.Stderr)
	//
	// Customization:
	// Pass go-prompt options to customize appearance and behavior:
	//
	//	sh.RunPrompt(out, err,
	//	    prompt.OptionPrefix(">>> "),
	//	    prompt.OptionTitle("My Shell"),
	//	    prompt.OptionPrefixTextColor(prompt.Blue))
	//
	// Thread-safety: Safe to call concurrently, though typically called once.
	// Multiple concurrent prompts will compete for terminal input.
	//
	// Example - Basic Interactive Shell:
	//
	//	// Create TTYSaver with signal handling
	//	ttySaver, err := tty.New(nil, true)
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	//
	//	sh := shell.New(ttySaver)
	//	sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
	//	    fmt.Fprintln(out, "Hello, World!")
	//	}))
	//	sh.Add("sys:", command.New("info", "System info", sysInfoFunc))
	//
	//	// Blocks until user types "quit" or "exit"
	//	sh.RunPrompt(os.Stdout, os.Stderr)
	//
	// Example - Customized Prompt:
	//
	//	ttySaver, _ := tty.New(nil, true)
	//	sh := shell.New(ttySaver)
	//	// ... register commands ...
	//
	//	sh.RunPrompt(os.Stdout, os.Stderr,
	//	    prompt.OptionPrefix("myapp> "),
	//	    prompt.OptionTitle("MyApp Admin Shell"),
	//	    prompt.OptionSuggestionBGColor(prompt.DarkGray))
	//
	// Example - Non-Interactive Mode (without TTYSaver):
	//
	//	// For non-interactive command execution, use nil TTYSaver
	//	sh := shell.New(nil)
	//	sh.Add("", myCommand)
	//	sh.Run(os.Stdout, os.Stderr, []string{"mycommand", "arg1"})
	//	// Don't call RunPrompt() without a TTYSaver
	//
	// See also:
	//   - github.com/c-bata/go-prompt for available options and customization
	//   - github.com/nabbar/golib/shell/tty for terminal state management details
	//   - New() for Shell creation with TTYSaver
	RunPrompt(out, err io.Writer, opt ...libshl.Option)

	// ExitRegister registers a custom exit function and/or custom exit command names.
	//
	// This method allows customizing how the shell handles exit requests in interactive mode.
	// You can define what happens before exit (e.g., cleanup) and what commands trigger exit.
	//
	// Parameters:
	//   - f: Function to call before exiting. If nil, defaults to os.Exit(0).
	//        If the function returns, os.Exit(0) is called immediately after.
	//   - name: Variadic list of command names that trigger exit.
	//           If empty, defaults to ["exit", "quit"].
	//
	// Behavior:
	//   - The exit function is called when an exit command is entered in RunPrompt.
	//   - The exit commands are added to the auto-completion list.
	//   - Case-insensitive matching is used for exit commands.
	//
	// Thread-safety: Safe for concurrent use via atomic operations.
	//
	// Example:
	//
	//	// Custom cleanup and commands
	//	sh.ExitRegister(func() {
	//	    fmt.Println("Cleaning up...")
	//	    db.Close()
	//	}, "bye", "logout")
	ExitRegister(f func(), name ...string)
}

// New creates a new Shell instance with the specified TTYSaver for terminal management.
//
// The function initializes a Shell with a thread-safe atomic map for command storage
// and optional terminal state management via the provided TTYSaver. The Shell is
// immediately ready for use after creation.
//
// Parameters:
//   - ts: TTYSaver for terminal state management. Can be nil for non-interactive use.
//   - nil: Shell works in non-interactive mode (Run() only, no RunPrompt())
//   - tty.New(nil, false): Basic TTYSaver without signal handling
//   - tty.New(nil, true): TTYSaver with signal handling for interactive mode
//
// The TTYSaver is used by RunPrompt() to:
//   - Save and restore terminal state (prevent corruption)
//   - Handle termination signals (Ctrl+C, SIGTERM, etc.)
//   - Manage terminal attributes for interactive input
//
// Return Value:
// Returns a Shell interface implementation with:
//   - Empty command registry (no commands registered yet)
//   - Thread-safe operations via github.com/nabbar/golib/atomic.MapTyped
//   - Optional terminal management via provided TTYSaver
//   - All methods (Add, Get, Run, Walk, RunPrompt, ExitRegister) available immediately
//
// Implementation:
// Uses github.com/nabbar/golib/atomic.NewMapTyped for lock-free concurrent access.
// The map uses command full names (prefix + name) as keys and Command instances as values.
// The TTYSaver is stored in an atomic.Value for thread-safe access.
//
// Thread-safety:
// The returned Shell is safe for concurrent use by multiple goroutines without
// additional synchronization. All operations (Add, Get, Run, Walk, RunPrompt) can
// be called concurrently.
//
// Memory:
// The Shell has minimal memory overhead - atomic map structure plus TTYSaver reference.
// Commands are stored by reference, so memory usage scales with the number of
// registered commands.
//
// Example - Non-Interactive Shell (nil TTYSaver):
//
//	// For simple command execution without terminal interaction
//	sh := shell.New(nil)
//	sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, "Hello, World!")
//	}))
//	sh.Run(os.Stdout, os.Stderr, []string{"hello"})
//
// Example - Interactive Shell (with TTYSaver):
//
//	// Create TTYSaver with signal handling for interactive mode
//	ttySaver, err := tty.New(nil, true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	sh := shell.New(ttySaver)
//	sh.Add("sys:", command.New("info", "System info", infoFunc))
//	sh.Add("user:", command.New("list", "List users", listFunc))
//
//	// Start interactive prompt (blocks until user exits)
//	sh.RunPrompt(os.Stdout, os.Stderr)
//
// Example - Administrative Shell with Namespaces:
//
//	ttySaver, _ := tty.New(nil, true)
//	sh := shell.New(ttySaver)
//
//	// System commands
//	sh.Add("sys:", sysInfo, sysRestart, sysStatus)
//
//	// User commands
//	sh.Add("user:", userList, userAdd, userDel)
//
//	// Database commands
//	sh.Add("db:", dbConnect, dbQuery, dbBackup)
//
//	// Start interactive mode
//	sh.RunPrompt(os.Stdout, os.Stderr,
//	    prompt.OptionPrefix("admin> "),
//	    prompt.OptionTitle("Admin Shell"))
//
// See also:
//   - github.com/nabbar/golib/shell/tty.New() for creating TTYSaver instances
//   - github.com/nabbar/golib/shell/command for creating commands
//   - RunPrompt() for interactive mode (requires non-nil TTYSaver)
func New(ts tty.TTYSaver) Shell {
	s := &shell{
		c:  libatm.NewMapTyped[string, shlcmd.Command](),
		s:  libatm.NewValue[tty.TTYSaver](),
		xf: libatm.NewValue[func()](),
		xn: libatm.NewValue[[]string](),
	}

	s.ExitRegister(nil)

	if ts != nil {
		s.s.Store(ts)
	}

	return s
}
