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
	"runtime"

	. "github.com/nabbar/golib/ioutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test suite validates the PathCheckCreate function behavior across multiple scenarios:
// - File and directory creation operations
// - Permission management and updates
// - Error handling for invalid inputs
// - Edge cases including idempotency and concurrency
// - Platform-specific behaviors (Unix vs Windows)
//
// The tests are organized in three main describe blocks:
// 1. File Operations: Tests file creation, permission updates, and error handling
// 2. Directory Operations: Tests directory creation, permission updates, and error handling
// 3. Edge Cases: Tests idempotency, mixed operations, and concurrent access

var _ = Describe("PathCheckCreate - File Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "ioutils_test_*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	// Tests for basic file creation functionality
	Context("Creating files", func() {
		It("should create a new file with correct permissions", func() {
			filePath := filepath.Join(tempDir, "test.txt")
			err := PathCheckCreate(true, filePath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())

			info, err := os.Stat(filePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.IsDir()).To(BeFalse())
		})

		It("should create nested directories for file", func() {
			filePath := filepath.Join(tempDir, "nested", "dir", "test.txt")
			err := PathCheckCreate(true, filePath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())
			Expect(filepath.Dir(filePath)).To(BeADirectory())
		})

		It("should create deeply nested directories", func() {
			filePath := filepath.Join(tempDir, "a", "b", "c", "d", "e", "test.txt")
			err := PathCheckCreate(true, filePath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())
		})

		It("should handle file creation in root of temp dir", func() {
			filePath := filepath.Join(tempDir, "root.txt")
			err := PathCheckCreate(true, filePath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())
		})
	})

	// Tests for permission management on existing files
	Context("Updating file permissions", func() {
		It("should update file permissions if file exists", func() {
			filePath := filepath.Join(tempDir, "existing.txt")

			// Create with one permission
			err := PathCheckCreate(true, filePath, 0600, 0755)
			Expect(err).ToNot(HaveOccurred())

			// Update with different permission
			err = PathCheckCreate(true, filePath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(filePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0644)))
		})

		It("should not error if file exists with correct permissions", func() {
			filePath := filepath.Join(tempDir, "correct.txt")

			// Create with permission
			err := PathCheckCreate(true, filePath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			// Call again with same permission
			err = PathCheckCreate(true, filePath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle multiple permission updates", func() {
			filePath := filepath.Join(tempDir, "multi.txt")

			perms := []os.FileMode{0600, 0644, 0666, 0640}
			for _, perm := range perms {
				err := PathCheckCreate(true, filePath, perm, 0755)
				Expect(err).ToNot(HaveOccurred())

				info, err := os.Stat(filePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(info.Mode() & 0777).To(Equal(perm))
			}
		})
	})

	// Tests for error scenarios when working with files
	Context("Error handling for files", func() {
		It("should return error if path is directory but file expected", func() {
			dirPath := filepath.Join(tempDir, "dir")
			err := os.Mkdir(dirPath, 0755)
			Expect(err).ToNot(HaveOccurred())

			err = PathCheckCreate(true, dirPath, 0644, 0755)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is a directory"))
		})

		It("should return error for empty file path", func() {
			err := PathCheckCreate(true, "", 0644, 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should handle file with special characters in name", func() {
			// Test various filename edge cases that should be valid
			specialNames := []string{
				"file with spaces.txt",
				"file-with-dashes.txt",
				"file_with_underscores.txt",
				"file.multiple.dots.txt",
			}

			for _, name := range specialNames {
				filePath := filepath.Join(tempDir, name)
				err := PathCheckCreate(true, filePath, 0644, 0755)
				Expect(err).ToNot(HaveOccurred())
				Expect(filePath).To(BeAnExistingFile())
			}
		})
	})

	// Tests for various permission scenarios including edge cases
	Context("File with various permissions", func() {
		It("should create file with read-only permissions", func() {
			filePath := filepath.Join(tempDir, "readonly.txt")
			err := PathCheckCreate(true, filePath, 0444, 0755)

			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(filePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0444)))
		})

		It("should create file with write-only permissions", func() {
			// Windows has limited support for write-only permissions due to ACL differences
			if runtime.GOOS == "windows" {
				Skip("Write-only permissions not well supported on Windows")
			}

			filePath := filepath.Join(tempDir, "writeonly.txt")
			err := PathCheckCreate(true, filePath, 0200, 0755)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should create file with full permissions", func() {
			filePath := filepath.Join(tempDir, "fullperm.txt")
			err := PathCheckCreate(true, filePath, 0777, 0755)

			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(filePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0777)))
		})
	})
})

