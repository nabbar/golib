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

package muxtun

import (
	"encoding/binary"
	"time"
)

type ctrl uint8

const (
	cmdSYN ctrl = iota + 1
	cmdSYNACK
	cmdACK
	cmdFIN
	cmdFINACK
	cmdRSYN
	cmdRACK
	cmdPING
	cmdPONG
)

func ctrlFromUint32(i uint32) ctrl {
	switch i {
	case cmdSYN.uint32():
		return cmdSYN
	case cmdSYNACK.uint32():
		return cmdSYNACK
	case cmdACK.uint32():
		return cmdACK
	case cmdFIN.uint32():
		return cmdFIN
	case cmdFINACK.uint32():
		return cmdFINACK
	case cmdRSYN.uint32():
		return cmdRSYN
	case cmdRACK.uint32():
		return cmdRACK
	case cmdPING.uint32():
		return cmdPING
	case cmdPONG.uint32():
		return cmdPONG
	}

	return 0
}

func ctrlFromUint8(i uint8) ctrl {
	return ctrlFromUint32(uint32(i))
}

func (c ctrl) uint8() uint8 {
	return uint8(c)
}

func (c ctrl) uint32() uint32 {
	return uint32(c)
}

func (o *mdl) parseReadCmd(p *packet) {
	defer o.putPacket(p)

	switch ctrlFromUint8(p.p[0]) {
	case cmdSYN:
		o.sendSYNACK(p.p[1:])
	case cmdSYNACK:
		o.readSYNACK(p.p[1:])
	case cmdACK:
		o.readACK(p.p[1:])
	case cmdFIN:
		o.sendFINACK(p.p[1:])
	case cmdFINACK:
		o.readFINACK(p.p[1:])
	case cmdRSYN:
		o.sendRACK()
	case cmdRACK:
		o.ao.clear()
		o.rs.Store(false)
	case cmdPING:
		o.sendPONG(p.p[1:])
	case cmdPONG:
		o.readPONG(p.p[1:])
	default:
		// do nothing
	}
}

func (o *mdl) sendSYN(id ...uint8) {
	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	lng := min(len(id), maxSizeBuffer-1)

	for lng > 0 {
		clear(*buf)
		pkt.clear()

		for i := 0; i < lng; i++ {
			(*buf)[i] = id[i]
			o.vo.add(id[i])
			o.s.new(id[i])
		}

		if e := pkt.sendCmdBuf(cmdSYN, (*buf)[:lng]); e != nil {
			o.sendRSYN()
			return
		}

		clear(*buf)
		if _, e := pkt.write(o.w, buf); e != nil {
			o.sendRSYN()
		}

		if lng < len(id) {
			id = id[lng:]
		} else {
			id = id[:0]
		}

		lng = min(len(id), maxSizeBuffer-1)
	}
}

func (o *mdl) sendSYNACK(p []byte) {
	for i := range p {
		o.vo.add(p[i])
		o.vo.state(p[i], true)
	}

	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdBuf(cmdSYNACK, p); e != nil {
		o.sendRSYN()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		o.sendRSYN()
	}
}

func (o *mdl) readSYNACK(p []byte) {
	for i := range p {
		o.vi.state(p[i], true)
	}
}

func (o *mdl) sendACK(id ...uint64) {
	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdUint64(cmdACK, id...); e != nil {
		o.sendRSYN()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		o.sendRSYN()
	}
}

func (o *mdl) readACK(p []byte) {
	for len(p) > 0 {
		o.ao.del(binary.BigEndian.Uint64(p[0:8]))
		p = p[8:]
	}
}

func (o *mdl) sendFIN(id ...uint8) {
	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	lng := min(len(id), maxSizeBuffer-1)

	for lng > 0 {
		clear(*buf)
		pkt.clear()

		for i := 0; i < lng; i++ {
			(*buf)[i] = id[i]
			o.vo.state(id[i], false)
			o.s.del(id[i])
		}

		if e := pkt.sendCmdBuf(cmdFIN, (*buf)[:lng]); e != nil {
			o.sendRSYN()
			return
		}

		clear(*buf)
		if _, e := pkt.write(o.w, buf); e != nil {
			o.sendRSYN()
		}

		if lng < len(id) {
			id = id[lng:]
		} else {
			id = id[:0]
		}

		lng = min(len(id), maxSizeBuffer-1)
	}
}

func (o *mdl) sendFINACK(p []byte) {
	for i := range p {
		o.vo.delete(p[i])
	}

	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdBuf(cmdFINACK, p); e != nil {
		o.sendRSYN()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		o.sendRSYN()
	}
}

func (o *mdl) readFINACK(p []byte) {
	for i := range p {
		o.vi.delete(p[i])
	}
}

func (o *mdl) sendRSYN() {
	o.rs.Store(true)

	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdUint64(cmdRSYN); e != nil {
		_ = o.Close()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		_ = o.Close()
		return
	}

	o.ao.clear()
}

func (o *mdl) sendRACK() {
	o.rs.Store(true)

	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdUint64(cmdRACK); e != nil {
		_ = o.Close()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		_ = o.Close()
		return
	}

	o.ao.clear()
}

func (o *mdl) sendPING() {
	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	tms := time.Now()

	if e := pkt.sendCmdUint64(cmdPING, uint64(tms.Unix()), uint64(tms.UnixNano())); e != nil {
		o.sendRSYN()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		o.sendRSYN()
		return
	}
}

func (o *mdl) sendPONG(p []byte) {
	buf := o.getBuffer(maxSizeBuffer)
	pkt := o.getPacket()

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if e := pkt.sendCmdBuf(cmdPONG, p); e != nil {
		o.sendRSYN()
		return
	}

	if _, e := pkt.write(o.w, buf); e != nil {
		o.sendRSYN()
		return
	}
}

func (o *mdl) readPONG(p []byte) {
	tms := binary.BigEndian.Uint64(p[0:8])
	tmn := binary.BigEndian.Uint64(p[8:16])
	tmu := time.Unix(int64(tms), int64(tmn))
	dur := time.Since(tmu)

	if dur > 0 {
		o.cl.Store(uint64(dur))
	} else {
		o.cl.Store(uint64(-dur))
	}
}
