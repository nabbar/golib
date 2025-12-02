/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options Model", func() {
	Describe("Clone", func() {
		Context("with empty options", func() {
			It("should return a valid clone", func() {
				original := &Options{}
				clone := original.Clone()

				Expect(clone.InheritDefault).To(Equal(original.InheritDefault))
				Expect(clone.TraceFilter).To(Equal(original.TraceFilter))
				Expect(clone.Stdout).To(BeNil())
			})
		})

		Context("with full options", func() {
			It("should clone all fields correctly", func() {
				original := &Options{
					InheritDefault: true,
					TraceFilter:    "/tmp/trace",
					Stdout: &OptionsStd{
						DisableStandard:  true,
						DisableColor:     true,
						DisableTimestamp: false,
						DisableStack:     true,
						EnableTrace:      true,
						EnableAccessLog:  false,
					},
					LogFileExtend: true,
					LogFile: OptionsFiles{
						{
							LogLevel: []string{"Debug", "Info"},
							Filepath: "/var/log/app.log",
							Create:   true,
						},
					},
					LogSyslogExtend: false,
					LogSyslog: OptionsSyslogs{
						{
							LogLevel: []string{"Error", "Fatal"},
							Network:  "tcp",
							Host:     "localhost:514",
							Tag:      "myapp",
						},
					},
				}

				clone := original.Clone()

				// Verify basic fields
				Expect(clone.InheritDefault).To(Equal(original.InheritDefault))
				Expect(clone.TraceFilter).To(Equal(original.TraceFilter))

				// Verify stdout is deep copied
				Expect(clone.Stdout).ToNot(BeNil())
				Expect(clone.Stdout).ToNot(BeIdenticalTo(original.Stdout))
				Expect(clone.Stdout.DisableStandard).To(Equal(original.Stdout.DisableStandard))
				Expect(clone.Stdout.DisableColor).To(Equal(original.Stdout.DisableColor))

				// Verify log files are deep copied
				Expect(clone.LogFile).To(HaveLen(len(original.LogFile)))
				Expect(clone.LogFile[0].Filepath).To(Equal(original.LogFile[0].Filepath))

				// Verify syslog are deep copied
				Expect(clone.LogSyslog).To(HaveLen(len(original.LogSyslog)))
				Expect(clone.LogSyslog[0].Tag).To(Equal(original.LogSyslog[0].Tag))

				// Modify clone and ensure original is not affected
				clone.TraceFilter = "/tmp/modified"
				Expect(original.TraceFilter).To(Equal("/tmp/trace"))
			})
		})
	})

	Describe("Merge", func() {
		Context("when merging trace filter", func() {
			It("should override trace filter if provided", func() {
				base := &Options{
					TraceFilter: "/original",
				}
				override := &Options{
					TraceFilter: "/override",
				}

				base.Merge(override)
				Expect(base.TraceFilter).To(Equal("/override"))
			})

			It("should keep original trace filter if override is empty", func() {
				base := &Options{
					TraceFilter: "/original",
				}
				override := &Options{
					TraceFilter: "",
				}

				base.Merge(override)
				Expect(base.TraceFilter).To(Equal("/original"))
			})
		})

		Context("when merging stdout options", func() {
			It("should merge all stdout flags", func() {
				base := &Options{
					Stdout: &OptionsStd{
						DisableStandard:  false,
						DisableColor:     false,
						DisableTimestamp: false,
						DisableStack:     false,
						EnableTrace:      false,
						EnableAccessLog:  false,
					},
				}
				override := &Options{
					Stdout: &OptionsStd{
						DisableStandard:  true,
						DisableColor:     true,
						DisableTimestamp: true,
						DisableStack:     true,
						EnableTrace:      true,
						EnableAccessLog:  true,
					},
				}

				base.Merge(override)

				Expect(base.Stdout.DisableStandard).To(BeTrue())
				Expect(base.Stdout.DisableColor).To(BeTrue())
				Expect(base.Stdout.DisableTimestamp).To(BeTrue())
				Expect(base.Stdout.DisableStack).To(BeTrue())
				Expect(base.Stdout.EnableTrace).To(BeTrue())
				Expect(base.Stdout.EnableAccessLog).To(BeTrue())
			})

			It("should create stdout if nil in base", func() {
				base := &Options{}
				override := &Options{
					Stdout: &OptionsStd{
						DisableStandard: true,
					},
				}

				base.Merge(override)

				Expect(base.Stdout).ToNot(BeNil())
				Expect(base.Stdout.DisableStandard).To(BeTrue())
			})

			It("should not merge false values in stdout", func() {
				base := &Options{
					Stdout: &OptionsStd{
						DisableStandard: true,
						EnableTrace:     true,
					},
				}
				override := &Options{
					Stdout: &OptionsStd{
						DisableStandard: false,
						EnableTrace:     false,
					},
				}

				base.Merge(override)

				// False values should not override true values
				Expect(base.Stdout.DisableStandard).To(BeTrue())
				Expect(base.Stdout.EnableTrace).To(BeTrue())
			})
		})

		Context("when merging log files", func() {
			It("should replace log files when extend is false", func() {
				base := &Options{
					LogFile: OptionsFiles{
						{Filepath: "/var/log/base.log"},
					},
				}
				override := &Options{
					LogFileExtend: false,
					LogFile: OptionsFiles{
						{Filepath: "/var/log/override.log"},
					},
				}

				base.Merge(override)

				Expect(base.LogFile).To(HaveLen(1))
				Expect(base.LogFile[0].Filepath).To(Equal("/var/log/override.log"))
			})

			It("should extend log files when extend is true", func() {
				base := &Options{
					LogFile: OptionsFiles{
						{Filepath: "/var/log/base.log"},
					},
				}
				override := &Options{
					LogFileExtend: true,
					LogFile: OptionsFiles{
						{Filepath: "/var/log/override.log"},
					},
				}

				base.Merge(override)

				Expect(base.LogFile).To(HaveLen(2))
				Expect(base.LogFile[0].Filepath).To(Equal("/var/log/base.log"))
				Expect(base.LogFile[1].Filepath).To(Equal("/var/log/override.log"))
			})
		})

		Context("when merging syslogs", func() {
			It("should replace syslogs when extend is false", func() {
				base := &Options{
					LogSyslog: OptionsSyslogs{
						{Tag: "base"},
					},
				}
				override := &Options{
					LogSyslogExtend: false,
					LogSyslog: OptionsSyslogs{
						{Tag: "override"},
					},
				}

				base.Merge(override)

				Expect(base.LogSyslog).To(HaveLen(1))
				Expect(base.LogSyslog[0].Tag).To(Equal("override"))
			})

			It("should extend syslogs when extend is true", func() {
				base := &Options{
					LogSyslog: OptionsSyslogs{
						{Tag: "base"},
					},
				}
				override := &Options{
					LogSyslogExtend: true,
					LogSyslog: OptionsSyslogs{
						{Tag: "override"},
					},
				}

				base.Merge(override)

				Expect(base.LogSyslog).To(HaveLen(2))
				Expect(base.LogSyslog[0].Tag).To(Equal("base"))
				Expect(base.LogSyslog[1].Tag).To(Equal("override"))
			})
		})

		Context("when merging function options", func() {
			It("should override function if provided", func() {
				baseFn := func() *Options {
					return &Options{TraceFilter: "base"}
				}
				overrideFn := func() *Options {
					return &Options{TraceFilter: "override"}
				}

				base := &Options{
					InheritDefault: true,
				}
				base.RegisterDefaultFunc(baseFn)

				override := &Options{}
				override.RegisterDefaultFunc(overrideFn)

				base.Merge(override)

				// Verify the function was overridden
				// Since base has InheritDefault=true, Options() will use the new function
				result := base.Options()
				Expect(result.TraceFilter).To(Equal("override"))
			})
		})
	})

	Describe("Options", func() {
		Context("without inheritance", func() {
			It("should return current options when InheritDefault is false", func() {
				opts := &Options{
					InheritDefault: false,
					TraceFilter:    "/current",
				}

				result := opts.Options()

				Expect(result.TraceFilter).To(Equal("/current"))
			})
		})

		Context("with inheritance", func() {
			It("should inherit from default function", func() {
				defaultFn := func() *Options {
					return &Options{
						TraceFilter: "/default",
						Stdout: &OptionsStd{
							DisableStandard: true,
						},
					}
				}

				opts := &Options{
					InheritDefault: true,
					TraceFilter:    "/override",
				}
				opts.RegisterDefaultFunc(defaultFn)

				result := opts.Options()

				// Current value should override default
				Expect(result.TraceFilter).To(Equal("/override"))
				// Default value should be inherited
				Expect(result.Stdout).ToNot(BeNil())
				Expect(result.Stdout.DisableStandard).To(BeTrue())
			})

			It("should merge stdout options with defaults", func() {
				defaultFn := func() *Options {
					return &Options{
						Stdout: &OptionsStd{
							DisableStandard: true,
							DisableColor:    false,
						},
					}
				}

				opts := &Options{
					InheritDefault: true,
					Stdout: &OptionsStd{
						DisableColor: true,
					},
				}
				opts.RegisterDefaultFunc(defaultFn)

				result := opts.Options()

				Expect(result.Stdout.DisableStandard).To(BeTrue()) // From default
				Expect(result.Stdout.DisableColor).To(BeTrue())    // From current
			})

			It("should extend log files when LogFileExtend is true", func() {
				defaultFn := func() *Options {
					return &Options{
						LogFile: OptionsFiles{
							{Filepath: "/default.log"},
						},
					}
				}

				opts := &Options{
					InheritDefault: true,
					LogFileExtend:  true,
					LogFile: OptionsFiles{
						{Filepath: "/current.log"},
					},
				}
				opts.RegisterDefaultFunc(defaultFn)

				result := opts.Options()

				Expect(result.LogFile).To(HaveLen(2))
				Expect(result.LogFile[0].Filepath).To(Equal("/default.log"))
				Expect(result.LogFile[1].Filepath).To(Equal("/current.log"))
			})

			It("should replace log files when LogFileExtend is false", func() {
				defaultFn := func() *Options {
					return &Options{
						LogFile: OptionsFiles{
							{Filepath: "/default.log"},
						},
					}
				}

				opts := &Options{
					InheritDefault: true,
					LogFileExtend:  false,
					LogFile: OptionsFiles{
						{Filepath: "/current.log"},
					},
				}
				opts.RegisterDefaultFunc(defaultFn)

				result := opts.Options()

				Expect(result.LogFile).To(HaveLen(1))
				Expect(result.LogFile[0].Filepath).To(Equal("/current.log"))
			})
		})

		Context("with nil default function", func() {
			It("should not panic when function is nil", func() {
				opts := &Options{
					InheritDefault: true,
					TraceFilter:    "/current",
				}

				result := opts.Options()

				Expect(result.TraceFilter).To(Equal("/current"))
			})
		})
	})

	Describe("RegisterDefaultFunc", func() {
		It("should register a default function", func() {
			opts := &Options{}
			fn := func() *Options {
				return &Options{TraceFilter: "/test"}
			}

			opts.RegisterDefaultFunc(fn)

			// Can't directly test private field, but we can test the behavior
			opts.InheritDefault = true
			result := opts.Options()
			Expect(result.TraceFilter).To(Equal("/test"))
		})

		It("should clear default function when nil is passed", func() {
			opts := &Options{}
			fn := func() *Options {
				return &Options{TraceFilter: "/test"}
			}

			opts.RegisterDefaultFunc(fn)
			opts.RegisterDefaultFunc(nil)

			opts.InheritDefault = true
			result := opts.Options()
			Expect(result.TraceFilter).To(BeEmpty())
		})
	})

	Describe("Validate", func() {
		Context("with valid options", func() {
			It("should return nil for valid options", func() {
				opts := &Options{
					InheritDefault: false,
					TraceFilter:    "/tmp",
					Stdout: &OptionsStd{
						DisableStandard: false,
					},
				}

				err := opts.Validate()
				Expect(err).To(BeNil())
			})

			It("should return nil for empty options", func() {
				opts := &Options{}

				err := opts.Validate()
				Expect(err).To(BeNil())
			})
		})
	})
})
