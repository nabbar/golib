# Golib Archive

**Golang utilities for archive and compression management**

---

## Overview

This package provides tools to manipulate archive files (ZIP, TAR, etc.), compress/decompress streams (GZIP, BZIP2, etc.), and helpers to simplify common file operations in Go.

---

## Installation

Add the dependency to your Go project:

```shell
go get github.com/nabbar/golib/archive
```

Or in your `go.mod`:

```go
require github.com/nabbar/golib/archive vX.Y.Z
```

---

## Package Structure

- **archive**: Create, extract, and list archives (ZIP, TAR, etc.), see [Subpackage `archive`](#archive-subpackage) for details.
- **compress**: Compress/decompress files or streams (GZIP, BZIP2, etc.), see [Subpackage `compress`](#compress-subpackage) for details.
- **helper**: Utility functions for file and stream manipulation, including a unified interface for compression and decompression, see [Subpackage `helper`](#helper-subpackage) for details.

---

## Quick Start

### Extract any type of Archive / Compressed File

```go
import "github.com/nabbar/golib/archive"

in, err := os.Open("archive.zip")

if err != nil {
    panic(err)
}

defer in.Close()

err := archive.ExtractAll("archive.zip", "archive", "destination_folder")

if err != nil {
    panic(err)
}

```

### Detect Archive Type

```go
package main 

import (
	"fmt"
    "os"
	
	libarc "github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"
)

func main() {
    var (
        in *os.File
		alg arcarc.Algorithm
        arc arctps.Reader
        lst []string
        err   error
    )
  
	in, err = os.Open("archive.zip")

	if err != nil {
		panic(err)
	}

	defer in.Close()

	alg, arc, _, err = libarc.DetectArchive(in)

	if err != nil {
		panic(err)
	}
	
	defer arc.Close()
	
	fmt.Println("Archive type detected:", alg.String())

	lst, err = arc.List()

	if err != nil {
		panic(err)
	}

	for _, f := range lst {
		fmt.Println("File in archive:", f)
	}
}

```

### Detect Compression Type

```go
package main 

import (
	"fmt"
    "io"
    "os"
	
	libarc "github.com/nabbar/golib/archive"
    arccmp "github.com/nabbar/golib/archive/compress"
)

func main() {
    var (
		in *os.File
        alg arccmp.Algorithm
        rdr io.ReadCloser
        err   error
    )

    in, err = os.Open("archive.gz")

	if err != nil {
		panic(err)
	}

	defer in.Close()

	alg, rdr, err = libarc.DetectCompression(in)

	if err != nil {
		panic(err)
	}
	
	defer rdr.Close()
	
	fmt.Println("Compression type detected:", alg.String())

    // Read the decompressed data
	_, err = io.Copy(io.Discard, rdr)

	if err != nil {
		panic(err)
	}
}

```

---

# `archive` Subpackage

The `archive` subpackage provides tools to manage archive files (such as ZIP and TAR), including creation, extraction, listing, and detection of archive types. It is designed to work with multiple files and directories, and supports both reading and writing operations.

## Features

- Create archives (ZIP, TAR) from files or directories
- Extract archives to a destination folder
- List the contents of an archive
- Detect archive type from a stream
- Read and write archives using unified interfaces
- Support for streaming and random access (when possible)

---

## Main Types

- **Algorithm**: Enum for supported archive formats (`Tar`, `Zip`, `None`)
- **Reader/Writer**: Interfaces for reading and writing archive entries

---

## API Overview

### Detect Archive Type

Detect the archive algorithm from an `io.ReadCloser` and get a compatible reader.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/archive"
)

file, _ := os.Open("archive.tar")
alg, reader, closer, err := archive.Detect(file)
if err != nil {
    // handle error
}
defer closer.Close()
fmt.Println("Archive type:", alg.String())
```

### Create an Archive

Create a new archive (ZIP or TAR) and add files/directories.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/archive"
)

out, _ := os.Create("myarchive.zip")
writer, err := archive.Zip.Writer(out)
if err != nil {
    // handle error
}
// Use writer to add files/directories (see Writer interface)
```

### Extract an Archive

Extract all files from an archive to a destination.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/archive"
)

in, _ := os.Open("myarchive.tar")
alg, reader, closer, err := archive.Detect(in)
if err != nil {
    // handle error
}
defer closer.Close()

// Use reader to walk through files and extract them
```

### List Archive Contents

List all files in an archive.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/archive"
)

in, _ := os.Open("archive.zip")
alg, reader, closer, err := archive.Detect(in)
if err != nil {
    // handle error
}
defer closer.Close()

files, err := reader.List()
if err != nil {
    // handle error
}
for _, f := range files {
    fmt.Println(f)
}
```

---

## Error Handling

All functions return an `error` value. Always check and handle errors when working with archives.

---

## Notes

- The archive type is detected by reading the file header.
- For ZIP archives, random access is required (`io.ReaderAt`).
- For TAR archives, streaming is supported.
- Use the `Walk` method (if available) to iterate over archive entries efficiently.

---

For more details, refer to the GoDoc or the source code in `archive/archive`.

---

# `compress` Subpackage

The `compress` subpackage provides utilities to handle compression and decompression of single files or data streams using various algorithms. It supports GZIP, BZIP2, LZ4, XZ, and offers a unified interface for detection, reading, and writing compressed data.

## Features

- Detect compression algorithm from a stream
- Compress and decompress data using GZIP, BZIP2, LZ4, XZ
- Unified `Reader` and `Writer` interfaces for all supported algorithms
- Simple API for marshaling/unmarshaling algorithm types
- Support for both streaming and random access (when possible)

---

## Supported Algorithms

- `None` (no compression)
- `Gzip`
- `Bzip2`
- `LZ4`
- `XZ`

---

## Main Types

- **Algorithm**: Enum for supported compression formats
- **Reader/Writer**: Interfaces for reading from and writing to compressed streams

---

## API Overview

### Detect Compression Algorithm

Detect the compression algorithm from an `io.Reader` and get a compatible decompressor.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/compress"
)

