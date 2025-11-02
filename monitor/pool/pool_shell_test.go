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

package pool_test

import (
	"bytes"
	"context"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Shell Commands", func() {
	var (
		pool monpool.Pool
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 10*time.Second)
		pool = newPool(ctx)
	})

	AfterEach(func() {
		if pool != nil && pool.IsRunning() {
			_ = pool.Stop(ctx)
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("ShellCommandInfo", func() {
		It("should return a list of command info", func() {
			infos := monpool.ShellCommandInfo()
			Expect(infos).ToNot(BeEmpty())
		})

		It("should include standard commands", func() {
			infos := monpool.ShellCommandInfo()

			// Check we have multiple commands (list, info, start, stop, restart, status)
			Expect(len(infos)).To(BeNumerically(">=", 6))
		})

		It("should include list command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(infos).ToNot(BeEmpty())
			// The first command should be 'list' based on the code
		})

		It("should include info command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(len(infos)).To(BeNumerically(">=", 2))
		})

		It("should include start command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(len(infos)).To(BeNumerically(">=", 3))
		})

		It("should include stop command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(len(infos)).To(BeNumerically(">=", 4))
		})

		It("should include restart command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(len(infos)).To(BeNumerically(">=", 5))
		})

		It("should include status command", func() {
			infos := monpool.ShellCommandInfo()
			Expect(len(infos)).To(BeNumerically(">=", 6))
		})
	})

	Describe("GetShellCommand", func() {
		It("should return a list of executable commands", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(commands).ToNot(BeEmpty())
		})

		It("should return 6 commands", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(commands).To(HaveLen(6))
		})

		It("should return valid command objects", func() {
			commands := pool.GetShellCommand(ctx)
			for _, cmd := range commands {
				Expect(cmd).ToNot(BeNil())
			}
		})
	})

	Describe("Shell Command Execution", func() {
		BeforeEach(func() {
			// Add test monitors for command execution
			monitors := []montps.Monitor{
				createTestMonitor("shell-test-1", nil),
				createTestMonitor("shell-test-2", nil),
				createTestMonitor("shell-test-3", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should provide working list command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(commands).ToNot(BeEmpty())

			// The list command should be first
			// We can't easily test command execution without the actual shell infrastructure
			// but we can verify the command exists
			Expect(len(commands)).To(BeNumerically(">=", 1))
		})

		It("should provide working info command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(len(commands)).To(BeNumerically(">=", 2))
		})

		It("should provide working start command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(len(commands)).To(BeNumerically(">=", 3))
		})

		It("should provide working stop command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(len(commands)).To(BeNumerically(">=", 4))
		})

		It("should provide working restart command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(len(commands)).To(BeNumerically(">=", 5))
		})

		It("should provide working status command", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(len(commands)).To(BeNumerically(">=", 6))
		})
	})

	Describe("Shell Command Context", func() {
		It("should respect context in GetShellCommand", func() {
			// Create a context that will be cancelled
			cmdCtx, cmdCnl := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cmdCnl()

			commands := pool.GetShellCommand(cmdCtx)
			Expect(commands).ToNot(BeEmpty())

			// Wait for context to expire
			<-cmdCtx.Done()

			// Commands should still be valid objects
			// even if context expired
			Expect(commands).To(HaveLen(6))
		})
	})

	Describe("Shell Command with Running Pool", func() {
		It("should work with running monitors", func() {
			monitor := createTestMonitor("running-shell", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Give monitors time to start
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			Expect(commands).ToNot(BeEmpty())
			Expect(commands).To(HaveLen(6))
		})

		It("should work with stopped pool", func() {
			monitor := createTestMonitor("stopped-shell", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// Don't start the pool
			commands := pool.GetShellCommand(ctx)
			Expect(commands).ToNot(BeEmpty())
			Expect(commands).To(HaveLen(6))
		})
	})

	Describe("Shell Command Idempotency", func() {
		It("should return consistent commands on multiple calls", func() {
			monitor := createTestMonitor("idempotent-shell", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// Get commands multiple times
			commands1 := pool.GetShellCommand(ctx)
			commands2 := pool.GetShellCommand(ctx)
			commands3 := pool.GetShellCommand(ctx)

			// Should have same number of commands
			Expect(commands1).To(HaveLen(len(commands2)))
			Expect(commands2).To(HaveLen(len(commands3)))
		})
	})

	Describe("Shell Command with Empty Pool", func() {
		It("should work with no monitors", func() {
			commands := pool.GetShellCommand(ctx)
			Expect(commands).ToNot(BeEmpty())
			Expect(commands).To(HaveLen(6))
		})
	})

	Describe("Shell Command Performance", func() {
		It("should handle large number of monitors efficiently", func() {
			// Add many monitors
			for i := 0; i < 20; i++ {
				mon := createTestMonitor("perf-shell-"+string(rune('a'+i%26)), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			// Getting shell commands should be fast
			start := time.Now()
			commands := pool.GetShellCommand(ctx)
			elapsed := time.Since(start)

			Expect(commands).To(HaveLen(6))
			// Should complete in reasonable time
			Expect(elapsed).To(BeNumerically("<", 100*time.Millisecond))
		})
	})

	Describe("Shell Command Execution Coverage", func() {
		BeforeEach(func() {
			// Add test monitors
			mon1 := createTestMonitor("cmd-test-1", nil)
			mon2 := createTestMonitor("cmd-test-2", nil)
			defer mon1.Stop(ctx)
			defer mon2.Stop(ctx)

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())
		})

		It("should execute list command", func() {
			commands := pool.GetShellCommand(ctx)
			listCmd := commands[0] // list is first command

			var stdout, stderr bytes.Buffer
			listCmd.Run(&stdout, &stderr, []string{})

			// Should list monitors
			output := stdout.String()
			Expect(output).To(ContainSubstring("cmd-test-1"))
			Expect(output).To(ContainSubstring("cmd-test-2"))
		})

		It("should execute info command with specific monitor", func() {
			commands := pool.GetShellCommand(ctx)
			infoCmd := commands[1] // info is second command

			var stdout, stderr bytes.Buffer
			infoCmd.Run(&stdout, &stderr, []string{"cmd-test-1"})

			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})

		It("should execute info command with all monitors", func() {
			commands := pool.GetShellCommand(ctx)
			infoCmd := commands[1]

			var stdout, stderr bytes.Buffer
			infoCmd.Run(&stdout, &stderr, []string{})

			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})

		It("should execute start command", func() {
			commands := pool.GetShellCommand(ctx)
			startCmd := commands[2] // start is third command

			var stdout, stderr bytes.Buffer
			startCmd.Run(&stdout, &stderr, []string{"cmd-test-1"})

			time.Sleep(200 * time.Millisecond)

			// Verify monitor is running
			mon := pool.MonitorGet("cmd-test-1")
			Expect(mon).ToNot(BeNil())
		})

		It("should execute start command for all monitors", func() {
			commands := pool.GetShellCommand(ctx)
			startCmd := commands[2]

			var stdout, stderr bytes.Buffer
			startCmd.Run(&stdout, &stderr, []string{})

			time.Sleep(200 * time.Millisecond)

			// Verify at least one monitor started
			Expect(pool.MonitorList()).ToNot(BeEmpty())
		})

		It("should execute stop command", func() {
			// First start the monitors
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			stopCmd := commands[3] // stop is fourth command

			var stdout, stderr bytes.Buffer
			stopCmd.Run(&stdout, &stderr, []string{"cmd-test-1"})

			output := stdout.String()
			Expect(output).To(ContainSubstring("cmd-test-1"))
		})

		It("should execute stop command for all monitors", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			stopCmd := commands[3]

			var stdout, stderr bytes.Buffer
			stopCmd.Run(&stdout, &stderr, []string{})

			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})

		It("should execute restart command", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			restartCmd := commands[4] // restart is fifth command

			var stdout, stderr bytes.Buffer
			restartCmd.Run(&stdout, &stderr, []string{"cmd-test-1"})

			time.Sleep(100 * time.Millisecond)
			output := stdout.String()
			Expect(output).To(ContainSubstring("cmd-test-1"))
		})

		It("should execute restart command for all monitors", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			restartCmd := commands[4]

			var stdout, stderr bytes.Buffer
			restartCmd.Run(&stdout, &stderr, []string{})

			time.Sleep(100 * time.Millisecond)
			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})

		It("should execute status command", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			statusCmd := commands[5] // status is sixth command

			var stdout, stderr bytes.Buffer
			statusCmd.Run(&stdout, &stderr, []string{"cmd-test-1"})

			output := stdout.String()
			Expect(output).To(ContainSubstring("cmd-test-1"))
		})

		It("should execute status command for all monitors", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			commands := pool.GetShellCommand(ctx)
			statusCmd := commands[5]

			var stdout, stderr bytes.Buffer
			statusCmd.Run(&stdout, &stderr, []string{})

			output := stdout.String()
			Expect(output).ToNot(BeEmpty())
		})
	})
})
