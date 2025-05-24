# Golib Artifact

**Golang utilities for artifact version management retrieve / download**

---

## Overview

The `artifact` package provides tools to list, search, retrieve or downlaod version of artifacts such as files, binaries... 

---

## Installation

Add the dependency to your Go project:

```shell
go get github.com/nabbar/golib/artifact
```

Or in your `go.mod`:

```go
require github.com/nabbar/golib/artifact vX.Y.Z
```

---

## Features

- List and search artifacts
- Retrieve artifacts by name and version
- Retrieve minor / major versions of artifacts
- Retrieve latest version of artifacts
- Retrieve download URL of artifacts
- Support for local and remote storage backends

---

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/nabbar/golib/artifact"
	"github.com/hashicorp/go-version"
	"github.com/nabbar/golib/artifact/github"
)

func main() {
	var (
		err error
		art artifact.Client
		lst version.Collection
    )
	
	// Create a new artifact manager
	art, err = github.NewGithub(context.Background(), &http.Client{}, "repo")
	if err != nil {
		fmt.Println("Error creating artifact manager:", err)
		return
	}

	// List Versions
	lst, err = art.ListReleases()
	if err != nil {
		fmt.Println("Error retrieving artifact:", err)
		return
	}
	
	for _, ver := range lst {
        fmt.Printf("Version: %s\n", ver.String())

		link, e := art.GetArtifact("linux", "", ver)

		if e != nil {
			fmt.Println("No linux version found")
			continue
		}
		
		fmt.Printf("Donwload Linux: %s\n", link)
    }
}
```

---

## Error Handling

All functions return an `error` value. Always check and handle errors when storing, retrieving, or verifying artifacts.

---

## Contributing

Contributions are welcome! Please submit issues or pull requests on [GitHub](https://github.com/nabbar/golib).

---

## License

MIT © Nicolas JUHEL