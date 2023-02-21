/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package bloom

import "sync"

type Collection interface {
	Add(metricName, value string)
	Contains(metricName, value string) bool
}

type colBf struct {
	m  sync.RWMutex
	bf map[string]BloomFilter
}

func New() Collection {
	return &colBf{
		m:  sync.RWMutex{},
		bf: make(map[string]BloomFilter, 0),
	}
}

func (c *colBf) Add(metricName, value string) {
	c.m.Lock()
	defer c.m.Unlock()

	if len(c.bf) < 1 {
		c.bf = make(map[string]BloomFilter, 0)
	}

	var (
		b  BloomFilter
		ok bool
	)

	if b, ok = c.bf[metricName]; !ok {
		b = NewBloomFilter()
	}

	b.Add(value)
	c.bf[metricName] = b
}

func (c *colBf) Contains(metricName, value string) bool {
	c.m.RLock()
	defer c.m.RUnlock()

	if len(c.bf) < 1 {
		return false
	}

	if b, ok := c.bf[metricName]; ok {
		return b.Contains(value)
	}

	return false
}
