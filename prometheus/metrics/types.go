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

package metrics

func (m *metrics) SetDesc(desc string) {
	m.m.Lock()
	defer m.m.Unlock()

	m.d = desc
}

func (m *metrics) GetDesc() string {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.d
}

func (m *metrics) AddLabel(label ...string) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.l) < 1 {
		m.l = make([]string, 0)
	}

	m.l = append(m.l, label...)
}

func (m *metrics) GetLabel() []string {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.l
}

func (m *metrics) AddBuckets(bucket ...float64) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.b) < 1 {
		m.b = make([]float64, 0)
	}

	m.b = append(m.b, bucket...)
}

func (m *metrics) GetBuckets() []float64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.b
}

func (m *metrics) AddObjective(key, value float64) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.o) < 1 {
		m.o = make(map[float64]float64, 0)
	}

	m.o[key] = value
}

func (m *metrics) GetObjectives() map[float64]float64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.o
}
