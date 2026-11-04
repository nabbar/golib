//go:build amd64 || arm64

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
	"io"
)

const (
	// use this constant to forge packet header + data like this :
	// +-----------++--------------++----------------++------------+
	// | stream ID | packet length | packet checksum | packet data |
	// |   1 byte  |    4 bytes    |     4 bytes     |   1MB max   |
	// +-----------++--------------++----------------++------------+
	//

	// header end index
	maxPosId       = 1
	maxPosLength   = maxPosId + 1<<2
	maxPosChecksum = maxPosLength + 1<<2

	// size
	maxSizeHeader  = maxPosChecksum
	maxSizePayload = 1 << 20
	maxSizeBuffer  = maxSizeHeader + maxSizePayload
)

type packet struct {
	p []byte
	h []byte
	c uint32
	l uint32
	i uint8
}

func (o *packet) sendCmdUint64(c ctrl, v ...uint64) error {
	var (
		e error
		p = 1
	)

	o.i = 0
	o.p[0] = c.uint8()

	for _, u := range v {
		binary.BigEndian.PutUint64(o.p[p:p+8], u)
		p += 8
	}

	if o.c, e = checksum(o.p); e != nil {
		return e
	}

	return nil
}

func (o *packet) sendCmdBuf(c ctrl, b []byte) error {
	var e error

	if len(b) > maxSizePayload-1 {
		return ErrPayloadTooBig
	}

	o.i = 0
	o.p[0] = c.uint8()
	copy(o.p[1:], b)

	if o.c, e = checksum(o.p); e != nil {
		return e
	}

	return nil
}

func (o *packet) encode(i uint8, p []byte) (n int, e error) {
	o.i = i

	if l := len(p); l > maxSizePayload {
		o.l = maxSizePayload
	} else {
		o.l = uint32(l)
	}

	copy(o.p[:o.l], p[:o.l])

	if o.c, e = checksum(o.p); e != nil {
		return 0, e
	}

	return int(o.l), nil
}

func (o *packet) write(w io.Writer, b *[]byte) (int, error) {
	(*b)[maxPosId-1] = o.i
	binary.BigEndian.PutUint32((*b)[maxPosId:maxPosLength], o.l)
	binary.BigEndian.PutUint32((*b)[maxPosLength:maxPosChecksum], o.c)
	copy((*b)[maxPosChecksum:], o.p)

	return w.Write(*b)
}

func (o *packet) read(r io.Reader) error {
	if _, e := io.ReadFull(r, o.h); e != nil {
		return e
	}

	o.i = o.h[maxPosId-1]
	o.l = binary.BigEndian.Uint32(o.h[maxPosId:maxPosLength])
	o.c = binary.BigEndian.Uint32(o.h[maxPosLength:maxPosChecksum])

	if _, e := io.ReadFull(r, o.p); e != nil {
		return e
	}

	if c, e := checksum(o.p); e != nil {
		return e
	} else if c != o.c {
		return ErrInvalidChecksum
	}

	return nil
}

func (o *packet) len() int {
	if o.l > 0 {
		return int(o.l) + maxSizeHeader
	}

	return 0
}

func (o *packet) clear() {
	clear(o.p)
	clear(o.h)
}

func (o *packet) check(n int) bool {
	if n < 0 {
		return false
	}

	return uint64(n) != maxSizeHeader+uint64(o.l)
}

func (o *packet) id() uint64 {
	return uint64(o.i)<<32 | uint64(o.c)
}

func (o *mdl) getPacket() *packet {
	var pCpy *packet

	// Attempt to retrieve a recycled pointer from the pool.
	if v := o.p.Get(); v != nil {
		pCpy = v.(*packet)
		pCpy.clear()
		return pCpy
	}

	// Fallback to fresh allocation if the pool is empty or the retrieved buffer is nil.
	return &packet{
		i: 0,
		l: 0,
		c: 0,
		h: make([]byte, maxSizeHeader),
		p: make([]byte, maxSizePayload),
	}
}

func (o *mdl) putPacket(buf *packet) {
	o.p.Put(buf)
}

func (o *mdl) setChanPacket() chan *packet {
	return make(chan *packet)
}

func (o *mdl) readPacket(c chan *packet, r io.Reader) {
	var e error

	for o.x.Err() == nil && e == nil {
		var pkt = o.getPacket()

		if e = pkt.read(r); e != nil {
			if e != io.EOF {
				o.logError(strReading, e)
			}

			continue
		}

		o.cntRdrAdd(pkt.len())
		c <- pkt
	}
}
