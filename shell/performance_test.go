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
 */

package shell_test

import (
	"fmt"
	"io"
	"time"

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Performance Benchmarks", func() {
	Describe("Add Performance", func() {
		It("should measure single command addition", func() {
			experiment := gmeasure.NewExperiment("Add - Single Command")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add-single", func() {
					cmd := command.New(fmt.Sprintf("cmd%d", idx), "Test", nil)
					sh.Add("", cmd)
				})
			}, gmeasure.SamplingConfig{N: 10000, Duration: 5 * time.Second})

			Expect(experiment.Get("add-single").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure batch command addition", func() {
			experiment := gmeasure.NewExperiment("Add - Batch Commands")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add-batch", func() {
					cmd1 := command.New(fmt.Sprintf("cmd1_%d", idx), "Test 1", nil)
					cmd2 := command.New(fmt.Sprintf("cmd2_%d", idx), "Test 2", nil)
					cmd3 := command.New(fmt.Sprintf("cmd3_%d", idx), "Test 3", nil)
					sh.Add("", cmd1, cmd2, cmd3)
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})
	})

	Describe("Get Performance", func() {
		It("should measure Get with few commands", func() {
			experiment := gmeasure.NewExperiment("Get - Few Commands")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 10; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("get-command", func() {
					_, _ = sh.Get("cmd5")
				})
			}, gmeasure.SamplingConfig{N: 10000, Duration: 5 * time.Second})

			Expect(experiment.Get("get-command").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure Get with many commands", func() {
			experiment := gmeasure.NewExperiment("Get - Many Commands")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 1000; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("get-1000", func() {
					_, _ = sh.Get("cmd500")
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})
	})

	Describe("Walk Performance", func() {
		It("should measure Walk with small command set", func() {
			experiment := gmeasure.NewExperiment("Walk - Small Set")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 10; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("walk-10", func() {
					sh.Walk(func(name string, item command.Command) bool {
						return true
					})
				})
			}, gmeasure.SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})

		It("should measure Walk with large command set", func() {
			experiment := gmeasure.NewExperiment("Walk - Large Set")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 1000; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("walk-1000", func() {
					sh.Walk(func(name string, item command.Command) bool {
						return true
					})
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Run Performance", func() {
		It("should measure simple command execution", func() {
			experiment := gmeasure.NewExperiment("Run - Simple Command")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				// Simple no-op command
			})
			sh.Add("", cmd)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("run-simple", func() {
					sh.Run(nil, nil, []string{"test"})
				})
			}, gmeasure.SamplingConfig{N: 10000, Duration: 5 * time.Second})

			Expect(experiment.Get("run-simple").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure command with output", func() {
			experiment := gmeasure.NewExperiment("Run - With Output")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			cmd := command.New("test", "Test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "test output")
			})
			sh.Add("", cmd)

			experiment.Sample(func(idx int) {
				outBuf := newSafeBuffer()
				experiment.MeasureDuration("run-output", func() {
					sh.Run(outBuf, nil, []string{"test"})
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})
	})

	Describe("Concurrent Operations", func() {
		It("should measure concurrent Get operations", func() {
			experiment := gmeasure.NewExperiment("Concurrent - Get")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 100; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent-get", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func(id int) {
							_, _ = sh.Get(fmt.Sprintf("cmd%d", id))
							done <- true
						}(i)
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})
		})

		It("should measure concurrent Add operations", func() {
			experiment := gmeasure.NewExperiment("Concurrent - Add")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				sh := shell.New(nil)
				experiment.MeasureDuration("concurrent-add", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func(id int) {
							cmd := command.New(fmt.Sprintf("cmd%d", id), "Test", nil)
							sh.Add("", cmd)
							done <- true
						}(i)
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})
		})

		It("should measure concurrent Walk operations", func() {
			experiment := gmeasure.NewExperiment("Concurrent - Walk")
			AddReportEntry(experiment.Name, experiment)

			sh := shell.New(nil)
			for i := 0; i < 50; i++ {
				sh.Add("", command.New(fmt.Sprintf("cmd%d", i), "Test", nil))
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent-walk", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func() {
							count := 0
							sh.Walk(func(name string, item command.Command) bool {
								count++
								return true
							})
							done <- true
						}()
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Memory Efficiency", func() {
		It("should measure memory usage with many commands", func() {
			experiment := gmeasure.NewExperiment("Memory - Many Commands")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				sh := shell.New(nil)
				for i := 0; i < 1000; i++ {
					sh.Add("", command.New(fmt.Sprintf("cmd%d", i), fmt.Sprintf("Command %d", i), func(out, err io.Writer, args []string) {
						fmt.Fprint(out, "test")
					}))
				}
				experiment.RecordValue("commands", 1000)
			}, gmeasure.SamplingConfig{N: 10, Duration: 5 * time.Second})
		})
	})
})
