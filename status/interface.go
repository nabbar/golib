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
	"sync/atomic"

	"github.com/gin-gonic/gin"
	librtr "github.com/nabbar/golib/router"
	libver "github.com/nabbar/golib/version"
)

const statusOK = "OK"
const statusKO = "KO"

type Response struct {
	InfoResponse
	StatusResponse

	Components []CptResponse `json:"components"`
}

func (r Response) IsOk() bool {
	if len(r.Components) < 1 {
		return true
	}

	for _, c := range r.Components {
		if c.Status != statusOK {
			return false
		}
	}

	return true
}

func (r Response) IsOkMandatory() bool {
	if len(r.Components) < 1 {
		return true
	}

	for _, c := range r.Components {
		if !c.Mandatory {
			continue
		}

		if c.Status != statusOK {
			return false
		}
	}

	return true
}

type RouteStatus interface {
	MiddlewareAdd(mdw ...gin.HandlerFunc)
	HttpStatusCode(codeOk, codeKO, codeWarning int)

	Get(c *gin.Context)
	Register(prefix string, register librtr.RegisterRouter)
	RegisterGroup(group, prefix string, register librtr.RegisterRouterInGroup)

	ComponentNew(key string, cpt Component)
	ComponentDel(key string)
	ComponentDelAll(containKey string)
}

func New(Name string, Release string, Hash string, msgOk string, msgKo string, msgWarm string) RouteStatus {
	return &rtrStatus{
		m:   make([]gin.HandlerFunc, 0),
		n:   Name,
		v:   Release,
		h:   Hash,
		mOK: msgOk,
		cOk: http.StatusOK,
		mKO: msgKo,
		cKO: http.StatusServiceUnavailable,
		mWM: msgWarm,
		cWM: http.StatusOK,
		c:   make(map[string]*atomic.Value),
	}
}

func NewVersion(version libver.Version, msgOk string, msgKO string, msgWarm string) RouteStatus {
	return New(version.GetPackage(), version.GetRelease(), version.GetBuild(), msgOk, msgKO, msgWarm)
}
