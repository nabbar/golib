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
	"context"
	"io/fs"
	"runtime"

	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

const (
	textEmbed = "Embed FS"
)

// Monitor creates and returns a health monitor for the static file handler.
// This integrates with github.com/nabbar/golib/monitor for health checks and status reporting.
//
// The monitor provides:
//   - Runtime version information
//   - Build metadata
//   - Filesystem health checks
//
// See github.com/nabbar/golib/monitor/types for configuration details.
func (s *staticHandler) Monitor(ctx context.Context, cfg montps.Config, vrs libver.Version) (montps.Monitor, error) {
	res := make(map[string]interface{}, 0)
	res["runtime"] = runtime.Version()[2:]
	res["release"] = vrs.GetRelease()
	res["build"] = vrs.GetBuild()
	res["date"] = vrs.GetDate()

	var (
		e   error
		i   fs.FileInfo
		inf moninf.Info
		mon montps.Monitor
	)

	if inf, e = moninf.New(textEmbed); e != nil {
		return nil, e
	} else {
		inf.RegisterName(func() (string, error) {
			return textEmbed, nil
		})
	}

	// Try to get filesystem info from the first base path if available
	basePaths := s.getBase()
	if len(basePaths) > 0 {
		if i, e = s.fileInfo(basePaths[0]); e == nil {
			res["path"] = i.Name()
		}
	}

	// Always register info with at least runtime and version data
	inf.RegisterInfo(func() (map[string]interface{}, error) {
		return res, nil
	})

	if mon, e = libmon.New(s.dwn, inf); e != nil {
		return nil, e
	} else if e = mon.SetConfig(ctx, cfg); e != nil {
		return nil, e
	} else {
		mon.SetHealthCheck(s.HealthCheck)
		if e = mon.Start(ctx); e != nil {
			return nil, e
		}
	}

	return mon, nil
}

// HealthCheck performs a health check on the embedded filesystem.
// It verifies that all base paths are accessible.
// Returns an error if any base path cannot be accessed.
func (s *staticHandler) HealthCheck(ctx context.Context) error {
	for _, p := range s.getBase() {
		if _, err := s.fileInfo(p); err != nil {
			return err
		}
	}

	return nil
}
