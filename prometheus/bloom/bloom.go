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

// Package bloom provides a thread-safe Bloom filter implementation for Prometheus metrics.
//
// A Bloom filter is a space-efficient probabilistic data structure used to test whether
// an element is a member of a set. It can tell you with certainty that an element is NOT
// in the set, but it can only tell you that an element MIGHT be in the set (false positives
// are possible, but false negatives are not).
//
// # Use Case in Prometheus
//
// This package is primarily used to track unique values in Prometheus metrics, particularly
// for counting unique IP addresses, request IDs, or other high-cardinality dimensions
// without storing every individual value.
//
// # Features
//
//   - Thread-safe operations with read-write mutexes
//   - Multiple hash functions (6 hash functions using different seeds)
//   - Default size optimized for typical metrics use cases
//   - Collection support for managing multiple Bloom filters per metric
//
// # Trade-offs
//
// Advantages:
//   - Constant space usage regardless of unique value count
//   - Fast lookups and insertions (O(k) where k is number of hash functions)
//   - Memory efficient compared to storing all values
//
// Disadvantages:
//   - False positive rate increases as more items are added
//   - Cannot remove items once added
//   - Cannot retrieve the original values
//
// # Basic Usage
//
//	// Single Bloom filter
//	bf := bloom.NewBloomFilter()
//	bf.Add("192.168.1.1")
//	bf.Add("10.0.0.1")
//
//	if bf.Contains("192.168.1.1") {
//	    // Might be present (or false positive)
//	}
//
//	// Collection of Bloom filters (one per metric)
//	collection := bloom.New()
//	collection.Add("request_ip_total", "192.168.1.1")
//	collection.Add("request_ip_total", "10.0.0.1")
//
//	if collection.Contains("request_ip_total", "192.168.1.1") {
//	    // Might be present
//	}
//
// # Integration with Prometheus Metrics
//
// This package is used by webmetrics to track unique request IPs efficiently:
//
//	// In webmetrics/requestIPTotal.go
//	bloom := bloom.New()
//	// Track if we've seen this IP for this endpoint
//	key := endpoint + ":" + clientIP
//	if !bloom.Contains(metricName, key) {
//	    bloom.Add(metricName, key)
//	    metric.Inc([]string{endpoint, clientIP})
//	}
//
// # Performance Characteristics
//
//   - Space complexity: O(m) where m is the bit array size (fixed)
//   - Time complexity: O(k) where k is the number of hash functions (6)
//   - False positive rate: ~0.8% with default configuration
//
// For more on Bloom filters: https://en.wikipedia.org/wiki/Bloom_filter
package bloom

import (
	"sync"

	"github.com/bits-and-blooms/bitset"
)

// defaultSize is the default size of the bit array for the Bloom filter.
// Set to 2^25 (33,554,432 bits or ~4MB) which provides a good balance between
// memory usage and false positive rate for typical metrics use cases.
const defaultSize = 2 << 24

// seeds are the prime numbers used as seeds for the hash functions.
// Using 6 different seeds provides 6 independent hash functions, which
// helps reduce the false positive rate while keeping computational cost reasonable.
var seeds = []uint{7, 11, 13, 31, 37, 61}

// bloomFilter is the internal implementation of the BloomFilter interface.
// It uses a bit array to track set membership and multiple hash functions
// to minimize false positive rates.
type bloomFilter struct {
	mu    sync.RWMutex   // Protects concurrent access to the bit set
	set   *bitset.BitSet // Bit array storing the filter data
	funcs []simpleHash   // Hash functions for distributing values
}

// BloomFilter is a probabilistic data structure used to test whether an element is a member of a set.
//
// A Bloom filter can definitively tell you if an element is NOT in the set (no false negatives),
// but can only tell you if an element MIGHT be in the set (false positives are possible).
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently from multiple goroutines.
//
// # Example
//
//	bf := bloom.NewBloomFilter()
//	bf.Add("user123")
//	bf.Add("user456")
//
//	if bf.Contains("user123") {
//	    // Definitely added (or false positive)
//	}
//	if !bf.Contains("user999") {
//	    // Definitely not added (guaranteed)
//	}
type BloomFilter interface {
	// Add inserts a value into the Bloom filter.
	//
	// The value is hashed using multiple hash functions, and the corresponding
	// bits in the bit array are set to 1.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Add(value string)

	// Contains checks if a value might be in the Bloom filter.
	//
	// Returns:
	//   - true: the value MIGHT be present (or could be a false positive)
	//   - false: the value is DEFINITELY NOT present
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Contains(value string) bool
}

// NewBloomFilter creates a new Bloom filter instance with default configuration.
//
// The filter is initialized with:
//   - Default size: 2^25 bits (~4MB)
//   - 6 hash functions using different prime number seeds
//
// Returns a new BloomFilter ready to use.
//
// Example:
//
//	bf := NewBloomFilter()
//	bf.Add("192.168.1.1")
//	if bf.Contains("192.168.1.1") {
//	    // Value might be present
//	}
func NewBloomFilter() BloomFilter {
	bf := &bloomFilter{
		set:   bitset.New(defaultSize),
		funcs: make([]simpleHash, len(seeds)),
	}

	for i := 0; i < len(seeds); i++ {
		bf.funcs[i] = simpleHash{defaultSize, seeds[i]}
	}

	return bf
}

// Add inserts a value into the Bloom filter by setting bits at positions
// determined by multiple hash functions.
func (bf *bloomFilter) Add(value string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for _, f := range bf.funcs {
		bf.set.Set(f.hash(value))
	}
}

// Contains checks if a value might be in the Bloom filter.
// Returns true if the value might be present (with possible false positives),
// false if the value is definitely not present.
func (bf *bloomFilter) Contains(value string) bool {
	if len(value) < 1 {
		return false
	}

	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for _, f := range bf.funcs {
		if !bf.set.Test(f.hash(value)) {
			return false
		}
	}

	return true
}

// simpleHash is a simple hash function implementation for the Bloom filter.
// It uses a seed-based approach to generate different hash values for the same input.
type simpleHash struct {
	Cap  uint // Capacity of the bit array (used to mod the hash result)
	Seed uint // Seed value to generate different hash functions
}

// hash computes a hash value for the given string.
//
// The hash function uses the seed to generate different distributions for the same input,
// allowing the Bloom filter to use multiple independent hash functions.
//
// The result is masked with (Cap - 1) to ensure it fits within the bit array bounds.
func (s *simpleHash) hash(value string) uint {
	var result uint = 0
	for i := 0; i < len(value); i++ {
		result = result*s.Seed + uint(value[i])
	}
	return (s.Cap - 1) & result
}
