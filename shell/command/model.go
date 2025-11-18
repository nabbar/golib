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

package command

import (
	"io"
)

// model is the internal implementation of the Command interface.
// It stores the command's metadata (name and description) and its execution function.
// All fields are immutable after creation, making the struct thread-safe.
type model struct {
	n string  // n holds the command name
	d string  // d holds the command description
	r FuncRun // r is the optional function to execute when Run is called
}

// Name returns the command's name.
// It implements the CommandInfo interface.
//
// Returns an empty string if the receiver is nil (defensive programming).
// This method is safe for concurrent use.
func (o *model) Name() string {
	// Defensive check: prevent panic if called on a nil receiver
	if o == nil {
		return ""
	}

	return o.n
}

// Describe returns the command's human-readable description.
// It implements the CommandInfo interface.
//
// Returns an empty string if the receiver is nil (defensive programming).
// This method is safe for concurrent use.
func (o *model) Describe() string {
	// Defensive check: prevent panic if called on a nil receiver
	if o == nil {
		return ""
	}

	return o.d
}

// Run executes the command with the provided output writers and arguments.
// It implements the Command interface.
//
// Parameters:
//   - buf: Writer for standard output (can be nil if the function handles it)
//   - err: Writer for error output (can be nil if the function handles it)
//   - args: Slice of string arguments to pass to the command (can be nil or empty)
//
// Behavior:
//   - If the receiver is nil, the method returns immediately (defensive programming)
//   - If the function (r) is nil, the method returns immediately (no-op command)
//   - Otherwise, the stored function is invoked with the provided parameters
//
// This method is safe for concurrent use as it only reads immutable fields.
// However, the actual function execution's thread-safety depends on the FuncRun implementation.
func (o *model) Run(buf io.Writer, err io.Writer, args []string) {
	// Defensive checks: prevent panic and handle no-op commands
	if o == nil || o.r == nil {
		return
	}

	// Invoke the stored function with the provided parameters
	o.r(buf, err, args)
}
