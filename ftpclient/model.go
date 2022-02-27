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
	liberr "github.com/nabbar/golib/errors"
)

type ftpClient struct {
	m sync.Mutex

	cfg *atomic.Value
	cli *atomic.Value
}

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

func (f *ftpClient) setConfig(cfg *Config) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cfg == nil {
		f.cfg = new(atomic.Value)
	}

	f.cfg.Store(cfg)
}

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

func (f *ftpClient) setClient(cli *libftp.ServerConn) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.cli == nil {
		f.cli = new(atomic.Value)
	}

	f.cli.Store(cli)
}

func (f *ftpClient) Connect() liberr.Error {
	var (
		e   error
		err liberr.Error
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

	if cli, err = cfg.New(); err != nil {
		return err
	}

	if e = cli.NoOp(); e != nil {
		return ErrorFTPConnectionCheck.ErrorParent(e)
	}

	f.setClient(cli)
	return nil
}

func (f *ftpClient) Check() liberr.Error {
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
		return ErrorFTPConnectionCheck.ErrorParent(e)
	}
	return nil
}

func (f *ftpClient) Close() {
	if cli := f.getClient(); cli != nil {
		_ = cli.Quit()
	}
}

func (f *ftpClient) NameList(path string) ([]string, liberr.Error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().NameList(path); e != nil {
		return nil, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "NameList", "NLST"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) List(path string) ([]*libftp.Entry, liberr.Error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().List(path); e != nil {
		return nil, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "List", "MLSD/LIST"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) ChangeDir(path string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().ChangeDir(path); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "ChangeDir", "CWD"))
	} else {
		return nil
	}
}

func (f *ftpClient) CurrentDir() (string, liberr.Error) {
	if err := f.Check(); err != nil {
		return "", err
	}

	if r, e := f.getClient().CurrentDir(); e != nil {
		return "", ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "CurrentDir", "PWD"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) FileSize(path string) (int64, liberr.Error) {
	if err := f.Check(); err != nil {
		return 0, err
	}

	if r, e := f.getClient().FileSize(path); e != nil {
		return 0, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "FileSize", "SIZE"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) GetTime(path string) (time.Time, liberr.Error) {
	if err := f.Check(); err != nil {
		return time.Time{}, err
	}

	if r, e := f.getClient().GetTime(path); e != nil {
		return time.Time{}, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "GetTime", "MDTM"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) SetTime(path string, t time.Time) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().SetTime(path, t); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "SetTime", "MFMT/MDTM"))
	} else {
		return nil
	}
}

func (f *ftpClient) Retr(path string) (*libftp.Response, liberr.Error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().Retr(path); e != nil {
		return nil, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "Retr", "RETR"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) RetrFrom(path string, offset uint64) (*libftp.Response, error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	if r, e := f.getClient().RetrFrom(path, offset); e != nil {
		return nil, ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "RetrFrom", "RETR"))
	} else {
		return r, nil
	}
}

func (f *ftpClient) Stor(path string, r io.Reader) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Stor(path, r); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "Stor", "STOR"))
	} else {
		return nil
	}
}

func (f *ftpClient) StorFrom(path string, r io.Reader, offset uint64) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().StorFrom(path, r, offset); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "StorFrom", "STOR"))
	} else {
		return nil
	}
}

func (f *ftpClient) Append(path string, r io.Reader) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Append(path, r); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "Append", "APPE"))
	} else {
		return nil
	}
}

func (f *ftpClient) Rename(from, to string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Rename(from, to); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "Rename", "RNFR/RNTO"))
	} else {
		return nil
	}
}

func (f *ftpClient) Delete(path string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().Delete(path); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "Delete", "DELE"))
	} else {
		return nil
	}
}

func (f *ftpClient) RemoveDirRecur(path string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().RemoveDirRecur(path); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "RemoveDirRecur", "DELE/RMD"))
	} else {
		return nil
	}
}

func (f *ftpClient) MakeDir(path string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().MakeDir(path); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "MakeDir", "MKD"))
	} else {
		return nil
	}
}

func (f *ftpClient) RemoveDir(path string) liberr.Error {
	if err := f.Check(); err != nil {
		return err
	}

	if e := f.getClient().RemoveDir(path); e != nil {
		return ErrorFTPCommand.ErrorParent(e, fmt.Errorf("command : %s = %s", "MakeDir", "RMD"))
	} else {
		return nil
	}
}

func (f *ftpClient) Walk(root string) (*libftp.Walker, liberr.Error) {
	if err := f.Check(); err != nil {
		return nil, err
	}

	return f.getClient().Walk(root), nil
}
