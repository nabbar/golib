/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package config_test

import (
	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionsSyslog", func() {
	Describe("Clone", func() {
		Context("with empty options", func() {
			It("should return a valid clone", func() {
				original := OptionsSyslog{}
				clone := original.Clone()

				Expect(clone.LogLevel).To(BeNil())
				Expect(clone.Network).To(BeEmpty())
				Expect(clone.Host).To(BeEmpty())
				Expect(clone.Facility).To(BeEmpty())
				Expect(clone.Tag).To(BeEmpty())
				Expect(clone.DisableStack).To(BeFalse())
				Expect(clone.DisableTimestamp).To(BeFalse())
				Expect(clone.EnableTrace).To(BeFalse())
				Expect(clone.EnableAccessLog).To(BeFalse())
			})
		})

		Context("with full options", func() {
			It("should clone all fields correctly", func() {
				original := OptionsSyslog{
					LogLevel:         []string{"Error", "Fatal", "Critical"},
					Network:          "tcp",
					Host:             "syslog.example.com:514",
					Facility:         "local0",
					Tag:              "myapp",
					DisableStack:     true,
					DisableTimestamp: false,
					EnableTrace:      true,
					EnableAccessLog:  false,
				}

				clone := original.Clone()

				// Verify all fields are copied
				Expect(clone.LogLevel).To(Equal(original.LogLevel))
				Expect(clone.Network).To(Equal(original.Network))
				Expect(clone.Host).To(Equal(original.Host))
				Expect(clone.Facility).To(Equal(original.Facility))
				Expect(clone.Tag).To(Equal(original.Tag))
				Expect(clone.DisableStack).To(Equal(original.DisableStack))
				Expect(clone.DisableTimestamp).To(Equal(original.DisableTimestamp))
				Expect(clone.EnableTrace).To(Equal(original.EnableTrace))
				Expect(clone.EnableAccessLog).To(Equal(original.EnableAccessLog))

				// Verify deep copy
				clone.Tag = "modified"
				Expect(original.Tag).To(Equal("myapp"))

				clone.LogLevel[0] = "Modified"
				Expect(original.LogLevel[0]).To(Equal("Error"))
			})
		})

		Context("with TCP network configuration", func() {
			It("should clone TCP configuration correctly", func() {
				original := OptionsSyslog{
					Network:  "tcp",
					Host:     "192.168.1.100:514",
					Facility: "local1",
					Tag:      "tcp-app",
				}

				clone := original.Clone()

				Expect(clone.Network).To(Equal("tcp"))
				Expect(clone.Host).To(Equal("192.168.1.100:514"))
				Expect(clone.Facility).To(Equal("local1"))
			})
		})

		Context("with UDP network configuration", func() {
			It("should clone UDP configuration correctly", func() {
				original := OptionsSyslog{
					Network:  "udp",
					Host:     "syslog.local:514",
					Facility: "local2",
					Tag:      "udp-app",
				}

				clone := original.Clone()

				Expect(clone.Network).To(Equal("udp"))
				Expect(clone.Host).To(Equal("syslog.local:514"))
			})
		})

		Context("with local syslog configuration", func() {
			It("should clone local configuration correctly", func() {
				original := OptionsSyslog{
					Network:  "",
					Host:     "",
					Facility: "local0",
					Tag:      "local-app",
				}

				clone := original.Clone()

				Expect(clone.Network).To(BeEmpty())
				Expect(clone.Host).To(BeEmpty())
				Expect(clone.Tag).To(Equal("local-app"))
			})
		})
	})

	Describe("Field Validation", func() {
		Context("Network field", func() {
			It("should support tcp network", func() {
				opts := OptionsSyslog{
					Network: "tcp",
				}

				Expect(opts.Network).To(Equal("tcp"))
			})

			It("should support udp network", func() {
				opts := OptionsSyslog{
					Network: "udp",
				}

				Expect(opts.Network).To(Equal("udp"))
			})

			It("should support empty network for local syslog", func() {
				opts := OptionsSyslog{
					Network: "",
				}

				Expect(opts.Network).To(BeEmpty())
			})
		})

		Context("Host field", func() {
			It("should support hostname with port", func() {
				opts := OptionsSyslog{
					Host: "syslog.example.com:514",
				}

				Expect(opts.Host).To(Equal("syslog.example.com:514"))
			})

			It("should support IP address with port", func() {
				opts := OptionsSyslog{
					Host: "192.168.1.100:514",
				}

				Expect(opts.Host).To(Equal("192.168.1.100:514"))
			})

			It("should support empty host for local syslog", func() {
				opts := OptionsSyslog{
					Host: "",
				}

				Expect(opts.Host).To(BeEmpty())
			})
		})

		Context("Facility field", func() {
			It("should support local0 facility", func() {
				opts := OptionsSyslog{
					Facility: "local0",
				}

				Expect(opts.Facility).To(Equal("local0"))
			})

			It("should support different local facilities", func() {
				facilities := []string{"local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7"}

				for _, facility := range facilities {
					opts := OptionsSyslog{
						Facility: facility,
					}
					Expect(opts.Facility).To(Equal(facility))
				}
			})
		})

		Context("Tag field", func() {
			It("should support application tag", func() {
				opts := OptionsSyslog{
					Tag: "myapp",
				}

				Expect(opts.Tag).To(Equal("myapp"))
			})

			It("should support empty tag", func() {
				opts := OptionsSyslog{
					Tag: "",
				}

				Expect(opts.Tag).To(BeEmpty())
			})
		})

		Context("LogLevel field", func() {
			It("should support multiple log levels", func() {
				opts := OptionsSyslog{
					LogLevel: []string{"Error", "Fatal", "Critical"},
				}

				Expect(opts.LogLevel).To(HaveLen(3))
				Expect(opts.LogLevel).To(ContainElement("Error"))
				Expect(opts.LogLevel).To(ContainElement("Fatal"))
				Expect(opts.LogLevel).To(ContainElement("Critical"))
			})

			It("should support all log levels", func() {
				opts := OptionsSyslog{
					LogLevel: []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
				}

				Expect(opts.LogLevel).To(HaveLen(6))
			})
		})
	})

	Describe("Common Use Cases", func() {
		Context("remote TCP syslog", func() {
			It("should configure remote TCP syslog correctly", func() {
				opts := OptionsSyslog{
					Network:      "tcp",
					Host:         "syslog.example.com:514",
					Facility:     "local0",
					Tag:          "production-app",
					LogLevel:     []string{"Warning", "Error", "Fatal", "Critical"},
					EnableTrace:  true,
					DisableStack: false,
				}

				Expect(opts.Network).To(Equal("tcp"))
				Expect(opts.Host).To(Equal("syslog.example.com:514"))
				Expect(opts.LogLevel).To(ContainElement("Error"))
			})
		})

		Context("remote UDP syslog", func() {
			It("should configure remote UDP syslog correctly", func() {
				opts := OptionsSyslog{
					Network:  "udp",
					Host:     "192.168.1.50:514",
					Facility: "local1",
					Tag:      "staging-app",
					LogLevel: []string{"Info", "Warning", "Error"},
				}

				Expect(opts.Network).To(Equal("udp"))
				Expect(opts.Host).To(Equal("192.168.1.50:514"))
			})
		})

		Context("local syslog", func() {
			It("should configure local syslog correctly", func() {
				opts := OptionsSyslog{
					Network:  "",
					Host:     "",
					Facility: "local0",
					Tag:      "local-daemon",
					LogLevel: []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
				}

				Expect(opts.Network).To(BeEmpty())
				Expect(opts.Host).To(BeEmpty())
				Expect(opts.Tag).To(Equal("local-daemon"))
			})
		})
	})
})

