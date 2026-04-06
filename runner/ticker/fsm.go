/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package ticker

// state is an internal type used to represent the current operational status of the ticker runner.
// It is used within a Finite State Machine (FSM) pattern to ensure that the runner transitions
// between lifecycle phases (starting, running, stopping) in a predictable and thread-safe manner.
type state uint32

const (
	// stopped (value 0) signifies that the ticker runner is completely inactive.
	// In this state, no background goroutine exists, no timers are ticking,
	// and the uptime is zeroed. This is the initial state of a new Ticker.
	stopped state = iota

	// started (value 1) is a transitional state indicating that the Start procedure
	// has been called and the background goroutine is currently being initialized.
	// This state prevents multiple concurrent Start calls from spawning multiple goroutines.
	started

	// running (value 2) indicates that the background goroutine is fully operational.
	// In this state, the execution loop is actively waiting for ticks from 'time.Ticker'
	// and will execute the user-defined function upon each tick.
	running

	// reqStop (value 4) is a request signal state. It acts as an asynchronous flag
	// to notify the active background goroutine that it should terminate gracefully.
	// The main loop checks for this state at each iteration or via a dedicated polling timer.
	reqStop
)

// getState is a helper function that casts a raw uint32 value (typically from an atomic Load)
// back into the internal 'state' type for easier comparison and readability.
func getState(i uint32) state {
	return state(i)
}

// Uint32 returns the numerical value of the state as a uint32.
// This is primarily used as an argument for 'sync/atomic' Store and Swap operations.
func (s state) Uint32() uint32 {
	return uint32(s)
}

// getState returns the current state of the ticker runner implementation ('run' struct).
// It uses an atomic Load operation on the internal state field 's' to ensure thread-safety
// without the need for a full mutex lock, making status checks extremely fast.
func (o *run) getState() state {
	return getState(o.s.Load())
}

// setState updates the current state of the ticker runner implementation ('run' struct).
// It uses an atomic Store operation to ensure the update is immediately visible to all
// goroutines, maintaining the integrity of the Finite State Machine.
func (o *run) setState(s state) {
	o.s.Store(s.Uint32())
}
