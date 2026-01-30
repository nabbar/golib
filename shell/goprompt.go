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
	"os"
	"strings"

	libshl "github.com/c-bata/go-prompt"
	shlcmd "github.com/nabbar/golib/shell/command"
)

// RunPrompt starts an interactive shell prompt using the go-prompt library.
// It implements the Shell interface's interactive mode.
//
// This method blocks until the user exits the shell (via "quit" or "exit" commands,
// or via SIGINT/SIGTERM signals if signal handling is enabled in the TTYSaver).
// It provides an interactive REPL (Read-Eval-Print-Loop) with auto-completion
// and command history.
//
// Prerequisites:
// The Shell must have been created with a TTYSaver (via New()) for proper terminal
// state management. If the Shell was created with nil TTYSaver, this method will
// still work but without terminal state preservation or signal handling.
//
// Terminal Safety:
// The method uses the TTYSaver provided during Shell creation to manage terminal state:
//   - Terminal state is automatically saved by the TTYSaver before RunPrompt starts
//   - go-prompt puts the terminal in raw mode for character-by-character input
//   - Terminal state is restored via defer on normal exit
//   - Signal handlers (if enabled in TTYSaver) ensure cleanup on interruption
//   - Fallback ANSI escape sequences are used if primary restoration fails
//
// Features:
//   - Auto-completion: Tab completion of registered command names
//   - Suggestions: Live dropdown with command descriptions as you type
//   - Command history: Arrow keys to navigate command history
//   - Built-in commands: "quit" and "exit" to terminate the prompt
//   - Namespace support: Autocompletes commands with prefixes (e.g., "sys:")
//   - Signal handling: Graceful shutdown on Ctrl+C (if TTYSaver configured)
//   - Empty line handling: Pressing Enter with no input does nothing (no-op)
//
// Parameters:
//   - out: Writer for command output (defaults to os.Stdout if nil)
//   - err: Writer for error output (defaults to os.Stderr if nil)
//   - opt: Optional go-prompt configuration options (see github.com/c-bata/go-prompt)
//
// Built-in Commands:
//   - "quit" (case-insensitive): Exit the shell gracefully
//   - "exit" (case-insensitive): Exit the shell gracefully
//
// Execution Flow:
//  1. Initialize completion suggestions from all registered commands
//  2. Set up defer for terminal restoration
//  3. Start signal handler goroutine (if TTYSaver has signal handling enabled)
//  4. Build suggestion list for auto-completion
//  5. Create executor and completer functions
//  6. Initialize go-prompt with custom executor and completer
//  7. Enter interactive loop (blocks here until user exits)
//  8. On exit, terminal is restored via defer
//
// Signal Handling:
// If the Shell was created with a TTYSaver that has signal handling enabled
// (sig=true in tty.New()), a background goroutine is started to handle signals:
//   - SIGINT (Ctrl+C): Restores terminal and exits
//   - SIGTERM: Graceful shutdown
//   - SIGQUIT (Ctrl+\): Quit with terminal restoration
//   - SIGHUP: Terminal hangup handling
//
// Example - Basic Interactive Shell:
//
//	// Create TTYSaver with signal handling enabled
//	ttySaver, err := tty.New(nil, true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	sh := shell.New(ttySaver)
//	sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
//	    fmt.Fprintln(out, "Hello!")
//	}))
//	sh.Add("sys:", command.New("info", "System info", infoFunc))
//
//	// Start interactive mode (blocks until user exits)
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
// See also:
//   - New() for creating Shell with TTYSaver
//   - github.com/c-bata/go-prompt for available customization options
//   - github.com/nabbar/golib/shell/tty for terminal state management
func (s *shell) RunPrompt(out, err io.Writer, opt ...libshl.Option) {
	// Step 1: Initialize variables and set defaults
	var (
		c = make([]libshl.Suggest, 0) // Suggestion list for auto-completion
		p *libshl.Prompt              // go-prompt instance

		fe libshl.Executor  // Executor function for command execution
		fc libshl.Completer // Completer function for suggestions
	)

	// Default to stdout/stderr if nil
	if out == nil {
		out = os.Stdout
	}

	if err == nil {
		err = os.Stderr
	}

	// Step 2: Setup deferred terminal restoration
	// This ensures cleanup on normal exit, panic, or error
	defer func() {
		// Attempt to restore terminal state via TTYSaver
		// If restoration fails, try fallback ANSI sequences
		if t := s.s.Load(); t != nil {
			if t.Restore() != nil {
				// Fallback: Use ANSI escape sequences for basic restoration
				_, _ = fmt.Fprint(out, "\033[?25h") // Show cursor (DECTCEM)
				_, _ = fmt.Fprint(out, "\033[0m")   // Reset attributes (SGR 0)
			}
		}
	}()

	// Step 3: Start signal handler goroutine (if TTYSaver configured)
	// This handles Ctrl+C, SIGTERM, etc. for graceful shutdown
	if t := s.s.Load(); t != nil {
		go func() {
			// Signal() blocks until a signal is received
			// then restores terminal and exits
			_ = t.Signal()
		}()
	}

	// Step 4: Build suggestion list from all registered commands
	// This list is used by the completer for auto-completion
	s.c.Range(func(k string, v shlcmd.Command) bool {
		c = append(c, libshl.Suggest{
			Text:        k,            // Command name for completion
			Description: v.Describe(), // Description shown in suggestions dropdown
		})
		return true // Continue iterating through all commands
	})

	for _, n := range s.xn.Load() {
		c = append(c, libshl.Suggest{
			Text:        n,
			Description: "Exit shell",
		})
	}

	// Step 5: Create executor closure
	// This wraps our executor method for go-prompt
	fe = func(in string) {
		s.executor(out, err, in)
	}

	// Step 6: Create completer closure
	// This wraps our completer method for go-prompt
	fc = func(doc libshl.Document) []libshl.Suggest {
		return s.completer(c, doc)
	}

	// Step 7: Create and configure the go-prompt instance
	// Apply user-provided options for customization
	p = libshl.New(fe, fc, opt...)

	// Step 8: Start the interactive prompt (blocks until exit)
	// This is the main REPL loop - only returns when user exits
	p.Run()
}

