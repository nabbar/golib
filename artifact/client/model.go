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
	"errors"
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
func (g *Helper) listReleasesOrderMajor() (map[int]hscvrs.Collection, error) {
	var (
		err error
		rel = make(map[int]hscvrs.Collection)
		vrs hscvrs.Collection
	)

	if g.F == nil {
		return rel, errors.New("invalid functions to retrieve list of releases")
	}

	if vrs, err = g.F(); err != nil {
		return rel, err
	}

	for _, v := range vrs {
		s := v.Segments()

		rel[s[0]] = append(rel[s[0]], v)
	}

	return rel, err
}

// ListReleasesOrder implements ArtHelper.ListReleasesOrder.
// Returns a nested map structure organizing versions by major and minor version numbers.
// Structure: map[major]map[minor]Collection
//
// Example:
//
//	{1: {0: [1.0.0, 1.0.1], 2: [1.2.0, 1.2.5]}, 2: {1: [2.1.3, 2.1.9]}}
func (g *Helper) ListReleasesOrder() (map[int]map[int]hscvrs.Collection, error) {
	var (
		err error
		rel = make(map[int]map[int]hscvrs.Collection)
		vrs map[int]hscvrs.Collection
	)

	if vrs, err = g.listReleasesOrderMajor(); err != nil {
		return rel, err
	}

	for major, list := range vrs {
		for _, v := range list {
			s := v.Segments()

			if len(rel[major]) == 0 {
				rel[major] = make(map[int]hscvrs.Collection)
			}

			rel[major][s[1]] = append(rel[major][s[1]], v)
		}
	}

	return rel, err
}

// ListReleasesMajor implements ArtHelper.ListReleasesMajor.
// Returns all versions with the specified major version number, sorted in ascending order.
// Returns an empty collection if the major version is not found.
func (g *Helper) ListReleasesMajor(major int) (hscvrs.Collection, error) {
	var (
		err error
		rel hscvrs.Collection
		vrs map[int]hscvrs.Collection
	)

	if vrs, err = g.listReleasesOrderMajor(); err != nil {
		return rel, err
	}

	if _, ok := vrs[major]; !ok {
		return rel, err
	} else if len(vrs[major]) > 0 {
		rel = vrs[major]
	}

	sort.Sort(rel)

	return rel, err
}

// ListReleasesMinor implements ArtHelper.ListReleasesMinor.
// Returns all versions matching the specified major and minor version numbers.
// The returned collection is sorted in ascending order.
// Returns an empty collection if the major/minor combination is not found.
func (g *Helper) ListReleasesMinor(major, minor int) (hscvrs.Collection, error) {
	var (
		ok bool

		err error
		rel hscvrs.Collection
		vrs map[int]map[int]hscvrs.Collection
	)

	if vrs, err = g.ListReleasesOrder(); err != nil {
		return rel, err
	} else if _, ok = vrs[major]; !ok {
		return rel, err
	} else if _, ok = vrs[major][minor]; !ok {
		return rel, err
	} else if len(vrs[major][minor]) > 0 {
		rel = vrs[major][minor]
	}

	sort.Sort(rel)

	return rel, err
}

// GetLatest implements ArtHelper.GetLatest.
// Returns the highest version across all major and minor versions.
// Determines the latest by finding the highest major version, then the highest minor
// within that major, and finally the highest patch version.
func (g *Helper) GetLatest() (*hscvrs.Version, error) {
	var (
		err error
		rel *hscvrs.Version
		vrs map[int]map[int]hscvrs.Collection
		maj int // major
		mnr int // minor
	)

	if vrs, err = g.ListReleasesOrder(); err != nil {
		return rel, err
	}

	for i := range vrs {
		if maj < i {
			maj = i
		}
	}

	for i := range vrs[maj] {
		if mnr < i {
			mnr = i
		}
	}

	return g.GetLatestMinor(maj, mnr)
}

// GetLatestMajor implements ArtHelper.GetLatestMajor.
// Returns the highest version within the specified major version number.
// First finds the highest minor version for the given major, then returns the
// highest patch version within that major.minor combination.
func (g *Helper) GetLatestMajor(major int) (*hscvrs.Version, error) {
	var (
		err error
		rel *hscvrs.Version
		vrs map[int]map[int]hscvrs.Collection
		mnr int // minor
	)

	if vrs, err = g.ListReleasesOrder(); err != nil {
		return rel, err
	} else if _, ok := vrs[major]; !ok {
		return rel, err
	}

	for i := range vrs[major] {
		if mnr < i {
			mnr = i
		}
	}

	return g.GetLatestMinor(major, mnr)
}

// GetLatestMinor implements ArtHelper.GetLatestMinor.
// Returns the highest version (by patch number) within the specified major.minor version.
// Iterates through all versions in the collection and returns the highest one.
func (g *Helper) GetLatestMinor(major, minor int) (*hscvrs.Version, error) {
	var (
		err error
		rel *hscvrs.Version
		vrs hscvrs.Collection
	)

	if vrs, err = g.ListReleasesMinor(major, minor); err != nil {
		return rel, err
	}

	for i := 0; i < len(vrs); i++ {
		if vrs[i] == nil {
			// continue
		} else if rel == nil {
			rel = vrs[i]
		} else if rel.LessThan(vrs[i]) {
			rel = vrs[i]
		}
	}

	return rel, err
}
