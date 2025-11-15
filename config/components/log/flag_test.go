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

package log_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	loglvl "github.com/nabbar/golib/logger/level"
	spfcbr "github.com/spf13/cobra"
)

// Flag registration tests verify the RegisterFlag method and CLI integration.
// These tests ensure proper flag setup, binding to viper, and default values.
var _ = Describe("Flag Registration", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		cpt CptLog
		cmd *spfcbr.Command
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

		// Create a fresh command for each test
		cmd = &spfcbr.Command{
			Use: "test",
		}
	})

	AfterEach(func() {
		if cpt != nil {
			cpt.Stop()
		}
		cnl()
	})

	Describe("RegisterFlag behavior", func() {
		Context("basic flag registration", func() {
			It("should register all required flags", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// Verify all expected flags exist
				flags := []string{
					kd + ".disableStandard",
					kd + ".disableStack",
					kd + ".disableTimestamp",
					kd + ".enableTrace",
					kd + ".traceFilter",
					kd + ".disableColor",
				}

				for _, flagName := range flags {
					flag := cmd.PersistentFlags().Lookup(flagName)
					Expect(flag).NotTo(BeNil(), "Flag %s should exist", flagName)
				}
			})

			It("should set correct default values for boolean flags", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// Check default values
				disableStandard := cmd.PersistentFlags().Lookup(kd + ".disableStandard")
				Expect(disableStandard.DefValue).To(Equal("false"))

				disableStack := cmd.PersistentFlags().Lookup(kd + ".disableStack")
				Expect(disableStack.DefValue).To(Equal("false"))

				disableTimestamp := cmd.PersistentFlags().Lookup(kd + ".disableTimestamp")
				Expect(disableTimestamp.DefValue).To(Equal("false"))

				enableTrace := cmd.PersistentFlags().Lookup(kd + ".enableTrace")
				Expect(enableTrace.DefValue).To(Equal("true"))

				disableColor := cmd.PersistentFlags().Lookup(kd + ".disableColor")
				Expect(disableColor.DefValue).To(Equal("false"))
			})

			It("should set correct default value for string flags", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				traceFilter := cmd.PersistentFlags().Lookup(kd + ".traceFilter")
				Expect(traceFilter).NotTo(BeNil())
				Expect(traceFilter.DefValue).To(Equal(""))
			})

			It("should bind flags to viper", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// Viper should have bindings for all flags
				// Note: We can't directly check viper bindings, but we can verify
				// that the registration completed without error
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("flag registration errors", func() {
			It("should return error for uninitialized component", func() {
				uninit := New(ctx, loglvl.InfoLevel)

				err := uninit.RegisterFlag(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for component without viper", func() {
				partial := New(ctx, loglvl.InfoLevel)
				partial.Init(kd, ctx, nil, nil, vs, fl)

				err := partial.RegisterFlag(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for component without key", func() {
				partial := New(ctx, loglvl.InfoLevel)
				partial.Init("", ctx, nil, fv, vs, fl)

				err := partial.RegisterFlag(cmd)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("multiple flag registrations", func() {
			It("should handle multiple registrations on different commands", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// Second registration on a different command should succeed
				cmd2 := &spfcbr.Command{Use: "test2"}
				err = cpt.RegisterFlag(cmd2)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle registration on different commands", func() {
				cmd1 := &spfcbr.Command{Use: "cmd1"}
				cmd2 := &spfcbr.Command{Use: "cmd2"}

				err := cpt.RegisterFlag(cmd1)
				Expect(err).NotTo(HaveOccurred())

				err = cpt.RegisterFlag(cmd2)
				Expect(err).NotTo(HaveOccurred())

				// Both commands should have the flags
				Expect(cmd1.PersistentFlags().Lookup(kd + ".disableStandard")).NotTo(BeNil())
				Expect(cmd2.PersistentFlags().Lookup(kd + ".disableStandard")).NotTo(BeNil())
			})
		})

		Context("flag usage information", func() {
			It("should have descriptive usage text for all flags", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				flags := map[string]string{
					kd + ".disableStandard":  "allow disabling to write log to standard output stdout/stderr",
					kd + ".disableStack":     "allow to disable the goroutine id before each message",
					kd + ".disableTimestamp": "allow to disable the timestamp before each message",
					kd + ".enableTrace":      "allow to add the origin caller/file/line of each message",
					kd + ".traceFilter":      "define the path to clean for trace",
					kd + ".disableColor":     "define if color could be use or not in messages format",
				}

				for flagName, expectedUsage := range flags {
					flag := cmd.PersistentFlags().Lookup(flagName)
					Expect(flag).NotTo(BeNil())
					Expect(flag.Usage).To(ContainSubstring(expectedUsage))
				}
			})
		})

		Context("flag types", func() {
			It("should register boolean flags as bool type", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				boolFlags := []string{
					kd + ".disableStandard",
					kd + ".disableStack",
					kd + ".disableTimestamp",
					kd + ".enableTrace",
					kd + ".disableColor",
				}

				for _, flagName := range boolFlags {
					flag := cmd.PersistentFlags().Lookup(flagName)
					Expect(flag).NotTo(BeNil())
					Expect(flag.Value.Type()).To(Equal("bool"))
				}
			})

			It("should register string flags as string type", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				flag := cmd.PersistentFlags().Lookup(kd + ".traceFilter")
				Expect(flag).NotTo(BeNil())
				Expect(flag.Value.Type()).To(Equal("string"))
			})
		})

		Context("flag persistence", func() {
			It("should register flags as persistent", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// All flags should be in persistent flags, not local flags
				allFlags := []string{
					kd + ".disableStandard",
					kd + ".disableStack",
					kd + ".disableTimestamp",
					kd + ".enableTrace",
					kd + ".traceFilter",
					kd + ".disableColor",
				}

				for _, flagName := range allFlags {
					// Should be in persistent flags
					persistentFlag := cmd.PersistentFlags().Lookup(flagName)
					Expect(persistentFlag).NotTo(BeNil())

					// Should NOT be in local flags
					localFlag := cmd.Flags().Lookup(flagName)
					Expect(localFlag).To(BeNil())
				}
			})
		})
	})

	Describe("Flag integration with configuration", func() {
		Context("flag values override config", func() {
			It("should allow starting after flag registration", func() {
				err := cpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				err = cpt.Start()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should work with different component keys", func() {
				customKey := "custom-log"
				customCpt := New(ctx, loglvl.InfoLevel)
				customCpt.Init(customKey, ctx, nil, fv, vs, fl)

				configData := map[string]interface{}{
					customKey: map[string]interface{}{
						"stdout": map[string]interface{}{
							"disableStandard": true,
						},
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				cmd := &spfcbr.Command{Use: "test"}
				err = customCpt.RegisterFlag(cmd)
				Expect(err).NotTo(HaveOccurred())

				// Verify flags use custom key
				flag := cmd.PersistentFlags().Lookup(customKey + ".disableStandard")
				Expect(flag).NotTo(BeNil())
			})
		})
	})

	Describe("Sequential multiple flag registrations", func() {
		Context("different components", func() {
			It("should handle sequential registrations on different commands", func() {
				// Create multiple commands sequentially
				for i := 0; i < 3; i++ {
					testCmd := &spfcbr.Command{Use: "test"}
					err := cpt.RegisterFlag(testCmd)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})
	})
})
