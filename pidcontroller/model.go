/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
 */

package pidcontroller

import (
	"context"
	"math"
	"time"
)

// pid is the concrete implementation of the PID controller.
// It uses proportional (Kp), integral (Ki), and derivative (Kd) terms to calculate
// the control variable output.
type pid struct {
	kp float64 // Proportional gain coefficient
	ki float64 // Integral gain coefficient
	kd float64 // Derivative gain coefficient

	prevError float64 // Stores the error from the previous iteration
	integral  float64 // Stores the accumulated error over time
}

// calc computes the output of the PID controller based on the error between the desired
// setpoint (end) and the actual process variable (actual).
//
// The calculation involves:
//   - Proportional Term (P): kp * error
//   - Integral Term (I): ki * accumulated_error
//   - Derivative Term (D): kd * (error - previous_error)
//
// The output is the sum of these three terms.
//
// Parameters:
//   - end: The target value (SetPoint).
//   - actual: The current value (Process Variable).
//
// Returns:
//
//	The control variable output, representing the adjustment needed.
func (p *pid) calc(end, actual float64) float64 {
	// Calculate the error: difference between target and current value
	pidError := end - actual

	// Accumulate the error for the Integral term
	p.integral += pidError

	// Calculate the change in error for the Derivative term
	derive := pidError - p.prevError

	// Compute the PID output
	output := p.kp*pidError + p.ki*p.integral + p.kd*derive

	// Update previous error for the next iteration
	p.prevError = pidError

	return output
}

// RangeCtx generates a sequence of values transitioning from a minimum to a maximum value,
// utilizing the PID controller to determine the step size at each iteration.
//
// This method iteratively applies the PID calculation to reach the target 'max' from 'min'.
// It respects the provided context for cancellation.
//
// Parameters:
//   - ctx: The context to manage the lifecycle of the operation.
//   - min: The starting value.
//   - max: The target value.
//
// Returns:
//
//	A slice of float64 containing the sequence of values generated.
func (p *pid) RangeCtx(ctx context.Context, min, max float64) []float64 {
	// Pre-allocate slice with a small buffer
	var res = make([]float64, 0, 100)

	// If min > max, the current logic assumes we are increasing towards max.
	// We need to return immediately if the condition is already met or violated.
	if min >= max {
		return []float64{max}
	}

	// Counter for batching context checks.
	// Checking the context error channel at every iteration in a tight loop can be expensive.
	// We use a uint8 counter to check only every 256 iterations.
	var check uint8

	for {
		// Check for context cancellation or timeout
		// We use a counter to amortize the cost of the interface check
		check++
		if check%100 == 0 {
			if ctx.Err() != nil {
				// Append the target value to ensure closure if cancelled
				return append(res, max)
			}
			check = 1
		}

		// Calculate the next step using the PID controller
		step := p.calc(max, min)

		// Check for convergence or tiny steps to prevent infinite loops near target
		if math.Abs(step) < 1e-9 && math.Abs(max-min) < 1e-9 {
			return append(res, max)
		}

		min += step

		// Check if we have reached or exceeded the target
		if min >= max {
			// If the calculated value reaches or overshoots the target, cap it at max and return
			return append(res, max)
		}

		// Append the current value to the result
		res = append(res, min)
	}
}

// Range acts as a wrapper around RangeCtx with a default timeout of 5 seconds.
// It generates a sequence of values from min to max.
//
// Parameters:
//   - min: The starting value.
//   - max: The target value.
//
// Returns:
//
//	A slice of float64 containing the sequence of values.
func (p *pid) Range(min, max float64) []float64 {
	// Create a context with a 5-second timeout
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return p.RangeCtx(ctx, min, max)
}
