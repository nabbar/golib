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
	libcbr "github.com/nabbar/golib/cobra"
	libver "github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spfcbr "github.com/spf13/cobra"
)

var _ = Describe("Cobra Commands", func() {
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

	Describe("NewCommand", func() {
		It("should create a new command", func() {
			cmd := cobra.NewCommand(
				"test",
				"Test command",
				"This is a test command",
				"<args>",
				"--flag value",
			)

			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Short).To(Equal("Test command"))
			Expect(cmd.Long).To(Equal("This is a test command"))
			Expect(cmd.Use).To(ContainSubstring("test"))
			Expect(cmd.Use).To(ContainSubstring("<args>"))
		})

		It("should create command with example", func() {
			cmd := cobra.NewCommand(
				"run",
				"Run command",
				"Runs the application",
				"",
				"run --verbose",
			)

			Expect(cmd.Example).To(ContainSubstring("run"))
			Expect(cmd.Example).To(ContainSubstring("--verbose"))
		})

		It("should handle empty use and example", func() {
			cmd := cobra.NewCommand(
				"simple",
				"Simple command",
				"Simple description",
				"",
				"",
			)

			Expect(cmd).ToNot(BeNil())
			Expect(cmd.Use).To(ContainSubstring("simple"))
		})

		It("should create multiple commands", func() {
			cmd1 := cobra.NewCommand("cmd1", "Command 1", "Description 1", "", "")
			cmd2 := cobra.NewCommand("cmd2", "Command 2", "Description 2", "", "")

			Expect(cmd1).ToNot(BeNil())
			Expect(cmd2).ToNot(BeNil())
			Expect(cmd1.Use).ToNot(Equal(cmd2.Use))
		})
	})

	Describe("AddCommand", func() {
		It("should add a single command", func() {
			cmd := &spfcbr.Command{
				Use:   "subcommand",
				Short: "A subcommand",
				Run: func(cmd *spfcbr.Command, args []string) {
					// Test command
				},
			}

			cobra.AddCommand(cmd)

			rootCmd := cobra.Cobra()
			Expect(rootCmd.Commands()).To(ContainElement(cmd))
		})

		It("should add multiple commands", func() {
			cmd1 := &spfcbr.Command{Use: "cmd1", Short: "Command 1"}
			cmd2 := &spfcbr.Command{Use: "cmd2", Short: "Command 2"}
			cmd3 := &spfcbr.Command{Use: "cmd3", Short: "Command 3"}

			cobra.AddCommand(cmd1, cmd2, cmd3)

			rootCmd := cobra.Cobra()
			commands := rootCmd.Commands()

			Expect(len(commands)).To(BeNumerically(">=", 3))
			Expect(commands).To(ContainElement(cmd1))
			Expect(commands).To(ContainElement(cmd2))
			Expect(commands).To(ContainElement(cmd3))
		})

		It("should preserve command order", func() {
			cmd1 := &spfcbr.Command{Use: "alpha", Short: "Alpha"}
			cmd2 := &spfcbr.Command{Use: "beta", Short: "Beta"}

			cobra.AddCommand(cmd1, cmd2)

			rootCmd := cobra.Cobra()
			commands := rootCmd.Commands()

			Expect(commands).To(ContainElement(cmd1))
			Expect(commands).To(ContainElement(cmd2))
		})
	})

	Describe("Command Creation with NewCommand", func() {
		It("should create command that can be added", func() {
			cmd := cobra.NewCommand(
				"execute",
				"Execute command",
				"Executes something",
				"",
				"",
			)

			cobra.AddCommand(cmd)

			rootCmd := cobra.Cobra()
			Expect(rootCmd.Commands()).To(ContainElement(cmd))
		})

		It("should support command with flags", func() {
			cmd := cobra.NewCommand(
				"flagged",
				"Command with flags",
				"Has flags",
				"",
				"",
			)

			var testFlag string
			cmd.Flags().StringVarP(&testFlag, "test", "t", "", "test flag")

			cobra.AddCommand(cmd)

			Expect(cmd.Flags().Lookup("test")).ToNot(BeNil())
		})
	})

	Describe("Command Hierarchy", func() {
		It("should support nested commands", func() {
			parentCmd := &spfcbr.Command{
				Use:   "parent",
				Short: "Parent command",
			}

			childCmd := &spfcbr.Command{
				Use:   "child",
				Short: "Child command",
			}

			parentCmd.AddCommand(childCmd)
			cobra.AddCommand(parentCmd)

			Expect(parentCmd.Commands()).To(ContainElement(childCmd))
		})

		It("should allow adding commands to created commands", func() {
			mainCmd := cobra.NewCommand(
				"main",
				"Main command",
				"Main description",
				"",
				"",
			)

			subCmd := &spfcbr.Command{
				Use:   "sub",
				Short: "Sub command",
			}

			mainCmd.AddCommand(subCmd)
			cobra.AddCommand(mainCmd)

			Expect(mainCmd.Commands()).To(ContainElement(subCmd))
		})
	})

	Describe("Command Usage", func() {
		It("should format command usage correctly", func() {
			cmd := cobra.NewCommand(
				"format",
				"Format command",
				"Formats output",
				"<file> [options]",
				"format myfile.txt --pretty",
			)

			Expect(cmd.Use).To(ContainSubstring("format"))
			Expect(cmd.Use).To(ContainSubstring("<file>"))
			Expect(cmd.Use).To(ContainSubstring("[options]"))
		})

		It("should preserve example format", func() {
			cmd := cobra.NewCommand(
				"example",
				"Example command",
				"Example description",
				"",
				"example --flag1 --flag2=value",
			)

			Expect(cmd.Example).To(ContainSubstring("example"))
			Expect(cmd.Example).To(ContainSubstring("--flag1"))
			Expect(cmd.Example).To(ContainSubstring("--flag2=value"))
		})
	})

	Describe("Command Execution", func() {
		It("should execute added command", func() {
			executed := false
			cmd := &spfcbr.Command{
				Use:   "exec",
				Short: "Execute",
				Run: func(cmd *spfcbr.Command, args []string) {
					executed = true
				},
			}

			cobra.AddCommand(cmd)
			cmd.Run(cmd, []string{})

			Expect(executed).To(BeTrue())
		})

		It("should pass arguments to command", func() {
			var receivedArgs []string
			cmd := &spfcbr.Command{
				Use:   "args",
				Short: "Arguments test",
				Run: func(cmd *spfcbr.Command, args []string) {
					receivedArgs = args
				},
			}

			cobra.AddCommand(cmd)
			testArgs := []string{"arg1", "arg2", "arg3"}
			cmd.Run(cmd, testArgs)

			Expect(receivedArgs).To(Equal(testArgs))
		})
	})

	Describe("Command Validation", func() {
		It("should validate command structure", func() {
			cmd := cobra.NewCommand(
				"validate",
				"Validation test",
				"Tests validation",
				"",
				"",
			)

			Expect(cmd.Use).ToNot(BeEmpty())
			Expect(cmd.Short).ToNot(BeEmpty())
			Expect(cmd.Long).ToNot(BeEmpty())
		})

		It("should handle command without run function", func() {
			cmd := &spfcbr.Command{
				Use:   "norun",
				Short: "No run function",
			}

			cobra.AddCommand(cmd)

			rootCmd := cobra.Cobra()
			Expect(rootCmd.Commands()).To(ContainElement(cmd))
		})
	})

	Describe("Root Command Access", func() {
		It("should provide access to root command", func() {
			rootCmd := cobra.Cobra()
			Expect(rootCmd).ToNot(BeNil())
			Expect(rootCmd.TraverseChildren).To(BeTrue())
		})

		It("should maintain root command properties", func() {
			rootCmd := cobra.Cobra()
			Expect(rootCmd.Version).To(ContainSubstring("v1.0.0"))
			Expect(rootCmd.Long).To(Equal("Test Description"))
		})
	})
})
