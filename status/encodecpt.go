/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package status

import (
	"bytes"
	"encoding/json"

	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	stsctr "github.com/nabbar/golib/status/control"
	stslmd "github.com/nabbar/golib/status/listmandatory"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

// encComponent defines an interface for objects that can be marshaled into
// both text and JSON formats. This is used to handle different component
// encoding strategies (e.g., list vs. map).
type encComponent interface {
	MarshalText() (text []byte, err error)
	MarshalJSON() ([]byte, error)
}

// encMod represents a single control mode group for encoding purposes. It contains
// the mode, the overall status of the group, and a map of the components within it.
// This structure is used when "map mode" is enabled.
type encMod struct {
	Mode stsctr.Mode                     `json:"mode"`
	Sts  monsts.Status                   `json:"status"`
	Cpt  map[string]montps.MonitorStatus `json:"components"`
}

// modControl is a helper struct that encapsulates the necessary information for
// encoding component statuses in "map mode". It holds the mandatory component
// list, the monitor pool, and a function to compute group status.
type modControl struct {
	ctr stslmd.ListMandatory
	cpt montps.Pool
	fct func(name ...string) (monsts.Status, string)
}

// MarshalText implements the `encoding.TextMarshaler` interface. It generates a
// plain text representation of the component groups, their modes, and their statuses.
//
// Format:
//
//	STATUS - Mode:
//	    Component1: Status - Message
//	    Component2: Status - Message
func (o modControl) MarshalText() (text []byte, err error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	for _, i := range o.getEncControl() {
		if p, e := i.Mode.MarshalText(); e != nil {
			return nil, e
		} else {
			buf.WriteString(i.Sts.String() + " - ")
			buf.Write(p)
			buf.WriteRune(':')
			buf.WriteRune('\n')
		}
		for _, c := range i.Cpt {
			if p, e := c.MarshalText(); e != nil {
				return nil, e
			} else {
				buf.WriteRune('\t')
				buf.Write(p)
				buf.WriteRune('\n')
			}
		}
	}

	return buf.Bytes(), nil
}

// MarshalJSON implements the `json.Marshaler` interface. It generates a JSON
// representation of the component groups.
func (o modControl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.getEncControl())
}

// getEncControl is the core logic for organizing components into their respective
// control groups for encoding. It walks through the mandatory list, groups
// components by their control mode, and then adds any remaining (ignored)
// components into a final "Ignore" group.
//
// This function ensures that every monitored component appears exactly once in
// the output, grouped by its effective validation mode.
func (o modControl) getEncControl() []encMod {
	var (
		res = make([]encMod, 0)
		fnd = make([]string, 0)
		ign = make([]string, 0)
	)

	// Group components based on the mandatory configuration.
	o.ctr.Walk(func(m stsmdt.Mandatory) bool {
		var ctr = encMod{
			Mode: m.GetMode(),
			Cpt:  make(map[string]montps.MonitorStatus),
		}

		for _, c := range m.KeyList() {
			fnd = append(fnd, c)
			if cpt := o.cpt.MonitorGet(c); cpt == nil {
				continue
			} else {
				ctr.Cpt[c] = cpt
			}
		}

		if len(ctr.Cpt) > 0 {
			ctr.Sts, _ = o.fct(m.KeyList()...)
			res = append(res, ctr)
		}

		return true
	})

	// Find any components that were not part of a mandatory group.
	o.cpt.MonitorWalk(func(name string, val montps.Monitor) bool {
		for _, v := range fnd {
			if v == name {
				return true
			}
		}
		ign = append(ign, name)
		return true
	})

	// Add the remaining components to an "Ignore" group.
	if len(ign) > 0 {
		var ctr = encMod{
			Mode: stsctr.Ignore,
			Sts:  monsts.OK,
			Cpt:  make(map[string]montps.MonitorStatus),
		}

		for _, c := range ign {
			fnd = append(fnd, c)
			if cpt := o.cpt.MonitorGet(c); cpt == nil {
				continue
			} else {
				ctr.Cpt[c] = cpt
			}
		}

		if len(ctr.Cpt) > 0 {
			res = append(res, ctr)
		}
	}

	return res
}
