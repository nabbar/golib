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

// rtr is the internal implementation of RouterList interface.
// It maintains a map of route groups to their respective route items.
type rtr struct {
	init func() *ginsdk.Engine // Function to initialize a new Gin engine
	list map[string][]itm      // Map of group names to route items
}

// Handler applies all registered routes to the given Gin engine.
// Routes without a group (EmptyHandlerGroup) are registered at root level.
// Grouped routes are registered within their respective Gin route groups.
func (l *rtr) Handler(engine *ginsdk.Engine) {
	for grpRoute, grpList := range l.list {
		if grpRoute == EmptyHandlerGroup {
			// Register routes at root level
			for _, r := range grpList {
				engine.Handle(r.method, r.relative, r.router...)
			}
		} else {
			// Register routes within a group
			var grp = engine.Group(grpRoute)
			for _, r := range grpList {
				grp.Handle(r.method, r.relative, r.router...)
			}
		}
	}
}

// RegisterInGroup adds a route to the specified group.
// If group is empty, the route is registered without a group (root level).
// Multiple routes can be registered with the same method and path in the same group.
func (l *rtr) RegisterInGroup(group, method, relativePath string, router ...ginsdk.HandlerFunc) {
	if group == "" {
		group = EmptyHandlerGroup
	}

	// Initialize group list if it doesn't exist
	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]itm, 0)
	}

	// Append new route to group
	l.list[group] = append(l.list[group], itm{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

// RegisterMergeInGroup adds or replaces a route in the specified group.
// If a route with the same method and relative path already exists in the group,
// its handlers are replaced with the new ones. Otherwise, a new route is added.
// This is useful for overriding default routes or updating handlers dynamically.
func (l *rtr) RegisterMergeInGroup(group, method, relativePath string, router ...ginsdk.HandlerFunc) {
	if group == "" {
		group = EmptyHandlerGroup
	}

	// Initialize group list if it doesn't exist
	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]itm, 0)
	}

	// Check if same route exists and merge handlers
	for i, r := range l.list[group] {
		if !r.Same(method, relativePath) {
			continue
		} else {
			// Replace existing handlers
			l.list[group][i].Merge(router...)
			return
		}
	}

	// No existing route found, add new one
	l.list[group] = append(l.list[group], itm{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

// Register adds a route at root level (without a group).
// This is a convenience method that calls RegisterInGroup with an empty group.
func (l *rtr) Register(method, relativePath string, router ...ginsdk.HandlerFunc) {
	l.RegisterInGroup("", method, relativePath, router...)
}

// Engine returns a new Gin engine instance.
// If an init function was provided during creation, it is used.
// Otherwise, DefaultGinInit is called to create a default engine.
func (l *rtr) Engine() *ginsdk.Engine {
	if l.init != nil {
		return l.init()
	} else {
		return DefaultGinInit()
	}
}
