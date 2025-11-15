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

	hscvrs "github.com/hashicorp/go-version"
)

// Helper wraps a function that returns version collections and provides
// convenient methods for organizing, filtering, and retrieving versions.
// This struct is embedded in all platform-specific implementations (GitHub, GitLab, JFrog, S3).
//
// Example usage:
//
//	h := &Helper{F: client.ListReleases}
//	latest, err := h.GetLatest()          // Get latest overall version
//	v2Latest, err := h.GetLatestMajor(2)  // Get latest v2.x.x version
type Helper struct {
	F func() (releases hscvrs.Collection, err error)
}

// listReleasesOrderMajor organizes releases by major version number.
// Returns a map where keys are major version numbers and values are collections of versions.
// This is an internal helper method used by other organization methods.
func (g *Helper) listReleasesOrderMajor() (releases map[int]hscvrs.Collection, err error) {
	var (
		vers hscvrs.Collection
	)

	if vers, err = g.F(); err != nil {
		return
	}

	for _, v := range vers {
		s := v.Segments()

		if releases == nil {
			releases = make(map[int]hscvrs.Collection)
		}

		releases[s[0]] = append(releases[s[0]], v)
	}

	return
}

// ListReleasesOrder implements ArtHelper.ListReleasesOrder.
// Returns a nested map structure organizing versions by major and minor version numbers.
// Structure: map[major]map[minor]Collection
//
// Example:
//
//	{1: {0: [1.0.0, 1.0.1], 2: [1.2.0, 1.2.5]}, 2: {1: [2.1.3, 2.1.9]}}
func (g *Helper) ListReleasesOrder() (releases map[int]map[int]hscvrs.Collection, err error) {
	var (
		vers map[int]hscvrs.Collection
	)

	if vers, err = g.listReleasesOrderMajor(); err != nil {
		return
	}

	for major, list := range vers {
		for _, v := range list {
			s := v.Segments()

			if releases == nil {
				releases = make(map[int]map[int]hscvrs.Collection)
			}

			if releases[major] == nil || len(releases[major]) == 0 {
				releases[major] = make(map[int]hscvrs.Collection)
			}

			releases[major][s[1]] = append(releases[major][s[1]], v)
		}
	}

	return
}

// ListReleasesMajor implements ArtHelper.ListReleasesMajor.
// Returns all versions with the specified major version number, sorted in ascending order.
// Returns an empty collection if the major version is not found.
func (g *Helper) ListReleasesMajor(major int) (releases hscvrs.Collection, err error) {
	var (
		vers map[int]hscvrs.Collection
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

// ListReleasesMinor implements ArtHelper.ListReleasesMinor.
// Returns all versions matching the specified major and minor version numbers.
// The returned collection is sorted in ascending order.
// Returns an empty collection if the major/minor combination is not found.
func (g *Helper) ListReleasesMinor(major, minor int) (releases hscvrs.Collection, err error) {
	var (
		vers map[int]map[int]hscvrs.Collection
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

// GetLatest implements ArtHelper.GetLatest.
// Returns the highest version across all major and minor versions.
// Determines the latest by finding the highest major version, then the highest minor
// within that major, and finally the highest patch version.
func (g *Helper) GetLatest() (release *hscvrs.Version, err error) {
	var (
		vers  map[int]map[int]hscvrs.Collection
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

// GetLatestMajor implements ArtHelper.GetLatestMajor.
// Returns the highest version within the specified major version number.
// First finds the highest minor version for the given major, then returns the
// highest patch version within that major.minor combination.
func (g *Helper) GetLatestMajor(major int) (release *hscvrs.Version, err error) {
	var (
		vers  map[int]map[int]hscvrs.Collection
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

// GetLatestMinor implements ArtHelper.GetLatestMinor.
// Returns the highest version (by patch number) within the specified major.minor version.
// Iterates through all versions in the collection and returns the highest one.
func (g *Helper) GetLatestMinor(major, minor int) (release *hscvrs.Version, err error) {
	var (
		vers hscvrs.Collection
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
