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

package compress_test

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/compress"
)

var _ = Describe("TC-CC-001: Concurrency", func() {
	Context("TC-CC-002: Concurrent Parse calls", func() {
		It("TC-CC-003: should handle concurrent Parse safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					algs := []string{"gzip", "bzip2", "lz4", "xz", "none"}
					alg := compress.Parse(algs[idx%len(algs)])
					Expect(alg).ToNot(Equal(compress.Algorithm(99)))
				}(i)
			}

			wg.Wait()
		})
	})

	Context("TC-CC-004: Concurrent DetectOnly calls", func() {
		It("TC-CC-005: should handle concurrent DetectOnly safely", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			iterations := 50

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					alg, reader, err := compress.DetectOnly(bytes.NewReader(compressed))
					Expect(err).ToNot(HaveOccurred())
					Expect(alg).To(Equal(compress.Gzip))
					if reader != nil {
						reader.Close()
					}
				}()
			}

			wg.Wait()
		})
	})

	Context("TC-CC-006: Concurrent Detect calls", func() {
		It("TC-CC-007: should handle concurrent Detect safely", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			iterations := 50
			var successCount atomic.Int32

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					alg, reader, err := compress.Detect(bytes.NewReader(compressed))
					if err == nil && alg == compress.Gzip {
						defer reader.Close()
						data, err := io.ReadAll(reader)
						if err == nil && bytes.Equal(data, testData.dat) {
							successCount.Add(1)
						}
					}
				}()
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(iterations)))
		})
	})

	Context("TC-CC-008: Concurrent Reader creation", func() {
		It("TC-CC-009: should handle concurrent Reader creation safely", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			iterations := 50

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					reader, err := compress.Gzip.Reader(bytes.NewReader(compressed))
					Expect(err).ToNot(HaveOccurred())
					defer reader.Close()

					data, err := io.ReadAll(reader)
					Expect(err).ToNot(HaveOccurred())
					Expect(data).To(Equal(testData.dat))
				}()
			}

			wg.Wait()
		})
	})

	Context("TC-CC-010: Concurrent Writer creation", func() {
		It("TC-CC-011: should handle concurrent Writer creation safely", func() {
			testData := newTestData(100)

			var wg sync.WaitGroup
			iterations := 50
			var successCount atomic.Int32

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
					if err != nil {
						return
					}
					defer writer.Close()

					_, err = writer.Write(testData.dat)
					if err != nil {
						return
					}

					err = writer.Close()
					if err != nil {
						return
					}

					successCount.Add(1)
				}()
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(iterations)))
		})
	})

	Context("TC-CC-012: Concurrent encoding operations", func() {
		It("TC-CC-013: should handle concurrent MarshalText safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					data, err := alg.MarshalText()
					Expect(err).ToNot(HaveOccurred())
					Expect(data).ToNot(BeEmpty())
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-014: should handle concurrent UnmarshalText safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			inputs := []string{"gzip", "bzip2", "lz4", "xz", "none"}

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					var alg compress.Algorithm
					err := alg.UnmarshalText([]byte(inputs[idx%len(inputs)]))
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-015: should handle concurrent MarshalJSON safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					data, err := alg.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())
					Expect(data).ToNot(BeEmpty())
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-016: should handle concurrent UnmarshalJSON safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			inputs := [][]byte{
				[]byte(`"gzip"`),
				[]byte(`"bzip2"`),
				[]byte(`"lz4"`),
				[]byte(`"xz"`),
				[]byte(`null`),
			}

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					var alg compress.Algorithm
					err := alg.UnmarshalJSON(inputs[idx%len(inputs)])
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()
		})
	})

	Context("TC-CC-017: Concurrent algorithm methods", func() {
		It("TC-CC-018: should handle concurrent String calls safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					str := alg.String()
					Expect(str).ToNot(BeEmpty())
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-019: should handle concurrent Extension calls safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					_ = alg.Extension()
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-020: should handle concurrent IsNone calls safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					_ = alg.IsNone()
				}(i)
			}

			wg.Wait()
		})

		It("TC-CC-021: should handle concurrent DetectHeader calls safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			headers := [][]byte{
				{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00}, // Gzip
				{'B', 'Z', 'h', '9', 0x00, 0x00},     // Bzip2
				{0x04, 0x22, 0x4D, 0x18, 0x00, 0x00}, // LZ4
				{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}, // XZ
			}

			algorithms := compress.List()

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					alg := algorithms[idx%len(algorithms)]
					header := headers[idx%len(headers)]
					_ = alg.DetectHeader(header)
				}(i)
			}

			wg.Wait()
		})
	})

	Context("TC-CC-022: Mixed concurrent operations", func() {
		It("TC-CC-023: should handle mixed operations safely", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			iterations := 20

			// Parse operations
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = compress.Parse("gzip")
				}()
			}

			// Detect operations
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					alg, reader, err := compress.Detect(bytes.NewReader(compressed))
					if err == nil && alg == compress.Gzip {
						reader.Close()
					}
				}()
			}

			// String operations
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = compress.Gzip.String()
				}()
			}

			// Marshal operations
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = compress.Gzip.MarshalJSON()
				}()
			}

			wg.Wait()
		})

		It("TC-CC-024: should handle full round-trip concurrently", func() {
			var wg sync.WaitGroup
			iterations := 20

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					testData := newTestData(100 + idx*10)
					algs := []compress.Algorithm{
						compress.Gzip,
						compress.Bzip2,
						compress.LZ4,
						compress.XZ,
					}
					alg := algs[idx%len(algs)]

					// Compress
					var buf bytes.Buffer
					writer, err := alg.Writer(nopWriteCloser{&buf})
					Expect(err).ToNot(HaveOccurred())

					_, err = writer.Write(testData.dat)
					Expect(err).ToNot(HaveOccurred())

					err = writer.Close()
					Expect(err).ToNot(HaveOccurred())

					// Decompress
					reader, err := alg.Reader(&buf)
					Expect(err).ToNot(HaveOccurred())
					defer reader.Close()

					decompressed, err := io.ReadAll(reader)
					Expect(err).ToNot(HaveOccurred())
					Expect(decompressed).To(Equal(testData.dat))
				}(i)
			}

			wg.Wait()
		})
	})

	Context("TC-CC-025: Data race scenarios", func() {
		It("TC-CC-026: should not race on List function", func() {
			var wg sync.WaitGroup
			iterations := 100

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					lst := compress.List()
					Expect(lst).To(HaveLen(5))
				}()
			}

			wg.Wait()
		})

		It("TC-CC-027: should not race on ListString function", func() {
			var wg sync.WaitGroup
			iterations := 100

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					lst := compress.ListString()
					Expect(lst).To(HaveLen(5))
				}()
			}

			wg.Wait()
		})
	})
})
