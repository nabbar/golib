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
 */

package duration

import (
	"context"
	"time"

	libpid "github.com/nabbar/golib/pidcontroller"
)

// DefaultRateProportional is the default proportional rate for the PID controller used in range generation.
var DefaultRateProportional float64 = 0.1

// DefaultRateIntegral is the default integral rate for the PID controller used in range generation.
var DefaultRateIntegral float64 = 0.01

// DefaultRateDerivative is the default derivative rate for the PID controller used in range generation.
var DefaultRateDerivative float64 = 0.05

// RangeCtxTo generates a slice of durations from the receiver 'd' to the 'dur' parameter.
// The spacing between durations is determined by a PID controller, allowing for non-linear intervals.
// This is useful for scenarios like exponential backoff or other adaptive timing strategies.
//
// The context 'ctx' can be used to cancel the generation process.
// The 'rateP', 'rateI', and 'rateD' parameters configure the Proportional, Integral, and Derivative
// components of the PID controller, respectively.
//
// The resulting slice is guaranteed to start with 'd' and end with 'dur'.
func (d Duration) RangeCtxTo(ctx context.Context, dur Duration, rateP, rateI, rateD float64) []Duration {
	return rangeCtx(ctx, d, dur, rateP, rateI, rateD)
}

// RangeTo is a convenience wrapper for RangeCtxTo that uses a background context with a 5-second timeout.
// It generates a slice of durations from the receiver 'd' to 'dur' using the specified PID controller rates.
func (d Duration) RangeTo(dur Duration, rateP, rateI, rateD float64) []Duration {
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return d.RangeCtxTo(ctx, dur, rateP, rateI, rateD)
}

// RangeDefTo is a convenience wrapper for RangeTo that uses the default PID controller rates
// (DefaultRateProportional, DefaultRateIntegral, DefaultRateDerivative).
// It generates a slice of durations from the receiver 'd' to 'dur'.
func (d Duration) RangeDefTo(dur Duration) []Duration {
	return d.RangeTo(dur, DefaultRateProportional, DefaultRateIntegral, DefaultRateDerivative)
}

// RangeCtxFrom generates a slice of durations from 'dur' up to the receiver 'd'.
// It is the reverse of RangeCtxTo, using the same PID-controlled spacing logic.
//
// The context 'ctx' can be used to cancel the generation process.
// The 'rateP', 'rateI', and 'rateD' parameters configure the PID controller.
//
// The resulting slice is guaranteed to start with 'dur' and end with 'd'.
func (d Duration) RangeCtxFrom(ctx context.Context, dur Duration, rateP, rateI, rateD float64) []Duration {
	return rangeCtx(ctx, dur, d, rateP, rateI, rateD)
}

// RangeFrom is a convenience wrapper for RangeCtxFrom that uses a background context with a 5-second timeout.
// It generates a slice of durations from 'dur' to the receiver 'd' using the specified PID controller rates.
func (d Duration) RangeFrom(dur Duration, rateP, rateI, rateD float64) []Duration {
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return d.RangeCtxFrom(ctx, dur, rateP, rateI, rateD)
}

// RangeDefFrom is a convenience wrapper for RangeFrom that uses the default PID controller rates.
// It generates a slice of durations from 'dur' to the receiver 'd'.
func (d Duration) RangeDefFrom(dur Duration) []Duration {
	return d.RangeFrom(dur, DefaultRateProportional, DefaultRateIntegral, DefaultRateDerivative)
}

// rangeCtx is the internal implementation for generating a PID-controlled range of durations.
// It automatically selects the most appropriate time unit (from nanoseconds to days) based on the
// magnitude of the 'from' and 'to' durations to ensure reasonable precision.
func rangeCtx(ctx context.Context, from, to Duration, rateP, rateI, rateD float64) []Duration {
	var (
		p = libpid.New(rateP, rateI, rateD)
		r = make([]Duration, 0)
		u func(int64) Duration
		f float64
		t float64
	)

	// Select the largest common unit for precision.
	switch {
	case from.IsDays() && to.IsDays():
		u = Days
		f = libpid.Int64ToFloat64(from.Days())
		t = libpid.Int64ToFloat64(to.Days())
	case from.IsHours() && to.IsHours():
		u = Hours
		f = libpid.Int64ToFloat64(from.Hours())
		t = libpid.Int64ToFloat64(to.Hours())
	case from.IsMinutes() && to.IsMinutes():
		u = Minutes
		f = libpid.Int64ToFloat64(from.Minutes())
		t = libpid.Int64ToFloat64(to.Minutes())
	case from.IsSeconds() && to.IsSeconds():
		u = Seconds
		f = libpid.Int64ToFloat64(from.Seconds())
		t = libpid.Int64ToFloat64(to.Seconds())
	case from.IsMilliseconds() && to.IsMilliseconds():
		u = Milliseconds
		f = libpid.Int64ToFloat64(from.Milliseconds())
		t = libpid.Int64ToFloat64(to.Milliseconds())
	case from.IsMicroseconds() && to.IsMicroseconds():
		u = Microseconds
		f = libpid.Int64ToFloat64(from.Microseconds())
		t = libpid.Int64ToFloat64(to.Microseconds())
	default: // Default to nanoseconds for highest precision if units are mixed or small.
		u = Nanoseconds
		f = libpid.Int64ToFloat64(from.Nanoseconds())
		t = libpid.Int64ToFloat64(to.Nanoseconds())
	}

	// Generate the range using the PID controller.
	for _, v := range p.RangeCtx(ctx, f, t) {
		r = append(r, u(libpid.Float64ToInt64(v)))
	}

	// Ensure the generated range is strictly within the [from, to] bounds.
	for len(r) > 0 && r[0] < from {
		r = r[1:]
	}

	if len(r) > 0 && r[0] > from {
		r = append([]Duration{from}, r...)
	}

	for len(r) > 0 && r[len(r)-1] > to {
		r = r[:len(r)-1]
	}

	if len(r) > 0 && r[len(r)-1] < to {
		r = append(r, to)
	}

	// Ensure at least the start and end points are in the slice.
	if len(r) < 2 {
		return []Duration{from, to}
	}

	return r
}
