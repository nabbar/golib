# Package Archive
This package will try to do all uncompress/unarchive and extract one file or all file.
This package will expose 2 functions :
- ExtractFile : for one file extracted
- ExtractAll : to extract all file 

## Example of implementation

### Example of one file extracted

To use `ExtractFile` function, you will need this parameters :
- `src` : is the source file into a `ioutils.FileProgress` struct to expose an `os.File` pointer with interface `io.WriteTo`, `io.ReadFrom`, `io.ReaderAt` and progress capabilities
- `dst` : is the source file into a `ioutils.FileProgress` struct to expose an `os.File` pointer with interface `io.WriteTo`, `io.ReadFrom`, `io.ReaderAt` and progress capabilities
- `filenameContain` : is a `string` to search in the file name to find it and extract it. This string will be search into the `strings.Contains` function
- `filenameRegex` : is a regex pattern `string` to search in the file name any match and extract it. This string will be search into the `regexp.MatchString` function

You can implement this function as it. This example is available in [`test/test-archive`](../test/test-archive/main.go) folder.
```go
 import (
	"io"
	"io/ioutil"

	"github.com/nabbar/golib/archive"
	"github.com/nabbar/golib/ioutils"
)

const fileName = "fullpath to my archive file"

func main() {
	var (
		src ioutils.FileProgress
		dst ioutils.FileProgress
		err errors.Error
	)

	// register closing function in output function callback
  defer func() {
		if src != nil {
			_ = src.Close()
		}
		if dst != nil {
			_ = dst.Close()
		}
	}()

	// open archive with a ioutils NewFileProgress function
  if src, err = ioutils.NewFileProgressPathOpen(fileName); err != nil {
		panic(err)
	}

	// open a destination with a ioutils NewFileProgress function, as a temporary file
	if dst, err = ioutils.NewFileProgressTemp(); err != nil {
		panic(err)
	}

  // call the extract file function
	if err = archive.ExtractFile(tmp, rio, "path/to/my/file/into/archive", "archive name regex"); err != nil {
		panic(err)
	}
  
}
```


### Example of all files extracted

To use `ExtractAll` function, you will need this parameters :
- `src` : is the source file into a `ioutils.FileProgress` struct to expose an `os.File` pointer with interface `io.WriteTo`, `io.ReadFrom`, `io.ReaderAt` and progress capabilities
- `originalName` : is a `string` to define the originalName of the archive. This params is used to create a unique file created into the outputPath if the archive is not an archive or just compressed with a not catalogued compress type like gzip or bzip2.
- `outputPath` : is a `string` to precise the destination output directory (full path). All extracted file will be extracted with this directory as base of path.
- `defaultDirPerm` : is a `os.FileMode` to precise the permission of directory. This parameters is usefull if the output directory is not existing.

You can implement this function as it. This example is available in [`test/test-archive-all`](../test/test-archive-all/main.go) folder.
```go
 import (
	"io"
	"io/ioutil"

	"github.com/nabbar/golib/archive"
	"github.com/nabbar/golib/ioutils"
)

const fileName = "fullpath to my archive file"

func main() {
	var (
		src ioutils.FileProgress
		tmp ioutils.FileProgress
		out string
		err error
	)
 
  // open archive with a ioutils NewFileProgress function 
	if src, err = ioutils.NewFileProgressPathOpen(fileName); err != nil {
		panic(err)
	}

  // create an new temporary file to use his name as output path
	if tmp, err = ioutils.NewFileProgressTemp(); err != nil {
		panic(err)
	} else {
    // get the filename of the temporary file
		out = tmp.FilePath()
    
    // close the temporary file will call the delete temporary file
		_ = tmp.Close()
	}
  
  if err = archive.ExtractAll(src, path.Base(src.FilePath()), out, 0775); err != nil {
		panic(err)
	}
  
}
```
