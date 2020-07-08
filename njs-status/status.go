/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package njs_status

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	njs_router "github.com/nabbar/golib/njs-router"

	njs_version "github.com/nabbar/golib/njs-version"

	"github.com/gin-gonic/gin"
)

type StatusItemResponse struct {
	Name      string
	Status    string
	Message   string
	Release   string
	HashBuild string
}

type StatusResponse struct {
	StatusItemResponse
	Partner []StatusItemResponse
}

const statusOK = "OK"
const statusKO = "KO"

type statusItem struct {
	name    string
	build   string
	msgOK   string
	msgKO   string
	health  func() error
	release string
}

type statusPartner struct {
	statusItem
	WarnIfErr bool
	later     *initLater
}

type mainPackage struct {
	statusItem
	ptn    []statusPartner
	header func(c *gin.Context)
	later  *initLater
	init   bool
}

type initLater struct {
	version func() njs_version.Version
	info    func() (name, release, build string)
	msg     func() (ok string, ko string)
	health  func() error
}

type Status interface {
	Register(prefix string, register njs_router.RegisterRouter)

	AddPartner(name, msgOK, msgKO, release, build string, WarnIfError bool, health func() error)
	AddVersionPartner(vers njs_version.Version, msgOK, msgKO string, WarnIfError bool, health func() error)

	LaterAddPartner(info func() (name, release, build string), msg func() (ok string, ko string), health func() error, WarnIfError bool)
	LaterAddVersionPartner(vers func() njs_version.Version, msg func() (ok string, ko string), health func() error, WarnIfError bool)

	Get(c *gin.Context)
}

func NewStatusLater(info func() (name, release, build string), msg func() (ok string, ko string), health func() error, Header func(c *gin.Context)) Status {
	return &mainPackage{
		ptn:    make([]statusPartner, 0),
		header: Header,
		later: &initLater{
			version: nil,
			info:    info,
			msg:     msg,
			health:  health,
		},
		init: false,
	}
}

func NewStatus(name, msgOK, msgKO, release, build string, health func() error, Header func(c *gin.Context)) Status {
	return &mainPackage{
		statusItem: newItem(name, msgOK, msgKO, release, build, health),
		ptn:        make([]statusPartner, 0),
		header:     Header,
		later:      nil,
		init:       false,
	}
}

func NewVersionStatusLater(vers func() njs_version.Version, msg func() (ok string, ko string), health func() error, Header func(c *gin.Context)) Status {
	return &mainPackage{
		ptn:    make([]statusPartner, 0),
		header: Header,
		later: &initLater{
			version: vers,
			info:    nil,
			msg:     msg,
			health:  health,
		},
		init: false,
	}
}

func NewVersionStatus(vers njs_version.Version, msgOK, msgKO string, health func() error, Header func(c *gin.Context)) Status {
	return NewStatus(vers.GetPackage(), msgOK, msgKO, vers.GetRelease(), vers.GetBuild(), health, Header)
}

func newItem(name, msgOK, msgKO, release, build string, health func() error) statusItem {
	return statusItem{
		name:    name,
		build:   build,
		msgOK:   msgOK,
		msgKO:   msgKO,
		health:  health,
		release: release,
	}
}

func (p *mainPackage) AddPartner(name, msgOK, msgKO, release, build string, WarnIfError bool, health func() error) {
	p.ptn = append(p.ptn, statusPartner{
		statusItem: newItem(name, msgOK, msgKO, release, build, health),
		WarnIfErr:  WarnIfError,
		later:      nil,
	})
}

func (p *mainPackage) LaterAddPartner(info func() (name, release, build string), msg func() (ok string, ko string), health func() error, WarnIfError bool) {
	p.ptn = append(p.ptn, statusPartner{
		WarnIfErr: WarnIfError,
		later: &initLater{
			version: nil,
			info:    info,
			msg:     msg,
			health:  health,
		},
	})
}

func (p *mainPackage) AddVersionPartner(vers njs_version.Version, msgOK, msgKO string, WarnIfError bool, health func() error) {
	p.AddPartner(vers.GetPackage(), msgOK, msgKO, vers.GetRelease(), vers.GetBuild(), WarnIfError, health)
}

