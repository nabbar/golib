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
 */

package types_test

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive/types"
)

type mockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

type mockReader struct {
	files map[string]string
}

func (m *mockReader) Close() error {
	return nil
}

func (m *mockReader) List() ([]string, error) {
	var result []string
	for path := range m.files {
		result = append(result, path)
	}
	return result, nil
}

func (m *mockReader) Info(path string) (fs.FileInfo, error) {
	content, exists := m.files[path]
	if !exists {
		return nil, fs.ErrNotExist
	}
	return &mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

func (m *mockReader) Get(path string) (io.ReadCloser, error) {
	content, exists := m.files[path]
	if !exists {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader([]byte(content))), nil
}

func (m *mockReader) Has(path string) bool {
	_, exists := m.files[path]
	return exists
}

func (m *mockReader) Walk(fn types.FuncExtract) {
	for path, content := range m.files {
		info := &mockFileInfo{
			name:    filepath.Base(path),
			size:    int64(len(content)),
			mode:    0644,
			modTime: time.Now(),
			isDir:   false,
		}
		reader := io.NopCloser(bytes.NewReader([]byte(content)))
		if !fn(info, reader, path, "") {
			return
		}
	}
}

type mockWriter struct {
	files map[string]string
}

func (m *mockWriter) Close() error {
	return nil
}

func (m *mockWriter) Add(info fs.FileInfo, r io.ReadCloser, forcePath, link string) error {
	if r == nil {
		return nil
	}
	defer r.Close()

	path := forcePath
	if path == "" {
		path = info.Name()
	}

	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	m.files[path] = string(content)
	return nil
}

func (m *mockWriter) FromPath(source, filter string, fn types.ReplaceName) error {
	return nil
}

var _ = Describe("TC-IF-001: Interface Definitions", func() {
	Describe("TC-IF-002: Reader Interface", func() {
		It("TC-IF-003: should satisfy Reader interface", func() {
			var r types.Reader = &mockReader{
				files: map[string]string{
					"test.txt": "content",
				},
			}
			Expect(r).ToNot(BeNil())
		})

		It("TC-IF-004: should implement Close method", func() {
			r := &mockReader{files: map[string]string{}}
			err := r.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-IF-005: should implement List method", func() {
			r := &mockReader{
				files: map[string]string{
					"file1.txt": "content1",
					"file2.txt": "content2",
				},
			}
			files, err := r.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(files).To(HaveLen(2))
		})

		It("TC-IF-006: should implement Info method", func() {
			r := &mockReader{
				files: map[string]string{
					"test.txt": "content",
				},
			}
			info, err := r.Info("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(info).ToNot(BeNil())
			Expect(info.Size()).To(Equal(int64(7)))
		})

		It("TC-IF-007: should implement Get method", func() {
			r := &mockReader{
				files: map[string]string{
					"test.txt": "content",
				},
			}
			rc, err := r.Get("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(rc).ToNot(BeNil())
			defer rc.Close()

			content, _ := io.ReadAll(rc)
			Expect(string(content)).To(Equal("content"))
		})

		It("TC-IF-008: should implement Has method", func() {
			r := &mockReader{
				files: map[string]string{
					"test.txt": "content",
				},
			}
			Expect(r.Has("test.txt")).To(BeTrue())
			Expect(r.Has("missing.txt")).To(BeFalse())
		})

		It("TC-IF-009: should implement Walk method", func() {
			r := &mockReader{
				files: map[string]string{
					"file1.txt": "content1",
					"file2.txt": "content2",
				},
			}

			count := 0
			r.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
				if rc != nil {
					rc.Close()
				}
				count++
				return true
			})

			Expect(count).To(Equal(2))
		})
	})

	Describe("TC-IF-010: Writer Interface", func() {
		It("TC-IF-011: should satisfy Writer interface", func() {
			var w types.Writer = &mockWriter{
				files: map[string]string{},
			}
			Expect(w).ToNot(BeNil())
		})

		It("TC-IF-012: should implement Close method", func() {
			w := &mockWriter{files: map[string]string{}}
			err := w.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-IF-013: should implement Add method", func() {
			w := &mockWriter{files: map[string]string{}}

			info := &mockFileInfo{
				name: "test.txt",
				size: 7,
			}
			reader := io.NopCloser(bytes.NewReader([]byte("content")))

			err := w.Add(info, reader, "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(w.files).To(HaveKey("test.txt"))
			Expect(w.files["test.txt"]).To(Equal("content"))
		})

		It("TC-IF-014: should implement Add with custom path", func() {
			w := &mockWriter{files: map[string]string{}}

			info := &mockFileInfo{
				name: "original.txt",
				size: 7,
			}
			reader := io.NopCloser(bytes.NewReader([]byte("content")))

			err := w.Add(info, reader, "custom/path.txt", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(w.files).To(HaveKey("custom/path.txt"))
		})

		It("TC-IF-015: should implement FromPath method", func() {
			w := &mockWriter{files: map[string]string{}}
			err := w.FromPath("/test", "*.txt", nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-IF-016: Function Types", func() {
		It("TC-IF-017: should define FuncExtract type", func() {
			var fn types.FuncExtract = func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
				return true
			}
			Expect(fn).ToNot(BeNil())

			result := fn(nil, nil, "", "")
			Expect(result).To(BeTrue())
		})

		It("TC-IF-018: should define ReplaceName type", func() {
			var fn types.ReplaceName = func(source string) string {
				return "prefix/" + source
			}
			Expect(fn).ToNot(BeNil())

			result := fn("test.txt")
			Expect(result).To(Equal("prefix/test.txt"))
		})

		It("TC-IF-019: should use FuncExtract to control iteration", func() {
			stopAfterTwo := func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
				if r != nil {
					r.Close()
				}
				return path != "stop"
			}

			r := &mockReader{
				files: map[string]string{
					"file1.txt": "content1",
					"stop":      "stop",
					"file2.txt": "content2",
				},
			}

			count := 0
			r.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
				if rc != nil {
					rc.Close()
				}
				count++
				return stopAfterTwo(info, rc, path, link)
			})

			Expect(count).To(BeNumerically("<=", 3))
		})

		It("TC-IF-020: should use ReplaceName for path transformation", func() {
			addPrefix := func(source string) string {
				return "backup/" + source
			}

			result := addPrefix("data/file.txt")
			Expect(result).To(Equal("backup/data/file.txt"))
		})

		It("TC-IF-021: should use ReplaceName to flatten paths", func() {
			flatten := func(source string) string {
				return filepath.Base(source)
			}

			result := flatten("deep/nested/file.txt")
			Expect(result).To(Equal("file.txt"))
		})
	})

	Describe("TC-IF-022: Error Handling", func() {
		It("TC-IF-023: should return fs.ErrNotExist for missing files in Info", func() {
			r := &mockReader{files: map[string]string{}}
			_, err := r.Info("missing.txt")
			Expect(err).To(Equal(fs.ErrNotExist))
		})

		It("TC-IF-024: should return fs.ErrNotExist for missing files in Get", func() {
			r := &mockReader{files: map[string]string{}}
			_, err := r.Get("missing.txt")
			Expect(err).To(Equal(fs.ErrNotExist))
		})

		It("TC-IF-025: should handle nil reader in Add", func() {
			w := &mockWriter{files: map[string]string{}}

			info := &mockFileInfo{
				name: "dir/",
				size: 0,
			}

			err := w.Add(info, nil, "", "")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-IF-026: Integration Scenarios", func() {
		It("TC-IF-027: should list and retrieve all files", func() {
			r := &mockReader{
				files: map[string]string{
					"file1.txt": "content1",
					"file2.txt": "content2",
				},
			}

			files, err := r.List()
			Expect(err).ToNot(HaveOccurred())

			for _, path := range files {
				rc, err := r.Get(path)
				Expect(err).ToNot(HaveOccurred())
				rc.Close()
			}
		})

		It("TC-IF-028: should add multiple files to writer", func() {
			w := &mockWriter{files: map[string]string{}}

			files := map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
				"file3.txt": "content3",
			}

			for name, content := range files {
				info := &mockFileInfo{
					name: name,
					size: int64(len(content)),
				}
				reader := io.NopCloser(bytes.NewReader([]byte(content)))
				err := w.Add(info, reader, "", "")
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(w.files).To(HaveLen(3))
		})

		It("TC-IF-029: should walk and process files selectively", func() {
			r := &mockReader{
				files: map[string]string{
					"readme.txt":  "readme",
					"data.json":   "{}",
					"config.yaml": "key: value",
				},
			}

			txtFiles := []string{}
			r.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
				if rc != nil {
					rc.Close()
				}
				if filepath.Ext(path) == ".txt" {
					txtFiles = append(txtFiles, path)
				}
				return true
			})

			Expect(txtFiles).To(ContainElement("readme.txt"))
			Expect(txtFiles).ToNot(ContainElement("data.json"))
		})
	})
})
