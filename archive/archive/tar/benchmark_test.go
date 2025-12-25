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

package tar_test

import (
	"bytes"
	"io"
	"io/fs"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/archive/archive/tar"
)

var _ = Describe("TC-BC-001: Benchmarks", func() {
	It("TC-BC-002: Reader operations", func() {
		experiment := gmeasure.NewExperiment("Reader Operations")
		AddReportEntry(experiment.Name, experiment)

		// Prepare test data
		testFiles := map[string]string{
			"file1.txt":     strings.Repeat("a", 1000),
			"file2.txt":     strings.Repeat("b", 1000),
			"file3.txt":     strings.Repeat("c", 1000),
			"dir/file4.txt": strings.Repeat("d", 1000),
			"dir/file5.txt": strings.Repeat("e", 1000),
		}
		archiveBuf := createTestArchive(testFiles)

		experiment.SampleDuration("List", func(idx int) {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()
			reader.List()
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Info", func(idx int) {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()
			reader.Info("file1.txt")
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Get", func(idx int) {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()
			rc, _ := reader.Get("file1.txt")
			if rc != nil {
				io.ReadAll(rc)
				rc.Close()
			}
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Has", func(idx int) {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()
			reader.Has("file1.txt")
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Walk", func(idx int) {
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			defer reader.Close()
			reader.Walk(func(_ fs.FileInfo, _ io.ReadCloser, _ string, _ string) bool {
				return true
			})
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})
	})

	It("TC-BC-003: Writer operations", func() {
		experiment := gmeasure.NewExperiment("Writer Operations")
		AddReportEntry(experiment.Name, experiment)

		experiment.SampleDuration("Add small file", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "test.txt", size: 100, mode: 0644, modTime: time.Now()}
			content := io.NopCloser(strings.NewReader(strings.Repeat("x", 100)))
			writer.Add(info, content, "test.txt", "")
			writer.Close()
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Add medium file (10KB)", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "test.txt", size: 10240, mode: 0644, modTime: time.Now()}
			content := io.NopCloser(strings.NewReader(strings.Repeat("x", 10240)))
			writer.Add(info, content, "test.txt", "")
			writer.Close()
		}, gmeasure.SamplingConfig{N: 100, Duration: 0})

		experiment.SampleDuration("Add large file (1MB)", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "test.txt", size: 1048576, mode: 0644, modTime: time.Now()}
			content := io.NopCloser(strings.NewReader(strings.Repeat("x", 1048576)))
			writer.Add(info, content, "test.txt", "")
			writer.Close()
		}, gmeasure.SamplingConfig{N: 10, Duration: 0})

		experiment.SampleDuration("Add multiple files", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			for i := 0; i < 10; i++ {
				info := &testFileInfo{name: "file.txt", size: 100, mode: 0644, modTime: time.Now()}
				content := io.NopCloser(strings.NewReader(strings.Repeat("x", 100)))
				writer.Add(info, content, "file.txt", "")
			}
			writer.Close()
		}, gmeasure.SamplingConfig{N: 100, Duration: 0})
	})

	It("TC-BC-004: Round-trip operations", func() {
		experiment := gmeasure.NewExperiment("Round-trip")
		AddReportEntry(experiment.Name, experiment)

		experiment.SampleDuration("Write and read small archive", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			info := &testFileInfo{name: "test.txt", size: 100, mode: 0644, modTime: time.Now()}
			content := io.NopCloser(strings.NewReader(strings.Repeat("x", 100)))
			writer.Add(info, content, "test.txt", "")
			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			rc, _ := reader.Get("test.txt")
			if rc != nil {
				io.ReadAll(rc)
				rc.Close()
			}
			reader.Close()
		}, gmeasure.SamplingConfig{N: 100, Duration: 0})

		experiment.SampleDuration("Write and read multiple files", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			for i := 0; i < 5; i++ {
				info := &testFileInfo{name: "file.txt", size: 100, mode: 0644, modTime: time.Now()}
				content := io.NopCloser(strings.NewReader(strings.Repeat("x", 100)))
				writer.Add(info, content, "file.txt", "")
			}
			writer.Close()

			reader, _ := tar.NewReader(io.NopCloser(&buf))
			reader.List()
			reader.Close()
		}, gmeasure.SamplingConfig{N: 100, Duration: 0})
	})

	It("TC-BC-005: Memory operations", func() {
		experiment := gmeasure.NewExperiment("Memory Operations")
		AddReportEntry(experiment.Name, experiment)

		experiment.SampleDuration("Create and close reader", func(idx int) {
			archiveBuf := createTestArchive(map[string]string{"test.txt": "content"})
			reader, _ := tar.NewReader(io.NopCloser(bytes.NewReader(archiveBuf.Bytes())))
			reader.Close()
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

		experiment.SampleDuration("Create and close writer", func(idx int) {
			var buf bytes.Buffer
			writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
			writer.Close()
		}, gmeasure.SamplingConfig{N: 1000, Duration: 0})
	})
})
