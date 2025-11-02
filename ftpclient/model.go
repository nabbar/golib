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

package ftpclient

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	libftp "github.com/jlaffaye/ftp"
)

// ftpClient is the internal implementation of the FTPClient interface.
// It uses atomic values for thread-safe configuration and connection management.
// All public methods automatically handle connection health checks and reconnection.
type ftpClient struct {
	m sync.Mutex // Protects access to atomic values

	cfg *atomic.Value // Stores *Config
	cli *atomic.Value // Stores *libftp.ServerConn
}

// getConfig retrieves the current configuration in a thread-safe manner.
// Returns nil if the configuration is not set or invalid.
func (f *ftpClient) getConfig() *Config {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cfg == nil {
		return nil
	} else if i := f.cfg.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Config); !ok {
		return nil
	} else {
		return o
	}
}

// setConfig stores a new configuration in a thread-safe manner.
// Creates a new atomic.Value if not already initialized.
func (f *ftpClient) setConfig(cfg *Config) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cfg == nil {
		f.cfg = new(atomic.Value)
	}

	f.cfg.Store(cfg)
}

// getClient retrieves the current FTP connection in a thread-safe manner.
// Returns nil if no connection is established or if the connection is invalid.
func (f *ftpClient) getClient() *libftp.ServerConn {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cli == nil {
		return nil
	} else if i := f.cli.Load(); i == nil {
		return nil
	} else if o, ok := i.(*libftp.ServerConn); !ok {
		return nil
	} else {
		return o
	}
}

// setClient stores a new FTP connection in a thread-safe manner.
// Creates a new atomic.Value if not already initialized.
func (f *ftpClient) setClient(cli *libftp.ServerConn) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cli == nil {
		f.cli = new(atomic.Value)
	}

	f.cli.Store(cli)
}

// Connect establishes a connection to the FTP server using the registered configuration.
// If a connection already exists and is healthy (NOOP check passes), it returns immediately.
// Otherwise, it creates a new connection, validates it with a NOOP command, and stores it.
//
// Returns an error if:
//   - The configuration is not initialized
//   - The connection cannot be established
//   - The NOOP health check fails
func (f *ftpClient) Connect() error {
	var (
		e   error
		cfg *Config
		cli *libftp.ServerConn
	)

	if cli = f.getClient(); cli != nil {
		if e = cli.NoOp(); e != nil {
			_ = cli.Quit()
		} else {
			return nil
		}
	}

	if cfg = f.getConfig(); cfg == nil {
		return ErrorNotInitialized.Error(nil)
	}

	if cli, e = cfg.New(); e != nil {
		return e
	}

	if e = cli.NoOp(); e != nil {
		return ErrorFTPConnectionCheck.Error(e)
	}

	f.setClient(cli)
	return nil
}

// Check validates the current connection to the FTP server.
// If no connection exists, it attempts to establish one via Connect().
// For existing connections, it sends a NOOP command to verify the connection is alive.
//
// This method is called automatically before all FTP operations to ensure
// connection health and trigger reconnection if needed.
//
// Returns an error if:
//   - Unable to establish a connection
//   - The NOOP health check fails
func (f *ftpClient) Check() error {
	var cli *libftp.ServerConn

	if cli = f.getClient(); cli == nil {
		if err := f.Connect(); err != nil {
			return err
		}
	}

	if cli = f.getClient(); cli == nil {
		return ErrorNotInitialized.Error(nil)
	}

	if e := cli.NoOp(); e != nil {
		return ErrorFTPConnectionCheck.Error(e)
	}
	return nil
}

// Close gracefully terminates the FTP connection by sending a QUIT command.
// This method is safe to call multiple times or when no connection exists.
// It does not return an error; any QUIT command errors are silently ignored.
//
// Always call Close() when done with the client to free server resources:
//
//	defer client.Close()
func (f *ftpClient) Close() {
	if cli := f.getClient(); cli != nil {
		_ = cli.Quit()
	}
}

func (f *ftpClient) NameList(path string) ([]string, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().NameList(path); e != nil {
		return nil, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "NameList", "NLST"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) List(path string) ([]*libftp.Entry, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().List(path); e != nil {
		return nil, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "List", "MLSD/LIST"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) ChangeDir(path string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().ChangeDir(path); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "ChangeDir", "CWD"))
	} else {
		return nil
	}
}

func (f *ftpClient) CurrentDir() (string, error) {
	if err := f.Check(); err != nil {
		return "", err
	}

	if r, e := f.getClient().CurrentDir(); e != nil {
		return "", ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "CurrentDir", "PWD"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) FileSize(path string) (int64, error) {
	if err := f.Check(); err != nil {
		return 0, err
	}

	if r, e := f.getClient().FileSize(path); e != nil {
		return 0, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "FileSize", "SIZE"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) GetTime(path string) (time.Time, error) {
	if err := f.Check(); err != nil {
		return time.Time{}, err
	}

	if r, e := f.getClient().GetTime(path); e != nil {
		return time.Time{}, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "GetTime", "MDTM"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) SetTime(path string, t time.Time) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().SetTime(path, t); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "SetTime", "MFMT/MDTM"))
	} else {
		return nil
	}
}

func (f *ftpClient) Retr(path string) (*libftp.Response, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().Retr(path); e != nil {
		return nil, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "Retr", "RETR"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) RetrFrom(path string, offset uint64) (*libftp.Response, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().RetrFrom(path, offset); e != nil {
		return nil, ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "RetrFrom", "RETR"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) Stor(path string, r io.Reader) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Stor(path, r); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "Stor", "STOR"))
	} else {
		return nil
	}
}

func (f *ftpClient) StorFrom(path string, r io.Reader, offset uint64) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().StorFrom(path, r, offset); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "StorFrom", "STOR"))
	} else {
		return nil
	}
}

func (f *ftpClient) Append(path string, r io.Reader) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Append(path, r); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "Append", "APPE"))
	} else {
		return nil
	}
}

func (f *ftpClient) Rename(from, to string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Rename(from, to); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "Rename", "RNFR/RNTO"))
	} else {
		return nil
	}
}

func (f *ftpClient) Delete(path string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Delete(path); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "Delete", "DELE"))
	} else {
		return nil
	}
}

func (f *ftpClient) RemoveDirRecur(path string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().RemoveDirRecur(path); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "RemoveDirRecur", "DELE/RMD"))
	} else {
		return nil
	}
}

func (f *ftpClient) MakeDir(path string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().MakeDir(path); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "MakeDir", "MKD"))
	} else {
		return nil
	}
}

func (f *ftpClient) RemoveDir(path string) error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().RemoveDir(path); e != nil {
		return ErrorFTPCommand.Error(e, fmt.Errorf("command : %s = %s", "MakeDir", "RMD"))
	} else {
		return nil
	}
}

func (f *ftpClient) Walk(root string) (*libftp.Walker, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	return f.getClient().Walk(root), nil
}
