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
)

// ArtHelper provides version organization and retrieval methods.
// This interface is embedded in the main artifact.Client interface and provides
// hierarchical access to versions organized by major/minor version numbers.
//
// All methods automatically filter out pre-release versions (alpha, beta, rc, etc.)
// and return only stable, production-ready releases.
type ArtHelper interface {
	// ListReleasesOrder returns a two-level map of releases organized by major and minor versions.
	// Structure: map[major]map[minor]Collection
	//
	// Example result:
	//   {
	//     1: {0: [1.0.0, 1.0.1], 2: [1.2.0, 1.2.1, 1.2.5]},
	//     2: {0: [2.0.0], 1: [2.1.3, 2.1.9]}
	//   }
	ListReleasesOrder() (releases map[int]map[int]hscvrs.Collection, err error)

	// ListReleasesMajor returns a sorted slice of versions with the given major version number.
	// The slice is sorted in ascending order by minor version number.
	//
	// Example: ListReleasesMajor(1) will return a slice of all versions with major version number 1,
	// sorted by minor version number in ascending order.
	//
	// It returns an error if the given major version number is not found in the list of releases.
	//
	// The returned slice is a subset of the results returned by ListReleasesOrder().
	ListReleasesMajor(major int) (releases hscvrs.Collection, err error)

	// ListReleasesMinor returns a sorted slice of versions with the given major and minor version numbers.
	//
	// Example: ListReleasesMinor(1, 2) will return a slice of all versions with major version number 1 and minor version number 2,
	// sorted by version in ascending order.
	//
	// It returns an error if the given major and minor version numbers are not found in the list of releases.
	//
	// The returned slice is a subset of the results returned by ListReleasesOrder().
	ListReleasesMinor(major, minor int) (releases hscvrs.Collection, err error)

	// GetLatest returns the highest version in the list of releases.
	// The version is sorted by major version number in descending order, and by minor version number in descending order.
	// If the list of releases is empty, it returns an error.
	//
	// Example: GetLatest() will return the highest version in the list of releases, sorted by major version number in descending order, and by minor version number in descending order.
	//
	// It returns an error if the list of releases is empty.
	GetLatest() (release *hscvrs.Version, err error)

	// GetLatestMajor returns the highest version in the list of releases with the given major version number.
	//
	// The version is sorted by minor version number in descending order.
	// If the list of releases with the given major version number is empty, it returns an error.
	//
	// Example: GetLatestMajor(1) will return the highest version in the list of releases with major version number 1,
	// sorted by minor version number in descending order.
	//
	// It returns an error if the list of releases with the given major version number is empty.
	GetLatestMajor(major int) (release *hscvrs.Version, err error)

	// GetLatestMinor returns the highest version in the list of releases with the given major and minor version numbers.
	//
	// The version is sorted by version in descending order.
	// If the list of releases with the given major and minor version numbers is empty, it returns an error.
	//
	// Example: GetLatestMinor(1, 2) will return the highest version in the list of releases with major version number 1 and minor version number 2,
	// sorted by version in descending order.
	//
	// It returns an error if the list of releases with the given major and minor version numbers is empty.
	GetLatestMinor(major, minor int) (release *hscvrs.Version, err error)
}