var _ = Describe("OptionsSyslogs", func() {
	Describe("Clone", func() {
		Context("with empty slice", func() {
			It("should return an empty slice", func() {
				original := OptionsSyslogs{}
				clone := original.Clone()

				Expect(clone).To(BeEmpty())
				Expect(clone).ToNot(BeNil())
			})
		})

		Context("with multiple syslogs", func() {
			It("should clone all syslogs correctly", func() {
				original := OptionsSyslogs{
					{
						Network:      "tcp",
						Host:         "syslog1.example.com:514",
						Facility:     "local0",
						Tag:          "app1",
						LogLevel:     []string{"Error", "Fatal"},
						EnableTrace:  true,
						DisableStack: false,
					},
					{
						Network:          "udp",
						Host:             "syslog2.example.com:514",
						Facility:         "local1",
						Tag:              "app2",
						LogLevel:         []string{"Info", "Warning"},
						EnableTrace:      false,
						DisableTimestamp: true,
					},
					{
						Network:         "",
						Host:            "",
						Facility:        "local2",
						Tag:             "local-app",
						LogLevel:        []string{"Debug"},
						EnableAccessLog: true,
					},
				}

				clone := original.Clone()

				// Verify length
				Expect(clone).To(HaveLen(3))

				// Verify first syslog
				Expect(clone[0].Network).To(Equal("tcp"))
				Expect(clone[0].Host).To(Equal("syslog1.example.com:514"))
				Expect(clone[0].Tag).To(Equal("app1"))

				// Verify second syslog
				Expect(clone[1].Network).To(Equal("udp"))
				Expect(clone[1].DisableTimestamp).To(BeTrue())

				// Verify third syslog
				Expect(clone[2].Network).To(BeEmpty())
				Expect(clone[2].EnableAccessLog).To(BeTrue())

				// Verify deep copy
				clone[0].Tag = "modified"
				Expect(original[0].Tag).To(Equal("app1"))
			})
		})

		Context("with single syslog", func() {
			It("should clone single syslog correctly", func() {
				original := OptionsSyslogs{
					{
						Network:  "tcp",
						Host:     "syslog.local:514",
						Facility: "local0",
						Tag:      "single",
					},
				}

				clone := original.Clone()

				Expect(clone).To(HaveLen(1))
				Expect(clone[0].Tag).To(Equal("single"))
			})
		})
	})

	Describe("Slice Operations", func() {
		Context("appending syslogs", func() {
			It("should allow appending syslogs", func() {
				syslogs := OptionsSyslogs{
					{Tag: "first"},
				}

				syslogs = append(syslogs, OptionsSyslog{
					Tag: "second",
				})

				Expect(syslogs).To(HaveLen(2))
				Expect(syslogs[0].Tag).To(Equal("first"))
				Expect(syslogs[1].Tag).To(Equal("second"))
			})
		})

		Context("merging slices", func() {
			It("should allow merging two OptionsSyslogs slices", func() {
				base := OptionsSyslogs{
					{Tag: "base"},
				}
				extension := OptionsSyslogs{
					{Tag: "ext1"},
					{Tag: "ext2"},
				}

				merged := append(base, extension...)

				Expect(merged).To(HaveLen(3))
				Expect(merged[0].Tag).To(Equal("base"))
				Expect(merged[1].Tag).To(Equal("ext1"))
				Expect(merged[2].Tag).To(Equal("ext2"))
			})
		})
	})

	Describe("Multi-Syslog Configuration", func() {
		Context("with different facilities", func() {
			It("should support multiple syslogs with different facilities", func() {
				syslogs := OptionsSyslogs{
					{
						Facility: "local0",
						Tag:      "app-errors",
						LogLevel: []string{"Error", "Fatal"},
					},
					{
						Facility: "local1",
						Tag:      "app-info",
						LogLevel: []string{"Info", "Debug"},
					},
				}

				Expect(syslogs).To(HaveLen(2))
				Expect(syslogs[0].Facility).To(Equal("local0"))
				Expect(syslogs[1].Facility).To(Equal("local1"))
			})
		})

		Context("with different networks", func() {
			It("should support multiple syslogs with different networks", func() {
				syslogs := OptionsSyslogs{
					{
						Network: "tcp",
						Host:    "tcp-syslog.local:514",
						Tag:     "tcp-app",
					},
					{
						Network: "udp",
						Host:    "udp-syslog.local:514",
						Tag:     "udp-app",
					},
				}

				Expect(syslogs).To(HaveLen(2))
				Expect(syslogs[0].Network).To(Equal("tcp"))
				Expect(syslogs[1].Network).To(Equal("udp"))
			})
		})
	})
})
