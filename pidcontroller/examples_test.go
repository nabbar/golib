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

package pidcontroller_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nabbar/golib/pidcontroller"
)

// ExamplePID_Range demonstrates how to use the Range function to generate a smooth transition.
func ExamplePID_Range() {
	// Initialize a PID controller with specific gains.
	// Kp = 0.5 (Proportional)
	// Ki = 0.1 (Integral)
	// Kd = 0.2 (Derivative)
	pid := pidcontroller.New(0.1, 0.05, 0.05)

	// Generate a range of values from 0 to 10.
	// The PID controller will determine the steps based on the gains.
	values := pid.Range(0, 10)

	// Output the generated sequence.
	fmt.Printf("Generated %d steps\n", len(values))
	if len(values) > 0 {
		fmt.Printf("First step: %.2f\n", values[0])
		fmt.Printf("Last step: %.2f\n", values[len(values)-1])
	}
	// Output:
	// Generated 6 steps
	// First step: 2.00
	// Last step: 10.00
}

// ExamplePID_RangeCtx demonstrates how to use RangeCtx with a timeout.
func ExamplePID_RangeCtx() {
	// Initialize a PID controller.
	pid := pidcontroller.New(0.1, 0.0, 0.0)

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Generate a range of values from 0 to 100.
	values := pid.RangeCtx(ctx, 0, 100)

	fmt.Printf("Start: %.2f\n", values[0])
	fmt.Printf("Target reached: %v\n", values[len(values)-1] == 100.0)
	// Output:
	// Start: 10.00
	// Target reached: true
}

// Example_duration demonstrates how to use the PID controller to generate a sequence of durations
// between a minimum and maximum wait time. This can be useful for implementing backoff strategies.
func Example_duration() {
	// Initialize a PID controller suitable for duration calculation.
	// A lower Kp will result in more, smaller steps.
	pid := pidcontroller.New(0.05, 0.01, 0.01)

	// Define the minimum and maximum durations (e.g., 5s to 1m).
	minDuration := 5 * time.Second
	maxDuration := 1 * time.Minute

	// Convert durations to float64 (nanoseconds) for the PID controller.
	start := float64(minDuration)
	end := float64(maxDuration)

	// Generate the sequence of durations.
	steps := pid.Range(start, end)

	fmt.Printf("Generating backoff steps from %s to %s\n", minDuration, maxDuration)
	fmt.Printf("Total steps: %d\n", len(steps))

	if len(steps) > 0 {
		// Convert the first and last step back to time.Duration for display.
		firstStep := time.Duration(steps[0])
		lastStep := time.Duration(steps[len(steps)-1])
		fmt.Printf("First backoff: %s\n", firstStep.Round(time.Second))
		fmt.Printf("Last backoff: %s\n", lastStep.Round(time.Second))
	}

	// Output:
	// Generating backoff steps from 5s to 1m0s
	// Total steps: 13
	// First backoff: 9s
	// Last backoff: 1m0s
}
