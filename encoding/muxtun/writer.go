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
	"fmt"
	"io"
	"math"
)

type wrt struct {
	w func(p []byte) (int, error)
}

func (o *wrt) Write(p []byte) (int, error) {
	return o.w(p)
}

func (o *mdl) Writer(id uint8) io.Writer {
	return &wrt{
		w: func(p []byte) (int, error) {
			return o.Write(id, p)
		},
	}
}

func (o *mdl) Write(id uint8, pay []byte) (int, error) {
	var (
		lng int
		lnp = len(pay)
		buf = o.getBuffer(maxSizeBuffer)
		pkt = o.getPacket()
	)

	defer o.putBuffer(buf)
	defer o.putPacket(pkt)

	if id > 0 && !o.vo.open(id) {
		return 0, io.ErrClosedPipe
	}

	pkt.i = id

	for lnp > 0 {
		if n, e := pkt.encode(id, pay); e != nil {
			return lng, e
		} else if n < maxSizePayload {
			pay = pay[:0]
		} else {
			pay = pay[n:]
		}

		n, e := pkt.write(o.w, buf)
		o.cntWrtAdd(n)

		if e != nil {
			o.logError(strWriting, e)
			return lng + n, e
		}

		if !pkt.check(n) {
			o.logError(strWriting, fmt.Errorf("packet length mismatch: %d != %d", n, pkt.l))
			return lng + n, e
		}

		if uint64(pkt.l) > math.MaxInt-uint64(lng) {
			o.logError(strWriting, ErrPayloadTooBig)
			return math.MaxInt, ErrPayloadTooBig
		}

		// inc total writ
		lng += int(pkt.l)
		lnp -= int(pkt.l)

		o.ao.add(pkt)
		clear(*buf)
		pkt.clear()
	}

	return lng, nil
}
