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

package status

import (
	"time"

	libver "github.com/nabbar/golib/version"
)

// SetInfo manually sets the application information displayed in status responses.
// This method stores the provided values as functions that return constant values.
// The build date is set to zero time when using this method.
//
// Use this method when not using the github.com/nabbar/golib/version package.
// For version package integration, use SetVersion instead.
//
// Parameters:
//   - name: application or service name (e.g., "MyAPI")
//   - release: version string (e.g., "v1.2.3", "1.0.0-beta")
//   - hash: build hash or commit identifier (e.g., "abc123def456")
//
// This method is thread-safe.
func (o *sts) SetInfo(name, release, hash string) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fn = func() string {
		return name
	}

	o.fr = func() string {
		return release
	}

	o.fh = func() string {
		return hash
	}

	o.fd = func() time.Time {
		return time.Time{}
	}
}

// SetVersion sets application information from a Version object.
// This is the preferred method when using github.com/nabbar/golib/version.
// It automatically extracts name, release, build hash, and build time.
//
// The Version object provides dynamic access to version information,
// which is useful if version data can change during runtime.
//
// Parameters:
//   - v: the Version object from github.com/nabbar/golib/version
//
// This method is thread-safe.
//
// See github.com/nabbar/golib/version for Version interface details.
func (o *sts) SetVersion(v libver.Version) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fn = v.GetPackage
	o.fr = v.GetRelease
	o.fh = v.GetBuild
	o.fd = v.GetTime
}
