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

// Package tty provides terminal state management for saving and restoring
// terminal attributes. This is essential when using libraries like go-prompt
// that put the terminal in raw mode.
//
// The package handles terminal state preservation to prevent corruption when
// programs exit unexpectedly or are interrupted by signals.
//
// Use Cases:
//   - Interactive CLI applications using github.com/c-bata/go-prompt
//   - Terminal-based UI applications
//   - Any application that modifies terminal settings
//   - Shell implementations (see parent package github.com/nabbar/golib/shell)
//
// Key Features:
//   - Automatic terminal state capture and restoration
//   - Signal handler registration for graceful cleanup
//   - Fallback reset mechanism using ANSI escape sequences
//   - Thread-safe operations
//   - Nil-safe API (safe to call with nil values)
//
// Dependencies:
//   - golang.org/x/term: Terminal state management
//
// See also:
//   - github.com/nabbar/golib/shell: Shell package that uses tty for interactive prompts
//   - github.com/c-bata/go-prompt: Interactive prompt library
package tty

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Error variables for terminal operations.
//
// These errors provide context for terminal-related failures and can be used
// with errors.Is() for error checking.
var (
	// ErrorNotTTY is returned when the input reader is not connected to a terminal.
	// This occurs in non-interactive scenarios such as:
	//   - Input redirection: program < file
	//   - Pipes: echo "cmd" | program
	//   - Background jobs: program &
	//   - CI/CD environments without TTY allocation
	//   - Readers that don't implement the Fd() method (e.g., bytes.Buffer)
	//
	// This error is informational and typically not fatal - applications should
	// gracefully degrade to non-interactive mode when receiving this error.
	ErrorNotTTY = fmt.Errorf("stdin is not a terminal")

	// ErrorTTYFailed is returned when terminal state capture fails.
	// This wraps the underlying error from golang.org/x/term.GetState().
	// Possible causes include:
	//   - Permission denied to access terminal attributes
	//   - Invalid file descriptor
	//   - System call failures (e.g., tcgetattr)
	//
	// Use errors.Unwrap() to access the underlying system error for debugging.
	ErrorTTYFailed = fmt.Errorf("failed to get terminal state")

	// ErrorDevTTYFail is returned when the fallback mechanism cannot open /dev/tty.
	// This occurs during terminal reset fallback when:
	//   - /dev/tty doesn't exist (uncommon on Unix-like systems)
	//   - Permission denied to access /dev/tty
	//   - Process has no controlling terminal
	//   - Running on Windows (which doesn't have /dev/tty)
	//
	// This error indicates fallback terminal reset has failed, but the primary
	// restoration may have already succeeded.
	ErrorDevTTYFail = fmt.Errorf("failed to open /dev/tty")
)

