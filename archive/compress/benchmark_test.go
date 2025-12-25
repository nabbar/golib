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
	"context"
	"io"
	"maps"
	"runtime"
	"slices"
	"time"

	iotclo "github.com/nabbar/golib/ioutils/mapCloser"
	iotnwc "github.com/nabbar/golib/ioutils/nopwritecloser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/archive/compress"
)

var _ = Describe("TC-BC-001: Benchmarks", func() {
	Context("TC-BC-002: Algorithm operations", func() {
		It("TC-BC-003: should benchmark Parse operations", func() {
			experiment := gmeasure.NewExperiment("Parse operations")
			AddReportEntry(experiment.Name, experiment)

			inputs := []string{"gzip", "bzip2", "lz4", "xz", "none", "unknown"}

			experiment.Sample(func(idx int) {
				for _, input := range inputs {
					experiment.MeasureDuration(input, func() {
						_ = compress.Parse(input)
					})
				}
			}, gmeasure.SamplingConfig{N: 100})
		})

		It("TC-BC-004: should benchmark String operations", func() {
			experiment := gmeasure.NewExperiment("String operations")
			AddReportEntry(experiment.Name, experiment)

			algorithms := compress.List()

			experiment.Sample(func(idx int) {
				for _, alg := range algorithms {
					experiment.MeasureDuration(alg.String(), func() {
						_ = alg.String()
					})
				}
			}, gmeasure.SamplingConfig{N: 1000})
		})

		It("TC-BC-005: should benchmark Extension operations", func() {
			experiment := gmeasure.NewExperiment("Extension operations")
			AddReportEntry(experiment.Name, experiment)

			algorithms := compress.List()

			experiment.Sample(func(idx int) {
				for _, alg := range algorithms {
					experiment.MeasureDuration(alg.String(), func() {
						_ = alg.Extension()
					})
				}
			}, gmeasure.SamplingConfig{N: 1000})
		})

		It("TC-BC-006: should benchmark DetectHeader operations", func() {
			experiment := gmeasure.NewExperiment("DetectHeader operations")
			AddReportEntry(experiment.Name, experiment)

			headers := map[string][]byte{
				"gzip":  {0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00},
				"bzip2": {'B', 'Z', 'h', '9', 0x00, 0x00},
				"lz4":   {0x04, 0x22, 0x4D, 0x18, 0x00, 0x00},
				"xz":    {0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00},
			}

			experiment.Sample(func(idx int) {
				for name, header := range headers {
					alg := compress.Parse(name)
					experiment.MeasureDuration(name, func() {
						_ = alg.DetectHeader(header)
					})
				}
			}, gmeasure.SamplingConfig{N: 1000})
		})
	})

	Context("TC-BC-007: Detection operations", func() {
		It("TC-BC-008: should benchmark DetectOnly with various formats", func() {
			experiment := gmeasure.NewExperiment("DetectOnly operations")
			AddReportEntry(experiment.Name, experiment)

			testData := newTestData(1000)
			compressedData := make(map[string][]byte)

			for _, alg := range []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ} {
				data, err := compressTestData(alg, testData.dat)
				Expect(err).ToNot(HaveOccurred())
				compressedData[alg.String()] = data
			}

			experiment.Sample(func(idx int) {
				for name, data := range compressedData {
					experiment.MeasureDuration(name, func() {
						alg, reader, err := compress.DetectOnly(bytes.NewReader(data))
						if err == nil && reader != nil {
							reader.Close()
						}
						_ = alg
					})
				}
			}, gmeasure.SamplingConfig{N: 100})
		})

		It("TC-BC-009: should benchmark Detect with decompression", func() {
			experiment := gmeasure.NewExperiment("Detect with decompression")
			AddReportEntry(experiment.Name, experiment)

			testData := newTestData(1000)
			compressedData := make(map[string][]byte)

			for _, alg := range []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ} {
				data, err := compressTestData(alg, testData.dat)
				Expect(err).ToNot(HaveOccurred())
				compressedData[alg.String()] = data
			}

			experiment.Sample(func(idx int) {
				for name, data := range compressedData {
					experiment.MeasureDuration(name, func() {
						_, reader, err := compress.Detect(bytes.NewReader(data))
						if err == nil && reader != nil {
							reader.Close()
						}
					})
				}
			}, gmeasure.SamplingConfig{N: 100})
		})
	})

	Context("TC-BC-010: Compression/Decompression operations", func() {
		defer GinkgoRecover()

		var (
			obj = make([]*bnc, 0, 4*3)
			alg = []compress.Algorithm{
				compress.Bzip2,
				compress.Gzip,
				compress.LZ4,
				compress.XZ,
			}
			txt = map[int]string{
				1024:       "Small Data (1KB)",
				1024 * 10:  "Medium Data (10KB)",
				1024 * 100: "Large Data (100KB)",
			}
		)

		for i := 0; i < len(alg); i++ {
			k := slices.Collect(maps.Keys(txt))
			for j := 0; j < len(k); j++ {
				obj = append(obj, newTestBenchDataOpe(alg[i], k[j], txt[k[j]]))
			}
		}

		It("TC-BC-011: should benchmark compression/decompression/ratio", func() {
			expCmp := gmeasure.NewExperiment("Compression")
			AddReportEntry(expCmp.Name, expCmp)

			expDmp := gmeasure.NewExperiment("Decompression")
			AddReportEntry(expDmp.Name, expDmp)

			expCmp.Sample(func(idx int) {
				var clo = iotclo.New(context.Background())
				defer func() {
					_ = clo.Close()
				}()

				for i := 0; i < len(obj); i++ {
					o := obj[i]

					b := bytes.NewBuffer(make([]byte, 0, o.nbr))
					Expect(b).ToNot(BeNil())

					w, e := o.alg.Writer(iotnwc.New(b))
					Expect(e).ToNot(HaveOccurred())
					Expect(w).ToNot(BeNil())
					clo.Add(w)

					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expCmp.MeasureDuration(o.alg.String()+" - "+o.txt, func() {
						n, e := w.Write(o.buf)
						Expect(e).ToNot(HaveOccurred())
						Expect(n).To(Equal(o.nbr))
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					expCmp.RecordValue(o.alg.String()+" - "+o.txt+" [CPU time]", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expCmp.RecordValue(o.alg.String()+" - "+o.txt+" [Memory]", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expCmp.RecordValue(o.alg.String()+" - "+o.txt+" [Allocs]", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))

					b.Reset()
				}
			}, gmeasure.SamplingConfig{N: 20})

			expDmp.Sample(func(idx int) {
				var clo = iotclo.New(context.Background())
				defer func() {
					_ = clo.Close()
				}()

				for i := 0; i < len(obj); i++ {
					o := obj[i]

					b := bytes.NewBuffer(make([]byte, 0, o.nbr))
					Expect(b).ToNot(BeNil())

					w, e := o.alg.Writer(iotnwc.New(b))
					Expect(e).ToNot(HaveOccurred())
					Expect(w).ToNot(BeNil())

					n, e := io.Copy(w, bytes.NewReader(o.buf))
					Expect(e).ToNot(HaveOccurred())
					Expect(n).To(Equal(int64(o.nbr)))

					e = w.Close()
					Expect(e).ToNot(HaveOccurred())

					r, e := o.alg.Reader(b)
					Expect(e).ToNot(HaveOccurred())
					Expect(r).ToNot(BeNil())
					clo.Add(r)

					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expDmp.MeasureDuration(o.alg.String()+" - "+o.txt, func() {
						n, e := io.Copy(io.Discard, r)
						Expect(n).To(Equal(int64(o.nbr)))
						Expect(e).ToNot(HaveOccurred())
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					expDmp.RecordValue(o.alg.String()+" - "+o.txt+" [CPU time]", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expDmp.RecordValue(o.alg.String()+" - "+o.txt+" [Memory]", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expDmp.RecordValue(o.alg.String()+" - "+o.txt+" [Allocs]", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))

					b.Reset()
				}
			}, gmeasure.SamplingConfig{N: 20})
		})
	})

	Context("TC-BC-011: Round-trip operations", func() {
		It("TC-BC-012: should benchmark full round-trip", func() {
			experiment := gmeasure.NewExperiment("Round-trip operations")
			AddReportEntry(experiment.Name, experiment)

			testData := newTestData(1024)
			algorithms := []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ}

			experiment.Sample(func(idx int) {
				for _, alg := range algorithms {
					experiment.MeasureDuration(alg.String(), func() {
						var buf bytes.Buffer
						w, err := alg.Writer(nopWriteCloser{&buf})
						if err == nil {
							w.Write(testData.dat)
							w.Close()

							r, err := alg.Reader(&buf)
							if err == nil {
								io.ReadAll(r)
								r.Close()
							}
						}
					})
				}
			}, gmeasure.SamplingConfig{N: 20})
		})
	})

	Context("TC-BC-013: Encoding operations", func() {
		It("TC-BC-014: should benchmark text marshaling", func() {
			experiment := gmeasure.NewExperiment("Text marshaling")
			AddReportEntry(experiment.Name, experiment)

			algorithms := compress.List()

			experiment.Sample(func(idx int) {
				for _, alg := range algorithms {
					experiment.MeasureDuration("Marshal-"+alg.String(), func() {
						alg.MarshalText()
					})

					data, _ := alg.MarshalText()
					var unmarshaled compress.Algorithm
					experiment.MeasureDuration("Unmarshal-"+alg.String(), func() {
						unmarshaled.UnmarshalText(data)
					})
				}
			}, gmeasure.SamplingConfig{N: 1000})
		})

		It("TC-BC-015: should benchmark JSON marshaling", func() {
			experiment := gmeasure.NewExperiment("JSON marshaling")
			AddReportEntry(experiment.Name, experiment)

			algorithms := compress.List()

			experiment.Sample(func(idx int) {
				for _, alg := range algorithms {
					experiment.MeasureDuration("Marshal-"+alg.String(), func() {
						alg.MarshalJSON()
					})

					data, _ := alg.MarshalJSON()
					var unmarshaled compress.Algorithm
					experiment.MeasureDuration("Unmarshal-"+alg.String(), func() {
						unmarshaled.UnmarshalJSON(data)
					})
				}
			}, gmeasure.SamplingConfig{N: 1000})
		})
	})

	Context("TC-BC-016: Compression ratio analysis", func() {
		It("TC-BC-017: should measure compression ratios", func() {
			sizes := []int{1024, 1024 * 10, 1024 * 100}
			algorithms := []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ}

			for _, size := range sizes {
				testData := newTestData(size)

				for _, alg := range algorithms {
					var buf bytes.Buffer

					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					w, err := alg.Writer(nopWriteCloser{&buf})
					Expect(err).ToNot(HaveOccurred())

					_, err = w.Write(testData.dat)
					Expect(err).ToNot(HaveOccurred())

					err = w.Close()
					Expect(err).ToNot(HaveOccurred())

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					compressed := buf.Len()
					ratio := (1 - float64(compressed)/float64(size)) * 100
					memUsed := float64(m1.TotalAlloc-m0.TotalAlloc) / 1024
					allocCount := m1.Mallocs - m0.Mallocs

					AddReportEntry(
						"Compression Ratio Analysis",
						map[string]interface{}{
							"Algorithm":       alg.String(),
							"Original Size":   size,
							"Compressed Size": compressed,
							"Ratio":           ratio,
							"CPU Time (ms)":   elapsed.Seconds() * 1000,
							"Memory (KB)":     memUsed,
							"Allocations":     allocCount,
						},
					)
				}
			}
		})
	})
})
