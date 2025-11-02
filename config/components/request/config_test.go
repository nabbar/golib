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

package request_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/request"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

var _ = Describe("Configuration", func() {
	var (
		cpt ComponentRequest
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		cpt = New(ctx, nil)
	})

	Describe("RegisterFlag", func() {
		It("should not panic with valid command", func() {
			cmd := &spfcbr.Command{}
			Expect(func() {
				_ = cpt.RegisterFlag(cmd)
			}).NotTo(Panic())
		})

		It("should handle nil command gracefully", func() {
			Expect(func() {
				_ = cpt.RegisterFlag(nil)
			}).NotTo(Panic())
		})
	})

	Describe("Start and Reload", func() {
		It("should fail without initialization", func() {
			err := cpt.Start()
			Expect(err).To(HaveOccurred())
		})

		It("should fail without viper", func() {
			key := "request"
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }

			cpt.Init(key, ctx, getCpt, vpr, vrs, log)

			err := cpt.Start()
			Expect(err).To(HaveOccurred())
		})
	})
})
