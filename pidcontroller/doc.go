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

// Package pidcontroller implements a Proportional-Integral-Derivative (PID) controller.
//
// A PID controller is a control loop mechanism employing feedback that is widely used in industrial control systems and a variety of other applications requiring continuously modulated control.
// A PID controller calculates an error value as the difference between a desired setpoint (SP) and a measured process variable (PV) and applies a correction based on proportional, integral, and derivative terms (denoted P, I, and D respectively).
//
// # Concepts
//
// The PID controller algorithm involves three separate constant parameters, and is accordingly sometimes called three-term control: the proportional, the integral and derivative values, denoted P, I, and D.
// Simply put, these values can be interpreted in terms of time: P depends on the present error, I on the accumulation of past errors, and D is a prediction of future errors, based on current rate of change.
// The weighted sum of these three actions is used to adjust the process via a control element such as the position of a control valve, a damper, or the power supplied to a heating element.
//
// # Data Flow
//
// The following diagram illustrates the data flow within the PID controller:
//
//	 Setpoint (SP)
//	      |
//	      v
//	+-----------+     Error (e)     +-----------+
//	|           | ----------------> |           |
//	| Comparator|                   |    PID    |
//	|           | <---------------- | Algorithm |
//	+-----------+                   |           |
//	      ^                         +-----------+
//	      |                               | Control Variable (u)
//	      |                               v
//	Process Variable (PV)          +-----------+
//	                               |  Process  |
//	                               +-----------+
//
// # Mathematical Model
//
// The overall control function u(t) can be expressed mathematically as:
//
//	u(t) = Kp * e(t) + Ki * ∫ e(t) dt + Kd * de(t)/dt
//
// Where:
//   - Kp is the proportional gain, a tuning parameter.
//   - Ki is the integral gain, a tuning parameter.
//   - Kd is the derivative gain, a tuning parameter.
//   - e(t) = SP - PV(t) is the error (SP is the setpoint, and PV(t) is the process variable).
//   - t is the time or instantaneous time (the present).
//   - τ is the variable of integration (takes on values from time 0 to the present t).
//
// # Usage
//
// This package provides a simple interface `PID` to create and use a PID controller.
// The primary use case is generating a smooth transition from a starting value to a target value,
// controlled by the PID parameters to simulate physical constraints or desired behaviors (e.g., easing, overshoot).
//
// # Quick Start
//
// To use the PID controller, first create a new instance using the `New` function, providing the P, I, and D coefficients.
// Then, use the `Range` or `RangeCtx` methods to generate a sequence of values.
//
// Example:
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/nabbar/golib/pidcontroller"
//	)
//
//	func main() {
//		// Initialize a PID controller with specific gains
//		// Kp = 0.5 (Proportional)
//		// Ki = 0.1 (Integral)
//		// Kd = 0.2 (Derivative)
//		pid := pidcontroller.New(0.5, 0.1, 0.2)
//
//		// Generate a range of values from 0 to 100
//		// The PID controller will determine the steps based on the gains.
//		values := pid.Range(0, 100)
//
//		// Output the generated sequence
//		for i, v := range values {
//			fmt.Printf("Step %d: %f\n", i, v)
//		}
//	}
//
// # Use Cases
//
// 1. Smooth Animation: Calculating intermediate frames for an animation where the transition needs to be natural and physics-based.
// 2. Process Simulation: Simulating a heating process where temperature rises to a setpoint.
// 3. Rate Limiting: Controlling the rate of resource consumption or request processing.
package pidcontroller
