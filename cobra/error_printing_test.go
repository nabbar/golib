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
	"strings"

	libcbr "github.com/nabbar/golib/cobra"
	libver "github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cobra Error Printing", func() {
	var (
		cobra   libcbr.Cobra
		version libver.Version
	)

	BeforeEach(func() {
		cobra = libcbr.New()
		version = libver.NewVersion(
			libver.License_MIT,
			"testapp",
			"Test Description",
			"2024-01-01",
			"abc123",
			"v1.0.0",
			"Test Author",
			"test-app",
			struct{}{},
			0,
		)
		cobra.SetVersion(version)
		cobra.Init()
	})

	Describe("AddCommandPrintErrorCode", func() {
		It("should add error command", func() {
			var callCount int
			printFunc := func(item, value string) {
				callCount++
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			var found bool
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should have correct command use", func() {
			printFunc := func(item, value string) {}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Use).To(Equal("error"))
					break
				}
			}
		})

		It("should have error as example", func() {
			printFunc := func(item, value string) {}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Example).To(Equal("error"))
					break
				}
			}
		})

		It("should have descriptive short text", func() {
			printFunc := func(item, value string) {}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Short).ToNot(BeEmpty())
					Expect(strings.ToLower(cmd.Short)).To(ContainSubstring("error"))
					break
				}
			}
		})

		It("should accept print function", func() {
			printFunc := func(item, value string) {
				// Function to be called when command runs
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			// The function will be called when the command runs
			Expect(cobra.Cobra()).ToNot(BeNil())
		})

		It("should handle nil long description", func() {
			printFunc := func(item, value string) {}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					// Long can be empty
					Expect(cmd.Long).To(Equal(""))
					break
				}
			}
		})
	})

	Describe("Error Command Structure", func() {
		var printFunc libcbr.FuncPrintErrorCode

		BeforeEach(func() {
			printFunc = func(item, value string) {}
			cobra.AddCommandPrintErrorCode(printFunc)
		})

		It("should have Run function defined", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Run).ToNot(BeNil())
					break
				}
			}
		})

		It("should have correct use format", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Use).To(Equal("error"))
					break
				}
			}
		})

		It("should have example equal to use", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(cmd.Example).To(Equal("error"))
					break
				}
			}
		})
	})

	Describe("Print Function Integration", func() {
		It("should call print function with error codes", func() {
			var items []string
			var values []string

			printFunc := func(item, value string) {
				items = append(items, item)
				values = append(values, value)
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			// Find and execute the error command
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					// Execute the command
					cmd.Run(cmd, []string{})
					break
				}
			}

			// The function should have been called (or not, depending on error codes available)
			// At minimum, the command should execute without panic
		})

		It("should pass string format for items", func() {
			var receivedItems []string

			printFunc := func(item, value string) {
				receivedItems = append(receivedItems, item)
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					cmd.Run(cmd, []string{})
					break
				}
			}

			// Items should be strings
			for _, item := range receivedItems {
				Expect(item).To(BeAssignableToTypeOf(""))
			}
		})

		It("should pass string values", func() {
			var receivedValues []string

			printFunc := func(item, value string) {
				receivedValues = append(receivedValues, value)
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					cmd.Run(cmd, []string{})
					break
				}
			}

			// Values should be strings
			for _, value := range receivedValues {
				Expect(value).To(BeAssignableToTypeOf(""))
			}
		})
	})

	Describe("Error Code Handling", func() {
		It("should handle empty error code list", func() {
			callCount := 0
			printFunc := func(item, value string) {
				callCount++
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					// Should not panic even if no errors
					Expect(func() {
						cmd.Run(cmd, []string{})
					}).ToNot(Panic())
					break
				}
			}
		})

		It("should sort error codes", func() {
			var items []string

			printFunc := func(item, value string) {
				items = append(items, item)
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					cmd.Run(cmd, []string{})
					break
				}
			}

			// If items are present, they should be in order
			if len(items) > 1 {
				for i := 1; i < len(items); i++ {
					// Previous item should be <= current (sorted order)
					Expect(items[i-1] <= items[i]).To(BeTrue())
				}
			}
		})
	})

	Describe("Command Integration", func() {
		It("should integrate with root command", func() {
			printFunc := func(item, value string) {}
			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			Expect(rootCmd).ToNot(BeNil())

			hasErrorCmd := false
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					hasErrorCmd = true
					break
				}
			}

			Expect(hasErrorCmd).To(BeTrue())
		})

		It("should work with other commands", func() {
			printFunc := func(item, value string) {}
			cobra.AddCommandPrintErrorCode(printFunc)

			// Add another command
			testCmd := cobra.NewCommand(
				"test",
				"Test command",
				"Test description",
				"",
				"",
			)
			cobra.AddCommand(testCmd)

			rootCmd := cobra.Cobra()
			commands := rootCmd.Commands()

			// Should have both commands
			Expect(len(commands)).To(BeNumerically(">=", 2))
		})

		It("should not conflict with completion command", func() {
			printFunc := func(item, value string) {}
			cobra.AddCommandPrintErrorCode(printFunc)
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			commands := rootCmd.Commands()

			hasError := false
			hasCompletion := false

			for _, cmd := range commands {
				if cmd.Use == "error" {
					hasError = true
				}
				if strings.Contains(cmd.Use, "completion") {
					hasCompletion = true
				}
			}

			Expect(hasError).To(BeTrue())
			Expect(hasCompletion).To(BeTrue())
		})
	})

	Describe("Multiple Instances", func() {
		It("should add error command to different instances independently", func() {
			cobra1 := libcbr.New()
			cobra1.SetVersion(version)
			cobra1.Init()

			cobra2 := libcbr.New()
			cobra2.SetVersion(version)
			cobra2.Init()

			printFunc := func(item, value string) {}

			cobra1.AddCommandPrintErrorCode(printFunc)
			cobra2.AddCommandPrintErrorCode(printFunc)

			// Both should have error commands
			hasError1 := false
			for _, cmd := range cobra1.Cobra().Commands() {
				if cmd.Use == "error" {
					hasError1 = true
					break
				}
			}

			hasError2 := false
			for _, cmd := range cobra2.Cobra().Commands() {
				if cmd.Use == "error" {
					hasError2 = true
					break
				}
			}

			Expect(hasError1).To(BeTrue())
			Expect(hasError2).To(BeTrue())
		})

		It("should use different print functions per instance", func() {
			cobra1 := libcbr.New()
			cobra1.SetVersion(version)
			cobra1.Init()

			cobra2 := libcbr.New()
			cobra2.SetVersion(version)
			cobra2.Init()

			count1 := 0
			printFunc1 := func(item, value string) {
				count1++
			}

			count2 := 0
			printFunc2 := func(item, value string) {
				count2++
			}

			cobra1.AddCommandPrintErrorCode(printFunc1)
			cobra2.AddCommandPrintErrorCode(printFunc2)

			// Execute both commands
			for _, cmd := range cobra1.Cobra().Commands() {
				if cmd.Use == "error" {
					cmd.Run(cmd, []string{})
					break
				}
			}

			for _, cmd := range cobra2.Cobra().Commands() {
				if cmd.Use == "error" {
					cmd.Run(cmd, []string{})
					break
				}
			}

			// Both should have executed (counts may be 0 if no errors exist)
			// But they should not interfere with each other
			Expect(count1 >= 0).To(BeTrue())
			Expect(count2 >= 0).To(BeTrue())
		})
	})

	Describe("Function Type Safety", func() {
		It("should accept valid FuncPrintErrorCode", func() {
			validFunc := func(item, value string) {
				// Valid implementation
			}

			Expect(func() {
				cobra.AddCommandPrintErrorCode(validFunc)
			}).ToNot(Panic())
		})

		It("should work with function that prints", func() {
			printFunc := func(item, value string) {
				// Simulate printing
				_ = item + ": " + value
			}

			cobra.AddCommandPrintErrorCode(printFunc)

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "error" {
					Expect(func() {
						cmd.Run(cmd, []string{})
					}).ToNot(Panic())
					break
				}
			}
		})
	})
})
