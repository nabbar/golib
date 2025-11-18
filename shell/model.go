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

package shell

import (
	"fmt"
	"io"

	libatm "github.com/nabbar/golib/atomic"
	shlcmd "github.com/nabbar/golib/shell/command"
	"github.com/nabbar/golib/shell/tty"
)

// shell is the internal implementation of the Shell interface.
// It maintains a thread-safe registry of commands with support for command prefixes
// and optional terminal state management via TTYSaver.
//
// Design Philosophy:
// The struct is intentionally minimal, containing only the command registry and
// optional TTYSaver reference. All logic is implemented in methods, keeping the
// struct lightweight and focused. The struct is unexported to enforce creation
// through New(), ensuring proper initialization.
//
// Fields:
//
//   - c: Thread-safe command registry using atomic.MapTyped
//     Stores commands with full name (prefix + name) as key
//     Example keys: "help", "sys:info", "user:list"
//
//   - s: Optional TTYSaver for terminal state management using atomic.Value
//
//   - nil: Shell operates in non-interactive mode
//
//   - non-nil: Enables RunPrompt() with terminal management
//     Used by RunPrompt() for terminal state save/restore and signal handling
//
// Thread-Safety:
// All operations on the shell struct are thread-safe through atomic operations:
//   - Read operations (Run, Get, Desc) use atomic.MapTyped.Load for lock-free reads
//   - Write operations (Add) use atomic.MapTyped.Store for atomic writes
//   - Walk operations use atomic.MapTyped.Range for consistent snapshots
//   - TTYSaver access uses atomic.Value.Load/Store for thread-safe access
//   - No explicit locking required in user code
//
// The atomic.MapTyped and atomic.Value provide lock-free synchronization using
// atomic operations under the hood, offering better performance than traditional
// mutex-based maps under high concurrent load.
//
// Command Storage:
// Commands are stored in an atomic map where:
//   - Key: Full command name (prefix + command.Name()), e.g., "sys:info", "help"
//   - Value: Command instance (github.com/nabbar/golib/shell/command.Command)
//
// The map is initialized by New() with atomic.NewMapTyped and persists for the
// shell's lifetime. Commands can be added and removed during operation without
// blocking concurrent reads.
//
// Memory Layout:
// The struct is small (two pointers on 64-bit systems, ~16 bytes) and cheap to pass
// by reference. The actual command storage and TTYSaver are heap-allocated and
// referenced by pointer.
//
// See also:
//   - New() for proper shell initialization
//   - github.com/nabbar/golib/atomic for atomic data structures
//   - github.com/nabbar/golib/shell/tty for terminal state management
type shell struct {
	c libatm.MapTyped[string, shlcmd.Command] // Thread-safe command registry (atomic map)
	s libatm.Value[tty.TTYSaver]              // Optional TTYSaver for terminal management (atomic value)
}

// Run executes a registered command by name.
// It implements the Shell interface.
//
// The method performs command lookup and execution in three phases:
//  1. Validation: Check if args is non-empty
//  2. Lookup: Find command by name using atomic Load
//  3. Execution: Call command's Run method with remaining args
//
// Parameters:
//   - buf: Writer for standard output (nil-safe - command may handle nil)
//   - err: Writer for error output (nil-safe - used for error messages)
//   - args: Command name (args[0]) and its arguments (args[1:])
//
// Command Name Resolution:
// The command name in args[0] must be the full name including any prefix:
//   - "help" looks for command registered without prefix
//   - "sys:info" looks for command registered with "sys:" prefix
//   - Lookup is case-sensitive and exact match
//
// Behavior:
//   - Returns immediately (no-op) if args is empty or nil
//   - Writes "Invalid command\n" to err if command not found (when err != nil)
//   - Writes "Command not runable...\n" to err if command is nil (defensive check)
//   - Passes args[1:] to command function (remaining arguments after command name)
//
// Error Handling:
// The method uses defensive programming to handle edge cases:
//   - Empty args: Silent return (no error written)
//   - Command not found: Error written to err writer if provided
//   - Nil command: Error written to err writer if provided (should not occur)
//   - Nil writers: Safe - errors are silently ignored
//
// Thread-Safety:
// Uses atomic Load for reading from the command registry, providing lock-free
// concurrent access. Multiple goroutines can safely call Run simultaneously,
// even for different commands. The command execution itself runs without locks,
// so command implementations should be thread-safe if called concurrently.
//
// Performance:
// The lookup is O(1) using the atomic map's hash table. The method adds minimal
// overhead beyond the command's own execution time.
//
// Example:
//
//	// Execute "hello" command without arguments
//	sh.Run(os.Stdout, os.Stderr, []string{"hello"})
//
//	// Execute "sys:info" command with arguments
//	sh.Run(os.Stdout, os.Stderr, []string{"sys:info", "--verbose"})
//
//	// Execute with custom writers
//	var buf bytes.Buffer
//	sh.Run(&buf, &buf, []string{"echo", "test"})
func (s *shell) Run(buf io.Writer, err io.Writer, args []string) {
	if len(args) == 0 {
		return
	}

	// Lock for reading the command from the map
	// RLock allows multiple concurrent reads
	cmd, ok := s.c.Load(args[0])

	if ok {
		// Command found in registry
		if cmd == nil {
			// Defensive check: command exists but is nil
			if err != nil {
				_, _ = fmt.Fprintf(err, "Command not runable...\n")
			}
			return
		}

		// Execute the command with remaining arguments
		cmd.Run(buf, err, args[1:])
	} else {
		// Command not found in registry
		if err != nil {
			_, _ = fmt.Fprintf(err, "Invalid command\n")
		}
	}
}

