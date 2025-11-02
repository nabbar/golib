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

package types

import (
	"context"
	"encoding"
	"encoding/json"

	shlcmd "github.com/nabbar/golib/shell/command"
)

// FuncPool is a function type that returns a Pool instance.
// This is used for dependency injection and lazy initialization of pools.
type FuncPool func() Pool

// PoolManage defines the interface for managing monitors within a pool.
// It provides CRUD operations and iteration capabilities for monitors.
type PoolManage interface {
	// MonitorAdd adds a monitor to the pool.
	// Returns an error if the monitor is nil or has an empty name.
	// If the pool is running, the monitor is automatically started.
	MonitorAdd(mon Monitor) error

	// MonitorGet retrieves a monitor from the pool by name.
	// Returns nil if the monitor is not found.
	MonitorGet(name string) Monitor

	// MonitorSet updates or adds a monitor in the pool.
	// If the monitor doesn't exist, it will be added.
	// Returns an error if the monitor is nil or has an empty name.
	MonitorSet(mon Monitor) error

	// MonitorDel removes a monitor from the pool by name.
	// Does nothing if the monitor doesn't exist.
	MonitorDel(name string)

	// MonitorList returns a slice of all monitor names in the pool.
	MonitorList() []string

	// MonitorWalk iterates over monitors in the pool, calling the provided function for each.
	// The iteration stops if the function returns false.
	// The validName parameter optionally filters which monitors to visit.
	MonitorWalk(fct func(name string, val Monitor) bool, validName ...string)
}

// PoolStatus combines pool management with encoding capabilities.
// It allows pools to be marshaled to text and JSON formats.
type PoolStatus interface {
	encoding.TextMarshaler
	json.Marshaler
	PoolManage
}

// PoolShell provides shell command interface for pool operations.
// This enables CLI-style control of the pool and its monitors.
type PoolShell interface {
	// GetShellCommand returns a list of available shell commands for pool operations.
	// Common commands include: list, info, start, stop, restart, and status.
	GetShellCommand(ctx context.Context) []shlcmd.Command
}

// Pool is the main interface for monitor pool management.
// It combines status tracking, monitor management, and shell command capabilities.
type Pool interface {
	PoolStatus
	PoolShell
}
