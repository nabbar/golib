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

// PID (Proportional Integral Derivative controller)
type pid struct {
	kp float64 // rate proportional
	ki float64 // rate integral
	kd float64 // rate derivative

	prevError float64
	integral  float64
}

func (p *pid) calc(end, actual float64) float64 {
	pidError := end - actual
	p.integral += pidError
	derive := pidError - p.prevError

	output := p.kp*pidError + p.ki*p.integral + p.kd*derive
	p.prevError = pidError

	return output
}

func (p *pid) Range(min, max float64) []float64 {
	var res = make([]float64, 0)

	for {
		min += p.calc(max, min)

		if min > max {
			return append(res, max)
		} else {
			res = append(res, min)
		}
	}
}
