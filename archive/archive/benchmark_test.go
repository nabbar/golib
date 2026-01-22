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

package archive_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/archive/archive"
)

var _ = Describe("TC-BC-001: Benchmarks", func() {
	Context("TC-BC-002: Algorithm operations", func() {
		It("TC-BC-003: should benchmark Parse operations", func() {
			experiment := gmeasure.NewExperiment("Parse operations")
			AddReportEntry(experiment.Name, experiment)

			inputs := []string{"tar", "zip", "none", "unknown"}

			experiment.Sample(func(idx int) {
				for _, input := range inputs {
					experiment.MeasureDuration(input, func() {
						_ = archive.Parse(input)
					})
				}
			}, gmeasure.SamplingConfig{N: 100})
		})

		It("TC-BC-004: should benchmark String operations", func() {
			experiment := gmeasure.NewExperiment("String operations")
			AddReportEntry(experiment.Name, experiment)

			algorithms := []archive.Algorithm{archive.None, archive.Tar, archive.Zip}

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

			algorithms := []archive.Algorithm{archive.None, archive.Tar, archive.Zip}

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

			// Create valid headers
			tarHeader := make([]byte, 263)
			copy(tarHeader[257:263], append([]byte("ustar"), 0x00))

			zipHeader := make([]byte, 263)
			zipHeader[0] = 0x50
			zipHeader[1] = 0x4b
			zipHeader[2] = 0x03
			zipHeader[3] = 0x04

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("tar", func() {
					_ = archive.Tar.DetectHeader(tarHeader)
				})
				experiment.MeasureDuration("zip", func() {
					_ = archive.Zip.DetectHeader(zipHeader)
				})
			}, gmeasure.SamplingConfig{N: 1000})
		})
	})

	Context("TC-BC-007: Detection operations", func() {
		It("TC-BC-008: should benchmark Detect with various formats", func() {
			experiment := gmeasure.NewExperiment("Detect operations")
			AddReportEntry(experiment.Name, experiment)

			// Prepare TAR archive
			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "test.txt", strings.Repeat("x", 1000))

			var tarBuf bytes.Buffer
			tarWriter, _ := archive.Tar.Writer(&nopWriteCloser{&tarBuf})
			_ = tarWriter.FromPath(tmpDir, "*.txt", nil)
			_ = tarWriter.Close()

			// Prepare ZIP archive
			tmpFile, _ := createTempArchiveFile(".zip")
			defer os.Remove(tmpFile.Name())
			zipWriter, _ := archive.Zip.Writer(tmpFile)
			_ = zipWriter.FromPath(tmpDir, "*.txt", nil)
			_ = zipWriter.Close()
			tmpFile.Close()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("tar", func() {
					_, reader, stream, err := archive.Detect(io.NopCloser(bytes.NewReader(tarBuf.Bytes())))
					if err == nil {
						if reader != nil {
							reader.Close()
						}
						if stream != nil {
							stream.Close()
						}
					}
				})

				experiment.MeasureDuration("zip", func() {
					f, _ := os.Open(tmpFile.Name())
					if f != nil {
						defer f.Close()
						_, reader, stream, err := archive.Detect(f)
						if err == nil {
							if reader != nil {
								reader.Close()
							}
							if stream != nil {
								stream.Close()
							}
						}
					}
				})
			}, gmeasure.SamplingConfig{N: 100})
		})
	})

	Context("TC-BC-009: Archive creation and extraction operations", func() {
		It("TC-BC-010: should benchmark archive creation with different sizes", func() {
			sizes := map[string]int{
				"Small Data (1KB)":   1024,
				"Medium Data (10KB)": 10240,
				"Large Data (100KB)": 102400,
			}

			for sizeLabel, size := range sizes {
				expTarCreate := gmeasure.NewExperiment("TAR Creation - " + sizeLabel)
				AddReportEntry(expTarCreate.Name, expTarCreate)

				expZipCreate := gmeasure.NewExperiment("ZIP Creation - " + sizeLabel)
				AddReportEntry(expZipCreate.Name, expZipCreate)

				// Prepare test data
				tmpDir, _ := createTempDir()
				_ = createTestFile(tmpDir, "test.txt", strings.Repeat("x", size))

				// Benchmark TAR creation
				expTarCreate.Sample(func(idx int) {
					var buf bytes.Buffer
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expTarCreate.MeasureDuration("create", func() {
						writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					archiveSize := buf.Len()
					ratio := (1 - float64(archiveSize)/float64(size)) * 100
					if ratio < 0 {
						ratio = 0
					}

					expTarCreate.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expTarCreate.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expTarCreate.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
					expTarCreate.RecordValue("Archive Size", float64(archiveSize), gmeasure.Units("bytes"))
					expTarCreate.RecordValue("Overhead", float64(archiveSize-size), gmeasure.Units("bytes"))
				}, gmeasure.SamplingConfig{N: 20})

				// Benchmark ZIP creation
				expZipCreate.Sample(func(idx int) {
					tmpFile, _ := createTempArchiveFile(".zip")
					defer os.Remove(tmpFile.Name())
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expZipCreate.MeasureDuration("create", func() {
						writer, _ := archive.Zip.Writer(tmpFile)
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)
					tmpFile.Close()

					stat, _ := os.Stat(tmpFile.Name())
					archiveSize := int(stat.Size())
					ratio := (1 - float64(archiveSize)/float64(size)) * 100
					if ratio < 0 {
						ratio = 0
					}

					expZipCreate.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expZipCreate.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expZipCreate.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
					expZipCreate.RecordValue("Archive Size", float64(archiveSize), gmeasure.Units("bytes"))
					expZipCreate.RecordValue("Overhead", float64(archiveSize-size), gmeasure.Units("bytes"))
				}, gmeasure.SamplingConfig{N: 20})

				os.RemoveAll(tmpDir)
			}
		})

		It("TC-BC-011: should benchmark archive extraction with different sizes", func() {
			sizes := map[string]int{
				"Small Data (1KB)":   1024,
				"Medium Data (10KB)": 10240,
				"Large Data (100KB)": 102400,
			}

			for sizeLabel, size := range sizes {
				expTarExtract := gmeasure.NewExperiment("TAR Extraction - " + sizeLabel)
				AddReportEntry(expTarExtract.Name, expTarExtract)

				expZipExtract := gmeasure.NewExperiment("ZIP Extraction - " + sizeLabel)
				AddReportEntry(expZipExtract.Name, expZipExtract)

				// Prepare test archives
				tmpDir, _ := createTempDir()
				_ = createTestFile(tmpDir, "test.txt", strings.Repeat("x", size))

				// Create TAR archive
				var tarBuf bytes.Buffer
				tarWriter, _ := archive.Tar.Writer(&nopWriteCloser{&tarBuf})
				_ = tarWriter.FromPath(tmpDir, "*.txt", nil)
				_ = tarWriter.Close()

				// Create ZIP archive
				tmpZipFile, _ := createTempArchiveFile(".zip")
				zipWriter, _ := archive.Zip.Writer(tmpZipFile)
				_ = zipWriter.FromPath(tmpDir, "*.txt", nil)
				_ = zipWriter.Close()
				tmpZipFile.Close()

				// Benchmark TAR extraction
				expTarExtract.Sample(func(idx int) {
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expTarExtract.MeasureDuration("extract", func() {
						reader, _ := archive.Tar.Reader(io.NopCloser(bytes.NewReader(tarBuf.Bytes())))
						rc, _ := reader.Get("test.txt")
						if rc != nil {
							_, _ = io.Copy(io.Discard, rc)
							rc.Close()
						}
						reader.Close()
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					expTarExtract.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expTarExtract.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expTarExtract.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
				}, gmeasure.SamplingConfig{N: 20})

				// Benchmark ZIP extraction
				expZipExtract.Sample(func(idx int) {
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expZipExtract.MeasureDuration("extract", func() {
						f, _ := os.Open(tmpZipFile.Name())
						if f != nil {
							defer f.Close()
							reader, _ := archive.Zip.Reader(f)
							if reader != nil {
								defer reader.Close()
								rc, _ := reader.Get("test.txt")
								if rc != nil {
									_, _ = io.Copy(io.Discard, rc)
									rc.Close()
								}
							}
						}
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					expZipExtract.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expZipExtract.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expZipExtract.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
				}, gmeasure.SamplingConfig{N: 20})

				os.RemoveAll(tmpDir)
				os.Remove(tmpZipFile.Name())
			}
		})
	})

	Context("TC-BC-012: Multiple files operations", func() {
		It("TC-BC-013: should benchmark multiple files archiving", func() {
			fileCounts := map[string]int{
				"5 files":  5,
				"10 files": 10,
				"25 files": 25,
			}

			for label, count := range fileCounts {
				expTar := gmeasure.NewExperiment("TAR Multiple Files - " + label)
				AddReportEntry(expTar.Name, expTar)

				expZip := gmeasure.NewExperiment("ZIP Multiple Files - " + label)
				AddReportEntry(expZip.Name, expZip)

				// Prepare test data
				tmpDir, _ := createTempDir()
				totalSize := 0
				for i := 0; i < count; i++ {
					content := strings.Repeat("x", 1000)
					_ = createTestFile(tmpDir, filepath.Join("file", "test"+strings.Repeat("0", 2-len(strings.Split(strings.Trim(strings.Repeat("0", i), "0"), "")))+strings.Trim(strings.Repeat("0", i), "0")+".txt"), content)
					totalSize += len(content)
				}

				// Benchmark TAR
				expTar.Sample(func(idx int) {
					var buf bytes.Buffer
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expTar.MeasureDuration("create", func() {
						writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)

					expTar.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expTar.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expTar.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
				}, gmeasure.SamplingConfig{N: 20})

				// Benchmark ZIP
				expZip.Sample(func(idx int) {
					tmpFile, _ := createTempArchiveFile(".zip")
					defer os.Remove(tmpFile.Name())
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					t0 := time.Now()

					expZip.MeasureDuration("create", func() {
						writer, _ := archive.Zip.Writer(tmpFile)
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
					})

					elapsed := time.Since(t0)
					runtime.ReadMemStats(&m1)
					tmpFile.Close()

					expZip.RecordValue("CPU time", elapsed.Seconds()*1000, gmeasure.Units("ms"))
					expZip.RecordValue("Memory", float64(m1.TotalAlloc-m0.TotalAlloc)/1024, gmeasure.Units("KB"))
					expZip.RecordValue("Allocs", float64(m1.Mallocs-m0.Mallocs), gmeasure.Units("allocs"))
				}, gmeasure.SamplingConfig{N: 20})

				os.RemoveAll(tmpDir)
			}
		})
	})

	Context("TC-BC-014: Round-trip operations", func() {
		It("TC-BC-015: should benchmark full round-trip", func() {
			experiment := gmeasure.NewExperiment("Round-trip operations")
			AddReportEntry(experiment.Name, experiment)

			tmpDir, _ := createTempDir()
			defer os.RemoveAll(tmpDir)
			_ = createTestFile(tmpDir, "test.txt", strings.Repeat("x", 1024))

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("tar", func() {
					var buf bytes.Buffer
					writer, _ := archive.Tar.Writer(&nopWriteCloser{&buf})
					_ = writer.FromPath(tmpDir, "*.txt", nil)
					_ = writer.Close()

					reader, _ := archive.Tar.Reader(io.NopCloser(&buf))
					rc, _ := reader.Get("test.txt")
					if rc != nil {
						_, _ = io.ReadAll(rc)
						rc.Close()
					}
					reader.Close()
				})

				experiment.MeasureDuration("zip", func() {
					tmpFile, _ := createTempArchiveFile(".zip")
					defer os.Remove(tmpFile.Name())

					writer, _ := archive.Zip.Writer(tmpFile)
					_ = writer.FromPath(tmpDir, "*.txt", nil)
					_ = writer.Close()
					tmpFile.Close()

					f, _ := os.Open(tmpFile.Name())
					f.Stat()
					if f != nil {
						reader, err := archive.Zip.Reader(f)
						if err == nil && reader != nil {
							rc, _ := reader.Get("test.txt")
							if rc != nil {
								_, _ = io.ReadAll(rc)
								rc.Close()
							}
							reader.Close()
						}
						f.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 20})
		})
	})

	Context("TC-BC-016: Size and overhead analysis", func() {
		It("TC-BC-017: should measure archive overhead", func() {
			sizes := []int{1024, 10240, 102400}
			algorithms := []archive.Algorithm{archive.Tar, archive.Zip}

			for _, size := range sizes {
				for _, alg := range algorithms {
					tmpDir, _ := createTempDir()
					_ = createTestFile(tmpDir, "test.txt", strings.Repeat("x", size))

					var archiveSize int64

					if alg == archive.Tar {
						var buf bytes.Buffer
						writer, _ := alg.Writer(&nopWriteCloser{&buf})
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
						archiveSize = int64(buf.Len())
					} else {
						tmpFile, _ := createTempArchiveFile(".zip")
						defer os.Remove(tmpFile.Name())
						writer, _ := alg.Writer(tmpFile)
						_ = writer.FromPath(tmpDir, "*.txt", nil)
						_ = writer.Close()
						tmpFile.Close()
						stat, _ := os.Stat(tmpFile.Name())
						archiveSize = stat.Size()
					}

					overhead := archiveSize - int64(size)
					overheadPercent := (float64(overhead) / float64(size)) * 100

					AddReportEntry(
						"Archive Overhead Analysis",
						map[string]interface{}{
							"Algorithm":        alg.String(),
							"Original Size":    size,
							"Archive Size":     archiveSize,
							"Overhead (bytes)": overhead,
							"Overhead (%)":     overheadPercent,
						},
					)

					os.RemoveAll(tmpDir)
				}
			}
		})
	})
})
