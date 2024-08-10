/*
 *  MIT License
 *
 *  Copyright (c) 2024 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package pusher

import (
	"io"
	"os"
)

func (o *psh) getFile() (*os.File, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if i := o.tmp.Load(); i != nil {
		if v, k := i.(*os.File); k {
			return v, nil
		}
	}

	if f, e := o.cfg.getWorkingFile(); f == nil || e != nil {
		return nil, e
	} else {
		o.tmp.Store(f)
		return f, nil
	}
}

func (o *psh) fileReset() error {
	if f, e := o.getFile(); e != nil {
		return e
	} else {
		_, e = f.Seek(0, io.SeekStart)
		o.tmp.Store(f)
		return e
	}
}

func (o *psh) fileTruncate() error {
	if f, e := o.getFile(); e != nil {
		return e
	} else {
		o.prtSize.Store(0)
		e = f.Truncate(0)
		o.tmp.Store(f)
		return e
	}
}

func (o *psh) fileClose() error {
	if f, e := o.getFile(); e != nil {
		return e
	} else {
		e = f.Close()
		o.tmp.Store(f)
		return e
	}
}

func (o *psh) fileRemove() error {
	if f, e := o.getFile(); e != nil {
		return e
	} else {
		n := f.Name()
		e = o.fileClose()

		o.resetPartList()
		o.resetUploadId()
		o.prtSize.Store(0)
		o.objSize.Store(0)
		o.nbrPart.Store(0)

		_ = o.md5Reset()
		_ = o.shaPartReset()

		if er := os.Remove(n); er != nil && e != nil {
			return e
		} else if er != nil {
			return er
		}

		return nil
	}
}

func (o *psh) fileWrite(p []byte) (n int, err error) {
	if f, e := o.getFile(); e != nil {
		return 0, e
	} else {
		n, err = f.Write(p)

		// updating size counters
		o.prtSize.Add(int64(n))
		o.objSize.Add(int64(n))

		return n, err
	}
}
