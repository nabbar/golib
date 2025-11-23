/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package static

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"reflect"
	"strings"

	ginsdk "github.com/gin-gonic/gin"
	ginrdr "github.com/gin-gonic/gin/render"
	liberr "github.com/nabbar/golib/errors"
	loglvl "github.com/nabbar/golib/logger/level"
	librtr "github.com/nabbar/golib/router"
	rtrhdr "github.com/nabbar/golib/router/header"

	_ "github.com/ugorji/go/codec"
)

func (s *staticHandler) makeRoute(group, route string) string {
	if group == "" {
		group = urlPathSeparator
	}
	return path.Join(group, route)
}

func (s *staticHandler) genRegRouter(route, group string, register any, router ...ginsdk.HandlerFunc) {
	var (
		ok  bool
		rte string
		reg librtr.RegisterRouter
		grp librtr.RegisterRouterInGroup
	)

	if register == nil {
		return
	} else if grp, ok = register.(librtr.RegisterRouterInGroup); ok {
		rte = s.makeRoute(group, route)
		reg = nil
	} else if reg, ok = register.(librtr.RegisterRouter); ok {
		rte = s.makeRoute(urlPathSeparator, route)
		grp = nil
	} else {
		return
	}

	if len(router) > 0 {
		router = append(router, s.Get)
	} else {
		router = append(make([]ginsdk.HandlerFunc, 0), s.Get)
	}

	if rtr := s.getRouter(); len(rtr) > 0 {
		s.setRouter(append(rtr, rte))
	} else {
		s.setRouter(append(make([]string, 0), rte))
	}

	if reg != nil {
		reg(http.MethodGet, path.Clean(route), router...)
		reg(http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
	}

	if grp != nil {
		grp(group, http.MethodGet, path.Clean(route), router...)
		grp(group, http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
	}
}

func (s *staticHandler) RegisterRouter(route string, register librtr.RegisterRouter, router ...ginsdk.HandlerFunc) {
	s.genRegRouter(route, "/", register, router...)
}

func (s *staticHandler) RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...ginsdk.HandlerFunc) {
	s.genRegRouter(route, group, register, router...)
}

// Get is the main HTTP handler for serving static files.
// It implements the complete request flow including:
//   - Rate limiting
//   - Path security validation
//   - Redirects
//   - Custom handlers
//   - Index files
//   - ETag caching
//   - Suspicious access detection
//   - MIME type validation
func (s *staticHandler) Get(c *ginsdk.Context) {
	// Check rate limiting first
	if !s.checkRateLimit(c) {
		return // Rate limit exceeded, 429 response already sent
	}

	calledFile := c.Request.URL.Path

	// Validate path security
	if err := s.validatePath(calledFile); err != nil {
		ent := s.getLogger().Entry(loglvl.WarnLevel, "Get Static file, path validation failed")
		ent.FieldAdd("requestPath", calledFile)
		ent.ErrorAdd(false, err)
		ent.Log()

		// Notify WAF/IDS/EDR
		s.notifyPathSecurityEvent(c, EventTypePathTraversal, err.Error())

		// Log suspicious access before returning
		s.checkAndLogSuspicious(c, http.StatusForbidden)

		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if dest := s.GetRedirect("", calledFile); dest != "" {
		url := c.Request.URL
		url.Path = dest

		ent := s.getLogger().Entry(loglvl.DebugLevel, "Get redirect to url")
		ent.FieldAdd("requestPath", calledFile)
		ent.FieldAdd("redirectPath", dest)
		ent.Log()

		c.Redirect(http.StatusPermanentRedirect, url.String())
		return
	}

	if router := s.GetSpecific("", calledFile); router != nil {
		ent := s.getLogger().Entry(loglvl.DebugLevel, "Get call specific")
		ent.FieldAdd("requestPath", calledFile)
		ent.FieldAdd("called", reflect.ValueOf(router).String())
		ent.Log()

		router(c)
		return
	}

	// Check if an index file is configured for this route
	if idx := s.GetIndex("", calledFile); idx != "" {
		ent := s.getLogger().Entry(loglvl.DebugLevel, "Get call index")
		ent.FieldAdd("requestPath", calledFile)
		ent.FieldAdd("index", idx)
		ent.Log()

		// Use the index file directly, no further path processing needed
		calledFile = idx
	} else {
		// Normal file path processing
		for _, p := range s.getRouter() {
			if p == urlPathSeparator {
				continue
			}
			calledFile = strings.TrimPrefix(calledFile, p)
		}
		calledFile = strings.Trim(calledFile, urlPathSeparator)

		ent := s.getLogger().Entry(loglvl.DebugLevel, "Get call index")
		ent.FieldAdd("requestPath", c.Request.URL.Path)
		ent.FieldAdd("cleanedPath", calledFile)
		ent.Log()
	}

	if !s.Has(calledFile) {
		old := calledFile

		for _, p := range s.getBase() {
			f := path.Join(p, calledFile)

			if s.Has(f) {
				calledFile = f
				break
			}
		}

		if old != calledFile {
			ent := s.getLogger().Entry(loglvl.DebugLevel, "Get search file")
			ent.FieldAdd("requestPath", c.Request.URL.Path)
			ent.FieldAdd("oldCalledPath", old)
			ent.FieldAdd("newCalledPath", calledFile)
			ent.Log()
		}
	}

	if !s.Has(calledFile) {
		ent := s.getLogger().Entry(loglvl.WarnLevel, "Get cannot find file")
		ent.FieldAdd("requestPath", c.Request.URL.Path)
		ent.FieldAdd("cleanedPath", calledFile)
		ent.Log()

		// Log suspicious access
		s.checkAndLogSuspicious(c, http.StatusNotFound)

		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var (
		err error
		buf io.ReadCloser
		inf fs.FileInfo
	)

	inf, buf, err = s.fileGet(calledFile)
	defer func() {
		if buf != nil {
			_ = buf.Close()
		}
	}()

	if err != nil {
		// Check if it'spc a "not found" error (including directories without index)
		if liberr.Has(err, ErrorFileNotFound) {
			ent := s.getLogger().Entry(loglvl.WarnLevel, "file not found or directory without index")
			ent.FieldAdd("requestPath", c.Request.URL.Path)
			ent.FieldAdd("cleanedPath", calledFile)
			ent.ErrorAdd(false, err)
			ent.Log()

			// Log suspicious access
			s.checkAndLogSuspicious(c, http.StatusNotFound)

			c.AbortWithStatus(http.StatusNotFound)
		} else {
			ent := s.getLogger().Entry(loglvl.ErrorLevel, "get file error")
			ent.FieldAdd("requestPath", c.Request.URL.Path)
			ent.FieldAdd("cleanedPath", calledFile)
			ent.ErrorAdd(true, err)
			ent.Log()

			// Log suspicious access
			s.checkAndLogSuspicious(c, http.StatusInternalServerError)

			c.AbortWithStatus(http.StatusInternalServerError)
		}
	} else if inf != nil && inf.IsDir() {
		// This should not happen anymore, but keep as safety net
		ent := s.getLogger().Entry(loglvl.WarnLevel, "directory access without index")
		ent.FieldAdd("requestPath", c.Request.URL.Path)
		ent.FieldAdd("cleanedPath", calledFile)
		ent.Log()

		c.AbortWithStatus(http.StatusNotFound)
	} else {
		ent := s.getLogger().Entry(loglvl.DebugLevel, "get file info")
		ent.FieldAdd("requestPath", c.Request.URL.Path)
		ent.FieldAdd("cleanedPath", calledFile)
		ent.FieldAdd("pathSize", inf.Size())
		ent.Log()

		// Check ETag and potentially return 304 Not Modified
		if s.setETagHeader(c, calledFile, inf.Size(), inf.ModTime()) {
			// Cache hit - client already has correct version
			ent = s.getLogger().Entry(loglvl.DebugLevel, "cache hit - returning 304")
			ent.FieldAdd("requestPath", c.Request.URL.Path)
			ent.FieldAdd("cleanedPath", calledFile)
			ent.Log()

			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// Log suspicious access (even for successful requests)
		s.checkAndLogSuspicious(c, http.StatusOK)

		s.SendFile(c, calledFile, inf.Size(), s.IsDownload(calledFile), buf)
	}
}

// SendFile sends a file to the client with appropriate headers.
// This method:
//   - Validates and sets Content-Type
//   - Sets cache headers
//   - Handles downloads (Content-Disposition)
//   - Streams the file content
func (s *staticHandler) SendFile(c *ginsdk.Context, filename string, size int64, isDownload bool, buf io.ReadCloser) {
	head := rtrhdr.NewHeaders()

	// Check and set Content-Type
	mimeType, err := s.setContentTypeHeader(c, filename)
	if err != nil {
		// MIME type denied
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Set cache headers
	s.setCacheHeaders(c)

	if isDownload {
		head.Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%spc\"", path.Base(filename)))
	}

	c.Render(http.StatusOK, ginrdr.Reader{
		ContentLength: size,
		ContentType:   mimeType,
		Headers:       head.Header(),
		Reader:        buf,
	})
}
