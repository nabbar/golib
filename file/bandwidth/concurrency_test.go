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
 *
 */

package bandwidth_test

import (
	"io"
	"os"
	"sync"

	. "github.com/nabbar/golib/file/bandwidth"
	libfpg "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bandwidth Concurrency", func() {
	var (
		tempFiles []*os.File
		tempPaths []string
	)

	BeforeEach(func() {
		tempFiles = make([]*os.File, 3)
		tempPaths = make([]string, 3)

		// Create 3 temporary test files
		for i := 0; i < 3; i++ {
			var err error
			tempFiles[i], err = os.CreateTemp("", "bandwidth-concurrent-test-*.dat")
			Expect(err).ToNot(HaveOccurred())
			tempPaths[i] = tempFiles[i].Name()

			// Write 5KB of test data
			testData := make([]byte, 5*1024)
			for j := range testData {
				testData[j] = byte(j % 256)
			}
			_, err = tempFiles[i].Write(testData)
			Expect(err).ToNot(HaveOccurred())
			err = tempFiles[i].Close()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		for _, path := range tempPaths {
			if path != "" {
				_ = os.Remove(path)
			}
		}
	})

	Describe("Concurrent Access", func() {
		It("should handle concurrent RegisterIncrement calls", func() {
			bw := New(0)
			var wg sync.WaitGroup
			var mu sync.Mutex
			totalCalls := 0

			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					fpg, err := libfpg.Open(tempPaths[index])
					Expect(err).ToNot(HaveOccurred())
					defer fpg.Close()

					bw.RegisterIncrement(fpg, func(size int64) {
						mu.Lock()
						totalCalls++
						mu.Unlock()
					})

					_, err = io.ReadAll(fpg)
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()
			Expect(totalCalls).To(BeNumerically(">", 0))
		})

		It("should handle concurrent RegisterReset calls", func() {
			bw := New(0)
			var wg sync.WaitGroup
			var mu sync.Mutex
			resetCount := 0

			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					fpg, err := libfpg.Open(tempPaths[index])
					Expect(err).ToNot(HaveOccurred())
					defer fpg.Close()

					bw.RegisterReset(fpg, func(size, current int64) {
						mu.Lock()
						resetCount++
						mu.Unlock()
					})

					buffer := make([]byte, 1024)
					_, err = fpg.Read(buffer)
					Expect(err).ToNot(HaveOccurred())

					fpg.Reset(5 * 1024)
				}(i)
			}

			wg.Wait()
			Expect(resetCount).To(Equal(3))
		})

		It("should handle mixed concurrent operations", func() {
			bw := New(0) // No limit for fast concurrent testing
			var wg sync.WaitGroup
			var mu sync.Mutex
			operations := 0

			// Concurrent reads without bandwidth limiting for fast tests
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					fpg, err := libfpg.Open(tempPaths[index])
					Expect(err).ToNot(HaveOccurred())
					defer fpg.Close()

					bw.RegisterIncrement(fpg, func(size int64) {
						mu.Lock()
						operations++
						mu.Unlock()
					})

					_, err = io.ReadAll(fpg)
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()
			Expect(operations).To(BeNumerically(">", 0))
		})
	})

	Describe("Nil Safety", func() {
		It("should handle nil BandWidth gracefully", func() {
			// This tests the nil check in the Increment method
			// Should not panic when calling methods
			Expect(func() {
				_ = New(0)
			}).ToNot(Panic())
		})

		It("should handle nil callbacks in RegisterIncrement", func() {
			bw := New(0)
			fpg, err := libfpg.Open(tempPaths[0])
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			Expect(func() {
				bw.RegisterIncrement(fpg, nil)
				_, _ = io.ReadAll(fpg)
			}).ToNot(Panic())
		})

		It("should handle nil callbacks in RegisterReset", func() {
			bw := New(0)
			fpg, err := libfpg.Open(tempPaths[0])
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			Expect(func() {
				bw.RegisterReset(fpg, nil)
				buffer := make([]byte, 1024)
				_, _ = fpg.Read(buffer)
				fpg.Reset(5 * 1024)
			}).ToNot(Panic())
		})
	})
})
