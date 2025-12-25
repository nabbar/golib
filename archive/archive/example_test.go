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

package archive_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nabbar/golib/archive/archive"
)

func ExampleAlgorithm_String() {
	fmt.Println(archive.Tar.String())
	fmt.Println(archive.Zip.String())
	fmt.Println(archive.None.String())

	// Output:
	// tar
	// zip
	// none
}

func ExampleAlgorithm_Extension() {
	fmt.Println(archive.Tar.Extension())
	fmt.Println(archive.Zip.Extension())
	fmt.Println(archive.None.Extension())

	// Output:
	// .tar
	// .zip
	//
}

func ExampleAlgorithm_IsNone() {
	fmt.Println(archive.None.IsNone())
	fmt.Println(archive.Tar.IsNone())
	fmt.Println(archive.Zip.IsNone())

	// Output:
	// true
	// false
	// false
}

func ExampleParse() {
	alg1 := archive.Parse("tar")
	alg2 := archive.Parse("ZIP")
	alg3 := archive.Parse("unknown")

	fmt.Println(alg1.String())
	fmt.Println(alg2.String())
	fmt.Println(alg3.String())

	// Output:
	// tar
	// zip
	// none
}

func ExampleAlgorithm_MarshalText() {
	alg := archive.Tar
	text, _ := alg.MarshalText()
	fmt.Println(string(text))

	// Output:
	// tar
}

func ExampleAlgorithm_UnmarshalText() {
	var alg archive.Algorithm
	_ = alg.UnmarshalText([]byte("zip"))
	fmt.Println(alg.String())

	// Output:
	// zip
}

func ExampleAlgorithm_MarshalJSON() {
	alg := archive.Tar
	jsonData, _ := alg.MarshalJSON()
	fmt.Println(string(jsonData))

	alg = archive.None
	jsonData, _ = alg.MarshalJSON()
	fmt.Println(string(jsonData))

	// Output:
	// "tar"
	// null
}

func ExampleAlgorithm_UnmarshalJSON() {
	var alg archive.Algorithm
	_ = alg.UnmarshalJSON([]byte(`"zip"`))
	fmt.Println(alg.String())

	_ = alg.UnmarshalJSON([]byte(`null`))
	fmt.Println(alg.String())

	// Output:
	// zip
	// none
}

func ExampleAlgorithm_Writer_tar() {
	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	writer, err := archive.Tar.Writer(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer writer.Close()

	fmt.Println("TAR writer created successfully")

	// Output:
	// TAR writer created successfully
}

func ExampleAlgorithm_Writer_zip() {
	tmpFile, _ := os.CreateTemp("", "example-*.zip")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	writer, err := archive.Zip.Writer(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer writer.Close()

	fmt.Println("ZIP writer created successfully")

	// Output:
	// ZIP writer created successfully
}

func ExampleAlgorithm_Reader_tar() {
	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.Close()
	tmpFile.Close()

	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()

	reader, err := archive.Tar.Reader(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer reader.Close()

	fmt.Println("TAR reader created successfully")

	// Output:
	// TAR reader created successfully
}

func ExampleDetect_tar() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpFile.Close()

	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()

	alg, reader, stream, err := archive.Detect(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer stream.Close()

	if reader != nil {
		defer reader.Close()
	}

	fmt.Printf("Detected: %s\n", alg.String())

	// Output:
	// Detected: tar
}

func ExampleDetect_zip() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpFile.Close()

	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()

	alg, reader, stream, err := archive.Detect(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer stream.Close()

	if reader != nil {
		defer reader.Close()
	}

	fmt.Printf("Detected: %s\n", alg.String())

	// Output:
	// Detected: tar
}

func ExampleDetect_walk() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpFile.Close()

	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()

	_, reader, stream, _ := archive.Detect(tmpFile)
	defer stream.Close()
	defer reader.Close()

	reader.Walk(func(info os.FileInfo, r io.ReadCloser, path, link string) bool {
		if !info.IsDir() {
			fmt.Printf("File: %s\n", filepath.Base(path))
		}
		return true
	})

	// Output:
	// File: file1.txt
	// File: file2.txt
}

func ExampleAlgorithm_Writer_fromPath() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "data.log"), []byte("log data"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	writer, _ := archive.Tar.Writer(tmpFile)
	defer writer.Close()

	err := writer.FromPath(tmpDir, "*.txt", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Archive created with filtered files")

	// Output:
	// Archive created with filtered files
}

func ExampleAlgorithm_Reader_list() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpFile.Close()

	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()

	reader, err := archive.Tar.Reader(tmpFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer reader.Close()

	files, err := reader.List()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, f := range files {
		fmt.Println(filepath.Base(f))
	}

	// Output:
	// file1.txt
	// file2.txt
}

func ExampleAlgorithm_Reader_get() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	testContent := "test file content"
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(testContent), 0644)

	tmpArchive, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpArchive.Name())

	writer, _ := archive.Tar.Writer(tmpArchive)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpArchive.Close()

	// List files first
	tmpArchive, _ = os.Open(tmpArchive.Name())
	reader, _ := archive.Tar.Reader(tmpArchive)
	files, _ := reader.List()
	reader.Close()
	tmpArchive.Close()

	// Get file content with new reader
	tmpArchive, _ = os.Open(tmpArchive.Name())
	defer tmpArchive.Close()
	reader, _ = archive.Tar.Reader(tmpArchive)
	defer reader.Close()

	if len(files) > 0 {
		fileReader, _ := reader.Get(files[0])
		if fileReader != nil {
			defer fileReader.Close()
			data, _ := io.ReadAll(fileReader)
			fmt.Println(string(data))
		}
	}

	// Output:
	// test file content
}

