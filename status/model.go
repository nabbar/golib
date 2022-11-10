/*
 * MIT License
 *
 * Copyright (c) 2021 Nicolas JUHEL
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
 */

package status

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin/render"

	"github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	librtr "github.com/nabbar/golib/router"
	libsem "github.com/nabbar/golib/semaphore"
)

type rtrStatus struct {
	m sync.Mutex
	f []gin.HandlerFunc

	n string
	v string
	h string

	mOK string
	cOk int
	mKO string
	cKO int
	mWM string
	cWM int

	c map[string]*atomic.Value
}

const (
	keyShortOutput   = "short"
	keyOneLineOutput = "oneline"
)

func (r *rtrStatus) HttpStatusCode(codeOk, codeKO, codeWarning int) {
	r.cOk = codeOk
	r.cKO = codeKO
	r.cWM = codeWarning
}

func (r *rtrStatus) MiddlewareAdd(mdw ...gin.HandlerFunc) {
	if len(r.f) < 1 {
		r.f = make([]gin.HandlerFunc, 0)
	}

	r.f = append(r.f, mdw...)
}

func (r *rtrStatus) cleanPrefix(prefix string) string {
	return path.Clean(strings.TrimRight(path.Join("/", prefix), "/"))
}

func (r *rtrStatus) Register(prefix string, register librtr.RegisterRouter) {
	prefix = r.cleanPrefix(prefix)

	var m = r.f
	m = append(m, r.Get)
	register(http.MethodGet, prefix, m...)

	if prefix != "/" {
		register(http.MethodGet, prefix+"/", m...)
	}
}

func (r *rtrStatus) RegisterGroup(group, prefix string, register librtr.RegisterRouterInGroup) {
	prefix = r.cleanPrefix(prefix)

	var m = r.f
	m = append(m, r.Get)
	register(group, http.MethodGet, prefix, m...)

	if prefix != "/" {
		register(group, http.MethodGet, prefix+"/", m...)
	}
}

func (r *rtrStatus) getInfo() (name string, release string, hash string) {
	r.m.Lock()
	defer r.m.Unlock()

	return r.n, r.v, r.h
}

func (r *rtrStatus) getMsgOk() string {
	r.m.Lock()
	defer r.m.Unlock()

	return r.mOK
}

func (r *rtrStatus) getMsgKo() string {
	r.m.Lock()
	defer r.m.Unlock()

	return r.mKO
}

func (r *rtrStatus) getMsgWarn() string {
	r.m.Lock()
	defer r.m.Unlock()

	return r.mWM
}

func (r *rtrStatus) Get(x *gin.Context) {
	var (
		key string
		err liberr.Error
		rsp *Response
		s   libsem.Sem
	)

	defer func() {
		if s != nil {
			s.DeferMain()
		}
	}()

	inf := InfoResponse{
		Mandatory: true,
	}
	inf.Name, inf.Release, inf.HashBuild = r.getInfo()

	sts := StatusResponse{
		Status: DefMessageOK,
	}
	sts.Message = r.getMsgOk()

	rsp = &Response{
		InfoResponse:   inf,
		StatusResponse: sts,
		Components:     make([]CptResponse, 0),
	}

	s = libsem.NewSemaphoreWithContext(x, 0)

	for _, key = range r.ComponentKeys() {
		var c Component

		if c = r.ComponentGet(key); c == nil {
			continue
		}

		err = s.NewWorker()
		if liblog.ErrorLevel.LogGinErrorCtxf(liblog.DebugLevel, "init new thread to collect data for component '%s'", err, x, key) {
			continue
		}

		go func(ctx *gin.Context, sem libsem.Sem, cpt Component, resp *Response) {
			defer sem.DeferWorker()
			resp.appendNewCpt(cpt.Get(ctx))
		}(x, s, c, rsp)
	}

	err = s.WaitAll()

	var (
		code int
	)

	if liblog.ErrorLevel.LogGinErrorCtx(liblog.DebugLevel, "waiting all thread to collect data component ", err, x) {
		rsp.Message = r.getMsgKo()
		rsp.Status = DefMessageKO
		code = r.cKO
	} else if !rsp.IsOkMandatory() {
		rsp.Message = r.getMsgKo()
		rsp.Status = DefMessageKO
		code = r.cKO
	} else if !rsp.IsOk() {
		rsp.Message = r.getMsgWarn()
		rsp.Status = DefMessageOK
		code = r.cWM
	} else {
		rsp.Message = r.getMsgOk()
		rsp.Status = DefMessageOK
		code = r.cOk
	}

	if x.Request.URL.Query().Has(keyShortOutput) {
		rsp.Components = make([]CptResponse, 0)
	}

	x.Header("Connection", "Close")

	if code == r.cKO {
		x.Abort()
	}

	if x.Request.URL.Query().Has(keyOneLineOutput) {
		var buf = bytes.NewBuffer(make([]byte, 0))
		buf.WriteString(fmt.Sprintf("%s: %s (%s - %s) : %s\n", rsp.Status, rsp.Name, rsp.Release, rsp.HashBuild, rsp.Message))

		for _, c := range rsp.Components {
			buf.WriteString(fmt.Sprintf("%s: %s (%s - %s) : %s\n", c.Status, c.Name, c.Release, c.HashBuild, c.Message))
		}

		x.Render(code, render.Data{
			ContentType: gin.MIMEPlain,
			Data:        buf.Bytes(),
		})
	} else {
		x.JSON(code, rsp)
	}
}

func (r *rtrStatus) ComponentKeys() []string {
	var l = make([]string, 0)

	r.m.Lock()
	defer r.m.Unlock()

	for k := range r.c {
		if len(k) > 0 {
			l = append(l, k)
		}
	}

	return l
}

func (r *rtrStatus) ComponentGet(key string) Component {
	var (
		v  *atomic.Value
		i  interface{}
		o  Component
		ok bool
	)

	if v, ok = r.c[key]; !ok || v == nil {
		return nil
	} else if i = v.Load(); i == nil {
		return nil
	} else if o, ok = i.(Component); !ok {
		return nil
	} else {
		return o
	}
}

func (r *rtrStatus) ComponentNew(key string, cpt Component) {
	if len(r.c) < 1 {
		r.c = make(map[string]*atomic.Value)
	}

	if _, ok := r.c[key]; !ok {
		r.c[key] = &atomic.Value{}
	}

	r.c[key].Store(cpt)
}

func (r *rtrStatus) ComponentDel(key string) {
	for k := range r.c {
		if k == key {
			r.c[k].Store(nil)
		}
	}
}

func (r *rtrStatus) ComponentDelAll(containKey string) {
	if containKey == "" {
		r.c = make(map[string]*atomic.Value)
		return
	}

	for k := range r.c {
		if strings.Contains(k, containKey) {
			r.c[k].Store(nil)
		}
	}
}