// Walk iterates over all registered commands for inspection and enumeration.
// It implements the Shell interface.
//
// The method provides a way to iterate through the command registry without exposing
// the internal map structure. It's designed for read-only inspection, though it also
// performs automatic cleanup of nil commands during iteration.
//
// Parameters:
//   - fct: Callback function invoked for each command. If nil, returns immediately.
//
// Callback Function:
// The function receives:
//   - name: Full command name including prefix (e.g., "sys:info", "help")
//   - item: The Command instance (guaranteed non-nil when callback is invoked)
//
// The function should return:
//   - true: Continue iterating to the next command
//   - false: Stop iteration immediately
//
// Iteration Behavior:
//   - Non-deterministic order (depends on atomic map implementation)
//   - Nil commands are silently deleted (cleanup phase)
//   - Iteration can be stopped early by returning false from callback
//   - Provides consistent snapshot via atomic.MapTyped.Range
//
// Automatic Cleanup:
// During iteration, if a nil command is encountered (should not happen in normal
// operation), it is automatically removed from the registry. This defensive measure
// ensures registry integrity.
//
// Thread-Safety:
// Uses atomic.MapTyped.Range which provides a consistent snapshot of the registry
// during iteration. The registry can be safely modified by other goroutines during
// walking without causing data races or panics. Changes made during iteration may
// or may not be visible depending on timing.
//
// Performance:
// The iteration is O(n) where n is the number of commands. The atomic operations
// have minimal overhead. For large registries (1000+ commands), iteration remains fast.
//
// Use Cases:
//   - Command counting and statistics
//   - Help generation (list all commands)
//   - Command validation (check all meet criteria)
//   - Prefix-based filtering
//   - Command export/serialization
//
// Example - Count Commands:
//
//	count := 0
//	sh.Walk(func(name string, item command.Command) bool {
//	    count++
//	    return true
//	})
//
// Example - Filter by Prefix:
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
//	found := sh.Walk(func(name string, item command.Command) bool {
//	    return name != "target" // stop when target found
//	})
func (s *shell) Walk(fct func(name string, item shlcmd.Command) bool) {
	if fct == nil {
		return
	}

	// Lock for reading and potentially writing the map
	// Full lock needed because we may modify the map
	s.c.Range(func(k string, c shlcmd.Command) bool {
		if c == nil {
			s.c.Delete(k)
			return true
		} else {
			return fct(k, c)
		}
	})
}

// Add registers one or more commands with an optional prefix.
// It implements the Shell interface.
//
// The method adds commands to the internal registry, optionally prepending
// a prefix to each command name for namespace organization. This enables logical
// grouping of commands (e.g., "sys:", "user:", "db:").
//
// Parameters:
//   - prefix: String to prepend to each command name (empty string = no prefix)
//   - cmd: Variadic list of commands to register (nil commands are skipped)
//
// Name Formation:
// The full command name is formed by concatenating:
//
//	prefix + command.Name()
//
// Examples:
//   - prefix="", command.Name()="help" → full name="help"
//   - prefix="sys:", command.Name()="info" → full name="sys:info"
//   - prefix="user:", command.Name()="list" → full name="user:list"
//
// Behavior:
//   - Skips nil commands in the variadic list (defensive programming)
//   - Replaces existing commands with the same full name (idempotent)
//   - Empty prefix is valid and common for top-level commands
//   - Prefix can be any string (not limited to colon-suffixed)
//   - Command's own Name() is never modified, only the registry key
//
// Command Replacement:
// When a command with the same full name already exists, it is replaced
// atomically. This allows command redefinition and dynamic updates.
//
// Thread-Safety:
// Uses atomic Store for each command, ensuring thread-safe concurrent additions.
// Multiple goroutines can safely call Add simultaneously without data races.
// The atomic operations guarantee each command is stored consistently.
//
// Performance:
// Each Add operation is O(k) where k is the number of commands in the variadic
// list. The atomic Store is O(1) per command. Bulk additions are efficient.
//
// Example - Basic Registration:
//
//	sh.Add("", command.New("help", "Show help", helpFunc))
//	sh.Add("", command.New("exit", "Exit shell", exitFunc))
//
// Example - Namespaced Commands:
//
//	// System commands
//	sh.Add("sys:",
//	    command.New("info", "System info", infoFunc),
//	    command.New("restart", "Restart system", restartFunc))
//
//	// User commands
//	sh.Add("user:",
//	    command.New("list", "List users", listFunc),
//	    command.New("add", "Add user", addFunc))
//
// Example - Nil Handling:
//
//	// Nil commands are safely skipped
//	sh.Add("", cmd1, nil, cmd2, nil, cmd3)
//	// Only cmd1, cmd2, cmd3 are registered
func (s *shell) Add(prefix string, cmd ...shlcmd.Command) {
	var fct = func(s string) string {
		return prefix + s
	}

	if len(prefix) < 1 {
		fct = func(s string) string {
			return s
		}
	}

	// Process each command in the variadic list
	for i := 0; i < len(cmd); i++ {
		if c := cmd[i]; c == nil {
			continue
		} else {
			s.c.Store(fct(c.Name()), c)
		}
	}
}

