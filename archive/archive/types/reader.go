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

// FuncExtract is a callback function type used by the Walk method to process each file in an archive.
//
// The function receives file information and content, and returns a boolean to control iteration:
//   - Return true to continue walking through the archive
//   - Return false to stop the iteration immediately
//
// Parameters:
//   - fs.FileInfo: File metadata including name, size, permissions, and modification time
//   - io.ReadCloser: Stream to read the file content. May be nil if the file cannot be opened.
//     The caller is responsible for closing this stream if it is not nil.
//   - string: The path of the file within the archive (relative path)
//   - string: The symlink target if the file is a symbolic link, empty string otherwise
//
// Returns:
//   - bool: true to continue iterating, false to stop
//
// Example:
//
//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
//	    if r != nil {
//	        defer r.Close()
//	    }
//	    fmt.Printf("Processing: %s (size: %d)\n", path, info.Size())
//	    return true
//	})
type FuncExtract func(fs.FileInfo, io.ReadCloser, string, string) bool

// Reader defines the interface for reading and extracting files from archives.
//
// Implementations of this interface provide access to archive contents regardless
// of the underlying format (ZIP, TAR, BZIP2, etc.). All implementations must:
//   - Support concurrent reads from different files (if thread-safe)
//   - Return fs.ErrNotExist for missing files
//   - Release resources properly in Close()
//   - Handle both regular files and directories
//
// The interface embeds io.Closer, requiring implementations to provide a Close method
// that releases all resources associated with the reader.
//
// Thread Safety:
//
// Implementations should document whether they are safe for concurrent use.
// Typically, reading different files concurrently is safe, but modifying the
// reader state (e.g., seeking) should be serialized.
//
// Example:
//
//	reader, err := someformat.NewReader(file)
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
//
//	// List all files
//	files, _ := reader.List()
//	for _, path := range files {
//	    if reader.Has(path) {
//	        rc, _ := reader.Get(path)
//	        defer rc.Close()
//	        // Process file...
//	    }
//	}
type Reader interface {
	io.Closer

	// List returns a slice containing the paths of all files in the archive.
	//
	// The returned paths are relative to the archive root. Directory entries
	// may or may not be included depending on the archive format.
	//
	// Returns:
	//   - []string: A slice of file paths within the archive. Empty slice if archive is empty.
	//   - error: An error if the file list cannot be retrieved (e.g., corrupted archive).
	//
	// Example:
	//
	//	files, err := reader.List()
	//	if err != nil {
	//	    return err
	//	}
	//	for _, path := range files {
	//	    fmt.Println(path)
	//	}
	List() ([]string, error)

	// Info returns the file metadata for the specified path within the archive.
	//
	// The returned fs.FileInfo provides access to file attributes such as size,
	// permissions, modification time, and whether it's a directory.
	//
	// Parameters:
	//   - string: The path of the file within the archive (relative path).
	//
	// Returns:
	//   - fs.FileInfo: File metadata including Name(), Size(), Mode(), ModTime(), IsDir(), and Sys().
	//   - error: Returns fs.ErrNotExist if the file is not found in the archive,
	//     or another error if metadata cannot be retrieved.
	//
	// Example:
	//
	//	info, err := reader.Info("config/settings.json")
	//	if err != nil {
	//	    if errors.Is(err, fs.ErrNotExist) {
	//	        fmt.Println("File not found")
	//	    }
	//	    return err
	//	}
	//	fmt.Printf("Size: %d bytes\n", info.Size())
	Info(string) (fs.FileInfo, error)

	// Get retrieves an io.ReadCloser to read the contents of a file from the archive.
	//
	// The caller is responsible for closing the returned ReadCloser to prevent resource leaks.
	// The content is typically decompressed automatically if the archive format supports compression.
	//
	// Parameters:
	//   - string: The path of the file within the archive (relative path).
	//
	// Returns:
	//   - io.ReadCloser: A stream to read the file content. Must be closed by the caller.
	//   - error: Returns fs.ErrNotExist if the file is not found in the archive,
	//     or another error if the file cannot be opened or read.
	//
	// Example:
	//
	//	rc, err := reader.Get("data.txt")
	//	if err != nil {
	//	    return err
	//	}
	//	defer rc.Close()
	//
	//	content, err := io.ReadAll(rc)
	//	if err != nil {
	//	    return err
	//	}
	//	fmt.Println(string(content))
	Get(string) (io.ReadCloser, error)

	// Has checks whether the archive contains a file with the specified path.
	//
	// This is a convenience method that should be more efficient than calling Info
	// or Get and checking for errors. Implementations should cache file lists to
	// make this operation fast.
	//
	// Parameters:
	//   - string: The path of the file within the archive (relative path).
	//
	// Returns:
	//   - bool: true if the file exists in the archive, false otherwise.
	//
	// Example:
	//
	//	if reader.Has("optional-config.json") {
	//	    rc, _ := reader.Get("optional-config.json")
	//	    defer rc.Close()
	//	    // Process optional config...
	//	}
	Has(string) bool

	// Walk iterates through all files in the archive, calling the provided function for each file.
	//
	// The iteration continues until all files have been processed or the callback function
	// returns false. If the callback returns false, the iteration stops immediately.
	//
	// For each file, Walk opens the file and passes its information to the callback.
	// If a file cannot be opened, the callback may receive a nil io.ReadCloser.
	// The callback should handle this case gracefully.
	//
	// Walk does not return errors directly. Implementations should decide whether to
	// continue or stop iteration when errors occur. Errors can be logged or passed
	// through the callback mechanism.
	//
	// Parameters:
	//   - FuncExtract: Callback function invoked for each file in the archive.
	//     The callback receives:
	//     - fs.FileInfo: File metadata
	//     - io.ReadCloser: Stream to read file content (may be nil on error)
	//     - string: File path within the archive
	//     - string: Symlink target (empty for regular files)
	//     Return true to continue iteration, false to stop.
	//
	// Example:
	//
	//	reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
	//	    if r != nil {
	//	        defer r.Close()
	//	        // Process file content...
	//	    }
	//	    fmt.Printf("File: %s, Size: %d\n", path, info.Size())
	//	    return true // Continue
	//	})
	Walk(FuncExtract)
}
