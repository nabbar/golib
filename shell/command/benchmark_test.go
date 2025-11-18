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

package command_test

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nabbar/golib/shell/command"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmarks", func() {
	Describe("Command Creation Performance", func() {
		It("should measure New function performance", func() {
			experiment := NewExperiment("Command Creation - New")
			AddReportEntry(experiment.Name, experiment)

			fn := func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "test")
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create", func() {
					_ = command.New("test", "description", fn)
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})

		It("should measure Info function performance", func() {
			experiment := NewExperiment("CommandInfo Creation - Info")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create-info", func() {
					_ = command.Info("test", "description")
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})

		It("should measure creation with varying name lengths", func() {
			experiment := NewExperiment("Command Creation - Variable Name Length")
			AddReportEntry(experiment.Name, experiment)

			nameLengths := []int{10, 100, 1000, 10000}

			for _, length := range nameLengths {
				name := strings.Repeat("a", length)
				measurementName := fmt.Sprintf("name-len-%d", length)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(measurementName, func() {
						_ = command.New(name, "description", nil)
					})
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})
			}
		})
	})

	Describe("Method Call Performance", func() {
		var cmd command.Command

		BeforeEach(func() {
			cmd = command.New("benchmark", "benchmark command", func(out, err io.Writer, args []string) {
				// Minimal function
			})
		})

		It("should measure Name method performance", func() {
			experiment := NewExperiment("Method Call - Name")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("name", func() {
					_ = cmd.Name()
				})
			}, SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})

		It("should measure Describe method performance", func() {
			experiment := NewExperiment("Method Call - Describe")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("describe", func() {
					_ = cmd.Describe()
				})
			}, SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})

		It("should measure Run method with nil function", func() {
			experiment := NewExperiment("Method Call - Run (nil function)")
			AddReportEntry(experiment.Name, experiment)

			cmdNil := command.New("test", "test", nil)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("run-nil", func() {
					cmdNil.Run(nil, nil, nil)
				})
			}, SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})

		It("should measure Run method with minimal function", func() {
			experiment := NewExperiment("Method Call - Run (minimal function)")
			AddReportEntry(experiment.Name, experiment)

			cmdMinimal := command.New("test", "test", func(out, err io.Writer, args []string) {
				// Do nothing
			})

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("run-minimal", func() {
					cmdMinimal.Run(nil, nil, nil)
				})
			}, SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})
	})

	Describe("Run Method Performance Scenarios", func() {
		It("should measure Run with output writing", func() {
			experiment := NewExperiment("Run - With Output Writing")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "test output")
			})

			outBuf := newSafeBuffer()

			experiment.Sample(func(idx int) {
				outBuf.Reset()
				experiment.MeasureDuration("run-with-output", func() {
					cmd.Run(outBuf, nil, nil)
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})

		It("should measure Run with argument processing", func() {
			experiment := NewExperiment("Run - With Argument Processing")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				for _, arg := range args {
					fmt.Fprint(out, arg)
				}
			})

			outBuf := newSafeBuffer()
			args := []string{"arg1", "arg2", "arg3", "arg4", "arg5"}

			experiment.Sample(func(idx int) {
				outBuf.Reset()
				experiment.MeasureDuration("run-with-args", func() {
					cmd.Run(outBuf, nil, args)
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})

		It("should measure Run with varying argument counts", func() {
			experiment := NewExperiment("Run - Variable Argument Count")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				for _, arg := range args {
					fmt.Fprint(out, arg)
				}
			})

			argCounts := []int{0, 1, 10, 100, 1000}
			outBuf := newSafeBuffer()

			for _, count := range argCounts {
				args := make([]string, count)
				for i := range args {
					args[i] = fmt.Sprintf("arg%d", i)
				}

				measurementName := fmt.Sprintf("args-%d", count)

				experiment.Sample(func(idx int) {
					outBuf.Reset()
					experiment.MeasureDuration(measurementName, func() {
						cmd.Run(outBuf, nil, args)
					})
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})
			}
		})

		It("should measure Run with large output", func() {
			experiment := NewExperiment("Run - Large Output")
			AddReportEntry(experiment.Name, experiment)

			largeData := strings.Repeat("x", 10000)
			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, largeData)
			})

			outBuf := newSafeBuffer()

			experiment.Sample(func(idx int) {
				outBuf.Reset()
				experiment.MeasureDuration("run-large-output", func() {
					cmd.Run(outBuf, nil, nil)
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Concurrent Performance", func() {
		It("should measure concurrent Name calls", func() {
			experiment := NewExperiment("Concurrent - Name Calls")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "description", nil)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent-name", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func() {
							_ = cmd.Name()
							done <- true
						}()
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})

		It("should measure concurrent Run calls", func() {
			experiment := NewExperiment("Concurrent - Run Calls")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "test")
			})

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent-run", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func() {
							buf := newSafeBuffer()
							cmd.Run(buf, nil, nil)
							done <- true
						}()
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Memory Efficiency", func() {
		It("should measure memory impact of command creation", func() {
			experiment := NewExperiment("Memory - Command Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.RecordValue("commands-created", float64(idx+1))

				// Create multiple commands
				for i := 0; i < 100; i++ {
					_ = command.New(
						fmt.Sprintf("cmd%d", i),
						fmt.Sprintf("description%d", i),
						func(out, err io.Writer, args []string) {
							fmt.Fprint(out, "test")
						},
					)
				}
			}, SamplingConfig{N: 10, Duration: 5 * time.Second})
		})

		It("should measure memory efficiency of repeated calls", func() {
			experiment := NewExperiment("Memory - Repeated Calls")
			AddReportEntry(experiment.Name, experiment)

			cmd := command.New("test", "test", func(out, err io.Writer, args []string) {
				fmt.Fprint(out, "output")
			})

			outBuf := newSafeBuffer()

			experiment.Sample(func(idx int) {
				outBuf.Reset()
				for i := 0; i < 1000; i++ {
					cmd.Run(outBuf, nil, nil)
					outBuf.Reset()
				}
				experiment.RecordValue("iterations", 1000)
			}, SamplingConfig{N: 10, Duration: 5 * time.Second})
		})
	})

	Describe("Comparison Benchmarks", func() {
		It("should compare New vs Info performance", func() {
			experiment := NewExperiment("Comparison - New vs Info")
			AddReportEntry(experiment.Name, experiment)

			fn := func(out, err io.Writer, args []string) {}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("new", func() {
					_ = command.New("test", "desc", fn)
				})
				experiment.MeasureDuration("info", func() {
					_ = command.Info("test", "desc")
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})
		})

		It("should compare Run with nil vs simple function", func() {
			experiment := NewExperiment("Comparison - Nil vs Simple Function")
			AddReportEntry(experiment.Name, experiment)

			cmdNil := command.New("nil", "nil", nil)
			cmdSimple := command.New("simple", "simple", func(out, err io.Writer, args []string) {})

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("run-nil", func() {
					cmdNil.Run(nil, nil, nil)
				})
				experiment.MeasureDuration("run-simple", func() {
					cmdSimple.Run(nil, nil, nil)
				})
			}, SamplingConfig{N: 10000, Duration: 5 * time.Second})
		})
	})
})
