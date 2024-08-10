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
	"crypto/md5"
	"crypto/sha256"
	"hash"
)

func (o *psh) md5Get() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if i := o.prtMD5.Load(); i != nil {
		if v, k := i.(hash.Hash); k {
			return v, nil
		}
	}

	return o.md5Init()
}

func (o *psh) md5Init() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else {
		return md5.New(), nil
	}
}

func (o *psh) md5Reset() error {
	if !o.IsReady() {
		return ErrInvalidInstance
	} else if h, e := o.md5Init(); e != nil {
		return e
	} else if i := o.prtMD5.Swap(h); i == nil {
		return nil
	} else if v, k := i.(hash.Hash); !k {
		return nil
	} else {
		v.Reset()
		return nil
	}
}

func (o *psh) md5Write(p []byte) (n int, err error) {
	var (
		e error
		h hash.Hash
	)

	if !o.IsReady() {
		return 0, ErrInvalidInstance
	} else if h, e = o.md5Get(); e != nil {
		return 0, e
	} else {
		n, e = h.Write(p)
		o.prtMD5.Store(h)
		return n, e
	}
}

func (o *psh) md5Checksum() ([]byte, error) {
	if !o.IsReady() {
		return make([]byte, 0), ErrInvalidInstance
	} else if i := o.prtMD5.Load(); i == nil {
		return make([]byte, 0), ErrInvalidInstance
	} else if v, k := i.(hash.Hash); !k {
		return make([]byte, 0), ErrInvalidInstance
	} else {
		return v.Sum(nil), nil
	}
}

func (o *psh) shaPartGet() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return nil, nil
	} else if i := o.prtSha2.Load(); i != nil {
		if v, k := i.(hash.Hash); k {
			return v, nil
		}
	}

	return o.shaPartInit()
}

func (o *psh) shaPartInit() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else {
		return sha256.New(), nil
	}
}

func (o *psh) shaPartReset() error {
	if !o.IsReady() {
		return ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return nil
	} else if h, e := o.shaPartInit(); e != nil {
		return e
	} else if i := o.prtSha2.Swap(h); i == nil {
		return nil
	} else if v, k := i.(hash.Hash); !k {
		return nil
	} else {
		v.Reset()
		return nil
	}
}

func (o *psh) shaPartWrite(p []byte) (n int, err error) {
	var (
		e error
		h hash.Hash
	)

	if !o.IsReady() {
		return 0, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return len(p), nil
	} else if h, e = o.shaPartGet(); e != nil {
		return 0, e
	} else {
		n, e = h.Write(p)
		o.prtSha2.Store(h)
		return n, e
	}
}

func (o *psh) shaPartChecksum() ([]byte, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return nil, nil
	} else if i := o.prtSha2.Load(); i == nil {
		return nil, ErrInvalidInstance
	} else if v, k := i.(hash.Hash); !k {
		return nil, ErrInvalidInstance
	} else {
		return v.Sum(nil), nil
	}
}

func (o *psh) shaObjGet() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return nil, nil
	} else if i := o.objSha2.Load(); i != nil {
		if v, k := i.(hash.Hash); k {
			return v, nil
		}
	}

	return o.shaObjInit()
}

func (o *psh) shaObjInit() (hash.Hash, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else {
		return sha256.New(), nil
	}
}

func (o *psh) shaObjWrite(p []byte) (n int, err error) {
	var (
		e error
		h hash.Hash
	)

	if !o.IsReady() {
		return 0, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return len(p), nil
	} else if h, e = o.shaObjGet(); e != nil {
		return 0, e
	} else {
		n, e = h.Write(p)
		o.objSha2.Store(h)
		return n, e
	}
}

func (o *psh) shaObjChecksum() ([]byte, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if !o.cfg.isCheckSum() {
		return nil, nil
	} else if i := o.objSha2.Load(); i == nil {
		return nil, ErrInvalidInstance
	} else if v, k := i.(hash.Hash); !k {
		return nil, ErrInvalidInstance
	} else {
		return v.Sum(nil), nil
	}
}
