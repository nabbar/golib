/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package config_test

import (
	"bytes"
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	shlcmd "github.com/nabbar/golib/shell/command"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

type fakeComponent struct {
	key           string
	deps          []string
	started       bool
	running       bool
	startCount    int
	reloadCount   int
	stopCount     int
	onStartBefore cfgtps.FuncCptEvent
	onStartAfter  cfgtps.FuncCptEvent
	onRelBefore   cfgtps.FuncCptEvent
	onRelAfter    cfgtps.FuncCptEvent
}

func (f *fakeComponent) Type() string { return "fake" }
func (f *fakeComponent) Init(key string, _ context.Context, _ cfgtps.FuncCptGet, _ libvpr.FuncViper, _ libver.Version, _ liblog.FuncLog) {
	f.key = key
}
func (f *fakeComponent) DefaultConfig(indent string) []byte {
	return []byte("{\n" + indent + "\"enabled\": true\n}")
}
func (f *fakeComponent) Dependencies() []string { return append([]string{}, f.deps...) }
func (f *fakeComponent) SetDependencies(d []string) error {
	f.deps = append([]string{}, d...)
	return nil
}
func (f *fakeComponent) RegisterFlag(_ *spfcbr.Command) error  { return nil }
func (f *fakeComponent) RegisterMonitorPool(_ montps.FuncPool) {}
func (f *fakeComponent) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {
	f.onStartBefore, f.onStartAfter = before, after
}
func (f *fakeComponent) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {
	f.onRelBefore, f.onRelAfter = before, after
}
func (f *fakeComponent) IsStarted() bool { return f.started }
func (f *fakeComponent) IsRunning() bool { return f.running }
func (f *fakeComponent) Start() error {
	if f.onStartBefore != nil {
		_ = f.onStartBefore(f)
	}
	f.started = true
	f.running = true
	f.startCount++
	if f.onStartAfter != nil {
		_ = f.onStartAfter(f)
	}
	return nil
}
func (f *fakeComponent) Reload() error {
	if f.onRelBefore != nil {
		_ = f.onRelBefore(f)
	}
	f.started = true
	f.reloadCount++
	if f.onRelAfter != nil {
		_ = f.onRelAfter(f)
	}
	return nil
}
func (f *fakeComponent) Stop() { f.running = false; f.started = false; f.stopCount++ }

var _ = Describe("config", func() {
	var c libcfg.Config
	var a, b *fakeComponent

	BeforeEach(func() {
		c = libcfg.New(nil)
		a = &fakeComponent{deps: []string{"b"}}
		b = &fakeComponent{}
		c.ComponentSet("b", b)
		c.ComponentSet("a", a)
	})

	It("orders dependencies (b then a)", func() {
		Expect(c.(interface{ ComponentDependencies() []string }).ComponentDependencies()).To(Equal([]string{"b", "a"}))
	})

	It("runs lifecycle start/reload/stop and component status", func() {
		// start
		Expect(c.Start()).To(Succeed())
		Expect(a.started).To(BeTrue())
		Expect(b.started).To(BeTrue())
		Expect(c.(interface{ ComponentIsStarted() bool }).ComponentIsStarted()).To(BeTrue())
		Expect(c.(interface{ ComponentIsRunning(bool) bool }).ComponentIsRunning(true)).To(BeTrue())

		// reload
		Expect(c.Reload()).To(Succeed())
		Expect(a.reloadCount).To(Equal(1))
		Expect(b.reloadCount).To(Equal(1))
		Expect(a.started).To(BeTrue())
		Expect(b.started).To(BeTrue())

		// stop
		c.Stop()
		Expect(a.started).To(BeFalse())
		Expect(b.started).To(BeFalse())
		Expect(a.running).To(BeFalse())
		Expect(b.running).To(BeFalse())
	})

	It("fires config start/reload/stop hooks", func() {
		var sb, sa, rb, ra, tb, ta int
		c.RegisterFuncStartBefore(func() error { sb++; return nil })
		c.RegisterFuncStartAfter(func() error { sa++; return nil })
		c.RegisterFuncReloadBefore(func() error { rb++; return nil })
		c.RegisterFuncReloadAfter(func() error { ra++; return nil })
		c.RegisterFuncStopBefore(func() error { tb++; return nil })
		c.RegisterFuncStopAfter(func() error { ta++; return nil })

		Expect(c.Start()).To(Succeed())
		Expect(sb).To(Equal(1))
		Expect(sa).To(Equal(1))

		Expect(c.Reload()).To(Succeed())
		Expect(rb).To(Equal(1))
		Expect(ra).To(Equal(1))

		c.Stop()
		Expect(tb).To(Equal(1))
		Expect(ta).To(Equal(1))
	})

	It("aggregates default config JSON", func() {
		rdr := c.(interface{ DefaultConfig() io.Reader }).DefaultConfig()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(rdr)
		s := buf.String()
		Expect(s).To(ContainSubstring("\"a\""))
		Expect(s).To(ContainSubstring("\"b\""))
		Expect(s).To(ContainSubstring("enabled"))
	})

	It("exposes shell commands list/start/stop/restart", func() {
		cmds := c.GetShellCommand()
		toMap := func(lst []shlcmd.Command) map[string]shlcmd.Command {
			m := map[string]shlcmd.Command{}
			for _, cc := range lst {
				m[cc.Name()] = cc
			}
			return m
		}
		m := toMap(cmds)

		// list (before start)
		out, errb := &bytes.Buffer{}, &bytes.Buffer{}
		m["list"].Run(out, errb, nil)
		Expect(out.String()).To(ContainSubstring("b"))
		Expect(out.String()).To(ContainSubstring("a"))
		Expect(errb.Len()).To(Equal(0))

		// start all
		out.Reset()
		errb.Reset()
		m["start"].Run(out, errb, nil)
		Expect(out.String()).To(ContainSubstring("Starting component 'b'"))
		Expect(out.String()).To(ContainSubstring("Starting component 'a'"))
		Expect(errb.Len()).To(Equal(0))

		// stop all
		out.Reset()
		errb.Reset()
		m["stop"].Run(out, errb, nil)
		Expect(out.String()).To(ContainSubstring("Stopping component 'a'"))
		Expect(out.String()).To(ContainSubstring("Stopping component 'b'"))

		// restart all
		out.Reset()
		errb.Reset()
		m["restart"].Run(out, errb, nil)
		Expect(out.String()).To(ContainSubstring("Stopping component 'a'"))
		Expect(out.String()).To(ContainSubstring("Stopping component 'b'"))
		Expect(out.String()).To(ContainSubstring("Starting component 'b'"))
		Expect(out.String()).To(ContainSubstring("Starting component 'a'"))
		Expect(errb.Len()).To(Equal(0))
	})
})
