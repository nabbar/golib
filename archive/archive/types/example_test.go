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
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/nabbar/golib/archive/archive/types"
)

// ExampleFuncExtract demonstrates the callback function used with Walk.
func ExampleFuncExtract() {
	// This example shows how to define a FuncExtract callback
	// that processes each file in an archive

	callback := func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
		if r != nil {
			defer r.Close()
		}

		fmt.Printf("Processing: %s\n", path)
		fmt.Printf("Size: %d bytes\n", info.Size())
		fmt.Printf("Is directory: %v\n", info.IsDir())

		if link != "" {
			fmt.Printf("Link target: %s\n", link)
		}

		// Return true to continue walking, false to stop
		return true
	}

	// Use the callback with a reader's Walk method
	_ = callback

	fmt.Println("Callback function defined")
	// Output:
	// Callback function defined
}

// ExampleReplaceName demonstrates path transformation for archive creation.
func ExampleReplaceName() {
	// This example shows various path transformation strategies

	// Example 1: Add prefix to all paths
	addPrefix := func(source string) string {
		return "backup/" + source
	}

	// Example 2: Flatten directory structure
	flatten := func(source string) string {
		return filepath.Base(source)
	}

	// Example 3: Organize by extension
	organizeByExt := func(source string) string {
		ext := filepath.Ext(source)
		if ext != "" {
			return ext[1:] + "/" + filepath.Base(source)
		}
		return "other/" + filepath.Base(source)
	}

	// Demonstrate transformations
	original := "data/docs/readme.txt"

	fmt.Printf("Original: %s\n", original)
	fmt.Printf("With prefix: %s\n", addPrefix(original))
	fmt.Printf("Flattened: %s\n", flatten(original))
	fmt.Printf("By extension: %s\n", organizeByExt(original))

	// Output:
	// Original: data/docs/readme.txt
	// With prefix: backup/data/docs/readme.txt
	// Flattened: readme.txt
	// By extension: txt/readme.txt
}

// Example_readerUsage demonstrates typical Reader interface usage patterns.
func Example_readerUsage() {
	// This example shows how to use a Reader implementation
	// Note: This is conceptual - actual implementations come from format-specific packages

	fmt.Println("Reader interface usage:")
	fmt.Println("1. Open archive with format-specific constructor")
	fmt.Println("2. Call reader.List() to get all files")
	fmt.Println("3. Use reader.Has(path) to check file existence")
	fmt.Println("4. Use reader.Info(path) to get file metadata")
	fmt.Println("5. Use reader.Get(path) to extract file content")
	fmt.Println("6. Use reader.Walk(callback) to iterate all files")
	fmt.Println("7. Call reader.Close() when done")

	// Output:
	// Reader interface usage:
	// 1. Open archive with format-specific constructor
	// 2. Call reader.List() to get all files
	// 3. Use reader.Has(path) to check file existence
	// 4. Use reader.Info(path) to get file metadata
	// 5. Use reader.Get(path) to extract file content
	// 6. Use reader.Walk(callback) to iterate all files
	// 7. Call reader.Close() when done
}

// Example_writerUsage demonstrates typical Writer interface usage patterns.
func Example_writerUsage() {
	// This example shows how to use a Writer implementation
	// Note: This is conceptual - actual implementations come from format-specific packages

	fmt.Println("Writer interface usage:")
	fmt.Println("1. Create archive with format-specific constructor")
	fmt.Println("2. Use writer.Add(info, reader, path, link) to add single files")
	fmt.Println("3. Use writer.FromPath(src, filter, fn) to add directories")
	fmt.Println("4. Call writer.Close() to finalize the archive")

	// Output:
	// Writer interface usage:
	// 1. Create archive with format-specific constructor
	// 2. Use writer.Add(info, reader, path, link) to add single files
	// 3. Use writer.FromPath(src, filter, fn) to add directories
	// 4. Call writer.Close() to finalize the archive
}

// Example_walkCallback demonstrates how to implement a Walk callback.
func Example_walkCallback() {
	// Example callback that filters and processes files

	processTextFiles := func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
		// Close reader if provided
		if r != nil {
			defer r.Close()
		}

		// Skip directories
		if info.IsDir() {
			return true
		}

		// Process only .txt files
		if filepath.Ext(path) == ".txt" {
			fmt.Printf("Found text file: %s (%d bytes)\n", path, info.Size())
		}

		// Continue walking
		return true
	}

	// This callback would be used with: reader.Walk(processTextFiles)
	_ = processTextFiles

	fmt.Println("Walk callback defined for processing text files")
	// Output:
	// Walk callback defined for processing text files
}

// Example_pathTransformation demonstrates path transformation strategies.
func Example_pathTransformation() {
	// Common path transformation patterns

	// 1. Preserve directory structure with prefix
	withPrefix := func(prefix string) types.ReplaceName {
		return func(source string) string {
			return filepath.Join(prefix, source)
		}
	}

	// 2. Flatten all files to root
	flattenAll := func(source string) string {
		return filepath.Base(source)
	}

	// 3. Remove specific prefix
	removePrefix := func(prefix string) types.ReplaceName {
		return func(source string) string {
			if len(source) > len(prefix) && source[:len(prefix)] == prefix {
				return source[len(prefix):]
			}
			return source
		}
	}

	// Demonstrate usage
	testPath := "data/logs/app.log"

	fmt.Printf("Original: %s\n", testPath)
	fmt.Printf("With prefix: %s\n", withPrefix("backup")(testPath))
	fmt.Printf("Flattened: %s\n", flattenAll(testPath))
	fmt.Printf("Remove prefix: %s\n", removePrefix("data/")(testPath))

	// Output:
	// Original: data/logs/app.log
	// With prefix: backup/data/logs/app.log
	// Flattened: app.log
	// Remove prefix: logs/app.log
}

// Example_errorHandling demonstrates error handling patterns.
func Example_errorHandling() {
	// Common error handling patterns with Reader/Writer interfaces

	fmt.Println("Error handling patterns:")
	fmt.Println()
	fmt.Println("For Reader:")
	fmt.Println("- Check fs.ErrNotExist for missing files")
	fmt.Println("- Handle nil readers in Walk callbacks")
	fmt.Println("- Always close extracted files")
	fmt.Println()
	fmt.Println("For Writer:")
	fmt.Println("- Check Add() errors for each file")
	fmt.Println("- Handle FromPath() errors")
	fmt.Println("- Always call Close() to finalize")

	// Output:
	// Error handling patterns:
	//
	// For Reader:
	// - Check fs.ErrNotExist for missing files
	// - Handle nil readers in Walk callbacks
	// - Always close extracted files
	//
	// For Writer:
	// - Check Add() errors for each file
	// - Handle FromPath() errors
	// - Always call Close() to finalize
}

// Example_filterPatterns demonstrates common filter patterns for FromPath.
func Example_filterPatterns() {
	// Common filter patterns for selecting files

	fmt.Println("Filter pattern examples:")
	fmt.Println()
	fmt.Println("*.txt       - All .txt files")
	fmt.Println("*.go        - All Go source files")
	fmt.Println("data_*      - Files starting with 'data_'")
	fmt.Println("*_test.go   - All Go test files")
	fmt.Println("README*     - Files starting with 'README'")
	fmt.Println("*           - All files (default)")

	// Output:
	// Filter pattern examples:
	//
	// *.txt       - All .txt files
	// *.go        - All Go source files
	// data_*      - Files starting with 'data_'
	// *_test.go   - All Go test files
	// README*     - Files starting with 'README'
	// *           - All files (default)
}
