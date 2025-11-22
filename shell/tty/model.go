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

package tty

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// tty is the internal implementation of TTYSaver interface.
//
// This struct encapsulates all state needed to manage terminal attributes:
// the file descriptor, saved terminal state, signal handling configuration,
// and terminal detection cache.
//
// Design Philosophy:
//
// The struct is intentionally unexported to enforce the factory pattern via New().
// This ensures:
//   - Proper initialization and validation
//   - Consistent error handling
//   - Safe defaults for all fields
//   - Prevention of invalid state combinations
//
// After creation, the struct is effectively immutable - all fields are set during
// initialization and never modified. This makes it safe to use across goroutines
// without additional synchronization.
//
// Fields:
//
//   - fd: Unix file descriptor of the terminal device. Typically 0 (stdin), but
//     can be 1 (stdout) or 2 (stderr) or any file descriptor from os.Open().
//     Used for all terminal operations via golang.org/x/term.
//
//   - state: Pointer to captured terminal attributes (termios structure).
//     Obtained via term.GetState() during creation. May be nil if the file
//     descriptor is not a terminal or state capture failed. Contains:
//
//   - Input flags (echo, canonical mode, signal processing)
//
//   - Output flags (post-processing, CR/LF conversion)
//
//   - Control flags (baud rate, character size)
//
//   - Local flags and special characters
//
//   - sg: Signal handling enablement flag. When true, Signal() will block waiting
//     for termination signals. When false, Signal() returns immediately (no-op).
//     Set via the sig parameter in New().
//
//   - tm: Terminal detection cache (currently unused). Reserved for future
//     optimization to avoid repeated IsTerminal() system calls.
//
// Memory Layout:
//
// The struct is compact (~24 bytes on 64-bit systems):
//   - fd: 8 bytes (int on 64-bit)
//   - state: 8 bytes (pointer)
//   - sg: 1 byte (bool)
//   - padding: ~6 bytes (struct alignment)
//
// The actual termios data is allocated separately and referenced by state pointer.
// This design keeps tty itself lightweight for passing by pointer.
//
// Thread Safety:
//
// The struct is safe for concurrent use because:
//   - All fields are read-only after initialization
//   - No internal mutexes needed
//   - Concurrent Restore() calls are safe (idempotent)
//   - Signal() should only be called once per instance
//
// See also:
//   - New() for creation
//   - TTYSaver interface for public API
//   - golang.org/x/term.State for termios details
type tty struct {
	fd    int         // Terminal file descriptor (usually stdin/0)
	state *term.State // Saved terminal attributes (termios structure)
	sg    bool        // Signal handling enabled (true = Signal() blocks)
}

// IsTerminal reports whether the file descriptor refers to a terminal device.
//
// This method implements the TTYSaver interface and provides real-time terminal
// detection. Unlike the cached tm field, this method queries the OS on every call
// to handle scenarios where the terminal state changes after creation.
//
// Implementation:
//
// The method delegates to golang.org/x/term.IsTerminal(), which on Unix systems
// uses the tcgetattr system call to check if the file descriptor supports terminal
// operations. On Windows, it checks if the handle is a console.
//
// Return values:
//
//   - false: Returned in these cases:
//
//   - Receiver is nil (defensive programming)
//
//   - state field is nil (no terminal state was captured)
//
//   - File descriptor is not a terminal device
//
//   - File descriptor has been closed
//
//   - Terminal was disconnected (SSH disconnect, PTY closed)
//
//   - true: File descriptor is currently connected to a terminal device
//
// The result may change over time if:
//   - The file descriptor is closed via Close()
//   - SSH session disconnects
//   - Terminal is detached (tmux, screen)
//   - Process loses controlling terminal
//
// Use Cases:
//
//  1. Conditional feature enablement:
//     if saver.IsTerminal() {
//     enableColorOutput()
//     enableInteractiveMode()
//     }
//
//  2. Terminal disconnection detection:
//     if !saver.IsTerminal() {
//     log.Println("Terminal disconnected, switching to batch mode")
//     }
//
//  3. Testing and mocking:
//     Mock implementations can return false to simulate non-terminal environments
//
// Performance:
//
// Each call performs a system call (tcgetattr on Unix). For performance-critical
// code that checks terminal status frequently, consider caching the result if
// terminal state changes are not expected.
//
// Thread Safety:
//
// This method is safe to call concurrently. The underlying system call (tcgetattr)
// is atomic from the caller's perspective.
//
// See also:
//   - golang.org/x/term.IsTerminal() for the underlying implementation
//   - New() for initial terminal detection during creation
func (t *tty) IsTerminal() bool {
	// Defensive nil checks prevent panics in error paths
	if t == nil || t.state == nil {
		return false
	}

	// Query current terminal status from OS
	// This may differ from creation time if terminal was disconnected
	return term.IsTerminal(t.fd)
}

