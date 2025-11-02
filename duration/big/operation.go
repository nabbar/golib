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

package big

import (
	"context"
	"time"

	libpid "github.com/nabbar/golib/pidcontroller"
)

var (
	DefaultRateProportional float64 = 0.1
	DefaultRateIntegral     float64 = 0.01
	DefaultRateDerivative   float64 = 0.05
)

// Abs returns the absolute value of the duration.
//
// If the duration is positive or zero, it returns the duration.
// If the duration is negative and not equal to the minimum duration, it returns the negation of the duration.
// If the duration is equal to the minimum duration, it returns the maximum duration.
//
// Note that the minimum and maximum durations are defined as constants in the big package.
func (d Duration) Abs() Duration {
	switch {
	case d >= 0:
		return d
	case d == minDuration:
		return maxDuration
	default:
		return -d
	}
}

// RangeCtxTo generates a list of durations from d to dur, spaced according to the given PID controller parameters.
//
// The first element of the list is the start duration (d), and the last element is the end duration (dur).
// If the list has less than 3 elements, the start and end durations are added to the list.
// If the first element of the list is greater than the start duration, the start duration is added to the beginning of the list.
// If the last element of the list is less than the end duration, the end duration is added to the end of the list.
//
// The PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
//
// This function could take long time depends of rate given to PID Controller.
// To prevent this long time, the context is used to cancel the calculation before ending.
// If the context is canceled before the range is fully generated, the function will return an empty list.
func (d Duration) RangeCtxTo(ctx context.Context, dur Duration, rateP, rateI, rateD float64) []Duration {
	var (
		p = libpid.New(rateP, rateI, rateD)
		r = make([]Duration, 0)
	)

	for _, v := range p.RangeCtx(ctx, d.Float64(), dur.Float64()) {
		r = append(r, ParseFloat64(v))
	}

	if len(r) < 3 {
		r = append(make([]Duration, 0), d, dur)
	}

	if r[0] > d {
		r = append(append(make([]Duration, 0), d), r...)
	}

	if r[len(r)-1] < dur {
		r = append(r, dur)
	}

	return r
}

// RangeTo generates a list of durations from d to dur, spaced according to the given PID controller parameters.
//
// The first element of the list is the start duration (d), and the last element is the end duration (dur).
// If the list has less than 3 elements, the start and end durations are added to the list.
//
// If the first element of the list is greater than the start duration, the start duration is added to the beginning of the list.
//
// If the last element of the list is less than the end duration, the end duration is added to the end of the list.
//
// The PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
//
// This function could take long time depends of rate given to PID Controller.
// To prevent this long time, a deadline context is used to cancel the calculation before ending for 5s max.
// If the context is canceled before the range is fully generated, the function will return an empty list.
//
// To custom this default value of timeout, see RangeCtxTo
func (d Duration) RangeTo(dur Duration, rateP, rateI, rateD float64) []Duration {
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return d.RangeCtxTo(ctx, dur, rateP, rateI, rateD)
}

// RangeDefTo generates a list of durations from d to dur, spaced according to the default PID controller parameters.
//
// The first element of the list is the start duration (d), and the last element is the end duration (dur).
// If the list has less than 3 elements, the start and end durations are added to the list.
//
// If the first element of the list is greater than the start duration, the start duration is added to the beginning of the list.
//
// If the last element of the list is less than the end duration, the end duration is added to the end of the list.
//
// The default PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
func (d Duration) RangeDefTo(dur Duration) []Duration {
	return d.RangeTo(dur, DefaultRateProportional, DefaultRateIntegral, DefaultRateDerivative)
}

// RangeCtxFrom generates a list of durations from dur to d, spaced according to the given PID controller parameters.
//
// The first element of the list is the end duration (dur), and the last element is the start duration (d).
// If the list has less than 3 elements, the start and end durations are added to the list.
//
// If the first element of the list is greater than the end duration, the end duration is added to the beginning of the list.
//
// If the last element of the list is less than the start duration, the start duration is added to the end of the list.
//
// The PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
//
// This function could take long time depends of rate given to PID Controller.
// To prevent this long time, the context is used to cancel the calculation before ending.
// If the context is canceled before the range is fully generated, the function will return an empty list.
func (d Duration) RangeCtxFrom(ctx context.Context, dur Duration, rateP, rateI, rateD float64) []Duration {
	var (
		p = libpid.New(rateP, rateI, rateD)
		r = make([]Duration, 0)
	)

	for _, v := range p.RangeCtx(ctx, dur.Float64(), d.Float64()) {
		r = append(r, ParseFloat64(v))
	}

	if len(r) < 3 {
		r = append(make([]Duration, 0), d, dur)
	}

	if r[0] > dur {
		r = append(append(make([]Duration, 0), dur), r...)
	}

	if r[len(r)-1] < d {
		r = append(r, d)
	}

	return r
}

// RangeFrom generates a list of durations from dur to d, spaced according to the given PID controller parameters.
//
// The first element of the list is the end duration (dur), and the last element is the start duration (d).
// If the list has less than 3 elements, the start and end durations are added to the list.
//
// If the first element of the list is greater than the end duration, the end duration is added to the beginning of the list.
//
// If the last element of the list is less than the start duration, the start duration is added to the end of the list.
//
// The PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
//
// The context is used to cancel the range generation if the context is canceled before the range is fully generated.
// To prevent this long time, a deadline context is used to cancel the calculation before ending for 5s max.
// If the context is canceled before the range is fully generated, the function will return an empty list.
//
// To custom this default value of timeout, see RangeCtxFrom
func (d Duration) RangeFrom(dur Duration, rateP, rateI, rateD float64) []Duration {
	ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl()

	return d.RangeCtxFrom(ctx, dur, rateP, rateI, rateD)
}

// RangeDefFrom generates a list of durations from dur to d, spaced according to the default PID controller parameters.
//
// The first element of the list is the end duration (dur), and the last element is the start duration (d).
// If the list has less than 3 elements, the start and end durations are added to the list.
//
// If the first element of the list is greater than the end duration, the end duration is added to the beginning of the list.
//
// If the last element of the list is less than the start duration, the start duration is added to the end of the list.
//
// The default PID controller parameters are:
// - rateP: the proportional rate
// - rateI: the integral rate
// - rateD: the derivative rate
func (d Duration) RangeDefFrom(dur Duration) []Duration {
	return d.RangeFrom(dur, DefaultRateProportional, DefaultRateIntegral, DefaultRateDerivative)
}
