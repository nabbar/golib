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

type encComponent interface {
	MarshalText() (text []byte, err error)
	MarshalJSON() ([]byte, error)
}

type encMod struct {
	Mode stsctr.Mode                     `json:"mode"`
	Sts  monsts.Status                   `json:"status"`
	Cpt  map[string]montps.MonitorStatus `json:"components"`
}

type modControl struct {
	ctr stslmd.ListMandatory
	cpt montps.Pool
	fct func(name ...string) (monsts.Status, string)
}

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

func (o modControl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.getEncControl())
}

func (o modControl) getEncControl() []encMod {
	var (
		res = make([]encMod, 0)
		fnd = make([]string, 0)
		ign = make([]string, 0)
	)

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

		ctr.Sts, _ = o.fct(m.KeyList()...)

		res = append(res, ctr)
		return true
	})

	o.cpt.MonitorWalk(func(name string, val montps.Monitor) bool {
		for _, v := range fnd {
			if v == name {
				return true
			}
		}
		ign = append(ign, name)
		return true
	})

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
	}

	return res
}
