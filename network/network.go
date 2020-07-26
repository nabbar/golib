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
		return nil, NET_INTERFACE.ErrorParent(e)
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
		return nil, NET_INTERFACE.ErrorParent(e)
	} else {
		for _, f := range l {
			if (name != "" && f.Name == name) || (physical != "" && physical == f.HardwareAddr) {
				ifs = f //nosec
				break
			}
		}

		if ifs.Name == "" {
			return nil, NET_NOTFOUND.Error(nil)
		}
	}

	if l, e := netlib.IOCountersWithContext(ctx, true); e != nil {
		return nil, NET_COUNTER.ErrorParent(e)
	} else {
		for _, f := range l {
			if f.Name == ifs.Name {
				ifc = f
				break
			}
		}

		if ifc.Name == "" {
			return nil, NET_NOTFOUND.Error(nil)
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

	return NET_RELOAD.Error(nil)
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
