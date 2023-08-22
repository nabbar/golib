/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package client

import (
	hscvrs "github.com/hashicorp/go-version"
	liberr "github.com/nabbar/golib/errors"
)

type ArtifactManagement interface {
	ListReleasesOrder() (releases map[int]map[int]hscvrs.Collection, err liberr.Error)
	ListReleasesMajor(major int) (releases hscvrs.Collection, err liberr.Error)
	ListReleasesMinor(major, minor int) (releases hscvrs.Collection, err liberr.Error)

	GetLatest() (release *hscvrs.Version, err liberr.Error)
	GetLatestMajor(major int) (release *hscvrs.Version, err liberr.Error)
	GetLatestMinor(major, minor int) (release *hscvrs.Version, err liberr.Error)
}
