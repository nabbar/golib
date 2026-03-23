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

// fctGetName defines a function type that returns the application name.
type fctGetName func() string

// fctGetRelease defines a function type that returns the release version.
type fctGetRelease func() string

// fctGetHash defines a function type that returns the build hash.
type fctGetHash func() string

// fctGetDateBuild defines a function type that returns the build date and time.
type fctGetDateBuild func() time.Time

// sts is the internal implementation of the Status interface. It orchestrates
// health checking by maintaining application information, a monitor pool,
// configuration, and a status cache. It manages the aggregation of health
// statuses from various components based on defined control modes.
type sts struct {
	m sync.RWMutex            // Protects concurrent access to the struct's fields.
	p montps.FuncPool         // A provider function that returns the monitor pool.
	n FuncGetCfgCpt           // A provider function to retrieve components by key for dynamic config.
	r func() liberr.ReturnGin // A function that returns an error formatter for Gin responses.
	x libctx.Config[string]   // A thread-safe key-value store for configuration (e.g., return codes, global info, mandatory groups).
	c ch                      // The status cache, storing the last computed health status.

	// Functions to retrieve application information (name, release, hash, build date).
	// These are typically set once during application initialization.
	fn fctGetName
	fr fctGetRelease
	fh fctGetHash
	fd fctGetDateBuild
}

// checkFunc verifies that all required application information functions (name,
// release, hash, and date) are properly configured. These functions are crucial
// for populating the `Info` section of the status response.
// It returns true if all functions are non-nil.
func (o *sts) checkFunc() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.fn != nil && o.fr != nil && o.fh != nil && o.fd != nil
}

// IsCacheHealthy checks if the cached status is healthy, meaning its state is
// `OK` or `WARN`. This is a "tolerant" check.
// This method is extremely fast as it reads directly from a cache, incurring
// minimal overhead (<10ns). It is suitable for high-frequency probes like
// Kubernetes readiness checks.
func (o *sts) IsCacheHealthy() bool {
	return o.c.IsCache() >= monsts.Warn
}

// IsCacheStrictlyHealthy checks if the cached status is strictly `OK`.
// This is a "strict" check.
// This method is extremely fast as it reads directly from a cache, incurring
// minimal overhead (<10ns). It is suitable for high-frequency probes like
// Kubernetes liveness checks.
func (o *sts) IsCacheStrictlyHealthy() bool {
	return o.c.IsCache() == monsts.OK
}

// IsHealthy performs a live (non-cached) health check to determine if the overall
// system or a specific set of components are "healthy," meaning their status is
// either `OK` or `WARN`. This is a "tolerant" check.
// If component names are provided, the check is limited to those components.
// This method forces a re-evaluation of all relevant monitors.
//
// Parameters:
//   - name: An optional list of component names to check. If empty, checks all components.
//
// Returns:
//
//	`true` if the aggregated status is `OK` or `WARN`, `false` otherwise.
func (o *sts) IsHealthy(name ...string) bool {
	s, _ := o.getStatus(name...)
	return s >= monsts.Warn
}

// IsStrictlyHealthy performs a live (non-cached) health check to determine if the
// overall system or specific components are "strictly healthy," meaning their
// status is `OK`. This is a "strict" check.
// If component names are provided, the check is limited to those components.
// This method forces a re-evaluation of all relevant monitors.
//
// Parameters:
//   - name: An optional list of component names to check. If empty, checks all components.
//
// Returns:
//
//	`true` only if the aggregated status is `OK`, `false` otherwise.
func (o *sts) IsStrictlyHealthy(name ...string) bool {
	s, _ := o.getStrictStatus(name...)
	return s == monsts.OK
}

