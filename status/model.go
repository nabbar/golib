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
	"slices"
	"strings"
	"sync"
	"time"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	stsctr "github.com/nabbar/golib/status/control"
)

// fctGetName is a function type that returns the application name.
type fctGetName func() string

// fctGetRelease is a function type that returns the release version.
type fctGetRelease func() string

// fctGetHash is a function type that returns the build hash.
type fctGetHash func() string

// fctGetDateBuild is a function type that returns the build date/time.
type fctGetDateBuild func() time.Time

// sts is the internal implementation of the Status interface.
// It maintains application information, monitor pool, configuration, and cache.
type sts struct {
	m sync.RWMutex            // Protects concurrent access to fields
	p montps.FuncPool         // Function returning the monitor pool
	r func() liberr.ReturnGin // Function returning error formatter
	x libctx.Config[string]   // Configuration storage
	c ch                      // Status cache

	// Application information functions
	fn fctGetName
	fr fctGetRelease
	fh fctGetHash
	fd fctGetDateBuild
}

// checkFunc verifies that all required application information functions are set.
// Returns true if all functions (name, release, hash, date) are configured.
func (o *sts) checkFunc() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.fn != nil && o.fr != nil && o.fh != nil && o.fd != nil
}

// IsCacheHealthy checks if the cached status is healthy (OK or Warn).
// This method uses the cached status value if available and not expired.
func (o *sts) IsCacheHealthy() bool {
	return o.c.IsCache() >= monsts.Warn
}

// IsCacheStrictlyHealthy checks if the cached status is strictly OK (no warnings).
// This method uses the cached status value if available and not expired.
func (o *sts) IsCacheStrictlyHealthy() bool {
	return o.c.IsCache() == monsts.OK
}

// IsHealthy checks if the overall status or specific components are healthy.
// Returns true if status is OK or Warn (>= Warn threshold).
// If component names are provided, only those components are checked.
func (o *sts) IsHealthy(name ...string) bool {
	s, _ := o.getStatus(name...)
	return s >= monsts.Warn
}

// IsStrictlyHealthy checks if the overall status or specific components are strictly healthy.
// Returns true only if status is OK (no warnings or errors).
// If component names are provided, only those components are checked.
func (o *sts) IsStrictlyHealthy(name ...string) bool {
	s, _ := o.getStrictStatus(name...)
	return s == monsts.OK
}

// getStatus computes the overall status by walking through all monitored components.
// It applies control modes (Ignore, Should, AnyOf, Quorum) defined in configuration.
//
// Control modes:
//   - Ignore: component is skipped
//   - Should: component warning downgrades overall status to Warn (not KO)
//   - AnyOf: at least one component in group must be OK
//   - Quorum: majority (>50%) of components in group must be OK/Warn
//
// Parameters:
//   - keys: optional component names to check; if empty, checks all components
//
// Returns the computed status and an associated message from the worst component.
func (o *sts) getStatus(keys ...string) (monsts.Status, string) {
	stt := monsts.OK
	msg := ""
	ign := make([]string, 0)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if len(keys) > 0 && !slices.Contains(keys, name) {
			return true
		} else if len(ign) > 0 && slices.Contains(ign, name) {
			return true
		}

		v := val.Status()
		c := o.cfgGetMode(name)

		switch c {
		case stsctr.Ignore:
			return true
		case stsctr.Should:
			// default stt is ok, problem only if KO (should)
			// should cannot set global status to KO
			if v < monsts.Warn && stt > monsts.Warn {
				stt = monsts.Warn
				msg = val.Message()
			}
		case stsctr.Must:
			// if global status if better than any must
			// so global status must having value status
			if v < stt {
				stt = v
				msg = val.Message()
			}
		case stsctr.AnyOf, stsctr.Quorum:
			lst := o.cfgGetOne(name)
			sta := map[monsts.Status]uint8{monsts.KO: 0, monsts.Warn: 0, monsts.OK: 0}
			res := make([]string, 0)

			o.MonitorWalk(func(nme string, val montps.Monitor) bool {
				if !slices.Contains(lst, nme) {
					return true
				}

				ign = append(ign, nme)

				sta[val.Status()]++
				if s := val.Message(); len(s) > 0 {
					res = append(res, val.Message())
				}

				return true
			})

			msg = strings.Join(res, ", ")

			switch c {
			case stsctr.AnyOf:
				if sta[monsts.OK] > 0 {
					// no problem
				} else if sta[monsts.Warn] > 0 && stt > monsts.Warn {
					// warn only if global status is OK
					stt = monsts.Warn
				} else {
					// block is KO, so global status if KO
					stt = monsts.KO
				}
			case stsctr.Quorum:
				// total KO + Warm + OK
				// quorum is at least more than 50% of list
				ct := (sta[monsts.OK] + sta[monsts.Warn] + sta[monsts.KO]) / 2
				if i := sta[monsts.OK]; i > ct {
					// no problem so don't touch global status
				} else if i += sta[monsts.Warn]; i > ct && stt > monsts.Warn {
					// number of OK + warm is more than 50% of full list
					// so if global status is better than warm (OK), set to warm
					stt = monsts.Warn
				} else {
					// so more KO than OK + Warm, so set global status to KO
					stt = monsts.KO
				}
			default:
				// nothing
			}
		default:
			// nothing
		}

		return true
	})

	return stt, msg
}

// getStrictStatus computes the overall status by walking through all monitored components.
// It does not applies control modes (Ignore, Should, AnyOf, Quorum) defined in configuration.
//
// Parameters:
//   - keys: optional component names to check; if empty, checks all components
//
// Returns the computed status and an associated message from the worst component.
func (o *sts) getStrictStatus(keys ...string) (monsts.Status, string) {
	stt := monsts.OK
	msg := ""
	ign := make([]string, 0)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if len(keys) > 0 && !slices.Contains(keys, name) {
			return true
		} else if len(ign) > 0 && slices.Contains(ign, name) {
			return true
		}

		v := val.Status()

		// if global status if better than any must
		// so global status must having value status
		if v < stt {
			stt = v
			msg = val.Message()
		}

		return true
	})

	return stt, msg
}
