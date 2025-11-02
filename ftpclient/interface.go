/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

// Package ftpclient provides a thread-safe FTP client with automatic reconnection
// and connection pooling capabilities.
//
// This package wraps the github.com/jlaffaye/ftp library with additional features:
//   - Thread-safe operations using atomic values and mutexes
//   - Automatic connection management with health checks (NOOP)
//   - Flexible configuration with TLS/SSL support
//   - Context-aware operations for timeout control
//   - Comprehensive error handling with custom error types
//
// Architecture:
//
//	┌──────────────┐
//	│   FTPClient  │ ← Public Interface
//	└──────────────┘
//	       ↓
//	┌──────────────┐
//	│  ftpClient   │ ← Thread-safe implementation
//	│  (atomic)    │
//	└──────────────┘
//	       ↓
//	┌──────────────┐
//	│ ServerConn   │ ← Underlying FTP connection
//	└──────────────┘
//
// Basic usage:
//
//	import (
//	    "context"
//	    "github.com/nabbar/golib/ftpclient"
//	)
//
//	cfg := &ftpclient.Config{
//	    Hostname:    "ftp.example.com:21",
//	    Login:       "user",
//	    Password:    "pass",
//	    ConnTimeout: 30 * time.Second,
//	}
//	cfg.RegisterContext(func() context.Context {
//	    return context.Background()
//	})
//
//	client, err := ftpclient.New(cfg)
//	if err != nil {
//	    panic(err)
//	}
//	defer client.Close()
//
//	// Upload a file
//	file, _ := os.Open("local.txt")
//	defer file.Close()
//	err = client.Stor("remote.txt", file)
//
// See the Config struct for all available configuration options.
package ftpclient

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	libftp "github.com/jlaffaye/ftp"
)

// FTPClient defines the interface for FTP operations.
// All methods are thread-safe and handle connection management automatically.
// Failed connections will be automatically re-established on the next operation.
type FTPClient interface {
	// Connect establish the connection to server with the given configuration registered.
	Connect() error

	// Check try to retrieve a valid connection to the server and send an NOOP command to check the connection.
	Check() error

	// Close send the QUID command to the server if the connection is valid (cf Check).
	Close()

	// NameList issues an NLST FTP command.
	NameList(path string) ([]string, error)

	// List issues a LIST FTP command.
	List(path string) ([]*libftp.Entry, error)

	// ChangeDir issues a CWD FTP command, which changes the current directory to the specified path.
	ChangeDir(path string) error

	// CurrentDir issues a PWD FTP command, which Returns the path of the current directory.
	CurrentDir() (string, error)

	// FileSize issues a SIZE FTP command, which Returns the size of the file.
	FileSize(path string) (int64, error)

	// GetTime issues the MDTM FTP command to obtain the file modification time.
	// It returns a UTC time.
	GetTime(path string) (time.Time, error)

	// SetTime issues the MFMT FTP command to set the file modification time.
	// Also it can use a non-standard form of the MDTM command supported by the VsFtpd server instead of MFMT for the same purpose.
	// See "mdtm_write" in https://security.appspot.com/vsftpd/vsftpd_conf.html
	SetTime(path string, t time.Time) error

	// Retr issues a RETR FTP command to fetch the specified file from the remote FTP server.
	// The returned ReadCloser must be closed to cleanup the FTP data connection.
	Retr(path string) (*libftp.Response, error)

	// RetrFrom issues a RETR FTP command to fetch the specified file from the remote FTP server,
	// the server will not send the offset first bytes of the file.
	// The returned ReadCloser must be closed to cleanup the FTP data connection.
	RetrFrom(path string, offset uint64) (*libftp.Response, error)

	// Stor issues a STOR FTP command to store a file to the remote FTP server.
	// Stor creates the specified file with the content of the io.Reader.
	// Hint: io.Pipe() can be used if an io.Writer is required.
	Stor(path string, r io.Reader) error

	// StorFrom issues a STOR FTP command to store a file to the remote FTP server.
	// Stor creates the specified file with the content of the io.Reader, writing on the server will start at the given file offset.
	// Hint: io.Pipe() can be used if an io.Writer is required.
	StorFrom(path string, r io.Reader, offset uint64) error

	// Append issues a APPE FTP command to store a file to the remote FTP server.
	// If a file already exists with the given path, then the content of the io.Reader is appended.
	// Otherwise, a new file is created with that content. Hint: io.Pipe() can be used if an io.Writer is required.
	Append(path string, r io.Reader) error

	// Rename renames a file on the remote FTP server.
	Rename(from, to string) error

	// Delete issues a DELE FTP command to delete the specified file from the remote FTP server.
	Delete(path string) error

	// RemoveDirRecur deletes a non-empty folder recursively using RemoveDir and Delete.
	RemoveDirRecur(path string) error

	// MakeDir issues a MKD FTP command to create the specified directory on the remote FTP server.
	MakeDir(path string) error

	// RemoveDir issues a RMD FTP command to remove the specified directory from the remote FTP server.
	RemoveDir(path string) error

	//Walk prepares the internal walk function so that the caller can begin traversing the directory.
	Walk(root string) (*libftp.Walker, error)
}

// New creates a new FTP client instance with the given configuration.
// It immediately attempts to connect and validate the connection using a NOOP command.
//
// The client uses atomic operations for thread-safe configuration and connection management.
// If the initial connection fails, an error is returned and the client is nil.
//
// Example:
//
//	cfg := &Config{
//	    Hostname: "ftp.example.com:21",
//	    Login:    "user",
//	    Password: "pass",
//	}
//	client, err := New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func New(cfg *Config) (FTPClient, error) {
	c := &ftpClient{
		m:   sync.Mutex{},
		cfg: new(atomic.Value),
		cli: new(atomic.Value),
	}

	c.setConfig(cfg)

	if err := c.Check(); err != nil {
		return nil, err
	}

	return c, nil
}
