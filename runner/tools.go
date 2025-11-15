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

package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
)

// FunCheck is a function type that checks if a condition is met.
// It returns true if the condition is satisfied, false otherwise.
// Used by RunNbr and RunTick for polling operations.
type FunCheck func() bool

// FunRun is a function type that performs an action between checks.
// Typically used for sleep operations or state updates in polling loops.
// Used by RunNbr and RunTick for actions between condition checks.
type FunRun func()

// RunNbr performs a polling operation up to max attempts, executing chk() to test
// a condition and run() between checks. This is useful for retry logic with a
// fixed number of attempts.
//
// The function executes in this order:
//  1. Check condition with chk()
//  2. If true, return true immediately
//  3. Otherwise, execute run() (typically a sleep or state update)
//  4. Repeat up to max times
//  5. Perform final check and return result
//
// Parameters:
//   - max: Maximum number of retry attempts (0 means only final check)
//   - chk: Function to check if condition is met
//   - run: Function to execute between checks (e.g., time.Sleep)
//
// Returns true if the condition is met within max attempts, false otherwise.
//
// Example:
//
//	success := runner.RunNbr(10,
//	    func() bool { return server.IsReady() },
//	    func() { time.Sleep(100 * time.Millisecond) },
//	)
func RunNbr(max uint8, chk FunCheck, run FunRun) bool {
	var i uint8

	for i = 0; i < max; i++ {
		if chk() {
			return true
		}

		run()
	}

	return chk()
}

// RunTick performs a polling operation with timeout and tick interval, executing
// chk() to test a condition and run() between checks. This is useful for waiting
// on async operations with both time limits and context cancellation support.
//
// The function uses a time.Ticker to check the condition at regular intervals
// until either the condition is met, the context is cancelled, or the maximum
// duration is exceeded.
//
// Parameters:
//   - ctx: Context for cancellation (returns false if cancelled)
//   - tick: Interval between checks
//   - max: Maximum total duration to wait
//   - chk: Function to check if condition is met
//   - run: Function to execute between checks
//
// Returns:
//   - true if the condition is met within max duration
//   - false if context is cancelled or max duration exceeded
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	ready := runner.RunTick(ctx,
//	    500*time.Millisecond,  // Check every 500ms
//	    10*time.Second,         // Give up after 10s
//	    func() bool { return db.IsConnected() },
//	    func() { log.Println("Waiting for database...") },
//	)
func RunTick(ctx context.Context, tick, max time.Duration, chk FunCheck, run FunRun) bool {
	var (
		s = time.Now()
		t = time.NewTicker(tick)
	)

	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return false

		case <-t.C:
			if chk() {
				return true
			}

			run()

			if time.Since(s) >= max {
				return chk()
			}
		}
	}
}

// RecoveryCaller logs a recovered panic with a stack trace to stderr.
// This function is safe to call with nil rec (it will do nothing), making it
// convenient for use in defer statements with recover().
//
// It prints:
//   - The process name and panic value
//   - Any additional data passed via the data parameter
//   - Up to 10 frames of stack trace with file paths and line numbers
//
// The output is written to stderr to avoid interfering with normal program output.
// This function is used internally by all runner implementations to prevent panics
// from crashing the entire process.
//
// Parameters:
//   - proc: A descriptive name for the process/function where the panic occurred
//   - rec: The value returned by recover() (can be nil)
//   - data: Optional additional data to include in the output
//
// Example usage in a defer statement:
//
//	defer func() {
//	    runner.RecoveryCaller("golib/server/startstop/start", recover())
//	}()
//
// Example with additional data:
//
//	defer func() {
//	    runner.RecoveryCaller("myservice/worker", recover(), "WorkerID:", workerID)
//	}()
//
// Output format:
//
//	Recovering process 'process-name': panic-value
//	additional-data
//	  trace #0 => Line: 123 - File: /path/to/file.go
//	  trace #1 => Line: 456 - File: /path/to/other.go
func RecoveryCaller(proc string, rec any, data ...any) {
	if rec == nil {
		return
	}

	var (
		buf = bytes.NewBuffer(make([]byte, 0))

		// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need.
		pCnt = make([]uintptr, 10, 255)
		nCnt = runtime.Callers(1, pCnt)
	)

	// Header line describing the recovered panic. Fix typo: "Recovering".
	buf.WriteString(fmt.Sprintf("Recovering process '%s': %v\n", proc, rec)) // nolint
	for _, d := range data {
		buf.WriteString(fmt.Sprintf("%v\n", d)) // nolint
	}

	if nCnt > 0 {
		var (
			frames = runtime.CallersFrames(pCnt[:nCnt])
			more   = true
			lCnt   = 0
		)

		for more && lCnt < 10 {
			var frame runtime.Frame
			frame, more = frames.Next()

			if len(frame.File) > 0 {
				buf.WriteString(fmt.Sprintf("  trace #%d => Line: %d - File: %s\n", lCnt, frame.Line, frame.File)) // nolint
				lCnt++
			} else if len(frame.Function) > 0 {
				buf.WriteString(fmt.Sprintf("  trace #%d => Line: %d - Func: %s\n", lCnt, frame.Line, frame.Function)) // nolint
				lCnt++
			}
		}
	}

	if buf.Len() > 0 {
		// Print as string to avoid byte slice numeric representation in output.
		_, _ = fmt.Fprint(os.Stderr, buf.String())
	}
}
