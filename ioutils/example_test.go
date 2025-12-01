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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nabbar/golib/ioutils"
)

// ExamplePathCheckCreate_basicDirectory demonstrates the simplest use case:
// creating a single directory with standard permissions.
func ExamplePathCheckCreate_basicDirectory() {
	tmpDir := os.TempDir()
	dirPath := filepath.Join(tmpDir, "example_dir")

	// Create a directory with standard permissions
	// isFile=false means we want a directory
	// 0644 is file permission (not used for directories)
	// 0755 is directory permission (rwxr-xr-x)
	err := ioutils.PathCheckCreate(false, dirPath, 0644, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Clean up
	defer os.RemoveAll(dirPath)

	// Verify it exists
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		fmt.Println("Directory created successfully")
	}

	// Output: Directory created successfully
}

// ExamplePathCheckCreate_basicFile demonstrates creating a single file
// with appropriate permissions.
func ExamplePathCheckCreate_basicFile() {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "example_file.txt")

	// Create a file with standard permissions
	// isFile=true means we want a file
	// 0644 is file permission (rw-r--r--)
	// 0755 is directory permission for parent dirs
	err := ioutils.PathCheckCreate(true, filePath, 0644, 0755)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	// Clean up
	defer os.Remove(filePath)

	// Verify it exists
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		fmt.Println("File created successfully")
	}

	// Output: File created successfully
}

// ExamplePathCheckCreate_nestedPath demonstrates creating a file with
// automatic parent directory creation.
func ExamplePathCheckCreate_nestedPath() {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "example_nested", "deep", "path", "file.log")

	// Create file in nested directories - parent dirs created automatically
	err := ioutils.PathCheckCreate(true, filePath, 0644, 0755)
	if err != nil {
		log.Fatalf("Failed to create nested file: %v", err)
	}

	// Clean up
	defer os.RemoveAll(filepath.Join(tmpDir, "example_nested"))

	// Verify the entire path exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Println("Nested file structure created successfully")
	}

	// Output: Nested file structure created successfully
}

// ExamplePathCheckCreate_permissionUpdate demonstrates how PathCheckCreate
// updates permissions if the path already exists.
func ExamplePathCheckCreate_permissionUpdate() {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "example_perm_update.txt")

	// Create file with restrictive permissions
	err := ioutils.PathCheckCreate(true, filePath, 0600, 0755)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	// Clean up
	defer os.Remove(filePath)

	// Update permissions by calling again with different perms
	err = ioutils.PathCheckCreate(true, filePath, 0644, 0755)
	if err != nil {
		log.Fatalf("Failed to update permissions: %v", err)
	}

	// Verify new permissions
	if info, err := os.Stat(filePath); err == nil {
		if info.Mode()&0777 == 0644 {
			fmt.Println("Permissions updated successfully")
		}
	}

	// Output: Permissions updated successfully
}

// ExamplePathCheckCreate_idempotent demonstrates that PathCheckCreate
// can be called multiple times safely.
func ExamplePathCheckCreate_idempotent() {
	tmpDir := os.TempDir()
	dirPath := filepath.Join(tmpDir, "example_idempotent")

	// Clean up
	defer os.RemoveAll(dirPath)

	// Call multiple times - should not error
	for i := 0; i < 3; i++ {
		err := ioutils.PathCheckCreate(false, dirPath, 0644, 0755)
		if err != nil {
			log.Fatalf("Failed on iteration %d: %v", i, err)
		}
	}

	fmt.Println("Idempotent calls successful")

	// Output: Idempotent calls successful
}

// ExamplePathCheckCreate_applicationInit demonstrates a realistic use case:
// initializing application directories during startup.
func ExamplePathCheckCreate_applicationInit() {
	tmpDir := os.TempDir()
	appRoot := filepath.Join(tmpDir, "myapp")

	// Define application directory structure
	dirs := []string{
		filepath.Join(appRoot, "data"),
		filepath.Join(appRoot, "config"),
		filepath.Join(appRoot, "logs"),
		filepath.Join(appRoot, "cache"),
		filepath.Join(appRoot, "tmp"),
	}

	// Create all directories
	for _, dir := range dirs {
		if err := ioutils.PathCheckCreate(false, dir, 0644, 0755); err != nil {
			log.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	// Create important files
	files := map[string]os.FileMode{
		filepath.Join(appRoot, "config", "app.conf"):  0640, // Restrictive config
		filepath.Join(appRoot, "logs", "app.log"):     0644, // Standard log
		filepath.Join(appRoot, "data", "state.json"):  0644, // Data file
	}

	for file, perm := range files {
		if err := ioutils.PathCheckCreate(true, file, perm, 0755); err != nil {
			log.Fatalf("Failed to create %s: %v", file, err)
		}
	}

	// Clean up
	defer os.RemoveAll(appRoot)

	// Verify structure
	allExist := true
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			allExist = false
			break
		}
	}
	for file := range files {
		if _, err := os.Stat(file); err != nil {
			allExist = false
			break
		}
	}

	if allExist {
		fmt.Println("Application structure initialized")
	}

	// Output: Application structure initialized
}

