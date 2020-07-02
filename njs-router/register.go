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

package njs_router

import (
	"github.com/gin-gonic/gin"
)

const _EMPTY_GROUP = "<nil>"

var (
	defaultRouters = NewRouterList()
)

type routerItem struct {
	method   string
	relative string
	router   []gin.HandlerFunc
}

type routerList struct {
	list map[string][]routerItem
}

type RegisterRouter func(method string, relativePath string, router ...gin.HandlerFunc)

type RouterList interface {
	Register(method string, relativePath string, router ...gin.HandlerFunc)
	RegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc)
	Handler(engine *gin.Engine)
}

func RoutersRegister(method string, relativePath string, router ...gin.HandlerFunc) {
	defaultRouters.Register(method, relativePath, router...)
}

func RoutersRegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc) {
	defaultRouters.RegisterInGroup(group, method, relativePath, router...)
}

func RoutersHandler(engine *gin.Engine) {
	defaultRouters.Handler(engine)
}

func NewRouterList() RouterList {
	return &routerList{
		list: make(map[string][]routerItem, 0),
	}
}

func (l routerList) Handler(engine *gin.Engine) {
	for grpRoute, grpList := range l.list {
		if grpRoute == _EMPTY_GROUP {
			for _, r := range grpList {
				engine.Handle(r.method, r.relative, r.router...)
			}
		} else {
			var grp = engine.Group(grpRoute)
			for _, r := range grpList {
				grp.Handle(r.method, r.relative, r.router...)
			}
		}
	}
}

func (l *routerList) RegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc) {
	if group == "" {
		group = _EMPTY_GROUP
	}

	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]routerItem, 0)
	}

	l.list[group] = append(l.list[group], routerItem{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

func (l *routerList) Register(method string, relativePath string, router ...gin.HandlerFunc) {
	l.RegisterInGroup("", method, relativePath, router...)
}
