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
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/nabbar/golib/archive/archive/tar"
)

// ExampleNewReader demonstrates creating a tar archive reader.
// This is the simplest usage pattern for reading tar archives.
func ExampleNewReader() {
	// Create a simple tar archive in memory
	var buf bytes.Buffer
	// ... (archive content would be written here)

	reader, err := tar.NewReader(io.NopCloser(&buf))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Reader created: %T\n", reader)
	// Output:
	// Reader created: *tar.rdr
}

// ExampleNewWriter demonstrates creating a tar archive writer.
// This is the simplest usage pattern for writing tar archives.
func ExampleNewWriter() {
	var buf bytes.Buffer

	writer, err := tar.NewWriter(&nopWriteCloser{&buf})
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	fmt.Printf("Writer created: %T\n", writer)
	// Output:
	// Writer created: *tar.wrt
}

// Example_listFiles demonstrates listing all files in a tar archive.
func Example_listFiles() {
	// Create a test archive with some files
	buf := createTestArchive(map[string]string{
		"file1.txt":     "content 1",
		"file2.txt":     "content 2",
		"dir/file3.txt": "content 3",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	files, err := reader.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
	// Output:
	// file1.txt
	// file2.txt
	// dir/file3.txt
}

// Example_getFileInfo demonstrates retrieving file metadata from an archive.
func Example_getFileInfo() {
	buf := createTestArchive(map[string]string{
		"test.txt": "hello world",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	info, err := reader.Info("test.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", info.Name())
	fmt.Printf("Size: %d bytes\n", info.Size())
	fmt.Printf("IsDir: %v\n", info.IsDir())
	// Output:
	// Name: test.txt
	// Size: 11 bytes
	// IsDir: false
}

// Example_extractFile demonstrates extracting a specific file from an archive.
func Example_extractFile() {
	buf := createTestArchive(map[string]string{
		"message.txt": "Hello from tar archive",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	rc, err := reader.Get("message.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output:
	// Hello from tar archive
}

// Example_checkFileExists demonstrates checking if a file exists in an archive.
func Example_checkFileExists() {
	buf := createTestArchive(map[string]string{
		"config.json": "{}",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	if reader.Has("config.json") {
		fmt.Println("config.json found")
	}

	if !reader.Has("missing.txt") {
		fmt.Println("missing.txt not found")
	}
	// Output:
	// config.json found
	// missing.txt not found
}

// Example_walkArchive demonstrates iterating through all files in an archive.
func Example_walkArchive() {
	buf := createTestArchive(map[string]string{
		"doc1.txt":  "First document",
		"doc2.txt":  "Second document",
		"readme.md": "README",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	count := 0
	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
		count++
		return true // Continue to next file
	})
	fmt.Printf("Walked %d files\n", count)
	// Output:
	// Walked 3 files
}

// Example_walkWithFilter demonstrates processing specific files during walk.
func Example_walkWithFilter() {
	buf := createTestArchive(map[string]string{
		"code.go":    "package main",
		"test.go":    "package main_test",
		"readme.txt": "Documentation",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	fmt.Println("Go files:")
	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
		if strings.HasSuffix(path, ".go") {
			fmt.Printf("  %s\n", path)
		}
		return true
	})
	// Output:
	// Go files:
	//   code.go
	//   test.go
}

// Example_createArchive demonstrates creating a simple tar archive.
func Example_createArchive() {
	var buf bytes.Buffer

	writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
	defer writer.Close()

	// Add a file to the archive
	content := strings.NewReader("file content")
	info := &testFileInfo{name: "example.txt", size: 12, mode: 0644, modTime: time.Now()}

	err := writer.Add(info, io.NopCloser(content), "example.txt", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Archive created successfully")
	fmt.Printf("Archive size: %d bytes\n", buf.Len())
	// Output:
	// Archive created successfully
	// Archive size: 524 bytes
}

// Example_addMultipleFiles demonstrates adding multiple files to an archive.
func Example_addMultipleFiles() {
	var buf bytes.Buffer

	writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
	defer writer.Close()

	files := map[string]string{
		"file1.txt": "Content of file 1",
		"file2.txt": "Content of file 2",
		"file3.txt": "Content of file 3",
	}

	for name, content := range files {
		rc := io.NopCloser(strings.NewReader(content))
		info := &testFileInfo{name: name, size: int64(len(content)), mode: 0644, modTime: time.Now()}
		writer.Add(info, rc, name, "")
	}

	fmt.Printf("Added %d files to archive\n", len(files))
	// Output:
	// Added 3 files to archive
}

// Example_handleMissingFile demonstrates proper error handling for missing files.
func Example_handleMissingFile() {
	buf := createTestArchive(map[string]string{
		"exists.txt": "I exist",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	_, err := reader.Get("missing.txt")
	if err == fs.ErrNotExist {
		fmt.Println("File not found in archive")
	}
	// Output:
	// File not found in archive
}

// Example_stopWalkEarly demonstrates stopping the walk operation early.
func Example_stopWalkEarly() {
	buf := createTestArchive(map[string]string{
		"file1.txt": "one",
		"file2.txt": "two",
		"file3.txt": "three",
		"file4.txt": "four",
	})

	reader, _ := tar.NewReader(io.NopCloser(buf))
	defer reader.Close()

	count := 0
	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
		count++
		return count < 2 // Stop after 2 files
	})

	fmt.Printf("Processed %d files (stopped early)\n", count)
	// Output:
	// Processed 2 files (stopped early)
}

// Example_readAndWrite demonstrates a complete read-write cycle.
func Example_readAndWrite() {
	// Create an archive
	var buf bytes.Buffer
	writer, _ := tar.NewWriter(&nopWriteCloser{&buf})

	content := strings.NewReader("archived data")
	info := &testFileInfo{name: "data.txt", size: 13, mode: 0644, modTime: time.Now()}
	writer.Add(info, io.NopCloser(content), "data.txt", "")
	writer.Close()

	// Read it back
	reader, _ := tar.NewReader(io.NopCloser(&buf))
	defer reader.Close()

	rc, _ := reader.Get("data.txt")
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output:
	// archived data
}

// Helper types and functions for examples
// (Shared helpers are in helper_test.go)
