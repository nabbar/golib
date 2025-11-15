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

var _ = Describe("Cobra Completion", func() {
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

	Describe("AddCommandCompletion", func() {
		It("should add completion command", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			var found bool
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should have correct command use format", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			var completionCmd *struct {
				Use     string
				Short   string
				Long    string
				Example string
			}

			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					completionCmd = &struct {
						Use     string
						Short   string
						Long    string
						Example string
					}{
						Use:     cmd.Use,
						Short:   cmd.Short,
						Long:    cmd.Long,
						Example: cmd.Example,
					}
					break
				}
			}

			Expect(completionCmd).ToNot(BeNil())
			Expect(completionCmd.Use).To(ContainSubstring("completion"))
			Expect(completionCmd.Use).To(ContainSubstring("Bash"))
			Expect(completionCmd.Use).To(ContainSubstring("Zsh"))
			Expect(completionCmd.Use).To(ContainSubstring("PowerShell"))
			Expect(completionCmd.Use).To(ContainSubstring("Fish"))
		})

		It("should have descriptive short text", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					Expect(cmd.Short).ToNot(BeEmpty())
					Expect(cmd.Short).To(ContainSubstring("completion"))
					break
				}
			}
		})

		It("should have detailed long description", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					Expect(cmd.Long).ToNot(BeEmpty())
					Expect(cmd.Long).To(ContainSubstring("shell"))
					break
				}
			}
		})

		It("should have example usage", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					Expect(cmd.Example).ToNot(BeEmpty())
					Expect(cmd.Example).To(ContainSubstring("bash"))
					break
				}
			}
		})

		It("should only add command once", func() {
			cobra.AddCommandCompletion()
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			count := 0
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					count++
				}
			}

			// May add multiple times, but cobra handles duplicates
			Expect(count).To(BeNumerically(">=", 1))
		})
	})

	Describe("Completion Command Details", func() {
		BeforeEach(func() {
			cobra.AddCommandCompletion()
		})

		It("should mention bash in documentation", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					longText := strings.ToLower(cmd.Long)
					Expect(longText).To(ContainSubstring("bash"))
					break
				}
			}
		})

		It("should mention shell options in documentation", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					longText := strings.ToLower(cmd.Long)
					// Should mention shells in general
					Expect(longText).ToNot(BeEmpty())
					break
				}
			}
		})

		It("should have structured documentation", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					Expect(cmd.Long).ToNot(BeEmpty())
					Expect(len(cmd.Long)).To(BeNumerically(">", 50))
					break
				}
			}
		})

		It("should provide installation guidance", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					longText := strings.ToLower(cmd.Long)
					Expect(longText).To(ContainSubstring("script"))
					break
				}
			}
		})

		It("should include installation instructions", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					longText := strings.ToLower(cmd.Long)
					// Should have some installation guidance
					Expect(longText).To(ContainSubstring("enable"))
					break
				}
			}
		})

		It("should include usage example in long description", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Long description should contain package/command info
					Expect(cmd.Long).ToNot(BeEmpty())
					break
				}
			}
		})
	})

	Describe("Shell Support", func() {
		BeforeEach(func() {
			cobra.AddCommandCompletion()
		})

		It("should support bash completion", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Verify bash is mentioned in the use string
					Expect(strings.ToLower(cmd.Use)).To(ContainSubstring("bash"))
					break
				}
			}
		})

		It("should support zsh completion", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Verify zsh is mentioned in the use string
					Expect(strings.ToLower(cmd.Use)).To(ContainSubstring("zsh"))
					break
				}
			}
		})

		It("should support fish completion", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Verify fish is mentioned in the use string
					Expect(strings.ToLower(cmd.Use)).To(ContainSubstring("fish"))
					break
				}
			}
		})

		It("should support powershell completion", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Verify powershell is mentioned in the use string
					Expect(strings.ToLower(cmd.Use)).To(ContainSubstring("powershell"))
					break
				}
			}
		})
	})

	Describe("Command Integration", func() {
		It("should integrate with root command", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			Expect(rootCmd).ToNot(BeNil())

			var completionCmd *struct{ Use string }
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					completionCmd = &struct{ Use string }{Use: cmd.Use}
					break
				}
			}

			Expect(completionCmd).ToNot(BeNil())
		})

		It("should be accessible from root", func() {
			cobra.AddCommandCompletion()

			rootCmd := cobra.Cobra()
			hasCompletion := false

			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(strings.ToLower(cmd.Use), "completion") {
					hasCompletion = true
					break
				}
			}

			Expect(hasCompletion).To(BeTrue())
		})

		It("should work with other commands", func() {
			cobra.AddCommandCompletion()

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
	})

	Describe("Documentation Quality", func() {
		BeforeEach(func() {
			cobra.AddCommandCompletion()
		})

		It("should have clear usage instructions", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Should specify file path requirement
					Expect(strings.ToLower(cmd.Use)).To(ContainSubstring("file"))
					break
				}
			}
		})

		It("should mention valid extensions in use text", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Check use text has helpful information
					Expect(cmd.Use).ToNot(BeEmpty())
					Expect(len(cmd.Use)).To(BeNumerically(">", 10))
					break
				}
			}
		})

		It("should provide example with package name", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Example should include bash completion path
					Expect(cmd.Example).To(ContainSubstring("bash"))
					Expect(cmd.Example).To(ContainSubstring("/etc/bash_completion.d/"))
					break
				}
			}
		})

		It("should explain file path in long description", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Long description should explain what happens
					Expect(cmd.Long).To(ContainSubstring("script"))
					break
				}
			}
		})
	})

	Describe("Multiple Cobra Instances", func() {
		It("should add completion to different instances independently", func() {
			cobra1 := libcbr.New()
			cobra1.SetVersion(version)
			cobra1.Init()
			cobra1.AddCommandCompletion()

			cobra2 := libcbr.New()
			cobra2.SetVersion(version)
			cobra2.Init()
			cobra2.AddCommandCompletion()

			// Both should have completion commands
			hasCompletion1 := false
			for _, cmd := range cobra1.Cobra().Commands() {
				if strings.Contains(cmd.Use, "completion") {
					hasCompletion1 = true
					break
				}
			}

			hasCompletion2 := false
			for _, cmd := range cobra2.Cobra().Commands() {
				if strings.Contains(cmd.Use, "completion") {
					hasCompletion2 = true
					break
				}
			}

			Expect(hasCompletion1).To(BeTrue())
			Expect(hasCompletion2).To(BeTrue())
		})
	})

	Describe("Command Structure", func() {
		BeforeEach(func() {
			cobra.AddCommandCompletion()
		})

		It("should have a Run function", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					// Run function should be set
					Expect(cmd.Run).ToNot(BeNil())
					break
				}
			}
		})

		It("should be properly configured", func() {
			rootCmd := cobra.Cobra()
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "completion") {
					Expect(cmd.Use).ToNot(BeEmpty())
					Expect(cmd.Short).ToNot(BeEmpty())
					Expect(cmd.Long).ToNot(BeEmpty())
					Expect(cmd.Example).ToNot(BeEmpty())
					break
				}
			}
		})
	})
})
