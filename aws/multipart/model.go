/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
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

package multipart

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libctx "github.com/nabbar/golib/context"
	libiot "github.com/nabbar/golib/ioutils"
	libsiz "github.com/nabbar/golib/size"
)

type mpu struct {
	m sync.RWMutex
	x libctx.FuncContext
	c FuncClientS3
	s libsiz.Size            // part size
	i string                 // upload id
	b string                 // bucket name
	o string                 // object name
	n int32                  // part counter
	l []sdktyp.CompletedPart // slice of sent part to prepare complete MPU
	w libiot.FileProgress    // working file or temporary file

	// trigger function
	fc func(nPart int, obj string, e error) // on complete
	fp func(eTag string, e error)           // on push part
	fa func(nPart int, obj string, e error) // on abort
}

func (m *mpu) RegisterContext(fct libctx.FuncContext) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.x = fct
}

func (m *mpu) getContext() context.Context {
	if m == nil {
		return context.Background()
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.x == nil {
		return context.Background()
	} else if x := m.x(); x == nil {
		return context.Background()
	} else {
		return x
	}
}

func (m *mpu) RegisterClientS3(fct FuncClientS3) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.c = fct
}

func (m *mpu) getClient() *sdksss.Client {
	if m == nil {
		return nil
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.c == nil {
		return nil
	} else if c := m.c(); c == nil {
		return nil
	} else {
		return c
	}
}

func (m *mpu) RegisterMultipartID(id string) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.i = id
}

func (m *mpu) getMultipartID() string {
	if m == nil {
		return ""
	}

	m.m.RLock()
	defer m.m.RUnlock()

	return m.i
}

func (m *mpu) RegisterWorkingFile(file string, truncate bool) error {
	if m == nil {
		return ErrInvalidInstance
	}

	m.m.Lock()
	defer m.m.Unlock()

	var e error

	if m.w != nil {
		m.m.Unlock()

		if e = m.CheckSend(true, false); e != nil {
			return e
		}

		m.m.Lock()
		_ = m.w.Close()
		m.w = nil
	}

	m.w, e = libiot.NewFileProgressPathWrite(filepath.Clean(file), true, truncate, 0600)

	if e != nil {
		return e
	} else if truncate {
		return m.w.Truncate(0)
	}

	return nil
}

func (m *mpu) getWorkingFile() (libiot.FileProgress, error) {
	if m == nil {
		return nil, ErrInvalidInstance
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.w != nil {
		return m.w, nil
	}

	m.m.RUnlock()
	e := m.setTempWorkingFile()
	m.m.RLock()

	if e != nil {
		return nil, e
	} else if m.w == nil {
		return nil, ErrInvalidTMPFile
	}

	return m.w, nil
}

func (m *mpu) setTempWorkingFile() error {
	if m == nil {
		return ErrInvalidInstance
	}

	m.m.Lock()
	defer m.m.Unlock()

	var e error
	m.w, e = libiot.NewFileProgressTemp()
	return e
}

func (m *mpu) closeWorkingFile() error {
	if m == nil {
		return nil
	}

	m.m.Lock()
	defer m.m.Unlock()

	if m.w == nil {
		return nil
	}

	var e error

	e = m.w.Truncate(0)

	if er := m.w.Close(); er != nil {
		if e != nil {
			e = fmt.Errorf("%v, %v", e, er)
		} else {
			e = er
		}
	}

	m.w = nil
	return e
}

func (m *mpu) getPartSize() libsiz.Size {
	if m == nil {
		return DefaultPartSize
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.s < 1 {
		return DefaultPartSize
	}

	return m.s
}

func (m *mpu) setPartSize(s libsiz.Size) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	if s < 1 {
		s = DefaultPartSize
	}

	m.s = s
}

func (m *mpu) getObject() string {
	if m == nil {
		return ""
	}

	m.m.RLock()
	defer m.m.RUnlock()

	return m.o
}

func (m *mpu) getBucket() string {
	if m == nil {
		return ""
	}

	m.m.RLock()
	defer m.m.RUnlock()

	return m.b
}
