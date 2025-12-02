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

package ioutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nabbar/golib/ioutils"
)

// BenchmarkPathCheckCreate_NewFile benchmarks the creation of a new file.
// This measures the baseline performance for creating files that don't exist yet.
func BenchmarkPathCheckCreate_NewFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tmpDir, filepath.Base(tmpDir)+"-file-"+string(rune('0'+i%10))+".txt")
		if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_ExistingFile benchmarks checking an existing file.
// This measures the performance of the idempotent case where the file already exists
// with correct permissions. This is the common case in long-running applications.
func BenchmarkPathCheckCreate_ExistingFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file once before benchmark
	filePath := filepath.Join(tmpDir, "existing.txt")
	if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
		b.Fatalf("Failed to create initial file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_PermissionUpdate benchmarks updating file permissions.
// This measures the cost of changing permissions on an existing file.
func BenchmarkPathCheckCreate_PermissionUpdate(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file once before benchmark
	filePath := filepath.Join(tmpDir, "update.txt")
	if err := ioutils.PathCheckCreate(true, filePath, 0600, 0755); err != nil {
		b.Fatalf("Failed to create initial file: %v", err)
	}

	perms := []os.FileMode{0600, 0644, 0666, 0640}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		perm := perms[i%len(perms)]
		if err := ioutils.PathCheckCreate(true, filePath, perm, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_NewDirectory benchmarks creating a new directory.
// This measures the baseline performance for directory creation.
func BenchmarkPathCheckCreate_NewDirectory(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dirPath := filepath.Join(tmpDir, filepath.Base(tmpDir)+"-dir-"+string(rune('0'+i%10)))
		if err := ioutils.PathCheckCreate(false, dirPath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_ExistingDirectory benchmarks checking an existing directory.
// This measures the performance of the idempotent case for directories.
func BenchmarkPathCheckCreate_ExistingDirectory(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory once before benchmark
	dirPath := filepath.Join(tmpDir, "existing")
	if err := ioutils.PathCheckCreate(false, dirPath, 0644, 0755); err != nil {
		b.Fatalf("Failed to create initial directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ioutils.PathCheckCreate(false, dirPath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_NestedPath benchmarks creating files in nested directories.
// This measures the overhead of creating parent directories recursively.
func BenchmarkPathCheckCreate_NestedPath(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tmpDir, filepath.Base(tmpDir)+"-nest"+string(rune('0'+i%10)), "deep", "path", "file.txt")
		if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_DeepNesting benchmarks creating files in very deeply nested directories.
// This tests the performance impact of recursive directory creation.
func BenchmarkPathCheckCreate_DeepNesting(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tmpDir, filepath.Base(tmpDir)+"-deep"+string(rune('0'+i%10)), "a", "b", "c", "d", "e", "f", "g", "file.txt")
		if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
			b.Fatalf("PathCheckCreate failed: %v", err)
		}
	}
}

// BenchmarkPathCheckCreate_Parallel benchmarks concurrent file creation.
// This measures the scalability of PathCheckCreate under parallel workload.
func BenchmarkPathCheckCreate_Parallel(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			filePath := filepath.Join(tmpDir, filepath.Base(tmpDir)+"-parallel", filepath.Base(tmpDir)+"-file"+string(rune('0'+idx%10))+".txt")
			if err := ioutils.PathCheckCreate(true, filePath, 0644, 0755); err != nil {
				b.Fatalf("PathCheckCreate failed: %v", err)
			}
			idx++
		}
	})
}
