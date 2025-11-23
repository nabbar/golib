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

package static_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	ginsdk "github.com/gin-gonic/gin"
	libdur "github.com/nabbar/golib/duration"
	libfpg "github.com/nabbar/golib/file/progress"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	montps "github.com/nabbar/golib/monitor/types"
	librtr "github.com/nabbar/golib/router"
	"github.com/nabbar/golib/static"
	libver "github.com/nabbar/golib/version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testdata/*
var testFS embed.FS

var (
	// Global context for all tests
	testCtx    context.Context
	testCancel context.CancelFunc

	// Test gin engine
	testGinEngine *ginsdk.Engine

	// Global logger for tests
	testLogger liblog.Logger
	testAccLog liblog.Logger
)

type staticDownload interface {
	SetDownload(string, bool)
	IsDownload(string) bool
}

type staticIndex interface {
	SetIndex(string, string, string)
	GetIndex(string, string) string
	IsIndex(string) bool
	IsIndexForRoute(string, string, string) bool
}

type staticRedirect interface {
	SetRedirect(string, string, string, string)
	GetRedirect(string, string) string
	IsRedirect(string, string) bool
}

type staticSpecific interface {
	SetSpecific(string, string, ginsdk.HandlerFunc)
	GetSpecific(string, string) ginsdk.HandlerFunc
}

type staticConfig interface {
	SetDownload(string, bool)
	SetIndex(string, string, string)
	IsDownload(string) bool
	IsIndex(string) bool
	GetIndex(string, string) string
}

type staticMap interface {
	Map(func(string, os.FileInfo) error) error
}

type staticFind interface {
	Find(string) (io.ReadCloser, error)
}

type staticList interface {
	List(rootPath string) ([]string, error)
}

type staticInfo interface {
	Info(string) (os.FileInfo, error)
}

type staticTemp interface {
	Temp(string) (libfpg.Progress, error)
}

type staticTempSize interface {
	UseTempForFileSize(int64)
}

type staticFindHas interface {
	Find(string) (io.ReadCloser, error)
	Has(string) bool
}

type staticFindTempSize interface {
	Find(string) (io.ReadCloser, error)
	UseTempForFileSize(int64)
}

// TestStatic is the entry point for the Ginkgo test suite
func TestStatic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Static Package Suite")
}

var _ = BeforeSuite(func() {
	testCtx, testCancel = context.WithCancel(context.Background())
	ginsdk.SetMode(ginsdk.TestMode)

	// Initialize logger
	testLogger = liblog.New(testCtx)
	Expect(testLogger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true, // Enable log output for security evt
		},
	})).ToNot(HaveOccurred())

	// Initialize logger
	testAccLog = liblog.New(testCtx)
	Expect(testAccLog.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			EnableAccessLog: true,
			DisableStandard: true, // Enable log output for access logs
		},
	})).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Close access logger
	if testAccLog != nil {
		_ = testAccLog.Close()
	}

	// Close logger
	if testLogger != nil {
		_ = testLogger.Close()
	}

	if testCancel != nil {
		testCancel()
	}
})

// Helper functions

// newTestVersion creates a test version instance for use in tests
func newTestVersion() libver.Version {
	return libver.NewVersion(
		libver.License_MIT,
		"static-test",
		"Static Package Test",
		"2024",
		"test-build",
		"1.0.0",
		"test-author",
		"",
		struct{}{},
		0,
	)
}

// newTestMonitorConfig creates a test monitor config instance
func newTestMonitorConfig() montps.Config {
	return montps.Config{
		Name:          "test",
		CheckTimeout:  libdur.Seconds(1),
		IntervalCheck: libdur.Seconds(1),
		IntervalFall:  libdur.Seconds(1),
		IntervalRise:  libdur.Seconds(1),
		FallCountKO:   1,
		FallCountWarn: 1,
		RiseCountKO:   1,
		RiseCountWarn: 1,
		Logger: logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		},
	}
}

// getTestLogger returns the global test logger
func getTestLogger() liblog.Logger {
	return testLogger
}

// getTestLogger returns the global test logger
func getTestAccessLogger() liblog.Logger {
	return testAccLog
}

// newTestStatic creates a new static handler with test data
func newTestStatic() static.Static {
	s := static.New(testCtx, testFS, "testdata")
	s.RegisterLogger(testLogger)
	return s
}

// newTestStaticWithRoot creates a new static handler with custom root
func newTestStaticWithRoot(roots ...string) static.Static {
	s := static.New(testCtx, testFS, roots...)
	s.RegisterLogger(testLogger)
	return s
}

// setupTestRouter creates a gin router with static handler registered
func setupTestRouter(handler static.Static, route string, middlewares ...ginsdk.HandlerFunc) *ginsdk.Engine {
	n, e := librtr.GinEngine("")
	Expect(e).NotTo(HaveOccurred())

	// Disable automatic trailing slash redirects for tests
	n.RedirectTrailingSlash = false

	n = librtr.GinAddGlobalMiddleware(n, librtr.GinAccessLog(func() liblog.Logger {
		return getTestAccessLogger()
	}))

	n = librtr.GinAddGlobalMiddleware(n, librtr.GinErrorLog(func() liblog.Logger {
		return getTestLogger()
	}))

	routerList := librtr.NewRouterList(func() *ginsdk.Engine {
		return n
	})

	handler.RegisterRouter(route, routerList.Register, middlewares...)

	routerList.Handler(n)

	return n
}

// setupTestRouterInGroup creates a gin router with static handler registered in a group
func setupTestRouterInGroup(handler static.Static, route, group string, middlewares ...ginsdk.HandlerFunc) *ginsdk.Engine {
	routerList := librtr.NewRouterList(librtr.DefaultGinInit)

	handler.RegisterRouterInGroup(route, group, routerList.RegisterInGroup, middlewares...)

	engine := routerList.Engine()
	// Disable automatic trailing slash redirects for tests
	engine.RedirectTrailingSlash = false
	routerList.Handler(engine)

	return engine
}

// performRequest performs an HTTP request and returns the response
func performRequest(engine *ginsdk.Engine, method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	Expect(err).ToNot(HaveOccurred())

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	return w
}

// testMiddleware is a simple middleware for testing
func testMiddleware(ctx *ginsdk.Context) {
	ctx.Header("X-Test-Middleware", "true")
	ctx.Next()
}

func newMiddleware(id int) ginsdk.HandlerFunc {
	var h = "X-Test-Middleware"

	if id > 0 {
		h = h + "-" + strconv.Itoa(id)
	}

	return func(ctx *ginsdk.Context) {
		ctx.Header(h, "true")
		ctx.Next()
	}
}

func customMiddlewareOK(response string, fct func()) ginsdk.HandlerFunc {
	if fct == nil {
		fct = func() {}
	}
	return func(c *ginsdk.Context) {
		fct()
		c.String(http.StatusOK, response)
	}
}

func customMiddlewareCreated(response string) ginsdk.HandlerFunc {
	return func(c *ginsdk.Context) {
		c.String(http.StatusCreated, response)
	}
}