// executor is the internal command execution handler for the go-prompt library.
// It processes user input from the interactive prompt and routes it to the appropriate handler.
//
// The method is called by go-prompt's REPL loop every time the user presses Enter.
// It handles three cases:
//  1. Built-in exit commands (quit/exit)
//  2. Empty input (no-op)
//  3. Normal commands (routed to Run)
//
// Parameters:
//   - out: Writer for command output
//   - err: Writer for error output
//   - in: Raw input string from the user (includes leading/trailing whitespace)
//
// Processing Flow:
//  1. Trim whitespace from input
//  2. Check for exit commands (case-insensitive)
//  3. Check for empty input
//  4. Parse input into words using strings.Fields
//  5. Execute command via Run() method
//
// Built-in Commands:
//   - "quit" or "exit" (any case): Print goodbye message and exit with status 0
//
// Input Parsing:
// Uses strings.Fields for simple whitespace splitting. This means:
//   - Multiple spaces are treated as single separator
//   - Leading/trailing whitespace is ignored
//   - Quoted arguments are NOT supported (e.g., "hello world" = 2 args, not 1)
//
// For more sophisticated parsing (quoted arguments, escape sequences), a proper
// shell parser would be needed (e.g., github.com/kballard/go-shellquote).
//
// Thread-Safety:
// This method is called from the go-prompt event loop (single goroutine).
// It's not designed for concurrent execution.
func (s *shell) executor(out, err io.Writer, in string) {
	// Trim leading/trailing whitespace
	if len(strings.TrimSpace(in)) < 1 {
		return
	}

	for _, n := range s.xn.Load() {
		if strings.EqualFold(n, in) {
			if f := s.xf.Load(); f != nil {
				f()
			}
			os.Exit(0)
		}
	}

	// Parse input into command and arguments
	// strings.Fields splits on whitespace and handles multiple spaces
	// Note: Does not handle quoted strings
	s.Run(out, err, strings.Fields(in))
}

// completer is the internal auto-completion handler for the go-prompt library.
// It filters and returns command suggestions based on the current user input.
//
// The method is called by go-prompt's completion system whenever the user types
// or presses Tab. It provides live command suggestions as the user types.
//
// Parameters:
//   - sug: Complete list of available commands (built from registered commands)
//   - doc: Document interface providing current input context
//
// Returns:
//   - []Suggest: Filtered list of command suggestions matching user input
//
// Completion Behavior:
// Uses go-prompt's FilterHasPrefix which:
//   - Filters suggestions by prefix match (case-insensitive)
//   - Returns commands that start with what the user has typed
//   - Handles partial words (e.g., "sys" matches "sys:info", "system")
//   - Ignores case for matching (e.g., "Help" matches "help")
//
// Document API:
// The doc parameter provides:
//   - GetWordBeforeCursor(): Returns the partial word being typed
//   - GetText(): Returns the complete input line
//   - CursorPosition: Current cursor position
//
// Suggestion Format:
// Each suggestion contains:
//   - Text: Command name to complete (e.g., "sys:info")
//   - Description: Help text shown in dropdown (e.g., "System information")
//
// Performance:
// The filtering is O(n) where n is the number of registered commands.
// For most use cases (< 1000 commands), this is fast enough for real-time
// completion. The suggestion list is pre-built once in RunPrompt.
//
// Thread-Safety:
// This method is called from the go-prompt event loop (single goroutine).
// The suggestion list is immutable after creation, so no synchronization needed.
//
// Example Completion Scenarios:
//   - User types "h" → suggests "help", "hello", etc.
//   - User types "sys:" → suggests "sys:info", "sys:status", etc.
//   - User types "" (empty) → suggests all commands
//   - User types "xyz" → suggests nothing if no match
func (s *shell) completer(sug []libshl.Suggest, doc libshl.Document) []libshl.Suggest {
	// Get the word being typed (before cursor position)
	// This is the prefix we'll match against
	word := doc.GetWordBeforeCursor()

	// Filter suggestions using go-prompt's built-in prefix matcher
	// Third parameter (true) enables case-insensitive matching
	return libshl.FilterHasPrefix(sug, word, true)
}