// Restore restores the terminal to its saved state.
// This method implements the TTYSaver interface and provides the primary mechanism
// for returning the terminal to its original configuration.
//
// Implementation Details:
// The method calls golang.org/x/term.Restore() which uses the tcsetattr system call
// to apply the saved termios structure to the terminal. This restores all terminal
// attributes including:
//   - Input modes (ICANON, ECHO, ISIG, etc.)
//   - Output modes (OPOST, ONLCR, etc.)
//   - Control modes (baud rate, character size, etc.)
//   - Local modes (canonical vs raw mode, etc.)
//   - Special characters (EOF, EOL, ERASE, etc.)
//
// Defensive Programming:
// The method includes nil checks for both the receiver and the state field,
// making it safe to call even on improperly initialized instances. This prevents
// panics in error recovery paths.
//
// Error Conditions:
// Restoration may fail if:
//   - The file descriptor is no longer valid (terminal closed)
//   - The process lacks permission to modify terminal attributes
//   - The terminal was disconnected (SSH disconnect, PTY closed)
//   - The file descriptor doesn't refer to a terminal anymore
//
// When restoration fails, the error is wrapped with context and returned to the
// caller, which typically triggers a fallback reset using ANSI escape sequences.
//
// Thread Safety:
// This method is safe to call concurrently on the same instance, though terminal
// operations themselves may not be thread-safe at the OS level.
//
// Returns:
//   - nil: Terminal successfully restored or instance is nil/invalid (no-op)
//   - error: Wrapped error from term.Restore with additional context
func (t *tty) Restore() error {
	// Defensive check: allow calling on nil or incomplete state
	if t == nil || t.state == nil {
		return nil
	} else if !term.IsTerminal(t.fd) {
		return nil
	}

	// Apply the saved terminal state
	if term.Restore(t.fd, t.state) == nil {
		return nil
	}

	if err := resetTerminalFallback(); err == nil {
		return nil
	} else {
		return err
	}
}

// Signal blocks waiting for termination signals and restores terminal state on receipt.
//
// This method implements the TTYSaver interface and provides the core signal handling
// mechanism for graceful terminal restoration. It's designed to be called from a
// background goroutine via SignalHandler().
//
// Behavior:
//
// The method registers handlers for common termination signals and blocks the calling
// goroutine until one is received. Upon signal receipt:
//  1. The signal handler is unregistered (defer signal.Stop)
//  2. Terminal state is restored via Restore()
//  3. The method returns with the restoration result
//
// No-op Conditions (immediate return):
//
// The method returns nil immediately without blocking if:
//   - Receiver is nil (defensive nil check)
//   - state field is nil (no terminal state captured)
//   - sg field is false (signal handling disabled in New())
//   - File descriptor is not a terminal (checked via IsTerminal)
//
// This design ensures the method is always safe to call and degrades gracefully
// in non-terminal environments.
//
// Signals Handled:
//
// When signal handling is active, the method waits for these signals:
//   - os.Interrupt (SIGINT): Ctrl+C from user, common in interactive sessions
//   - syscall.SIGTERM: Graceful shutdown request from systemd, docker, or kill
//   - syscall.SIGQUIT: Ctrl+\ from user, typically requests core dump
//   - syscall.SIGHUP: Terminal hangup from SSH disconnect or terminal close
//
// On Windows, only os.Interrupt (SIGINT) is reliably supported. Other signals
// may not be delivered or may have different semantics.
//
// Signal Channel:
//
// The method uses a buffered channel (capacity 1) to prevent signal loss if a
// signal arrives between registration and the blocking receive. This ensures
// reliable signal delivery even in race conditions.
//
// Signal Handler Cleanup:
//
// The defer signal.Stop(sigChan) ensures signal handlers are properly unregistered
// when the method returns. This prevents:
//   - Signal handler leaks in long-running processes
//   - Goroutine leaks from orphaned signal goroutines
//   - Unexpected behavior if Signal() is called multiple times
//
// Returns:
//
//   - nil: One of these occurred:
//
//   - Signal received and terminal successfully restored
//
//   - No-op conditions met (see above)
//
//   - error: Terminal restoration failed after signal receipt.
//     The error is from Restore() and may indicate:
//
//   - File descriptor closed
//
//   - Permission denied
//
//   - Terminal disconnected
//
// Thread Safety:
//
// While the method itself is thread-safe, it should only be called once per
// TTYSaver instance. Multiple concurrent calls will result in:
//   - Duplicate signal handler registrations
//   - Undefined behavior regarding which handler receives the signal
//   - Potential goroutine leaks
//
// The typical pattern is:
//
//	go func() { _ = saver.Signal() }()  // Call once in a goroutine
//
// Usage Pattern:
//
// This method is rarely called directly. Instead, use the SignalHandler() wrapper:
//
//	saver, _ := tty.New(nil, true)  // Enable signal handling
//	tty.SignalHandler(saver)        // Spawns goroutine calling Signal()
//	defer tty.Restore(saver)        // Also restore on normal exit
//
// Direct usage (advanced):
//
//	saver, _ := tty.New(nil, true)
//	go func() {
//	    if err := saver.Signal(); err != nil {
//	        log.Printf("Failed to restore terminal: %v", err)
//	    }
//	}()
//
// Warning:
//
// This method blocks indefinitely when signal handling is enabled and the file
// descriptor is a valid terminal. Always call it in a separate goroutine unless
// you intend to block the current goroutine.
//
// SIGKILL Limitation:
//
// SIGKILL (kill -9) cannot be caught and will terminate the process immediately
// without calling signal handlers. Terminal state will not be restored in this case.
// Use SIGTERM for graceful shutdown whenever possible.
//
// See also:
//   - SignalHandler() for the recommended high-level wrapper
//   - Restore() for manual terminal restoration
//   - New() with sig=true to enable signal handling
//   - os/signal package for signal handling details
func (t *tty) Signal() error {
	// Fast path: return immediately if signal handling is disabled or impossible
	if t == nil || t.state == nil || !t.sg {
		return nil
	}

	// Check if file descriptor is still a terminal
	// Terminal may have been disconnected after creation
	if !term.IsTerminal(t.fd) {
		return nil
	}

	// Create buffered channel to avoid missing signals during setup
	// Buffer size of 1 is sufficient since we only need one signal to trigger
	sigChan := make(chan os.Signal, 1)

	// Register for common termination signals
	// Signal delivery is asynchronous and may occur immediately
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	// Ensure we unregister signal handlers when done
	// This prevents goroutine/handler leaks in long-running processes
	defer signal.Stop(sigChan)

	// Block until a signal is received
	// This is the main blocking point of the method
	<-sigChan

	// Signal received - restore terminal and return result
	return t.Restore()
}

