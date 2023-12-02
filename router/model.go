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

package router

import (
	ginsdk "github.com/gin-gonic/gin"
)

type rtr struct {
	init func() *ginsdk.Engine
	list map[string][]itm
}

func (l *rtr) Handler(engine *ginsdk.Engine) {
	for grpRoute, grpList := range l.list {
		if grpRoute == EmptyHandlerGroup {
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

func (l *rtr) RegisterInGroup(group, method, relativePath string, router ...ginsdk.HandlerFunc) {
	if group == "" {
		group = EmptyHandlerGroup
	}

	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]itm, 0)
	}

	l.list[group] = append(l.list[group], itm{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

func (l *rtr) RegisterMergeInGroup(group, method, relativePath string, router ...ginsdk.HandlerFunc) {
	if group == "" {
		group = EmptyHandlerGroup
	}

	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]itm, 0)
	}

	// if same route exists for same relative path and same method, so replace router list
	for i, r := range l.list[group] {
		if !r.Same(method, relativePath) {
			continue
		} else {
			l.list[group][i].Merge(router...)
			return
		}
	}

	l.list[group] = append(l.list[group], itm{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

func (l *rtr) Register(method, relativePath string, router ...ginsdk.HandlerFunc) {
	l.RegisterInGroup("", method, relativePath, router...)
}

func (l *rtr) Engine() *ginsdk.Engine {
	if l.init != nil {
		return l.init()
	} else {
		return DefaultGinInit()
	}
}