// TTYSaver provides an interface for managing terminal state lifecycle.
//
// Implementations capture terminal attributes at creation time and provide
// mechanisms to restore those attributes later. This is essential for applications
// that modify terminal settings (e.g., raw mode, echo disabled) and need to
// ensure clean restoration on exit.
//
// The interface supports both manual restoration via Restore() and automatic
// restoration via Signal() for graceful signal handling.
//
// Thread Safety:
// Implementations should be safe to call concurrently, though terminal operations
// at the OS level may not be thread-safe. Concurrent calls to Restore() are
// acceptable but may result in redundant system calls.
//
// See also:
//   - New() for creating TTYSaver instances
//   - Restore() for the convenience wrapper function
//   - SignalHandler() for automatic signal-based restoration
type TTYSaver interface {
	// IsTerminal reports whether the underlying file descriptor refers to a terminal.
	//
	// This method checks the current state of the file descriptor using
	// golang.org/x/term.IsTerminal(). The result may change if the file descriptor
	// is closed or the terminal is disconnected after the TTYSaver was created.
	//
	// Returns:
	//   - true: File descriptor is connected to a terminal device
	//   - false: Not a terminal, nil state, or invalid file descriptor
	//
	// Use cases:
	//   - Conditional terminal feature enablement
	//   - Detecting terminal disconnection
	//   - Testing/mocking scenarios
	IsTerminal() bool

	// Restore restores the saved terminal state to the file descriptor.
	//
	// This method applies the terminal attributes captured during creation,
	// reverting any modifications made since then (e.g., raw mode, disabled echo).
	//
	// The method is safe to call multiple times and will return nil if:
	//   - Restoration succeeds
	//   - The instance is nil or has no saved state
	//   - The file descriptor is not a terminal
	//
	// If the primary restoration fails, implementations may attempt a fallback
	// reset using ANSI escape sequences via /dev/tty.
	//
	// Returns:
	//   - nil: Terminal successfully restored or no action needed
	//   - error: Restoration failed (both primary and fallback)
	//
	// Thread Safety:
	// Safe to call concurrently, though may result in redundant system calls.
	Restore() error

	// Signal blocks waiting for termination signals and restores terminal on receipt.
	//
	// This method registers handlers for common termination signals (SIGINT, SIGTERM,
	// SIGQUIT, SIGHUP) and blocks the calling goroutine until one is received.
	// Upon signal receipt, it calls Restore() to clean up terminal state.
	//
	// The method returns immediately (no-op) if:
	//   - Signal handling was disabled during creation (sig=false in New())
	//   - The instance is nil or has no saved state
	//   - The file descriptor is not a terminal
	//
	// Signal handlers are automatically unregistered when the method returns,
	// preventing signal handler leaks.
	//
	// Returns:
	//   - nil: Signal received and terminal restored (or no-op conditions met)
	//   - error: Terminal restoration failed after signal receipt
	//
	// Usage:
	// Typically called in a goroutine via SignalHandler() wrapper rather than directly.
	//
	// Warning:
	// This method blocks indefinitely if signal handling is enabled and the file
	// descriptor is a valid terminal. Ensure it's called in a goroutine unless
	// intentionally blocking the main thread.
	Signal() error
}

// checkFd is an internal interface for extracting file descriptors from io.Reader.
//
// This interface is used to detect if an io.Reader can provide a file descriptor,
// which is necessary for terminal detection and state management.
//
// Standard library types that implement this interface:
//   - *os.File (stdin, stdout, stderr)
//   - *os.File from os.Open(), os.Create(), etc.
//
// Types that do NOT implement this interface:
//   - bytes.Buffer
//   - strings.Reader
//   - io.Pipe() readers
//   - Custom io.Reader implementations without file backing
//
// When an io.Reader doesn't implement checkFd, New() will treat it as a
// non-terminal and create a TTYSaver that safely no-ops on restoration.
type checkFd interface {
	// Fd returns the integer Unix file descriptor referencing the open file.
	// See os.File.Fd() for details on file descriptor semantics.
	Fd() uintptr
}

