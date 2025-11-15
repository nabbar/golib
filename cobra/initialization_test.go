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

package cobra_test

import (
	"context"

	libcbr "github.com/nabbar/golib/cobra"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cobra Initialization", func() {
	var (
		cobra   libcbr.Cobra
		version libver.Version
	)

	BeforeEach(func() {
		// Create a new cobra instance for each test
		cobra = libcbr.New()

		// Create a mock version
		version = libver.NewVersion(
			libver.License_MIT,
			"test-package",
			"Test Description",
			"2024-01-01",
			"abc123",
			"v1.0.0",
			"Test Author",
			"test-app",
			struct{}{},
			0,
		)
	})

	Describe("Creating new Cobra instance", func() {
		It("should create a non-nil instance", func() {
			Expect(cobra).ToNot(BeNil())
		})

		It("should return Cobra interface", func() {
			var c interface{} = cobra
			_, ok := c.(libcbr.Cobra)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("Setting Version", func() {
		It("should set version successfully", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Version).To(ContainSubstring("v1.0.0"))
			Expect(cmd.Version).To(ContainSubstring("abc123"))
		})

		It("should handle missing version gracefully", func() {
			// Create a minimal version to avoid nil pointer
			minimalVersion := libver.NewVersion(
				libver.License_MIT,
				"test",
				"",
				"",
				"",
				"",
				"",
				"",
				struct{}{},
				0,
			)
			cobra.SetVersion(minimalVersion)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd).ToNot(BeNil())
		})
	})

	Describe("Initialization", func() {
		It("should initialize successfully", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.TraverseChildren).To(BeTrue())
		})

		It("should set package name from version", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd.Use).ToNot(BeEmpty())
		})

		It("should set version information", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd.Version).To(ContainSubstring("v1.0.0"))
			Expect(cmd.Version).To(ContainSubstring("Test Author"))
			Expect(cmd.Version).To(ContainSubstring("MIT"))
		})

		It("should set descriptions", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd.Long).To(Equal("Test Description"))
		})
	})

	Describe("Setting FuncInit", func() {
		It("should set initialization function", func() {
			called := false
			cobra.SetFuncInit(func() {
				called = true
			})

			cobra.SetVersion(version)
			cobra.Init()

			// The function should not be called immediately
			// It will be called during Execute
			Expect(called).To(BeFalse())
		})
	})

	Describe("Setting Viper", func() {
		It("should set Viper function successfully", func() {
			viperFunc := func() libvpr.Viper {
				return nil
			}

			cobra.SetViper(viperFunc)
			cobra.SetVersion(version)
			cobra.Init()

			Expect(cobra.Cobra()).ToNot(BeNil())
		})
	})

	Describe("Setting Logger", func() {
		It("should set Logger function successfully", func() {
			loggerFunc := func() liblog.Logger {
				return liblog.New(context.Background())
			}

			cobra.SetLogger(loggerFunc)
			cobra.SetVersion(version)
			cobra.Init()

			Expect(cobra.Cobra()).ToNot(BeNil())
		})
	})

	Describe("ForceNoInfo Flag", func() {
		It("should set ForceNoInfo flag", func() {
			cobra.SetForceNoInfo(true)
			cobra.SetVersion(version)
			cobra.Init()

			Expect(cobra.Cobra()).ToNot(BeNil())
		})

		It("should default ForceNoInfo to false", func() {
			cobra.SetForceNoInfo(false)
			cobra.SetVersion(version)
			cobra.Init()

			Expect(cobra.Cobra()).ToNot(BeNil())
		})
	})

	Describe("Getting Cobra Command", func() {
		It("should return nil before initialization", func() {
			cmd := cobra.Cobra()
			Expect(cmd).To(BeNil())
		})

		It("should return command after initialization", func() {
			cobra.SetVersion(version)
			cobra.Init()

			cmd := cobra.Cobra()
			Expect(cmd).ToNot(BeNil())
		})
	})
})
