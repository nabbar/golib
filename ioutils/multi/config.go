/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package multi

// Config defines configuration for adaptive writer behavior.
type Config struct {
	// SampleWrite is the number of write operations to sample before evaluating mode switch.
	// Default: 100
	SampleWrite int

	// ThresholdLatency is the latency threshold in nanoseconds above which parallel mode is enabled.
	// If average write latency exceeds this value, the writer switches to parallel mode.
	// Default: 5000 (5µs)
	ThresholdLatency int64

	// MinimalWriter is the minimum number of writers required to consider parallel mode.
	// Parallel mode is only enabled if writer count >= MinimalWriter.
	// Default: 3
	MinimalWriter int

	// MinimalSize is the minimum data size in bytes that justifies parallel writes.
	// Parallel mode is only used if data size >= MinimalSize.
	// Default: 512 bytes
	MinimalSize int
}

// DefaultConfig returns default adaptive configuration based on benchmark results.
func DefaultConfig() Config {
	return Config{
		SampleWrite:      100,
		ThresholdLatency: 5000, // 5µs
		MinimalWriter:    3,
		MinimalSize:      512,
	}
}
