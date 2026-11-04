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
	"sync"
)

type rdr struct {
	r func(p []byte) (int, error)
	c func() error
}

func (o *rdr) Read(p []byte) (int, error) {
	return o.r(p)
}

func (o *rdr) Close() error {
	return o.c()
}

func (o *mdl) Reader(id uint8) io.ReadCloser {
	o.sendSYN(id)
	return &rdr{
		r: func(p []byte) (int, error) {
			return o.Read(id, p)
		},
		c: func() error {
			o.sendFIN(id)
			return nil
		},
	}
}

func (o *mdl) Read(id uint8, pay []byte) (int, error) {
	if id > 0 && !o.vi.open(id) {
		return 0, io.ErrClosedPipe
	}

	n, e := o.s.Read(id, pay)
	o.cntRdrAdd(n)

	if e != nil && e != io.EOF {
		o.logError(strWriting, e)
	}

	return n, e
}

func (o *mdl) readLoop(r io.Reader) {
	var (
		ch = o.setChanPacket()
		wg sync.WaitGroup
	)

	go func() {
		defer close(ch)
		defer o.n()

		o.readPacket(ch, r)
	}()

	for {
		select {
		case <-o.x.Done():
			return
		case pkt := <-ch:
			if pkt.i == 0 {
				// steam 0 : control stream, priority
				o.parseReadCmd(pkt)
				continue
			}

			if o.rs.Load() {
				// RSYN: trash all packets for other stream
				o.putPacket(pkt)
				continue
			}

			if !o.vi.open(pkt.i) {
				o.putPacket(pkt)
				continue
			}

			// other stream : delay statement whit io.pipe
			wg.Go(func() {
				defer o.putPacket(pkt)

				o.sendACK(pkt.id())
				pn, pe := o.s.Write(pkt.i, pkt.p)

				if uint64(pn) != uint64(pkt.l) {
					o.logError(strReading, fmt.Errorf("packet length mismatch on writing to stream: %d != %d", pn, pkt.l))
				}

				if pe != nil {
					o.logError("error on writing packet to stream", pe)
				}
			})
		}
	}
}