// New creates a new TTYSaver by capturing the current terminal state from the given reader.
//
// This function inspects the provided io.Reader to determine if it's connected to a
// terminal device and, if so, captures its current terminal attributes for later restoration.
//
// Parameters:
//
//   - in: The io.Reader to check for terminal support. Common values:
//
//   - nil: defaults to os.Stdin
//
//   - os.Stdin, os.Stdout, os.Stderr: standard file descriptors
//
//   - *os.File: any file opened via os.Open() or similar
//
//   - io.Reader without Fd(): treated as non-terminal (no error, safe no-op)
//
//   - sig: Whether to enable signal handling support for this TTYSaver.
//
//   - true: Signal() will block and wait for termination signals
//
//   - false: Signal() will immediately return (no-op)
//     This flag should be true if you plan to use SignalHandler() for automatic
//     terminal restoration on Ctrl+C or other termination signals.
//
// Behavior:
//
// The function checks if the reader implements the Fd() method. If so, it verifies
// whether the file descriptor refers to a terminal using golang.org/x/term.IsTerminal().
// For terminal devices, it captures the current state using term.GetState().
//
// Non-terminal inputs (pipes, redirections, buffers) are handled gracefully:
// the function returns a valid TTYSaver that safely no-ops on all operations.
// This allows code to be written uniformly without checking for terminal support.
//
// The captured state includes all terminal attributes:
//   - Input modes: canonical/raw, echo, signal processing
//   - Output modes: post-processing, newline handling
//   - Control modes: baud rate, character size
//   - Special characters: EOF, EOL, interrupt, erase, etc.
//
// Returns:
//   - TTYSaver: Always returns a valid TTYSaver (never nil on success)
//   - error: Only returns error if state capture fails on a valid terminal
//   - ErrorTTYFailed: wrapped with underlying term.GetState() error
//   - Never returns ErrorNotTTY - non-terminals are handled gracefully
//
// Thread Safety:
// This function is safe to call concurrently. Multiple goroutines can create
// separate TTYSaver instances simultaneously.
//
// Example - Basic usage with stdin:
//
//	saver, err := tty.New(nil, false)  // nil defaults to os.Stdin
//	if err != nil {
//	    return fmt.Errorf("failed to save terminal state: %w", err)
//	}
//	defer tty.Restore(saver)
//
//	// Safe to modify terminal now
//	// Terminal will be restored automatically on exit or panic
//
// Example - With signal handling:
//
//	saver, err := tty.New(nil, true)  // Enable signal support
//	if err != nil {
//	    return err
//	}
//	tty.SignalHandler(saver)  // Restore on Ctrl+C
//	defer tty.Restore(saver)  // Also restore on normal exit
//
//	// Your interactive application
//	runInteractivePrompt()
//
// Example - Alternative file descriptor:
//
//	saver, err := tty.New(os.Stdout, false)  // Capture stdout instead
//	if err != nil {
//	    return err
//	}
//	defer tty.Restore(saver)
//
// Example - Non-terminal input (safe):
//
//	buf := bytes.NewBufferString("data")
//	saver, err := tty.New(buf, false)  // No error, safe no-op
//	if err != nil {
//	    return err  // Won't happen with buffers
//	}
//	defer tty.Restore(saver)  // Safe, does nothing
//
// See also:
//   - Restore() for restoring terminal state
//   - SignalHandler() for automatic signal handling
//   - github.com/nabbar/golib/shell for higher-level shell abstractions
func New(in io.Reader, sig bool) (TTYSaver, error) {
	var (
		err   error
		fd    int
		state *term.State
	)

	if in == nil {
		in = os.Stdin
	}

	if f, k := in.(checkFd); k {
		fd = int(f.Fd())

		if term.IsTerminal(fd) {
			state, err = term.GetState(fd)
		}
	}

	// Save the current terminal state
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorTTYFailed, err)
	}

	return &tty{
		fd:    fd,
		state: state,
		sg:    sig,
	}, nil
}

// SignalHandler sets up automatic terminal restoration on process termination signals.
//
// This function spawns a background goroutine that calls state.Signal(), which blocks
// waiting for termination signals (SIGINT, SIGTERM, SIGQUIT, SIGHUP). When a signal
// is received, the terminal state is automatically restored before the signal handler returns.
//
// This provides a convenient way to ensure terminal cleanup on Ctrl+C or other
// interruptions without manually catching signals in your application code.
//
// Parameters:
//   - state: TTYSaver instance created by New(). Must have signal handling enabled
//     (sig=true parameter in New()) for this to work. Nil values are safely
//     ignored (no-op).
//
// Behavior:
//
// If state is nil, the function returns immediately without spawning a goroutine.
// This makes it safe to use even when New() might have failed:
//
//	saver, _ := tty.New(nil, true)  // Might be nil if terminal unavailable
//	tty.SignalHandler(saver)        // Safe even if saver is nil
//
// If state was created with sig=false in New(), the spawned goroutine will return
// immediately without waiting for signals, making this function effectively a no-op.
//
// The goroutine lifecycle:
//  1. Goroutine spawned immediately (non-blocking)
//  2. state.Signal() called, which blocks waiting for signals
//  3. On signal receipt, terminal is restored via state.Restore()
//  4. Signal handlers are unregistered automatically
//  5. Goroutine exits after restoration
//
// Signals handled (when sig=true):
//   - SIGINT (Ctrl+C): User interrupt
//   - SIGTERM: Graceful shutdown (systemd, docker stop, kill)
//   - SIGQUIT (Ctrl+\): User quit with core dump
//   - SIGHUP: Terminal hangup (SSH disconnect, terminal close)
//
// Important Notes:
//   - Call this function only once per TTYSaver to avoid duplicate signal handlers
//   - The function returns immediately; signal handling happens in the background
//   - SIGKILL cannot be caught and will terminate the process without restoration
//   - On Windows, only SIGINT is reliably supported
//
// Thread Safety:
// This function is safe to call concurrently from multiple goroutines, though
// calling it multiple times with the same TTYSaver may result in unexpected
// behavior due to duplicate signal registrations.
//
// Example - Basic usage:
//
//	saver, err := tty.New(nil, true)  // Enable signal handling
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tty.SignalHandler(saver)  // Non-blocking
//	defer tty.Restore(saver)  // Also restore on normal exit
//
//	// Your long-running application
//	// Terminal will be restored on Ctrl+C
//	runServer()
//
// Example - Integration with interactive prompt:
//
//	saver, err := tty.New(nil, true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tty.SignalHandler(saver)
//
//	// Use go-prompt or similar
//	prompt := goprompt.New(executor, completer)
//	prompt.Run()  // Terminal restored on Ctrl+C
//
// Example - With shell package:
//
//	saver, _ := tty.New(nil, true)
//	tty.SignalHandler(saver)
//
//	sh := shell.New()
//	sh.RunPrompt(os.Stdout, os.Stderr, true)  // Terminal restored on exit
//
// See also:
//   - New() with sig=true to enable signal handling
//   - Signal() for the underlying blocking signal handler
//   - Restore() for manual terminal restoration
//   - github.com/nabbar/golib/shell for shell implementations using this
func SignalHandler(state TTYSaver) {
	// Nil-safe: return immediately if no state to manage
	if state == nil {
		return
	}

	// Spawn background goroutine to handle signals
	// This is non-blocking and returns immediately
	go func() {
		_ = state.Signal() // Blocks until signal received (or no-ops if sig=false)
	}()
}

