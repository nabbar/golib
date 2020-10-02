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

package network

import (
	"context"
	"net"

	"github.com/nabbar/golib/errors"
	netlib "github.com/shirou/gopsutil/net"
)

type Interface interface {
	ReLoad(ctx context.Context) errors.Error

	GetIndex() int
	GetName() string
	GetHardwareAddr() string
	GetMTU() int
	GetAddr() []string

	IsUp() bool
	HasPhysical() bool

	GetFlags() []string

	GetStatIn() map[Stats]Number
	GetStatOut() map[Stats]Number
}

type iface struct {
	ifs *netlib.InterfaceStat
	ifc *netlib.IOCountersStat
}

func GetAllInterfaces(ctx context.Context, onlyPhysical, hasAddr bool, atLeastMTU int, withFlags ...net.Flags) ([]Interface, errors.Error) {
	var res = make([]string, 0)

	if l, e := netlib.InterfacesWithContext(ctx); e != nil {
		return nil, ErrorNetInterface.ErrorParent(e)
	} else {
		for _, f := range l {
			if onlyPhysical && f.HardwareAddr == "" {
				continue
			}

			if hasAddr && len(f.Addrs) < 1 {
				continue
			}

			if atLeastMTU > 0 && f.MTU < atLeastMTU {
				continue
			}

			if len(withFlags) > 0 && !FindAllFlagInList(f.Flags, withFlags) {
				continue
			}

			res = append(res, f.Name)
		}
	}

	var r = make([]Interface, 0)

	for _, n := range res {
		if i, e := NewInterface(ctx, n, ""); e != nil {
			return nil, e
		} else {
			r = append(r, i)
		}
	}

	return r, nil
}

func NewInterface(ctx context.Context, name, physical string) (Interface, errors.Error) {
	var (
		ifs netlib.InterfaceStat
		ifc netlib.IOCountersStat
	)

	if l, e := netlib.InterfacesWithContext(ctx); e != nil {
		return nil, ErrorNetInterface.ErrorParent(e)
	} else {
		for _, f := range l {
			if (name != "" && f.Name == name) || (physical != "" && physical == f.HardwareAddr) {
				ifs = f //nosec
				break
			}
		}

		if ifs.Name == "" {
			return nil, ErrorNetNotFound.Error(nil)
		}
	}

	if l, e := netlib.IOCountersWithContext(ctx, true); e != nil {
		return nil, ErrorNetCounter.ErrorParent(e)
	} else {
		for _, f := range l {
			if f.Name == ifs.Name {
				ifc = f
				break
			}
		}

		if ifc.Name == "" {
			return nil, ErrorNetNotFound.Error(nil)
		}
	}

	return &iface{
		ifs: &ifs,
		ifc: &ifc,
	}, nil
}

func (i *iface) ReLoad(ctx context.Context) errors.Error {
	if c, e := NewInterface(ctx, i.ifs.Name, ""); e != nil {
		return e
	} else if p, ok := c.(*iface); ok && p != nil {
		i.ifs = p.ifs
		i.ifc = p.ifc
		return nil
	}

	return ErrorNetReload.Error(nil)
}

func (i iface) GetIndex() int {
	return i.ifs.Index
}

func (i iface) GetName() string {
	return i.ifs.Name
}

func (i iface) GetHardwareAddr() string {
	return i.ifs.HardwareAddr
}

func (i iface) GetMTU() int {
	return i.ifs.MTU
}

func (i iface) GetAddr() []string {
	var a = make([]string, 0)

	for _, f := range i.ifs.Addrs {
		a = append(a, f.String())
	}

	return a
}

func (i iface) IsUp() bool {
	for _, f := range i.ifs.Flags {
		if f == net.FlagUp.String() {
			return true
		}
	}

	return false
}

func (i iface) HasPhysical() bool {
	return i.ifs.HardwareAddr != ""
}

func (i iface) GetFlags() []string {
	return i.ifs.Flags
}

func (i iface) GetStatIn() map[Stats]Number {
	var r = make(map[Stats]Number)

	r[StatBytes] = Number(i.ifc.BytesRecv)
	r[StatPackets] = Number(i.ifc.PacketsRecv)
	r[StatFifo] = Number(i.ifc.Fifoin)
	r[StatDrop] = Number(i.ifc.Dropin)
	r[StatErr] = Number(i.ifc.Errin)

	return r
}

func (i iface) GetStatOut() map[Stats]Number {
	var r = make(map[Stats]Number)

	r[StatBytes] = Number(i.ifc.BytesSent)
	r[StatPackets] = Number(i.ifc.PacketsSent)
	r[StatFifo] = Number(i.ifc.Fifoout)
	r[StatDrop] = Number(i.ifc.Dropout)
	r[StatErr] = Number(i.ifc.Errout)

	return r
}
