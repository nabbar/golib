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

package status

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/nabbar/golib/router"
	"github.com/nabbar/golib/version"

	"github.com/gin-gonic/gin"
)

// @TODO : see compliant with https://tools.ietf.org/html/draft-inadarei-api-health-check-02

// Model for function that return 3 string for message :
// ok : no error found for component and/or main
// ko : error found for component and/or main
// cpt : message for main status message only that say some of component are in error and are mandatory
type FctMessagesAll func() (ok string, ko string, cpt string)
type FctMessageItem func() (ok string, ko string)

type FctHealth func() error
type FctInfo func() (name, release, build string)
type FctVersion func() version.Version

type StatusItemResponse struct {
	Name      string
	Status    string
	Message   string
	Release   string
	HashBuild string
}

type StatusResponse struct {
	StatusItemResponse
	Component []StatusItemResponse
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

type statusComponent struct {
	statusItem
	WarnIfErr bool
	later     *initLater
}

type mainPackage struct {
	statusItem
	msgCptErr string
	cpt       []statusComponent
	header    gin.HandlerFunc
	later     *initLater
	init      bool
}

type initLater struct {
	version FctVersion
	info    FctInfo
	msgAll  FctMessagesAll
	msgItm  FctMessageItem
	health  FctHealth
}

type Status interface {
	Register(prefix string, register router.RegisterRouter)
	RegisterGroup(group, prefix string, register router.RegisterRouterInGroup)

	AddComponent(info FctInfo, msg FctMessageItem, health FctHealth, WarnIfError bool, later bool)
	AddVersionComponent(vers FctVersion, msg FctMessageItem, health FctHealth, WarnIfError bool, later bool)

	Get(c *gin.Context)
}

func NewStatus(info FctInfo, msg FctMessagesAll, health FctHealth, Header gin.HandlerFunc, later bool) Status {
	if later {
		return &mainPackage{
			cpt:    make([]statusComponent, 0),
			header: Header,
			later: &initLater{
				version: nil,
				info:    info,
				msgItm:  nil,
				msgAll:  msg,
				health:  health,
			},
			init: false,
		}
	} else {
		msgOk, msgKo, msgCpt := msg()
		name, rel, build := info()
		return &mainPackage{
			statusItem: newItem(name, msgOk, msgKo, rel, build, health),
			msgCptErr:  msgCpt,
			cpt:        make([]statusComponent, 0),
			header:     Header,
			later:      nil,
			init:       false,
		}
	}
}

func NewVersionStatus(vers FctVersion, msg FctMessagesAll, health FctHealth, Header gin.HandlerFunc, later bool) Status {
	if later {
		return &mainPackage{
			cpt:    make([]statusComponent, 0),
			header: Header,
			later: &initLater{
				version: vers,
				info:    nil,
				msgItm:  nil,
				msgAll:  msg,
				health:  health,
			},
			init: false,
		}
	} else {
		msgOk, msgKo, msgCpt := msg()
		return &mainPackage{
			statusItem: newItem(vers().GetPackage(), msgOk, msgKo, vers().GetRelease(), vers().GetBuild(), health),
			msgCptErr:  msgCpt,
			cpt:        make([]statusComponent, 0),
			header:     Header,
			later:      nil,
			init:       false,
		}
	}
}

func newItem(name, msgOK, msgKO, release, build string, health FctHealth) statusItem {
	return statusItem{
		name:    name,
		build:   build,
		msgOK:   msgOK,
		msgKO:   msgKO,
		health:  health,
		release: release,
	}
}

func (p *mainPackage) AddComponent(info FctInfo, msg FctMessageItem, health FctHealth, WarnIfError bool, later bool) {
	if later {
		p.cpt = append(p.cpt, statusComponent{
			WarnIfErr: WarnIfError,
			later: &initLater{
				version: nil,
				info:    info,
				msgItm:  msg,
				health:  health,
			},
		})
	} else {
		name, release, build := info()
		msgOK, msgKO := msg()
		p.cpt = append(p.cpt, statusComponent{
			statusItem: newItem(name, msgOK, msgKO, release, build, health),
			WarnIfErr:  WarnIfError,
			later:      nil,
		})
	}
}

func (p *mainPackage) AddVersionComponent(vers FctVersion, msg FctMessageItem, health FctHealth, WarnIfError bool, later bool) {
	if later {
		p.cpt = append(p.cpt, statusComponent{
			WarnIfErr: WarnIfError,
			later: &initLater{
				version: vers,
				info:    nil,
				msgItm:  msg,
				health:  health,
			},
		})
	} else {
		msgOK, msgKO := msg()
		p.cpt = append(p.cpt, statusComponent{
			statusItem: newItem(vers().GetPackage(), msgOK, msgKO, vers().GetRelease(), vers().GetBuild(), health),
			WarnIfErr:  WarnIfError,
			later:      nil,
		})
	}
}

func (p *mainPackage) initStatus() {
	if p.later != nil {
		var ok, ko string

		if p.later.msgAll != nil {
			ok, ko, p.msgCptErr = p.later.msgAll()
		} else if p.later.msgItm != nil {
			ok, ko = p.later.msgItm()
		}

		if p.later.info != nil {
			name, release, build := p.later.info()
			p.statusItem = newItem(name, ok, ko, release, build, p.health)
		} else if p.later.version != nil {
			vers := p.later.version()
			p.statusItem = newItem(vers.GetPackage(), ok, ko, vers.GetRelease(), vers.GetBuild(), p.health)
		}

		if p.later.health != nil {
			p.health = p.later.health
		}

		p.later = nil
	}

	var cpt = make([]statusComponent, 0)

	for _, part := range p.cpt {
		h := part.health
		if part.later != nil {

			if part.later.health != nil {
				h = part.later.health
			}

			if part.later.info != nil {
				name, release, build := part.later.info()
				ok, ko := part.later.msgItm()
				part = statusComponent{
					statusItem: newItem(name, ok, ko, release, build, h),
					WarnIfErr:  part.WarnIfErr,
					later:      nil,
				}
			} else if p.later.version != nil {
				v := p.later.version()
				ok, ko := p.later.msgItm()

				part = statusComponent{
					statusItem: newItem(v.GetPackage(), ok, ko, v.GetRelease(), v.GetBuild(), h),
					WarnIfErr:  part.WarnIfErr,
					later:      nil,
				}
			}
		}

		cpt = append(cpt, part)
	}

	p.init = true
	p.cpt = cpt
}

func (p *mainPackage) cleanPrefix(prefix string) string {
	return path.Clean(strings.TrimRight(path.Join("/", prefix), "/"))
}

func (p *mainPackage) Register(prefix string, register router.RegisterRouter) {
	prefix = p.cleanPrefix(prefix)

	register(http.MethodGet, prefix, p.header, p.Get)

	if prefix != "/" {
		register(http.MethodGet, prefix+"/", p.header, p.Get)
	}
}

func (p *mainPackage) RegisterGroup(group, prefix string, register router.RegisterRouterInGroup) {
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

	for _, pkg := range p.cpt {
		pres := pkg.GetStatusResponse(c)

		if res.Status == statusOK && pres.Status == statusKO && pkg.WarnIfErr {
			res.Status = statusKO
		}

		if pres.Status == statusKO {
			hasError = true
		}

		res.Component = append(res.Component, pres)
	}

	if res.Status != statusOK {
		if res.Message == p.msgOK {
			res.Message = p.msgCptErr
		} else if res.Message != p.msgKO {
			res.Message = strings.Join([]string{res.Message, p.msgCptErr}, ", ")
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, &res)
	} else if hasError {
		c.JSON(http.StatusAccepted, &res)
	} else {
		c.JSON(http.StatusOK, &res)
	}
}
