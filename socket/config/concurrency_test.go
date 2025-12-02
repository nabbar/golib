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

// Concurrency Tests - Socket Configuration Package
//
// This file contains thread-safety tests to verify that the socket/config package
// can be safely used from multiple goroutines concurrently.
//
// Test Coverage:
//   - Concurrent validation: Multiple goroutines validating the same configuration
//   - Concurrent creation: Multiple goroutines creating separate configurations
//   - Concurrent read access: Multiple goroutines reading configuration fields
//   - Mixed client/server operations: Concurrent operations on different config types
//   - Multiple protocol validation: Concurrent validation of different protocols
//   - High concurrency stress: 1000+ goroutines performing concurrent operations
//   - Rapid create-validate cycles: Repeated creation and validation under load
//
// Thread Safety Model:
// The configuration structures are safe to read from multiple goroutines after
// creation. Validation methods are read-only operations and can be called
// concurrently. Modifications should not be performed concurrently without
// external synchronization.
//
// These tests verify that there are no data races when using the package
// correctly and should be run with the -race detector enabled.
package config_test

import (
	"sync"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/nabbar/golib/socket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrent Client Operations", func() {
	Context("Concurrent validation", func() {
		It("should safely validate from multiple goroutines", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 100
			errCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := c.Validate(); err != nil {
						errCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})

		It("should handle concurrent validation of different clients", func() {
			clients := make([]config.Client, 100)
			for i := range clients {
				clients[i] = config.Client{
					Network: libptc.NetworkTCP,
					Address: "localhost:8080",
				}
			}

			var wg sync.WaitGroup
			errCount := atomic.Int32{}

			for i := range clients {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					if err := clients[idx].Validate(); err != nil {
						errCount.Add(1)
					}
				}(i)
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})
	})

	Context("Concurrent creation", func() {
		It("should safely create multiple clients concurrently", func() {
			var wg sync.WaitGroup
			numGoroutines := 100
			successCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c := config.Client{
						Network: libptc.NetworkTCP,
						Address: "localhost:8080",
					}
					if err := c.Validate(); err == nil {
						successCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(numGoroutines)))
		})
	})

	Context("Concurrent read access", func() {
		It("should safely read client fields from multiple goroutines", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 100
			matchCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if c.Network == libptc.NetworkTCP && c.Address == "localhost:8080" {
						matchCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(matchCount.Load()).To(Equal(int32(numGoroutines)))
		})
	})
})

var _ = Describe("Concurrent Server Operations", func() {
	Context("Concurrent validation", func() {
		It("should safely validate from multiple goroutines", func() {
			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 100
			errCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := s.Validate(); err != nil {
						errCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})

		It("should handle concurrent validation of different servers", func() {
			servers := make([]config.Server, 100)
			for i := range servers {
				servers[i] = config.Server{
					Network: libptc.NetworkTCP,
					Address: ":8080",
				}
			}

			var wg sync.WaitGroup
			errCount := atomic.Int32{}

			for i := range servers {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					if err := servers[idx].Validate(); err != nil {
						errCount.Add(1)
					}
				}(i)
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})
	})

	Context("Concurrent creation", func() {
		It("should safely create multiple servers concurrently", func() {
			var wg sync.WaitGroup
			numGoroutines := 100
			successCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					s := config.Server{
						Network: libptc.NetworkTCP,
						Address: ":8080",
					}
					if err := s.Validate(); err == nil {
						successCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(numGoroutines)))
		})
	})

	Context("Concurrent read access", func() {
		It("should safely read server fields from multiple goroutines", func() {
			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 100
			matchCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if s.Network == libptc.NetworkTCP && s.Address == ":8080" {
						matchCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(matchCount.Load()).To(Equal(int32(numGoroutines)))
		})
	})
})

var _ = Describe("Concurrent Mixed Operations", func() {
	Context("Client and server validation", func() {
		It("should handle concurrent validation of clients and servers", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 50
			errCount := atomic.Int32{}

			// Validate clients
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := c.Validate(); err != nil {
						errCount.Add(1)
					}
				}()
			}

			// Validate servers
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := s.Validate(); err != nil {
						errCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})
	})

	Context("Multiple protocol validation", func() {
		It("should handle concurrent validation of different protocols", func() {
			protocols := []libptc.NetworkProtocol{
				libptc.NetworkTCP,
				libptc.NetworkTCP4,
				libptc.NetworkTCP6,
				libptc.NetworkUDP,
				libptc.NetworkUDP4,
				libptc.NetworkUDP6,
			}

			var wg sync.WaitGroup
			errCount := atomic.Int32{}

			for _, proto := range protocols {
				wg.Add(1)
				go func(p libptc.NetworkProtocol) {
					defer wg.Done()

					var addr string
					switch p {
					case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
						addr = "localhost:8080"
					case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
						addr = "localhost:9000"
					}

					c := config.Client{
						Network: p,
						Address: addr,
					}

					if err := c.Validate(); err != nil {
						errCount.Add(1)
					}
				}(proto)
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})
	})
})

var _ = Describe("Concurrent Stress Tests", func() {
	Context("High concurrency validation", func() {
		It("should handle 1000 concurrent validations", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			var wg sync.WaitGroup
			numGoroutines := 1000
			errCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := c.Validate(); err != nil {
						errCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(errCount.Load()).To(BeZero())
		})
	})

	Context("Rapid creation and validation", func() {
		It("should handle rapid concurrent create-validate cycles", func() {
			var wg sync.WaitGroup
			numGoroutines := 500
			successCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					// Create and validate multiple times
					for j := 0; j < 10; j++ {
						c := config.Client{
							Network: libptc.NetworkTCP,
							Address: "localhost:8080",
						}
						if err := c.Validate(); err == nil {
							successCount.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(numGoroutines * 10)))
		})
	})
})
