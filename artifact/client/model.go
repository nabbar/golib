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
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/nabbar/golib/errors"
)

type ClientHelper struct {
	F func() (releases version.Collection, err errors.Error)
}

func (g *ClientHelper) listReleasesOrderMajor() (releases map[int]version.Collection, err errors.Error) {
	var (
		vers version.Collection
	)

	if vers, err = g.F(); err != nil {
		return
	}

	for _, v := range vers {
		s := v.Segments()

		if releases == nil {
			releases = make(map[int]version.Collection)
		}

		releases[s[0]] = append(releases[s[0]], v)
	}

	return
}

func (g *ClientHelper) ListReleasesOrder() (releases map[int]map[int]version.Collection, err errors.Error) {
	var (
		vers map[int]version.Collection
	)

	if vers, err = g.listReleasesOrderMajor(); err != nil {
		return
	}

	for major, list := range vers {
		for _, v := range list {
			s := v.Segments()

			if releases == nil {
				releases = make(map[int]map[int]version.Collection)
			}

			if releases[major] == nil || len(releases[major]) == 0 {
				releases[major] = make(map[int]version.Collection)
			}

			releases[major][s[1]] = append(releases[major][s[1]], v)
		}
	}

	return
}

func (g *ClientHelper) ListReleasesMajor(major int) (releases version.Collection, err errors.Error) {
	var (
		vers map[int]version.Collection
	)

	if vers, err = g.listReleasesOrderMajor(); err != nil {
		return
	}

	if _, ok := vers[major]; !ok {
		return
	} else if len(vers[major]) > 0 {
		releases = vers[major]
	}

	sort.Sort(releases)

	return
}

func (g *ClientHelper) ListReleasesMinor(major, minor int) (releases version.Collection, err errors.Error) {
	var (
		vers map[int]map[int]version.Collection
	)

	if vers, err = g.ListReleasesOrder(); err != nil {
		return
	}

	if _, ok := vers[major]; !ok {
		return
	}

	if _, ok := vers[major][minor]; !ok {
		return
	} else if len(vers[major][minor]) > 0 {
		releases = vers[major][minor]
	}

	sort.Sort(releases)

	return
}

func (g *ClientHelper) GetLatest() (release *version.Version, err errors.Error) {
	var (
		vers  map[int]map[int]version.Collection
		major int
		minor int
	)

	if vers, err = g.ListReleasesOrder(); err != nil {
		return
	}

	for i := range vers {
		if major < i {
			major = i
		}
	}

	for i := range vers[major] {
		if minor < i {
			minor = i
		}
	}

	return g.GetLatestMinor(major, minor)
}

func (g *ClientHelper) GetLatestMajor(major int) (release *version.Version, err errors.Error) {
	var (
		vers  map[int]map[int]version.Collection
		minor int
	)

	if vers, err = g.ListReleasesOrder(); err != nil {
		return
	}

	if _, ok := vers[major]; !ok {
		return
	}

	for i := range vers[major] {
		if minor < i {
			minor = i
		}
	}

	return g.GetLatestMinor(major, minor)
}

func (g *ClientHelper) GetLatestMinor(major, minor int) (release *version.Version, err errors.Error) {
	var (
		vers version.Collection
	)

	if vers, err = g.ListReleasesMinor(major, minor); err != nil {
		return
	}

	for i := 0; i < len(vers); i++ {
		if vers[i] == nil {
			continue
		}

		if release == nil || release.LessThan(vers[i]) {
			release = vers[i]
			continue
		}
	}

	return
}