// ExamplePathCheckCreate_logRotation demonstrates using PathCheckCreate
// for log file rotation scenarios.
func ExamplePathCheckCreate_logRotation() {
	tmpDir := os.TempDir()
	logDir := filepath.Join(tmpDir, "example_logs")
	currentLog := filepath.Join(logDir, "app.log")

	// Ensure log directory exists
	if err := ioutils.PathCheckCreate(false, logDir, 0644, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Create or verify current log file
	if err := ioutils.PathCheckCreate(true, currentLog, 0644, 0755); err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	// Simulate rotation - archive old log
	archivedLog := filepath.Join(logDir, "app.log.1")
	if err := os.Rename(currentLog, archivedLog); err == nil {
		// Create new log file
		if err := ioutils.PathCheckCreate(true, currentLog, 0644, 0755); err != nil {
			log.Fatalf("Failed to create new log file: %v", err)
		}
	}

	// Clean up
	defer os.RemoveAll(logDir)

	// Verify both files exist
	if _, err := os.Stat(currentLog); err == nil {
		if _, err := os.Stat(archivedLog); err == nil {
			fmt.Println("Log rotation successful")
		}
	}

	// Output: Log rotation successful
}

// ExamplePathCheckCreate_secureConfig demonstrates creating configuration
// files with restrictive permissions for security.
func ExamplePathCheckCreate_secureConfig() {
	tmpDir := os.TempDir()
	configDir := filepath.Join(tmpDir, "example_secure_config")
	secretsFile := filepath.Join(configDir, "secrets.conf")

	// Create config directory with standard permissions
	if err := ioutils.PathCheckCreate(false, configDir, 0644, 0750); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// Create secrets file with restrictive permissions (owner read/write only)
	if err := ioutils.PathCheckCreate(true, secretsFile, 0600, 0750); err != nil {
		log.Fatalf("Failed to create secrets file: %v", err)
	}

	// Clean up
	defer os.RemoveAll(configDir)

	// Verify restrictive permissions
	if info, err := os.Stat(secretsFile); err == nil {
		if info.Mode()&0777 == 0600 {
			fmt.Println("Secure configuration file created")
		}
	}

	// Output: Secure configuration file created
}

// ExamplePathCheckCreate_errorHandling demonstrates proper error handling
// when path conflicts occur.
func ExamplePathCheckCreate_errorHandling() {
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, "example_conflict")

	// Create as directory first
	if err := ioutils.PathCheckCreate(false, path, 0644, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Clean up
	defer os.RemoveAll(path)

	// Try to create as file - should fail
	err := ioutils.PathCheckCreate(true, path, 0644, 0755)
	if err != nil {
		fmt.Println("Expected error: path exists as directory")
	}

	// Output: Expected error: path exists as directory
}

// ExamplePathCheckCreate_multipleFiles demonstrates creating multiple files
// efficiently in a batch operation.
func ExamplePathCheckCreate_multipleFiles() {
	tmpDir := os.TempDir()
	baseDir := filepath.Join(tmpDir, "example_batch")

	// Create base directory
	if err := ioutils.PathCheckCreate(false, baseDir, 0644, 0755); err != nil {
		log.Fatalf("Failed to create base directory: %v", err)
	}

	// Clean up
	defer os.RemoveAll(baseDir)

	// Create multiple files with different permissions
	files := []struct {
		path string
		perm os.FileMode
	}{
		{filepath.Join(baseDir, "public.txt"), 0644},
		{filepath.Join(baseDir, "private.txt"), 0600},
		{filepath.Join(baseDir, "executable.sh"), 0755},
		{filepath.Join(baseDir, "data", "nested.dat"), 0644},
	}

	for _, f := range files {
		if err := ioutils.PathCheckCreate(true, f.path, f.perm, 0755); err != nil {
			log.Fatalf("Failed to create %s: %v", f.path, err)
		}
	}

	// Verify all created
	allCreated := true
	for _, f := range files {
		if _, err := os.Stat(f.path); err != nil {
			allCreated = false
			break
		}
	}

	if allCreated {
		fmt.Println("Multiple files created successfully")
	}

	// Output: Multiple files created successfully
}