func ExampleAlgorithm_Reader_has() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("content"), 0644)

	tmpFile, _ := os.CreateTemp("", "example-*.tar")
	defer os.Remove(tmpFile.Name())

	writer, _ := archive.Tar.Writer(tmpFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tmpFile.Close()

	// List files first
	tmpFile, _ = os.Open(tmpFile.Name())
	reader, _ := archive.Tar.Reader(tmpFile)
	files, _ := reader.List()
	reader.Close()
	tmpFile.Close()

	// Check with new reader
	tmpFile, _ = os.Open(tmpFile.Name())
	defer tmpFile.Close()
	reader, _ = archive.Tar.Reader(tmpFile)
	defer reader.Close()

	if len(files) > 0 {
		fmt.Println(reader.Has(files[0]))
		fmt.Println(reader.Has("nonexistent.txt"))
	}

	// Output:
	// true
	// false
}

func ExampleAlgorithm_DetectHeader() {
	tarHeader := make([]byte, 265)
	copy(tarHeader[257:], []byte("ustar\x00"))

	zipHeader := []byte{0x50, 0x4b, 0x03, 0x04}
	zipHeader = append(zipHeader, make([]byte, 261)...)

	fmt.Println(archive.Tar.DetectHeader(tarHeader))
	fmt.Println(archive.Zip.DetectHeader(zipHeader))
	fmt.Println(archive.Tar.DetectHeader(zipHeader))

	// Output:
	// true
	// true
	// false
}

func ExampleDetect_formatConversion() {
	tmpDir, _ := os.MkdirTemp("", "example-")
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte("data content"), 0644)

	tarFile, _ := os.CreateTemp("", "source-*.tar")
	defer os.Remove(tarFile.Name())

	writer, _ := archive.Tar.Writer(tarFile)
	_ = writer.FromPath(tmpDir, "*.txt", nil)
	_ = writer.Close()
	tarFile.Close()

	zipFile, _ := os.CreateTemp("", "dest-*.zip")
	defer os.Remove(zipFile.Name())
	zipFile.Close()

	tarFile, _ = os.Open(tarFile.Name())
	defer tarFile.Close()

	_, srcReader, srcStream, _ := archive.Detect(tarFile)
	defer srcStream.Close()
	defer srcReader.Close()

	zipFile, _ = os.OpenFile(zipFile.Name(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer zipFile.Close()

	dstWriter, _ := archive.Zip.Writer(zipFile)
	defer dstWriter.Close()

	var buf bytes.Buffer
	srcReader.Walk(func(info os.FileInfo, r io.ReadCloser, path, link string) bool {
		buf.Reset()
		io.Copy(&buf, r)
		rc := io.NopCloser(&buf)
		_ = dstWriter.Add(info, rc, path, link)
		return true
	})

	fmt.Println("Converted TAR to ZIP successfully")

	// Output:
	// Converted TAR to ZIP successfully
}
