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

package zip_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/archive/archive/zip"
)

var _ = Describe("TC-BC-001: ZIP Performance Benchmarks", func() {
	Describe("TC-BC-002: Reader Operations", func() {
		It("TC-BC-003: should benchmark Reader operations with varying file counts", func() {
			experiment := NewExperiment("Reader Operations")
			AddReportEntry(experiment.Name, experiment)

			// Small archive (5 files)
			smallFiles := map[string]string{
				"file1.txt": strings.Repeat("a", 1000),
				"file2.txt": strings.Repeat("b", 1000),
				"file3.txt": strings.Repeat("c", 1000),
				"file4.txt": strings.Repeat("d", 1000),
				"file5.txt": strings.Repeat("e", 1000),
			}
			smallArchive, err := createTestZipFile(smallFiles)
			if err != nil {
				Skip("Failed to create test archive: " + err.Error())
			}
			defer os.Remove(smallArchive)

			experiment.SampleDuration("List - 5 files", func(idx int) {
				f, err := os.Open(smallArchive)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()
				reader.List()
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Info - 5 files", func(idx int) {
				f, err := os.Open(smallArchive)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()
				reader.Info("file1.txt")
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Get - 5 files", func(idx int) {
				f, err := os.Open(smallArchive)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()
				rc, _ := reader.Get("file1.txt")
				if rc != nil {
					io.ReadAll(rc)
					rc.Close()
				}
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Has - 5 files", func(idx int) {
				f, err := os.Open(smallArchive)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()
				reader.Has("file1.txt")
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Walk - 5 files", func(idx int) {
				f, err := os.Open(smallArchive)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()
				reader.Walk(func(_ fs.FileInfo, rc io.ReadCloser, _ string, _ string) bool {
					if rc != nil {
						rc.Close()
					}
					return true
				})
			}, SamplingConfig{N: 500, Duration: 0})
		})
	})

	Describe("TC-BC-004: Writer Operations", func() {
		It("TC-BC-005: should benchmark Writer operations with varying file sizes", func() {
			experiment := NewExperiment("Writer Operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Add - Small file (100B)", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				content := []byte(strings.Repeat("x", 100))
				info := createTestFileInfo("test.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
				writer.Close()
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Add - Medium file (10KB)", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				content := []byte(strings.Repeat("x", 10240))
				info := createTestFileInfo("test.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
				writer.Close()
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Add - Large file (1MB)", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				content := []byte(strings.Repeat("x", 1048576))
				info := createTestFileInfo("test.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
				writer.Close()
			}, SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Add - Multiple small files (10x100B)", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				for i := 0; i < 10; i++ {
					content := []byte(strings.Repeat("x", 100))
					info := createTestFileInfo("file.txt", int64(len(content)))
					reader := io.NopCloser(bytes.NewReader(content))
					writer.Add(info, reader, "", "")
				}
				writer.Close()
			}, SamplingConfig{N: 100, Duration: 0})
		})
	})

	Describe("TC-BC-006: Round-trip Operations", func() {
		It("TC-BC-007: should benchmark complete write-read cycle", func() {
			experiment := NewExperiment("Write-Read Round-trip")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Write and read small archive", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				content := []byte(strings.Repeat("x", 100))
				info := createTestFileInfo("test.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
				writer.Close()

				zipReader := newReaderWithSize(buf.Bytes())
				r, _ := zip.NewReader(zipReader)
				rc, _ := r.Get("test.txt")
				if rc != nil {
					io.ReadAll(rc)
					rc.Close()
				}
				r.Close()
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Write and list 5 files", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				for i := 0; i < 5; i++ {
					content := []byte(strings.Repeat("x", 100))
					info := createTestFileInfo("file.txt", int64(len(content)))
					reader := io.NopCloser(bytes.NewReader(content))
					writer.Add(info, reader, "", "")
				}
				writer.Close()

				zipReader := newReaderWithSize(buf.Bytes())
				r, _ := zip.NewReader(zipReader)
				r.List()
				r.Close()
			}, SamplingConfig{N: 100, Duration: 0})
		})
	})

	Describe("TC-BC-008: Memory Operations", func() {
		It("TC-BC-009: should benchmark creation and closure", func() {
			experiment := NewExperiment("Memory Operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("NewReader + Close", func(idx int) {
				testFiles := map[string]string{"test.txt": "content"}
				archiveFile, err := createTestZipFile(testFiles)
				if err != nil {
					return
				}
				defer os.Remove(archiveFile)

				f, err := os.Open(archiveFile)
				if err != nil {
					return
				}
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				reader.Close()
			}, SamplingConfig{N: 500, Duration: 0})

			experiment.SampleDuration("NewWriter + Close", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				writer.Close()
			}, SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("NewWriter + Add + Close", func(idx int) {
				buf := newBufferWriteCloser()
				writer, _ := zip.NewWriter(buf)
				content := []byte("test")
				info := createTestFileInfo("test.txt", int64(len(content)))
				reader := io.NopCloser(bytes.NewReader(content))
				writer.Add(info, reader, "", "")
				writer.Close()
			}, SamplingConfig{N: 500, Duration: 0})
		})
	})

	Describe("TC-BC-010: Real-world Scenarios", func() {
		It("TC-BC-011: should benchmark backup scenario", func() {
			experiment := NewExperiment("Backup Scenario")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Create backup archive (10 files, 1KB each)", func(idx int) {
				tmpFile, _ := os.CreateTemp("", "backup-*.zip")
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

				writer, _ := zip.NewWriter(tmpFile)
				for i := 0; i < 10; i++ {
					content := []byte(strings.Repeat("x", 1024))
					info := createTestFileInfo("file.txt", int64(len(content)))
					reader := io.NopCloser(bytes.NewReader(content))
					writer.Add(info, reader, "", "")
				}
				writer.Close()
			}, SamplingConfig{N: 50, Duration: 0})
		})

		It("TC-BC-012: should benchmark extraction scenario", func() {
			experiment := NewExperiment("Extraction Scenario")
			AddReportEntry(experiment.Name, experiment)

			// Prepare archive
			testFiles := map[string]string{
				"file1.txt": strings.Repeat("a", 1000),
				"file2.txt": strings.Repeat("b", 1000),
				"file3.txt": strings.Repeat("c", 1000),
			}
			archiveFile, err := createTestZipFile(testFiles)
			if err != nil {
				Skip("Failed to create test archive: " + err.Error())
			}
			defer os.Remove(archiveFile)

			experiment.SampleDuration("Extract all files", func(idx int) {
				f, err := os.Open(archiveFile)
				if err != nil {
					return
				}
				defer f.Close()
				reader, err := zip.NewReader(f)
				if err != nil {
					return
				}
				defer reader.Close()

				reader.Walk(func(_ fs.FileInfo, rc io.ReadCloser, _ string, _ string) bool {
					if rc != nil {
						io.ReadAll(rc)
						rc.Close()
					}
					return true
				})
			}, SamplingConfig{N: 100, Duration: 0})
		})
	})
})
