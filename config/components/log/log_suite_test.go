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

package log_test

import (
	"context"
	"testing"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	monpol "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kd   = "logger"
	fp   montps.FuncPool
	fl   liblog.FuncLog
	x, n = context.WithCancel(context.Background())

	v  = libvpr.New(x, fl)
	fv = func() libvpr.Viper {
		return v
	}

	vs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
)

var _ = BeforeSuite(func() {
	p := monpol.New(x)
	fp = func() montps.Pool {
		return p
	}

	l := liblog.New(x)
	Expect(l.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})).NotTo(HaveOccurred())
	fl = func() liblog.Logger {
		return l
	}
})

var _ = AfterSuite(func() {
	n()
})

// TestLog runs the Ginkgo test suite for the Log component package.
// This suite tests logger configuration management, component lifecycle,
// and integration with the logging system.
//
// Test coverage includes:
//   - Component lifecycle (Init, Start, Reload, Stop)
//   - Configuration management and validation
//   - Logger creation and manipulation
//   - Log level management (Get/Set)
//   - Logger fields management
//   - Logger options management
//   - Default configuration handling
//   - Error conditions and edge cases
//   - Concurrent access scenarios
//   - Integration with logger package
//   - Flag registration for CLI
//
// The tests use standalone implementations without external dependencies
// to avoid billing or security issues. All tests are designed to be
// human-readable and maintainable with separate files per scope.
//
// Run tests with:
//
//	go test -v
//	go test -v -cover
//	CGO_ENABLED=1 go test -v -race
//
// For detailed coverage:
//
//	go test -v -coverprofile=coverage.out
//	go tool cover -html=coverage.out
func TestLog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Log Component Suite")
}
