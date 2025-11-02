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

import (
	"sync"
)

// Collection manages multiple Bloom filters, one per metric name.
//
// This allows tracking unique values independently for different metrics without
// cross-contamination. Each metric gets its own Bloom filter, created on-demand
// when the first value is added.
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently from multiple goroutines.
//
// # Use Case
//
// Commonly used in Prometheus webmetrics to track unique values per metric:
//   - Unique IP addresses per endpoint
//   - Unique user IDs per API route
//   - Unique request IDs per service
//
// # Example
//
//	collection := bloom.New()
//
//	// Track unique IPs for different endpoints
//	collection.Add("api_requests", "192.168.1.1")
//	collection.Add("api_requests", "10.0.0.1")
//	collection.Add("admin_requests", "192.168.1.100")
//
//	if collection.Contains("api_requests", "192.168.1.1") {
//	    // IP might have been seen for api_requests
//	}
type Collection interface {
	// Add inserts a value into the Bloom filter for the specified metric.
	//
	// If the metric doesn't exist yet, a new Bloom filter is automatically created.
	// This operation is thread-safe and idempotent (adding the same value multiple
	// times has the same effect as adding it once).
	//
	// Parameters:
	//   - metricName: the name of the metric to associate the value with
	//   - value: the value to add (e.g., IP address, user ID, request ID)
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Add(metricName, value string)

	// Contains checks if a value might exist in the Bloom filter for the specified metric.
	//
	// Returns:
	//   - false: if the metric doesn't exist OR the value is definitely not present
	//   - true: if the value MIGHT be present (or could be a false positive)
	//
	// Parameters:
	//   - metricName: the name of the metric to check
	//   - value: the value to check for
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Contains(metricName, value string) bool
}

// colBf is the internal implementation of the Collection interface.
// It uses a map to store separate Bloom filters for each metric name.
type colBf struct {
	mu sync.RWMutex           // Protects concurrent access to the map
	b  map[string]BloomFilter // Map of metric name to Bloom filter
}

// New creates a new Collection of Bloom filters.
//
// The collection starts empty and creates Bloom filters on-demand as metrics
// are added. Each metric gets its own independent Bloom filter.
//
// Returns a new Collection ready to use.
//
// Example:
//
//	collection := bloom.New()
//	collection.Add("http_requests", "192.168.1.1")
//	collection.Add("grpc_requests", "10.0.0.1")
func New() Collection {
	return &colBf{
		b: make(map[string]BloomFilter),
	}
}

// Add inserts a value into the Bloom filter for the specified metric name.
// If the metric doesn't exist, a new Bloom filter is created for it.
func (c *colBf) Add(metricName, value string) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	bf, exists := c.b[metricName]
	if !exists {
		bf = NewBloomFilter()
		c.b[metricName] = bf
	}

	bf.Add(value)
}

// Contains checks if a value might exist in the Bloom filter for the specified metric.
// Returns false if the metric doesn't exist or the value is definitely not present.
func (c *colBf) Contains(metricName, value string) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	bf, exists := c.b[metricName]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	return bf.Contains(value)
}
