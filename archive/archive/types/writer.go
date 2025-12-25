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

package types

import (
	"io"
	"io/fs"
)

// ReplaceName is a callback function type used by FromPath to transform file paths before adding them to the archive.
//
// This function receives the original source path and returns the path that should be used
// in the archive. This is useful for:
//   - Reorganizing directory structure in the archive
//   - Adding prefixes or suffixes to file names
//   - Flattening directory hierarchies
//   - Creating custom archive layouts
//
// Parameters:
//   - string: The original source path from the filesystem
//
// Returns:
//   - string: The transformed path to use in the archive
//
// Example:
//
//	// Add "backup/" prefix to all files
//	replaceFn := func(source string) string {
//	    return "backup/" + filepath.Base(source)
//	}
//	writer.FromPath("/data", "*.txt", replaceFn)
type ReplaceName func(string) string

// Writer defines the interface for creating archives and adding files to them.
//
// Implementations of this interface provide archive creation capabilities regardless
// of the underlying format (ZIP, TAR, BZIP2, etc.). All implementations must:
//   - Handle both regular files and directories
//   - Support compression if the format allows it
//   - Flush all data in Close()
//   - Handle nil readers for directory entries
//   - Apply path transformations correctly
//
// The interface embeds io.Closer, requiring implementations to provide a Close method
// that finalizes the archive and releases all resources.
//
// Thread Safety:
//
// Implementations should document whether they are safe for concurrent use.
// Typically, archive writers are NOT thread-safe and should be used from a single
// goroutine. Concurrent calls to Add() may result in corrupted archives.
//
// Example:
//
//	writer, err := someformat.NewWriter(file)
//	if err != nil {
//	    return err
//	}
//	defer writer.Close()
//
//	// Add single file
//	info, _ := os.Stat("data.txt")
//	f, _ := os.Open("data.txt")
//	defer f.Close()
//	writer.Add(info, f, "", "")
//
//	// Add directory recursively
//	writer.FromPath("/source", "*.go", nil)
type Writer interface {
	io.Closer

	// Add adds a single file to the archive with the given metadata and content.
	//
	// The method reads the content from the provided io.ReadCloser and writes it to the archive
	// using the metadata from fs.FileInfo. The caller should close the reader after calling Add,
	// or the implementation may close it automatically.
	//
	// If the reader is nil, the implementation should create a directory entry (if supported)
	// or skip the entry. This is useful for adding empty directories to the archive.
	//
	// Parameters:
	//   - fs.FileInfo: File metadata including name, size, permissions, and modification time.
	//     The Name() from this info is used as the default path in the archive unless forcePath is provided.
	//   - io.ReadCloser: Stream to read the file content. May be nil for directory entries.
	//     Some implementations automatically close this stream, while others expect the caller to close it.
	//   - string: forcePath - if non-empty, overrides the file name from fs.FileInfo.
	//     Use this to place the file at a specific path in the archive (e.g., "config/settings.json").
	//   - string: linkTarget - if non-empty, creates a symbolic link entry pointing to this target.
	//     Only relevant for archive formats that support symlinks (e.g., TAR). Ignored for other formats.
	//
	// Returns:
	//   - error: An error if the file cannot be added (e.g., write failure, invalid metadata).
	//     Returns nil on success.
	//
	// Example:
	//
	//	info, err := os.Stat("source.txt")
	//	if err != nil {
	//	    return err
	//	}
	//
	//	file, err := os.Open("source.txt")
	//	if err != nil {
	//	    return err
	//	}
	//	defer file.Close()
	//
	//	// Add file with original name
	//	err = writer.Add(info, file, "", "")
	//	if err != nil {
	//	    return err
	//	}
	//
	//	// Or add with custom path
	//	file2, _ := os.Open("local.txt")
	//	defer file2.Close()
	//	info2, _ := os.Stat("local.txt")
	//	err = writer.Add(info2, file2, "archive/renamed.txt", "")
	Add(fs.FileInfo, io.ReadCloser, string, string) error

	// FromPath recursively adds files from a filesystem path to the archive.
	//
	// This method walks the directory tree starting from the source path and adds all
	// files that match the filter pattern. The filter uses glob syntax (e.g., "*.txt", "data_*.json").
	// If the filter is empty, all files are added.
	//
	// For each file added, the ReplaceName function (if provided) is called to transform
	// the file path before adding it to the archive. This allows reorganizing the archive
	// structure without modifying the source filesystem.
	//
	// The method typically:
	//   - Walks the directory tree recursively
	//   - Applies the filter pattern to file names
	//   - Calls ReplaceName to transform paths (if provided)
	//   - Opens each file and calls Add() with the transformed path
	//   - Handles errors for individual files (may continue or stop depending on implementation)
	//
	// Parameters:
	//   - string: source - the filesystem path to add. Can be a file or directory.
	//     If it's a file, only that file is added (subject to filter).
	//     If it's a directory, all matching files in the tree are added recursively.
	//   - string: filter - glob pattern for filtering files (e.g., "*.txt", "data_*", "**/*.go").
	//     Uses filepath.Match for pattern matching. If empty, defaults to "*" (all files).
	//     Directory traversal is not affected by the filter - it only filters files.
	//   - ReplaceName: fn - optional function to transform file paths before adding to archive.
	//     If nil, files are added with their original paths relative to source.
	//     The function receives the source path and returns the desired archive path.
	//
	// Returns:
	//   - error: An error if the operation fails (e.g., cannot read directory, cannot add files).
	//     The behavior for partial failures varies by implementation - some may continue
	//     adding other files, while others may stop on the first error.
	//
	// Example:
	//
	//	// Add all .txt files from a directory
	//	err := writer.FromPath("/data/docs", "*.txt", nil)
	//	if err != nil {
	//	    return err
	//	}
	//
	//	// Add with path transformation
	//	err = writer.FromPath("/source", "*.go", func(source string) string {
	//	    // Flatten directory structure
	//	    return "src/" + filepath.Base(source)
	//	})
	//
	//	// Add all files
	//	err = writer.FromPath("/backup", "", func(source string) string {
	//	    // Add timestamp prefix
	//	    return time.Now().Format("2006-01-02") + "/" + source
	//	})
	FromPath(string, string, ReplaceName) error
}