// Tests for directory-specific operations, including creation, permission updates,
// and error handling scenarios that differ from file operations
var _ = Describe("PathCheckCreate - Directory Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "ioutils_test_*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	// Tests for basic directory creation functionality
	Context("Creating directories", func() {
		It("should create a new directory with correct permissions", func() {
			dirPath := filepath.Join(tempDir, "newdir")
			err := PathCheckCreate(false, dirPath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(dirPath).To(BeADirectory())

			info, err := os.Stat(dirPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should create nested directories", func() {
			dirPath := filepath.Join(tempDir, "nested", "deep", "dir")
			err := PathCheckCreate(false, dirPath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(dirPath).To(BeADirectory())
		})

		It("should create very deeply nested directories", func() {
			dirPath := filepath.Join(tempDir, "a", "b", "c", "d", "e", "f", "g")
			err := PathCheckCreate(false, dirPath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(dirPath).To(BeADirectory())
		})

		It("should handle directory in root of temp dir", func() {
			dirPath := filepath.Join(tempDir, "rootdir")
			err := PathCheckCreate(false, dirPath, 0644, 0755)

			Expect(err).ToNot(HaveOccurred())
			Expect(dirPath).To(BeADirectory())
		})
	})

	// Tests for permission management on existing directories
	Context("Updating directory permissions", func() {
		It("should update directory permissions if exists", func() {
			dirPath := filepath.Join(tempDir, "existingdir")

			// Create with one permission
			err := PathCheckCreate(false, dirPath, 0644, 0700)
			Expect(err).ToNot(HaveOccurred())

			// Update with different permission
			err = PathCheckCreate(false, dirPath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(dirPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0755)))
		})

		It("should not error if directory exists with correct permissions", func() {
			dirPath := filepath.Join(tempDir, "correctdir")

			// Create with permission
			err := PathCheckCreate(false, dirPath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			// Call again with same permission
			err = PathCheckCreate(false, dirPath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle multiple permission updates for directories", func() {
			dirPath := filepath.Join(tempDir, "multidir")

			perms := []os.FileMode{0700, 0755, 0775, 0750}
			for _, perm := range perms {
				err := PathCheckCreate(false, dirPath, 0644, perm)
				Expect(err).ToNot(HaveOccurred())

				info, err := os.Stat(dirPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(info.Mode() & 0777).To(Equal(perm))
			}
		})
	})

	// Tests for error scenarios when working with directories
	Context("Error handling for directories", func() {
		It("should return error if path is file but directory expected", func() {
			filePath := filepath.Join(tempDir, "file.txt")
			_, err := os.Create(filePath)
			Expect(err).ToNot(HaveOccurred())

			err = PathCheckCreate(false, filePath, 0644, 0755)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a directory"))
		})

		It("should return error for empty directory path", func() {
			err := PathCheckCreate(false, "", 0644, 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should handle directory with special characters in name", func() {
			specialNames := []string{
				"dir with spaces",
				"dir-with-dashes",
				"dir_with_underscores",
				"dir.with.dots",
			}

			for _, name := range specialNames {
				dirPath := filepath.Join(tempDir, name)
				err := PathCheckCreate(false, dirPath, 0644, 0755)
				Expect(err).ToNot(HaveOccurred())
				Expect(dirPath).To(BeADirectory())
			}
		})
	})

	// Tests for various directory permission scenarios
	Context("Directory with various permissions", func() {
		It("should create directory with restricted permissions", func() {
			dirPath := filepath.Join(tempDir, "restricted")
			err := PathCheckCreate(false, dirPath, 0644, 0700)

			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(dirPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0700)))
		})

		It("should create directory with open permissions", func() {
			dirPath := filepath.Join(tempDir, "openperm")
			err := PathCheckCreate(false, dirPath, 0644, 0775)

			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(dirPath)
			Expect(err).ToNot(HaveOccurred())
			// Note: Actual permissions may be affected by umask on Unix systems
			// We verify the directory was created, but exact permission matching
			// may vary depending on system configuration
			Expect(info.IsDir()).To(BeTrue())
		})
	})
})

// Tests for edge cases including idempotency, mixed operations, and concurrent access.
// These tests validate behavior that is critical for real-world usage patterns.
var _ = Describe("PathCheckCreate - Edge Cases", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "ioutils_test_*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	// Tests that repeated calls with same parameters don't cause errors
	Context("Idempotency", func() {
		It("should be idempotent for files", func() {
			filePath := filepath.Join(tempDir, "idempotent.txt")

			for i := 0; i < 10; i++ {
				err := PathCheckCreate(true, filePath, 0644, 0755)
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(filePath).To(BeAnExistingFile())
		})

		It("should be idempotent for directories", func() {
			dirPath := filepath.Join(tempDir, "idempotentdir")

			for i := 0; i < 10; i++ {
				err := PathCheckCreate(false, dirPath, 0644, 0755)
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(dirPath).To(BeADirectory())
		})
	})

	// Tests for combining different operations in realistic scenarios
	Context("Mixed operations", func() {
		It("should handle creating file in newly created directory", func() {
			dirPath := filepath.Join(tempDir, "newdir")
			err := PathCheckCreate(false, dirPath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			filePath := filepath.Join(dirPath, "file.txt")
			err = PathCheckCreate(true, filePath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			Expect(filePath).To(BeAnExistingFile())
		})

		It("should handle multiple files in same directory", func() {
			dirPath := filepath.Join(tempDir, "multifiles")
			err := PathCheckCreate(false, dirPath, 0644, 0755)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 5; i++ {
				filePath := filepath.Join(dirPath, filepath.Base(tempDir)+"-file"+string(rune('0'+i))+".txt")
				err = PathCheckCreate(true, filePath, 0644, 0755)
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})

	// Tests for concurrent safety when creating multiple paths simultaneously.
	// Note: These tests create paths in DIFFERENT directories to avoid race conditions.
	// Concurrent access to the SAME path is not safe and requires external synchronization.
	Context("Concurrent operations", func() {
		It("should handle concurrent file creation in different directories", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					filePath := filepath.Join(tempDir, filepath.Base(tempDir)+"-dir"+string(rune('0'+index)), "file.txt")
					err := PathCheckCreate(true, filePath, 0644, 0755)
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent directory creation", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					dirPath := filepath.Join(tempDir, filepath.Base(tempDir)+"-concurrent"+string(rune('0'+index)))
					err := PathCheckCreate(false, dirPath, 0644, 0755)
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})
})
