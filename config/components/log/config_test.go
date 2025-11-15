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
 */

package log_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	loglvl "github.com/nabbar/golib/logger/level"
	spfcbr "github.com/spf13/cobra"
)

// Configuration management tests verify config loading, validation and flag registration.
var _ = Describe("Configuration Management", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		cpt CptLog
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, loglvl.NilLevel)
		cpt.Init(kd, ctx, nil, fv, vs, fl)

		v.Viper().SetConfigType("json")

		configData := map[string]interface{}{
			kd: map[string]interface{}{
				"stdout": map[string]interface{}{
					"disableStandard": true,
				},
			},
		}

		configJSON, err := json.Marshal(configData)
		Expect(err).To(BeNil())

		err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if cpt != nil {
			cpt.Stop()
		}
		cnl()
	})

	Describe("RegisterFlag method", func() {
		Context("flag registration", func() {
			It("should register flags on initialized component", func() {
				cmd := &spfcbr.Command{}
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should register flags with command", func() {
				cmd := &spfcbr.Command{}
				err := cpt.RegisterFlag(cmd)
				Expect(err).To(BeNil())

				Expect(cmd.PersistentFlags().Lookup(kd + ".disableStandard")).NotTo(BeNil())
				Expect(cmd.PersistentFlags().Lookup(kd + ".disableStack")).NotTo(BeNil())
				Expect(cmd.PersistentFlags().Lookup(kd + ".disableTimestamp")).NotTo(BeNil())
				Expect(cmd.PersistentFlags().Lookup(kd + ".enableTrace")).NotTo(BeNil())
				Expect(cmd.PersistentFlags().Lookup(kd + ".traceFilter")).NotTo(BeNil())
				Expect(cmd.PersistentFlags().Lookup(kd + ".disableColor")).NotTo(BeNil())
			})

			It("should bind flags to viper", func() {
				key := "log"
				getCpt := func(k string) cfgtps.Component { return nil }

				cpt.Init(key, ctx, getCpt, fv, vs, fl)

				cmd := &spfcbr.Command{}
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})

	Describe("Start and Reload", func() {
		Context("component lifecycle", func() {
			It("should start successfully when initialized", func() {
				err := cpt.Start()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reload successfully when initialized", func() {
				err := cpt.Start()
				Expect(err).NotTo(HaveOccurred())
				err = cpt.Reload()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