// resetTerminalFallback attempts emergency terminal reset using ANSI escape sequences.
//
// This is a last-resort fallback mechanism invoked when normal terminal restoration
// via term.Restore() fails. It provides best-effort terminal cleanup by sending
// standardized ANSI control sequences directly to the controlling terminal device.
//
// Why This Fallback Exists:
//
// The primary restoration method (term.Restore) may fail if:
//   - File descriptor becomes invalid (closed, disconnected)
//   - Process loses permission to modify terminal attributes
//   - Terminal driver is in an inconsistent state
//   - SSH connection breaks during restoration
//
// In these cases, the terminal may be left in an unusable state (no echo, raw mode,
// hidden cursor, etc.). This fallback provides a recovery mechanism that works even
// when the file descriptor is unusable, as long as the process has a controlling terminal.
//
// Implementation Strategy:
//
// Instead of using the (potentially invalid) original file descriptor, this function:
//  1. Opens /dev/tty directly (the controlling terminal device)
//  2. Sends ANSI reset sequences to restore terminal to sane defaults
//  3. Closes /dev/tty regardless of success or failure
//
// Why /dev/tty:
//
// The function specifically uses /dev/tty rather than stdin/stdout because:
//   - stdin/stdout/stderr may be redirected to files, pipes, or sockets
//   - /dev/tty always refers to the process's controlling terminal
//   - Direct terminal access bypasses all redirections
//   - Works even when stdio descriptors are closed or invalid
//   - Ensures reset sequences reach the actual terminal device user is viewing
//
// Example scenario where /dev/tty is essential:
//
//	./program < input.txt > output.txt  # stdin/stdout redirected
//	# But /dev/tty still points to the terminal where user sees output
//
// ANSI Control Sequences:
//
// The function sends three carefully chosen reset sequences in order:
//
//  1. ESC c (RIS - Reset to Initial State)
//     - Hex: 0x1B 0x63
//     - Effect: Full terminal reset equivalent to power cycling
//     - Clears screen buffer and scrollback
//     - Resets all modes to power-on defaults
//     - Restores default character sets
//     - Most comprehensive reset available
//     - Supported since VT100 (1978), universal on modern terminals
//
//  2. ESC [?25h (DECTCEM - DEC Text Cursor Enable Mode)
//     - Hex: 0x1B 0x5B 0x3F 0x32 0x35 0x68
//     - Effect: Makes cursor visible
//     - Critical for user interaction (invisible cursor is very confusing)
//     - Prevents "blind typing" where user can't see cursor position
//     - Some applications hide cursor for UI effects and may not restore it
//
//  3. ESC [0m (SGR 0 - Select Graphic Rendition Reset)
//     - Hex: 0x1B 0x5B 0x30 0x6D
//     - Effect: Resets all text attributes to defaults
//     - Clears foreground/background colors
//     - Removes bold, dim, italic, underline, blink, reverse, strikethrough
//     - Ensures subsequent text is rendered normally
//
// These sequences are ordered from most to least comprehensive. If RIS succeeds,
// the others are redundant but harmless. If RIS fails, the others provide partial recovery.
//
// Error Handling Philosophy:
//
// This function follows a "best-effort, never fail catastrophically" approach:
//   - Returns ErrorDevTTYFail only if /dev/tty cannot be opened
//   - Write errors are silently ignored (terminal may not support sequences)
//   - File descriptor is always closed via defer
//   - No panics, no fatal errors
//
// The rationale: if we've reached this fallback, the terminal is already in a bad
// state. Partial success is better than no attempt, and we should never make things
// worse by panicking or failing completely.
//
// Platform Support:
//
// Unix-like systems (Linux, macOS, BSD, Solaris):
//   - /dev/tty exists and is the standard controlling terminal device
//   - ANSI escape sequences are universally supported
//   - Function works as designed
//
// Windows:
//   - No /dev/tty equivalent (Windows uses console handles differently)
//   - os.OpenFile("/dev/tty", ...) will fail with "file not found"
//   - Function returns ErrorDevTTYFail
//   - Windows terminal state is typically managed differently anyway
//
// Containers/Virtual Environments:
//   - /dev/tty availability depends on container configuration
//   - Docker with -t flag: /dev/tty available
//   - Docker without -t flag: /dev/tty not available (returns ErrorDevTTYFail)
//   - SSH sessions: /dev/tty available
//
// Invocation Context:
//
// This function is called only by Restore() when term.Restore() fails:
//
//	if term.Restore(t.fd, t.state) == nil {
//	    return nil
//	}
//	// Primary restoration failed, try fallback
//	if err := resetTerminalFallback(); err == nil {
//	    return nil  // Fallback succeeded
//	}
//
// It's not called for non-terminals or when primary restoration succeeds.
//
// Limitations:
//
// This fallback cannot restore:
//   - Specific baud rates (though rarely modified)
//   - Custom special characters (EOF, EOL, ERASE, etc.)
//   - Specific input/output flags
//   - Terminal window size
//   - Any application-specific terminal state
//
// It only provides a "factory reset" to standard defaults, which is usually
// sufficient to make the terminal usable again.
//
// Terminal Compatibility:
//
// The sequences are chosen for maximum compatibility:
//   - VT100 and all descendants (xterm, gnome-terminal, konsole, iTerm2, etc.)
//   - Linux console
//   - macOS Terminal.app
//   - Most SSH terminal emulators
//
// Very old or specialized terminals might not support all sequences, but this is
// increasingly rare. Unsupported sequences are typically ignored (safe).
//
// Returns:
//
//   - nil: /dev/tty opened and reset sequences sent successfully.
//     Terminal likely restored to usable state.
//
//   - ErrorDevTTYFail: Could not open /dev/tty. Occurs when:
//
//   - Running on Windows
//
//   - No controlling terminal (daemon, background job)
//
//   - Permission denied
//
//   - Inside container without TTY allocation
//
// See also:
//   - Restore() which calls this as fallback
//   - golang.org/x/term.Restore() for primary restoration
//   - ECMA-48 standard for ANSI escape sequence specifications
func resetTerminalFallback() error {
	// Open /dev/tty directly to bypass stdio redirections
	// os.O_WRONLY: We only need to write reset sequences
	// Mode 0: Not creating file, so mode is ignored
	ttyFile, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		// /dev/tty doesn't exist or can't be opened
		// This is expected on Windows or in non-terminal environments
		return ErrorDevTTYFail
	}

	// Ensure file is closed regardless of write success
	// Using named function for clarity in defer
	defer func() {
		_ = ttyFile.Close()
	}()

	// Send reset sequences (errors intentionally ignored - best effort)
	// Even if one write fails, others might succeed
	_, _ = ttyFile.WriteString("\033c")     // RIS: Full terminal reset
	_, _ = ttyFile.WriteString("\033[?25h") // DECTCEM: Show cursor
	_, _ = ttyFile.WriteString("\033[0m")   // SGR: Reset text attributes

	// If we got here, /dev/tty was opened and sequences were sent
	// Whether the terminal actually reset depends on terminal emulator support
	return nil
}