// getStatus computes the overall health status by evaluating all monitored components
// against the configured control modes (Ignore, Should, Must, AnyOf, Quorum).
//
// This method implements the core logic for status aggregation:
//   - It iterates through monitors, applying the corresponding control mode.
//   - `Should` mode can downgrade the status to `WARN` but not `KO`.
//   - `Must` mode can downgrade the status to `WARN` or `KO`.
//   - `AnyOf` and `Quorum` modes are evaluated for their respective groups.
//
// The final status is the "worst" status encountered, as permitted by the rules.
//
// Parameters:
//   - keys: An optional list of component names to include in the status computation.
//     If empty, all configured components are considered.
//
// Returns:
//
//	The computed `monsts.Status` and a message from the component that caused the degradation.
func (o *sts) getStatus(keys ...string) (monsts.Status, string) {
	stt := monsts.OK
	msg := ""
	ign := make([]string, 0) // Tracks components already processed as part of a group.

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		// Filter logic: if specific keys are requested, skip others.
		// Also skip components that have already been processed (e.g., as part of a group).
		if len(keys) > 0 && !slices.Contains(keys, name) {
			return true
		} else if len(ign) > 0 && slices.Contains(ign, name) {
			return true
		}

		v := val.Status()
		c := o.cfgGetMode(name) // Get the control mode for this component.

		switch c {
		case stsctr.Ignore:
			// Component is explicitly ignored, so we skip it.
			return true
		case stsctr.Should:
			// 'Should' mode: The component is desirable but not critical.
			// If it's KO or WARN, the global status becomes WARN (unless it's already KO).
			// It cannot cause a global KO.
			if v < monsts.Warn && stt > monsts.Warn {
				stt = monsts.Warn
				msg = val.Message()
			}
		case stsctr.Must:
			// 'Must' mode: The component is critical.
			// The global status adopts the status of this component if it's worse
			// than the current global status.
			// OK -> OK, WARN -> WARN, KO -> KO.
			if v < stt {
				stt = v
				msg = val.Message()
			}
		case stsctr.AnyOf, stsctr.Quorum:
			// Group modes: These require evaluating a set of components together.
			// First, we identify all members of the group.
			lst := o.cfgGetOne(name)
			sta := map[monsts.Status]uint8{monsts.KO: 0, monsts.Warn: 0, monsts.OK: 0}
			res := make([]string, 0)

			// Iterate through the pool to find all members of this group.
			o.MonitorWalk(func(nme string, val montps.Monitor) bool {
				if !slices.Contains(lst, nme) {
					return true
				}

				// Mark as ignored for the outer loop so we don't process them again individually.
				ign = append(ign, nme)

				// Tally the status.
				sta[val.Status()]++
				if s := val.Message(); len(s) > 0 {
					res = append(res, val.Message())
				}

				return true
			})

			msg = strings.Join(res, ", ")

			switch c {
			case stsctr.AnyOf:
				// 'AnyOf' mode: At least one component must be healthy.
				if sta[monsts.OK] > 0 {
					// At least one is OK, so the group is OK. Global status is unaffected.
				} else if sta[monsts.Warn] > 0 && stt > monsts.Warn {
					// No OK, but at least one WARN. Global status becomes WARN (if not already KO).
					stt = monsts.Warn
				} else {
					// All components are KO. Global status becomes KO.
					stt = monsts.KO
				}
			case stsctr.Quorum:
				// 'Quorum' mode: More than 50% of components must be healthy.
				// Calculate the threshold (half of the total count).
				ct := (sta[monsts.OK] + sta[monsts.Warn] + sta[monsts.KO]) / 2
				if i := sta[monsts.OK]; i > ct {
					// Majority are OK. Global status unaffected.
				} else if i += sta[monsts.Warn]; i > ct && stt > monsts.Warn {
					// Majority are OK or WARN. Global status becomes WARN (if not already KO).
					stt = monsts.Warn
				} else {
					// Majority are KO. Global status becomes KO.
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

// getStrictStatus computes the overall health status by checking all specified
// components without applying any control modes. In this mode, any component
// that is not `OK` will cause the overall status to be downgraded. This is a
// "strict" evaluation, where only a perfect `OK` status is acceptable.
//
// Parameters:
//   - keys: An optional list of component names to include in the status computation.
//     If empty, all configured components are considered.
//
// Returns:
//
//	The "worst" `monsts.Status` found among the checked components and its corresponding message.
func (o *sts) getStrictStatus(keys ...string) (monsts.Status, string) {
	stt := monsts.OK
	msg := ""
	ign := make([]string, 0) // Tracks components already processed as part of a group.

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if len(keys) > 0 && !slices.Contains(keys, name) {
			return true
		} else if len(ign) > 0 && slices.Contains(ign, name) {
			return true
		}

		v := val.Status()

		// Strict check: simply take the worst status encountered.
		// If any component is WARN, global becomes WARN.
		// If any component is KO, global becomes KO.
		if v < stt {
			stt = v
			msg = val.Message()
		}

		return true
	})

	return stt, msg
}