file, _ := os.Open("file.txt.gz")
alg, reader, err := compress.Detect(file)
if err != nil {
    // handle error
}
defer reader.Close()
fmt.Println("Compression type:", alg.String())
```

### Compress Data

Create a compressed file using a specific algorithm.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/compress"
)

in, _ := os.Open("file.txt")
out, _ := os.Create("file.txt.gz")
writer, err := compress.Gzip.Writer(out)
if err != nil {
    // handle error
}
defer writer.Close()
_, err = io.Copy(writer, in)
```

### Decompress Data

Decompress a file using the detected algorithm.

```go
import (
    "os"
    "github.com/nabbar/golib/archive/compress"
)

in, _ := os.Open("file.txt.gz")
alg, reader, err := compress.Detect(in)
if err != nil {
    // handle error
}
defer reader.Close()
out, _ := os.Create("file.txt")
_, err = io.Copy(out, reader)
```

### Parse and Marshal Algorithm

Convert between string and `Algorithm` type.

```go
alg := compress.Parse("gzip")
fmt.Println(alg.String()) // Output: gzip
```

---

## Error Handling

All functions return an `error` value. Always check and handle errors when working with compression streams.

---

## Notes

- The algorithm is detected by reading the file header.
- For writing, use the `Writer` method of the chosen algorithm.
- For reading, use the `Reader` method or `Detect` for auto-detection.
- The package supports both marshaling to/from text and JSON for the `Algorithm` type.

---

For more details, refer to the GoDoc or the source code in `archive/compress`.

---

# `helper` Subpackage

The `helper` subpackage provides advanced utilities to simplify compression and decompression workflows for files and streams. It offers a unified interface to handle both compression and decompression using various algorithms, and can be used as a drop-in `io.ReadWriteCloser` for flexible data processing.

## Features

- Unified `Helper` interface for compressing and decompressing data
- Supports all algorithms from the `compress` subpackage (GZIP, BZIP2, LZ4, XZ, None)
- Can be used as a reader or writer, depending on the operation
- Handles both file and stream sources/destinations
- Thread-safe buffer management for streaming operations
- Error handling for invalid sources or operations

---

## Main Types

- **Helper**: Interface implementing `io.ReadWriteCloser` for compression/decompression
- **Operation**: Enum (`Compress`, `Decompress`) to specify the desired operation

---

## API Overview

### Create a Helper

Create a new helper for compression or decompression, using a specific algorithm and source (reader or writer).

```go
package main

import (
    "io"
    "os"
    "strings"

    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
	// For compression (writing compressed data)
	out, err := os.Create("file.txt.gz")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	h, err := helper.NewWriter(compress.Gzip, helper.Compress, out)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	_, err = io.Copy(h, strings.NewReader("data to compress"))
	if err != nil {
		panic(err)
	}

	// For decompression (reading decompressed data)
	in, err := os.Open("file.txt.gz")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	h, err = helper.NewReader(compress.Gzip, helper.Decompress, in)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	_, err = io.Copy(os.Stdout, h)
	if err != nil {
		panic(err)
	}

	// For compression (writing compressed data)
	out, err := os.Create("file.txt.gz")

	if err != nil {
		panic(err)
	}

	defer out.Close()

	h, err := helper.NewWriter(compress.Gzip, helper.Compress, out)

	if err != nil {
		panic(err)
	}

	defer h.Close()

	_, err = io.copy(h, strings.NewReader("data to compress"))
	if err != nil {
		panic(err)
	}

	// For decompression (reading decompressed data)
	in, err := os.Open("file.txt.gz")

	if err != nil {
		panic(err)
	}

	defer in.Close()

	h, err = helper.NewReader(compress.Gzip, helper.Decompress, in)

	if err != nil {
		panic(err)
	}

	defer h.Close()

	_, err = io.Copy(os.Stdout, h)

	if err != nil {
		panic(err)
	}

}
```

### Use as Reader or Writer

You can use the helper as a standard `io.Reader`, `io.Writer`, or `io.ReadWriteCloser` depending on the operation.

```go
// Compress data from a file to another file
in, _ := os.Open("file.txt")
out, _ := os.Create("file.txt.gz")
h, _ := helper.New(compress.Gzip, helper.Compress, out)
defer h.Close()
io.Copy(h, in)

// Decompress data from a file to another file
in, _ := os.Open("file.txt.gz")
out, _ := os.Create("file.txt")
h, _ := helper.New(compress.Gzip, helper.Decompress, in)
defer h.Close()
io.Copy(out, h)
```

### Error Handling

All helper constructors and methods return errors. Typical errors include invalid source, closed resource, or invalid operation.

---

## Notes

- The `Helper` interface adapts to the source type: use a `Reader` for decompression, a `Writer` for compression.
- The `New` function auto-detects the source type and operation.
- Thread-safe buffers are used internally for streaming and chunked operations.
- For advanced use, you can directly use `NewReader` or `NewWriter` to specify the direction.

---

For more details, refer to the GoDoc or the source code in `archive/helper`.