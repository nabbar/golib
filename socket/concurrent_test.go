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

package socket_test

import (
	"fmt"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("[TC-CC] Socket Concurrent Operations", func() {
	Describe("ErrorFilter concurrency", func() {
		Context("concurrent calls to ErrorFilter", func() {
			It("[TC-CC-001] should handle concurrent filtering safely", func() {
				var wg sync.WaitGroup
				iterations := 100

				errors := []error{
					nil,
					fmt.Errorf("connection timeout"),
					fmt.Errorf("use of closed network connection"),
					fmt.Errorf("connection refused"),
					fmt.Errorf("broken pipe"),
				}

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						err := errors[idx%len(errors)]
						_ = libsck.ErrorFilter(err)
					}(i)
				}

				wg.Wait()
			})

			It("[TC-CC-002] should maintain correctness under concurrent load", func() {
				var wg sync.WaitGroup
				var nilCount, filteredCount, passedCount atomic.Int32
				iterations := 200

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						var err error
						switch idx % 3 {
						case 0:
							err = nil
						case 1:
							err = fmt.Errorf("use of closed network connection")
						case 2:
							err = fmt.Errorf("connection timeout")
						}

						result := libsck.ErrorFilter(err)
						if err == nil {
							if result == nil {
								nilCount.Add(1)
							}
						} else if err.Error() == "use of closed network connection" {
							if result == nil {
								filteredCount.Add(1)
							}
						} else {
							if result != nil {
								passedCount.Add(1)
							}
						}
					}(i)
				}

				wg.Wait()

				expectedPerType := int32(iterations / 3)
				Expect(nilCount.Load()).To(BeNumerically(">=", expectedPerType-2))
				Expect(filteredCount.Load()).To(BeNumerically(">=", expectedPerType-2))
				Expect(passedCount.Load()).To(BeNumerically(">=", expectedPerType-2))
			})
		})
	})

	Describe("ConnState.String concurrency", func() {
		Context("concurrent String conversions", func() {
			It("[TC-CC-003] should handle concurrent String calls safely", func() {
				var wg sync.WaitGroup
				iterations := 100

				states := []libsck.ConnState{
					libsck.ConnectionDial,
					libsck.ConnectionNew,
					libsck.ConnectionRead,
					libsck.ConnectionCloseRead,
					libsck.ConnectionHandler,
					libsck.ConnectionWrite,
					libsck.ConnectionCloseWrite,
					libsck.ConnectionClose,
				}

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						state := states[idx%len(states)]
						_ = state.String()
					}(i)
				}

				wg.Wait()
			})

			It("[TC-CC-004] should maintain string consistency under concurrent load", func() {
				var wg sync.WaitGroup
				var successCount atomic.Int32
				iterations := 200
				state := libsck.ConnectionNew
				expectedStr := "New Connection"

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						result := state.String()
						if result == expectedStr {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(iterations)))
			})
		})

		Context("mixed state conversions", func() {
			It("[TC-CC-005] should handle random state conversions concurrently", func() {
				var wg sync.WaitGroup
				iterations := 150

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						state := libsck.ConnState(idx % 10)
						str := state.String()
						Expect(str).NotTo(BeEmpty())
					}(i)
				}

				wg.Wait()
			})
		})
	})

	Describe("Mixed concurrent operations", func() {
		Context("ErrorFilter and ConnState.String together", func() {
			It("[TC-CC-006] should handle mixed operations concurrently", func() {
				var wg sync.WaitGroup
				iterations := 100

				for i := 0; i < iterations; i++ {
					wg.Add(2)

					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						err := fmt.Errorf("error %d", idx)
						_ = libsck.ErrorFilter(err)
					}(i)

					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						state := libsck.ConnState(idx % 8)
						_ = state.String()
					}(i)
				}

				wg.Wait()
			})

			It("[TC-CC-007] should maintain independent correctness", func() {
				var wg sync.WaitGroup
				var errorFilterOk, stateStringOk atomic.Int32
				iterations := 150

				for i := 0; i < iterations; i++ {
					wg.Add(2)

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						err := fmt.Errorf("use of closed network connection")
						if libsck.ErrorFilter(err) == nil {
							errorFilterOk.Add(1)
						}
					}()

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						state := libsck.ConnectionHandler
						if state.String() == "Run HandlerFunc" {
							stateStringOk.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(errorFilterOk.Load()).To(Equal(int32(iterations)))
				Expect(stateStringOk.Load()).To(Equal(int32(iterations)))
			})
		})
	})

	Describe("High-concurrency stress tests", func() {
		Context("heavy concurrent load", func() {
			It("[TC-CC-008] should handle high concurrent ErrorFilter load", func() {
				var wg sync.WaitGroup
				iterations := 1000

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						err := fmt.Errorf("error %d: use of closed network connection", idx)
						result := libsck.ErrorFilter(err)
						_ = result
					}(i)
				}

				wg.Wait()
			})

			It("[TC-CC-009] should handle high concurrent String conversion load", func() {
				var wg sync.WaitGroup
				iterations := 1000

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()
						state := libsck.ConnState(idx % 8)
						str := state.String()
						Expect(str).NotTo(BeEmpty())
					}(i)
				}

				wg.Wait()
			})
		})
	})
})