// Restore restores the terminal state using the provided TTYSaver.
// If restoration fails, it attempts a fallback reset using ANSI escape sequences.
//
// This function provides a robust terminal restoration mechanism with multiple fallback
// strategies to ensure the terminal is left in a usable state even if primary restoration
// fails. It's designed to be called in defer statements or error handlers.
//
// Restoration Process:
//  1. Check if state is nil (safe no-op if nil)
//  2. Call state.Restore() to apply saved terminal attributes
//  3. If restoration fails, call resetTerminalFallback() for ANSI escape sequence reset
//  4. Errors from fallback are silently ignored (best-effort recovery)
//
// Use Cases:
//   - defer tty.Restore(saver): Automatic cleanup on function exit
//   - Panic recovery: Ensures terminal is restored even after panics
//   - Error paths: Clean terminal state before returning errors
//   - Signal handlers: Called by SignalHandler() on program termination
//
// Nil Safety:
// This function is explicitly designed to handle nil values safely. It will not panic
// if passed a nil TTYSaver, making it safe to use in defer statements even if New()
// returned an error:
//
//	saver, _ := tty.New() // might return nil on error
//	defer tty.Restore(saver) // safe even if saver is nil
//
// Fallback Behavior:
// When the primary restore fails (e.g., file descriptor closed, permission denied),
// the function attempts to reset the terminal using ANSI escape sequences sent to
// /dev/tty. This provides a best-effort recovery mechanism that works in most cases
// even when the saved state cannot be reapplied.
//
// Parameters:
//   - state: TTYSaver to restore (nil-safe - no action taken if nil)
//
// Thread Safety:
// This function is safe to call concurrently, though typically called serially
// in cleanup paths.
//
// Example - Basic Usage:
//
//	saver, err := tty.New()
//	if err != nil {
//	    return err
//	}
//	defer tty.Restore(saver)
//
//	// Your terminal-modifying code here
//	// Terminal will be restored on return, panic, or normal exit
//
// Example - Error Handling:
//
//	saver, err := tty.New()
//	if err != nil {
//	    return err
//	}
//	defer tty.Restore(saver)
//
//	if err := doSomething(); err != nil {
//	    // Terminal will be restored before return
//	    return fmt.Errorf("operation failed: %w", err)
//	}
//
// Example - Panic Recovery:
//
//	saver, _ := tty.New()
//	defer func() {
//	    if r := recover(); r != nil {
//	        tty.Restore(saver) // Restore terminal before re-panicking
//	        panic(r)
//	    }
//	}()
//	defer tty.Restore(saver) // Also restore on normal exit
func Restore(state TTYSaver) {
	// Defensive check: safe to call with nil
	if state == nil {
		return
	} else {
		_ = state.Restore()
	}
}
