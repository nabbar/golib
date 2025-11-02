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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	shlcmd "github.com/nabbar/golib/shell/command"
)

var _ = Describe("Config Shell Commands", func() {
	var (
		cfg    libcfg.Config
		comp1  *simpleComponent
		comp2  *simpleComponent
		stdout *bytes.Buffer
		stderr *bytes.Buffer
	)

	BeforeEach(func() {
		cfg = libcfg.New(nil)
		comp1 = &simpleComponent{name: "comp1"}
		comp2 = &simpleComponent{name: "comp2", deps: []string{"comp1"}}
		cfg.ComponentSet("comp1", comp1)
		cfg.ComponentSet("comp2", comp2)
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	Describe("GetShellCommand", func() {
		It("should return shell commands", func() {
			cmds := cfg.GetShellCommand()
			Expect(cmds).ToNot(BeEmpty())
		})

		It("should include list command", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)
			Expect(cmdMap).To(HaveKey("list"))
		})

		It("should include start command", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)
			Expect(cmdMap).To(HaveKey("start"))
		})

		It("should include stop command", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)
			Expect(cmdMap).To(HaveKey("stop"))
		})

		It("should include restart command", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)
			Expect(cmdMap).To(HaveKey("restart"))
		})
	})

	Describe("list command", func() {
		It("should list all components", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["list"].Run(stdout, stderr, nil)

			output := stdout.String()
			Expect(output).To(ContainSubstring("comp1"))
			Expect(output).To(ContainSubstring("comp2"))
			Expect(stderr.Len()).To(Equal(0))
		})

		It("should list components in dependency order", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["list"].Run(stdout, stderr, nil)

			// comp1 should appear before comp2 in the output
			comp1Idx := bytes.Index(stdout.Bytes(), []byte("comp1"))
			comp2Idx := bytes.Index(stdout.Bytes(), []byte("comp2"))
			Expect(comp1Idx).To(BeNumerically("<", comp2Idx))
		})

		It("should show component status", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			// Start one component
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			stdout.Reset()
			cmdMap["list"].Run(stdout, stderr, nil)

			// Should show running status
			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})

		It("should handle empty component list", func() {
			emptyCfg := libcfg.New(nil)
			cmds := emptyCfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["list"].Run(stdout, stderr, nil)

			// Should not panic and should produce some output
			Expect(stderr.Len()).To(Equal(0))
		})
	})

	Describe("start command", func() {
		It("should start all components when no args", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["start"].Run(stdout, stderr, nil)

			Expect(comp1.started).To(BeTrue())
			Expect(comp2.started).To(BeTrue())
			Expect(stderr.Len()).To(Equal(0))
		})

		It("should show starting messages", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["start"].Run(stdout, stderr, nil)

			output := stdout.String()
			Expect(output).To(ContainSubstring("Starting"))
			Expect(output).To(ContainSubstring("comp1"))
			Expect(output).To(ContainSubstring("comp2"))
		})

		It("should start components in dependency order", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["start"].Run(stdout, stderr, nil)

			// comp1 should start before comp2
			startComp1Idx := bytes.Index(stdout.Bytes(), []byte("comp1"))
			startComp2Idx := bytes.Index(stdout.Bytes(), []byte("comp2"))
			Expect(startComp1Idx).To(BeNumerically("<", startComp2Idx))
		})

		It("should handle already started components", func() {
			// Start manually first
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			// Start again via shell command
			stdout.Reset()
			cmdMap["start"].Run(stdout, stderr, nil)

			// Should not panic or error
			Expect(stderr.Len()).To(Equal(0))
		})
	})

	Describe("stop command", func() {
		BeforeEach(func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop all components when no args", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["stop"].Run(stdout, stderr, nil)

			Expect(comp1.started).To(BeFalse())
			Expect(comp2.started).To(BeFalse())
			Expect(stderr.Len()).To(Equal(0))
		})

		It("should show stopping messages", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["stop"].Run(stdout, stderr, nil)

			output := stdout.String()
			Expect(output).To(ContainSubstring("Stopping"))
			Expect(output).To(ContainSubstring("comp1"))
			Expect(output).To(ContainSubstring("comp2"))
		})

		It("should handle already stopped components", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			// Stop twice
			cmdMap["stop"].Run(stdout, stderr, nil)
			stdout.Reset()
			stderr.Reset()
			cmdMap["stop"].Run(stdout, stderr, nil)

			// Should not panic
			Expect(stderr.Len()).To(Equal(0))
		})
	})

	Describe("restart command", func() {
		It("should stop and start all components", func() {
			// Start first
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["restart"].Run(stdout, stderr, nil)

			output := stdout.String()
			Expect(output).To(ContainSubstring("Stopping"))
			Expect(output).To(ContainSubstring("Starting"))
			Expect(comp1.started).To(BeTrue())
			Expect(comp2.started).To(BeTrue())
		})

		It("should restart even when not started", func() {
			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["restart"].Run(stdout, stderr, nil)

			// Should start components
			Expect(comp1.started).To(BeTrue())
			Expect(comp2.started).To(BeTrue())
			Expect(stderr.Len()).To(Equal(0))
		})

		It("should show both stop and start messages", func() {
			err := cfg.Start()
			Expect(err).ToNot(HaveOccurred())

			cmds := cfg.GetShellCommand()
			cmdMap := commandsToMap(cmds)

			cmdMap["restart"].Run(stdout, stderr, nil)

			output := stdout.String()
			Expect(output).To(ContainSubstring("Stopping"))
			Expect(output).To(ContainSubstring("Starting"))
			Expect(output).To(ContainSubstring("comp1"))
			Expect(output).To(ContainSubstring("comp2"))
		})
	})

	Describe("Command structure", func() {
		It("should have valid command names", func() {
			cmds := cfg.GetShellCommand()
			for _, cmd := range cmds {
				Expect(cmd.Name()).ToNot(BeEmpty())
			}
		})

		It("should have valid command descriptions", func() {
			cmds := cfg.GetShellCommand()
			for _, cmd := range cmds {
				Expect(cmd.Describe()).ToNot(BeEmpty())
			}
		})

		It("should have executable run functions", func() {
			cmds := cfg.GetShellCommand()
			for _, cmd := range cmds {
				// Should not panic when calling Run
				Expect(func() {
					cmd.Run(stdout, stderr, nil)
				}).ToNot(Panic())
			}
		})
	})
})

// commandsToMap converts a slice of commands to a map by name
func commandsToMap(cmds []shlcmd.Command) map[string]shlcmd.Command {
	result := make(map[string]shlcmd.Command)
	for _, cmd := range cmds {
		result[cmd.Name()] = cmd
	}
	return result
}
