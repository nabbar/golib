# `ftpclient` Package

The `ftpclient` package provides a high-level, thread-safe FTP client abstraction for Go, built on top of [jlaffaye/ftp](https://github.com/jlaffaye/ftp). It simplifies FTP operations, connection management, error handling, and configuration, including TLS support and advanced options.

## Features

- Thread-safe FTP client with connection pooling
- Rich configuration struct with validation (hostname, login, password, timeouts, TLS, etc.)
- Support for explicit/implicit TLS, time zones, and FTP protocol options (UTF8, EPSV, MLSD, MDTM)
- Full set of FTP commands: file/directory listing, upload, download, rename, delete, recursive removal, and more
- Custom error codes for precise diagnostics
- Context and TLS configuration registration
- Automatic connection checking and recovery

---

## Main Types & Functions

### `Config` Struct

Defines all connection and protocol options:

- `Hostname`: FTP server address (required, RFC1123)
- `Login` / `Password`: Credentials for authentication
- `ConnTimeout`: Timeout for all operations
- `TimeZone`: Force a specific time zone for the connection
- `DisableUTF8`, `DisableEPSV`, `DisableMLSD`, `EnableMDTM`: Protocol feature toggles
- `ForceTLS`: Require TLS for the connection
- `TLS`: TLS configuration (see `github.com/nabbar/golib/certificates`)
- Methods to register context and default TLS providers

#### Example

```go
import (
    "github.com/nabbar/golib/ftpclient"
    "time"
)

cfg := &ftpclient.Config{
    Hostname:    "ftp.example.com:21",
    Login:       "user",
    Password:    "pass",
    ConnTimeout: 10 * time.Second,
    ForceTLS:    true,
    // ... other options
}
if err := cfg.Validate(); err != nil {
    // handle config error
}
```

---

### `FTPClient` Interface

Main interface for FTP operations:

- `Connect() error`: Establish connection
- `Check() error`: Check and validate connection (NOOP)
- `Close()`: Close connection (QUIT)
- `NameList(path string) ([]string, error)`: List file names (NLST)
- `List(path string) ([]*ftp.Entry, error)`: List directory entries (LIST/MLSD)
- `ChangeDir(path string) error`: Change working directory (CWD)
- `CurrentDir() (string, error)`: Get current directory (PWD)
- `FileSize(path string) (int64, error)`: Get file size (SIZE)
- `GetTime(path string) (time.Time, error)`: Get file modification time (MDTM)
- `SetTime(path string, t time.Time) error`: Set file modification time (MFMT/MDTM)
- `Retr(path string) (*ftp.Response, error)`: Download file (RETR)
- `RetrFrom(path string, offset uint64) (*ftp.Response, error)`: Download from offset
- `Stor(path string, r io.Reader) error`: Upload file (STOR)
- `StorFrom(path string, r io.Reader, offset uint64) error`: Upload from offset
- `Append(path string, r io.Reader) error`: Append to file (APPE)
- `Rename(from, to string) error`: Rename file (RNFR/RNTO)
- `Delete(path string) error`: Delete file (DELE)
- `RemoveDirRecur(path string) error`: Recursively delete directory
- `MakeDir(path string) error`: Create directory (MKD)
- `RemoveDir(path string) error`: Remove directory (RMD)
- `Walk(root string) (*ftp.Walker, error)`: Walk directory tree

#### Example

```go
cli, err := ftpclient.New(cfg)
if err != nil {
    // handle error
}
defer cli.Close()

files, err := cli.List("/")
if err != nil {
    // handle error
}
for _, entry := range files {
    // process entry
}
```

---

### Error Handling

All errors are wrapped with custom codes (see `errors.go`):

- `ErrorParamsEmpty`
- `ErrorValidatorError`
- `ErrorEndpointParser`
- `ErrorNotInitialized`
- `ErrorFTPConnection`
- `ErrorFTPConnectionCheck`
- `ErrorFTPLogin`
- `ErrorFTPCommand`

Use `err.Error()` for user-friendly messages and check error codes for diagnostics.

---

### Notes

- The client is thread-safe and manages connection state automatically.
- TLS and context can be customized via registration methods.
- All FTP commands are wrapped and errors are contextualized.
- Designed for Go 1.18+.