// Get retrieves a command by its full name (including prefix).
// It implements the Shell interface.
//
// The method performs an O(1) lookup in the command registry using atomic Load.
// Command names are case-sensitive and must match exactly, including any prefix.
//
// Parameters:
//   - cmd: Full command name to search for (e.g., "help", "sys:info")
//
// Returns:
//   - command: The Command instance if found, nil if not found
//   - found: true if command exists in registry, false otherwise
//
// Lookup Rules:
//   - Exact match required: "info" != "sys:info"
//   - Case-sensitive: "Help" != "help"
//   - Prefix must be included: "info" won't find "sys:info"
//   - Empty string searches for command with empty name (if registered)
//
// Thread-Safety:
// Uses atomic Load providing lock-free concurrent access. Multiple goroutines
// can safely call Get simultaneously without blocking each other.
//
// Performance:
// O(1) hash table lookup via atomic map. Extremely fast even with thousands
// of registered commands.
//
// Example - Simple Lookup:
//
//	cmd, found := sh.Get("help")
//	if found {
//	    fmt.Println("Description:", cmd.Describe())
//	} else {
//	    fmt.Println("Command not found")
//	}
//
// Example - Prefixed Lookup:
//
//	sysInfo, found := sh.Get("sys:info")
//	if !found {
//	    return fmt.Errorf("sys:info command not available")
//	}
//	sysInfo.Run(out, err, args)
func (s *shell) Get(cmd string) (shlcmd.Command, bool) {
	return s.c.Load(cmd)
}

// Desc retrieves the description of a command by its full name.
// It implements the Shell interface.
//
// This is a convenience method that combines Get() and Command.Describe().
// It returns an empty string if the command doesn't exist or has no description.
//
// Parameters:
//   - cmd: Full command name to search for (including prefix if any)
//
// Returns:
//   - string: The command's description, or empty string if not found
//
// Behavior:
//   - Returns empty string if command doesn't exist
//   - Returns empty string if command is nil (defensive)
//   - Returns the result of command.Describe() otherwise
//   - Case-sensitive lookup: "help" != "Help"
//   - Prefix must match: "info" != "sys:info"
//
// Thread-Safety:
// Uses atomic Load for lock-free concurrent access. Safe to call from
// multiple goroutines simultaneously.
//
// Performance:
// O(1) lookup via atomic map plus one method call. Very efficient.
//
// Use Cases:
//   - Generating help text for a specific command
//   - Validating command existence before execution
//   - Building dynamic UI with command descriptions
//   - Command documentation generation
//
// Example - Simple Description Retrieval:
//
//	desc := sh.Desc("help")
//	if desc != "" {
//	    fmt.Println("Help command:", desc)
//	}
//
// Example - With Prefix:
//
//	sysDesc := sh.Desc("sys:info")
//	if sysDesc == "" {
//	    log.Printf("sys:info command not available")
//	} else {
//	    fmt.Printf("sys:info: %s\n", sysDesc)
//	}
//
// Example - Help Generation:
//
//	var commands []string
//	sh.Walk(func(name string, _ command.Command) bool {
//	    commands = append(commands, name)
//	    return true
//	})
//	for _, name := range commands {
//	    desc := sh.Desc(name)
//	    fmt.Printf("  %-20s %s\n", name, desc)
//	}
func (s *shell) Desc(cmd string) string {
	if c, k := s.c.Load(cmd); k && c != nil {
		return c.Describe()
	}
	return ""
}
