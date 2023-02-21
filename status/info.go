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

func (o *sts) SetVersion(v libver.Version) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fn = v.GetPackage
	o.fr = v.GetRelease
	o.fh = v.GetBuild
	o.fd = v.GetTime
}