func (p *mainPackage) LaterAddVersionPartner(vers func() njs_version.Version, msg func() (ok string, ko string), health func() error, WarnIfError bool) {
	p.ptn = append(p.ptn, statusPartner{
		WarnIfErr: WarnIfError,
		later: &initLater{
			version: vers,
			info:    nil,
			msg:     msg,
			health:  health,
		},
	})
}

func (p *mainPackage) initStatus() {
	if p.later != nil {
		if p.later.info != nil {
			name, release, build := p.later.info()
			ok, ko := p.later.msg()
			p.statusItem = newItem(name, ok, ko, release, build, p.health)
		} else if p.later.version != nil {
			vers := p.later.version()
			ok, ko := p.later.msg()
			p.statusItem = newItem(vers.GetPackage(), ok, ko, vers.GetRelease(), vers.GetBuild(), p.health)
		}

		if p.later.health != nil {
			p.health = p.later.health
		}

		p.later = nil
	}

	var ptn = make([]statusPartner, 0)

	for _, part := range p.ptn {
		if part.later != nil {
			if part.later.info != nil {
				name, release, build := part.later.info()
				ok, ko := part.later.msg()
				part = statusPartner{
					statusItem: newItem(name, ok, ko, release, build, part.health),
					WarnIfErr:  part.WarnIfErr,
					later:      nil,
				}
			} else if p.later.version != nil {
				v := p.later.version()
				ok, ko := p.later.msg()
				part = statusPartner{
					statusItem: newItem(v.GetPackage(), ok, ko, v.GetRelease(), v.GetBuild(), part.health),
					WarnIfErr:  part.WarnIfErr,
					later:      nil,
				}
			}

			if part.later.health != nil {
				part.health = part.later.health
			}
		}

		ptn = append(ptn, part)
	}

	p.init = true
	p.ptn = ptn
}

func (p *mainPackage) cleanPrefix(prefix string) string {
	return path.Clean(strings.TrimRight(path.Join("/", prefix), "/"))
}

func (p *mainPackage) Register(prefix string, register njs_router.RegisterRouter) {
	prefix = p.cleanPrefix(prefix)

	register(http.MethodGet, prefix, p.header, p.Get)

	if prefix != "/" {
		register(http.MethodGet, prefix+"/", p.header, p.Get)
	}
}

func (p *mainPackage) RegisterGroup(group, prefix string, register njs_router.RegisterRouterInGroup) {
	prefix = p.cleanPrefix(prefix)

	register(group, http.MethodGet, prefix, p.header, p.Get)

	if prefix != "/" {
		register(group, http.MethodGet, prefix+"/", p.header, p.Get)
	}
}

func (p *statusItem) GetStatusResponse(c *gin.Context) StatusItemResponse {
	res := StatusItemResponse{
		Name:      p.name,
		Status:    statusOK,
		Message:   p.msgOK,
		Release:   p.release,
		HashBuild: p.build,
	}

	if p.health != nil {
		if err := p.health(); err != nil {
			msg := fmt.Sprintf("%s: %v", p.msgKO, err)
			c.Errors = append(c.Errors, &gin.Error{
				Err:  fmt.Errorf(msg),
				Type: gin.ErrorTypePrivate,
			})
			res = StatusItemResponse{
				Name:      p.name,
				Status:    statusKO,
				Message:   msg,
				Release:   p.release,
				HashBuild: p.build,
			}
		}
	}

	return res
}

func (p *mainPackage) Get(c *gin.Context) {
	if !p.init {
		p.initStatus()
	}

	hasError := false
	res := StatusResponse{
		p.GetStatusResponse(c),
		make([]StatusItemResponse, 0),
	}

	for _, pkg := range p.ptn {
		pres := pkg.GetStatusResponse(c)
		if res.Status == statusOK && pres.Status == statusKO && !pkg.WarnIfErr {
			res.Status = statusKO
		} else if pres.Status == statusKO {
			hasError = true
		}
		res.Partner = append(res.Partner, pres)
	}

	if res.Status != statusOK {
		c.AbortWithStatusJSON(http.StatusInternalServerError, &res)
	} else if hasError {
		c.JSON(http.StatusMultiStatus, &res)
	} else {
		c.JSON(http.StatusOK, &res)
	}
}
