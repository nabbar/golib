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
 *
 */

package static

import (
	"embed"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
	liblog "github.com/nabbar/golib/logger"
	librtr "github.com/nabbar/golib/router"
	libsts "github.com/nabbar/golib/status"
)

type Static interface {
	RegisterRouter(route string, register librtr.RegisterRouter, router ...gin.HandlerFunc)
	RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...gin.HandlerFunc)

	RegisterLogger(log func() liblog.Logger)

	SetDownload(pathFile string, flag bool)
	SetIndex(group, route, pathFile string)
	GetIndex(group, route string) string
	SetRedirect(srcGroup, srcRoute, dstGroup, dstRoute string)
	GetRedirect(srcGroup, srcRoute string) string
	SetSpecific(group, route string, router gin.HandlerFunc)
	GetSpecific(group, route string) gin.HandlerFunc

	IsDownload(pathFile string) bool
	IsIndex(pathFile string) bool
	IsIndexForRoute(pathFile, group, route string) bool
	IsRedirect(group, route string) bool

	Has(pathFile string) bool
	List(rootPath string) ([]string, liberr.Error)
	Find(pathFile string) (io.ReadCloser, liberr.Error)
	Info(pathFile string) (os.FileInfo, liberr.Error)
	Temp(pathFile string) (libiot.FileProgress, liberr.Error)

	Map(func(pathFile string, inf os.FileInfo) error) liberr.Error
	UseTempForFileSize(size int64)

	StatusInfo() (name string, release string, hash string)
	StatusHealth() error
	StatusComponent(mandatory bool, message libsts.FctMessage, infoCacheTimeout, healthCacheTimeout time.Duration, sts libsts.RouteStatus)

	Get(c *gin.Context)
	SendFile(c *gin.Context, filename string, size int64, isDownload bool, buf io.ReadCloser)
}

func New(content embed.FS, embedRootDir ...string) Static {
	return &staticHandler{
		m: sync.Mutex{},
		l: nil,
		c: content,
		b: embedRootDir,
		z: 0,
		i: nil,
		d: nil,
		f: nil,
		r: nil,
	}
}
