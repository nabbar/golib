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
	"net/http"
	"path"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	librtr "github.com/nabbar/golib/router"
	libsem "github.com/nabbar/golib/semaphore"
)

type rtrStatus struct {
	m []gin.HandlerFunc

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

func (r *rtrStatus) HttpStatusCode(codeOk, codeKO, codeWarning int) {
	r.cOk = codeOk
	r.cKO = codeKO
	r.cWM = codeWarning
}

func (r *rtrStatus) MiddlewareAdd(mdw ...gin.HandlerFunc) {
	if len(r.m) < 1 {
		r.m = make([]gin.HandlerFunc, 0)
	}

	r.m = append(r.m, mdw...)
}

func (r *rtrStatus) cleanPrefix(prefix string) string {
	return path.Clean(strings.TrimRight(path.Join("/", prefix), "/"))
}

func (r *rtrStatus) Register(prefix string, register librtr.RegisterRouter) {
	prefix = r.cleanPrefix(prefix)

	var m = r.m
	m = append(m, r.Get)
	register(http.MethodGet, prefix, m...)

	if prefix != "/" {
		register(http.MethodGet, prefix+"/", m...)
	}
}

func (r *rtrStatus) RegisterGroup(group, prefix string, register librtr.RegisterRouterInGroup) {
	prefix = r.cleanPrefix(prefix)

	var m = r.m
	m = append(m, r.Get)
	register(group, http.MethodGet, prefix, m...)

	if prefix != "/" {
		register(group, http.MethodGet, prefix+"/", m...)
	}
}

func (r *rtrStatus) Get(x *gin.Context) {
	var (
		atm *atomic.Value
		key string
		err liberr.Error
		rsp *Response
		sem libsem.Sem
	)

	defer func() {
		if sem != nil {
			sem.DeferMain()
		}
	}()

	rsp = &Response{
		InfoResponse: InfoResponse{
			Name:      r.n,
			Release:   r.v,
			HashBuild: r.h,
			Mandatory: true,
		},
		StatusResponse: StatusResponse{
			Status:  DefMessageOK,
			Message: r.mOK,
		},
		Components: make([]CptResponse, 0),
	}

	sem = libsem.NewSemaphoreWithContext(x, 0)

	for key, atm = range r.c {
		var (
			ok bool
			c  Component
		)

		if atm == nil {
			continue
		}

		if c, ok = atm.Load().(Component); !ok {
			continue
		}

		err = sem.NewWorker()
		if liblog.ErrorLevel.LogGinErrorCtxf(liblog.DebugLevel, "init new thread to collect data for component '%s'", err, x, key) {
			continue
		}

		go func(c Component) {
			defer sem.DeferWorker()
			rsp.appendNewCpt(c.Get(x))
		}(c)
	}

	err = sem.WaitAll()

	if liblog.ErrorLevel.LogGinErrorCtx(liblog.DebugLevel, "waiting all thread to collect data component ", err, x) {
		rsp.Message = r.mKO
		rsp.Status = DefMessageKO
		x.AbortWithStatusJSON(r.cKO, rsp)
	} else if !rsp.IsOkMandatory() {
		rsp.Message = r.mKO
		rsp.Status = DefMessageKO
		x.AbortWithStatusJSON(r.cKO, rsp)
	} else if !rsp.IsOk() {
		rsp.Message = r.mWM
		rsp.Status = DefMessageOK
		x.JSON(r.cWM, rsp)
	} else {
		rsp.Message = r.mOK
		rsp.Status = DefMessageOK
		x.JSON(r.cOk, rsp)
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
