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

package prometheus

import (
	"sync"

	"github.com/bits-and-blooms/bitset"
)

const defaultSize = 2 << 24

var seeds = []uint{7, 11, 13, 31, 37, 61}

type bloomFilter struct {
	m     sync.Mutex
	Set   *bitset.BitSet
	Funcs [6]simpleHash
}

type BloomFilter interface {
	Add(value string)
	Contains(value string) bool
}

func NewBloomFilter() BloomFilter {
	bf := new(bloomFilter)

	for i := 0; i < len(bf.Funcs); i++ {
		bf.Funcs[i] = simpleHash{defaultSize, seeds[i]}
	}

	bf.Set = bitset.New(defaultSize)

	return bf
}

func (bf *bloomFilter) Add(value string) {
	bf.m.Lock()
	defer bf.m.Unlock()

	for _, f := range bf.Funcs {
		bf.Set.Set(f.hash(value))
	}
}

func (bf *bloomFilter) Contains(value string) bool {
	if value == "" {
		return false
	}

	ret := true

	bf.m.Lock()
	defer bf.m.Unlock()

	for _, f := range bf.Funcs {
		ret = ret && bf.Set.Test(f.hash(value))
	}

	return ret
}

type simpleHash struct {
	Cap  uint
	Seed uint
}

func (s *simpleHash) hash(value string) uint {
	var result uint = 0
	for i := 0; i < len(value); i++ {
		result = result*s.Seed + uint(value[i])
	}
	return (s.Cap - 1) & result
}
