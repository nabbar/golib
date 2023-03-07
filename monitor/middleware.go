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

package monitor

import (
	"context"

	"github.com/nabbar/golib/monitor/types"
)

type fctMiddleWare func(m middleWare) error

type middleWare interface {
	Context() context.Context
	Config() *runCfg
	Run(ctx context.Context)
	Next() error
	Add(fct fctMiddleWare)
}

type mdl struct {
	ctx context.Context
	cfg *runCfg
	crs int
	mdl []fctMiddleWare
}

func newMiddleware(cfg *runCfg, fct types.HealthCheck) middleWare {
	o := &mdl{
		ctx: nil,
		cfg: cfg,
		crs: 0,
		mdl: make([]fctMiddleWare, 0),
	}

	o.Add(func(m middleWare) error {
		return fct(m.Context())
	})

	return o
}

func (m *mdl) Context() context.Context {
	return m.ctx
}

func (m *mdl) Config() *runCfg {
	return m.cfg
}

func (m *mdl) Run(ctx context.Context) {
	var cnl context.CancelFunc

	m.ctx, cnl = context.WithTimeout(ctx, m.cfg.checkTimeout)
	defer cnl()

	m.crs = len(m.mdl)
	_ = m.Next()
}

func (m *mdl) Next() error {
	m.crs--

	if m.crs >= 0 && m.crs < len(m.mdl) {
		return m.mdl[m.crs](m)
	}

	return nil
}

func (m *mdl) Add(fct fctMiddleWare) {
	m.mdl = append(m.mdl, fct)
}
