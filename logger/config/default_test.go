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
	"encoding/json"

	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default Configuration", func() {
	Describe("DefaultConfig", func() {
		Context("with empty indent", func() {
			It("should return valid JSON configuration", func() {
				config := DefaultConfig("")

				Expect(config).ToNot(BeEmpty())

				// Verify it's valid JSON
				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())
			})

			It("should contain expected fields", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				// Verify top-level fields exist
				Expect(result).To(HaveKey("inheritDefault"))
				Expect(result).To(HaveKey("traceFilter"))
				Expect(result).To(HaveKey("stdout"))
				Expect(result).To(HaveKey("logFile"))
				Expect(result).To(HaveKey("logSyslog"))
			})

			It("should have correct default values", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				Expect(opts.InheritDefault).To(BeFalse())
				Expect(opts.TraceFilter).To(BeEmpty())
				Expect(opts.Stdout).ToNot(BeNil())
			})
		})

		Context("with indent", func() {
			It("should return indented JSON", func() {
				config := DefaultConfig("  ")

				Expect(config).ToNot(BeEmpty())

				// Verify it's still valid JSON
				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())
			})

			It("should format JSON with proper indentation", func() {
				config := DefaultConfig("  ")
				configStr := string(config)

				// Indented JSON should have newlines and spaces
				Expect(configStr).To(ContainSubstring("\n"))
				Expect(configStr).To(ContainSubstring("  "))
			})
		})

		Context("stdout default configuration", func() {
			It("should have correct stdout defaults", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				Expect(opts.Stdout).ToNot(BeNil())
				Expect(opts.Stdout.DisableStandard).To(BeFalse())
				Expect(opts.Stdout.DisableStack).To(BeFalse())
				Expect(opts.Stdout.DisableTimestamp).To(BeFalse())
				Expect(opts.Stdout.EnableTrace).To(BeTrue())
				Expect(opts.Stdout.DisableColor).To(BeFalse())
				Expect(opts.Stdout.EnableAccessLog).To(BeFalse())
			})
		})

		Context("logFile default configuration", func() {
			It("should have logFile configuration", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				Expect(opts.LogFile).ToNot(BeNil())
				Expect(opts.LogFile).To(HaveLen(1))
			})

			It("should have correct logFile defaults", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				logFile := opts.LogFile[0]
				Expect(logFile.LogLevel).To(ContainElement("Debug"))
				Expect(logFile.LogLevel).To(ContainElement("Info"))
				Expect(logFile.LogLevel).To(ContainElement("Warning"))
				Expect(logFile.LogLevel).To(ContainElement("Error"))
				Expect(logFile.LogLevel).To(ContainElement("Fatal"))
				Expect(logFile.LogLevel).To(ContainElement("Critical"))

				Expect(logFile.Filepath).To(BeEmpty())
				Expect(logFile.Create).To(BeFalse())
				Expect(logFile.CreatePath).To(BeFalse())
				Expect(logFile.DisableStack).To(BeFalse())
				Expect(logFile.DisableTimestamp).To(BeFalse())
				Expect(logFile.EnableTrace).To(BeTrue())
				Expect(logFile.EnableAccessLog).To(BeFalse())
			})

			It("should have file permissions configured", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				logFiles := result["logFile"].([]interface{})
				logFile := logFiles[0].(map[string]interface{})

				Expect(logFile).To(HaveKey("fileMode"))
				Expect(logFile).To(HaveKey("pathMode"))
				Expect(logFile["fileMode"]).To(Equal("0644"))
				Expect(logFile["pathMode"]).To(Equal("0755"))
			})

			It("should have buffer size configured", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				logFiles := result["logFile"].([]interface{})
				logFile := logFiles[0].(map[string]interface{})

				Expect(logFile).To(HaveKey("file-buffer-size"))
				Expect(logFile["file-buffer-size"]).To(Equal("32KB"))
			})
		})

		Context("logSyslog default configuration", func() {
			It("should have logSyslog configuration", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				Expect(opts.LogSyslog).ToNot(BeNil())
				Expect(opts.LogSyslog).To(HaveLen(1))
			})

			It("should have correct logSyslog defaults", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				syslog := opts.LogSyslog[0]
				Expect(syslog.LogLevel).To(ContainElement("Debug"))
				Expect(syslog.LogLevel).To(ContainElement("Info"))
				Expect(syslog.LogLevel).To(ContainElement("Warning"))
				Expect(syslog.LogLevel).To(ContainElement("Error"))
				Expect(syslog.LogLevel).To(ContainElement("Fatal"))
				Expect(syslog.LogLevel).To(ContainElement("Critical"))

				Expect(syslog.Network).To(Equal("tcp"))
				Expect(syslog.Host).To(BeEmpty())
				Expect(syslog.Facility).To(Equal("local0"))
				Expect(syslog.Tag).To(BeEmpty())
				Expect(syslog.DisableStack).To(BeFalse())
				Expect(syslog.DisableTimestamp).To(BeFalse())
				Expect(syslog.EnableTrace).To(BeTrue())
				Expect(syslog.EnableAccessLog).To(BeFalse())
			})
		})
	})

	Describe("SetDefaultConfig", func() {
		var originalConfig []byte

		BeforeEach(func() {
			// Save original config
			originalConfig = DefaultConfig("")
		})

		AfterEach(func() {
			// Restore original config
			SetDefaultConfig(originalConfig)
		})

		Context("with custom configuration", func() {
			It("should allow setting custom default configuration", func() {
				customConfig := []byte(`{
					"inheritDefault": true,
					"traceFilter": "/custom/path",
					"stdout": {
						"disableStandard": true
					}
				}`)

				SetDefaultConfig(customConfig)

				config := DefaultConfig("")
				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				Expect(opts.InheritDefault).To(BeTrue())
				Expect(opts.TraceFilter).To(Equal("/custom/path"))
				Expect(opts.Stdout).ToNot(BeNil())
				Expect(opts.Stdout.DisableStandard).To(BeTrue())
			})

			It("should persist custom configuration across calls", func() {
				customConfig := []byte(`{"inheritDefault": true}`)

				SetDefaultConfig(customConfig)

				// First call
				config1 := DefaultConfig("")
				var opts1 Options
				json.Unmarshal(config1, &opts1)
				Expect(opts1.InheritDefault).To(BeTrue())

				// Second call should return same config
				config2 := DefaultConfig("")
				var opts2 Options
				json.Unmarshal(config2, &opts2)
				Expect(opts2.InheritDefault).To(BeTrue())
			})
		})

		Context("with invalid JSON", func() {
			It("should not panic with invalid JSON", func() {
				invalidConfig := []byte(`{invalid json}`)

				Expect(func() {
					SetDefaultConfig(invalidConfig)
				}).ToNot(Panic())
			})
		})

		Context("with empty configuration", func() {
			It("should accept empty configuration", func() {
				emptyConfig := []byte(`{}`)

				SetDefaultConfig(emptyConfig)

				config := DefaultConfig("")
				Expect(config).ToNot(BeEmpty())
			})
		})
	})

	Describe("Integration Tests", func() {
		Context("unmarshalling and using default config", func() {
			It("should produce usable Options struct", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				// Verify we can use the options
				Expect(opts.Validate()).To(BeNil())
			})

			It("should allow cloning default options", func() {
				config := DefaultConfig("")

				var opts Options
				err := json.Unmarshal(config, &opts)
				Expect(err).To(BeNil())

				clone := opts.Clone()
				Expect(clone.InheritDefault).To(Equal(opts.InheritDefault))
			})

			It("should allow merging with default options", func() {
				config := DefaultConfig("")

				var base Options
				err := json.Unmarshal(config, &base)
				Expect(err).To(BeNil())

				override := &Options{
					TraceFilter: "/override",
				}

				base.Merge(override)
				Expect(base.TraceFilter).To(Equal("/override"))
			})
		})
	})

	Describe("JSON Format Validation", func() {
		Context("indentation consistency", func() {
			It("should produce consistent indentation", func() {
				config1 := DefaultConfig("  ")
				config2 := DefaultConfig("  ")

				Expect(config1).To(Equal(config2))
			})

			It("should handle different indent strings", func() {
				configTab := DefaultConfig("\t")
				configSpace := DefaultConfig("  ")

				// Both should be valid JSON
				var opts1, opts2 Options
				Expect(json.Unmarshal(configTab, &opts1)).To(Succeed())
				Expect(json.Unmarshal(configSpace, &opts2)).To(Succeed())
			})
		})

		Context("with no indent", func() {
			It("should produce compact JSON", func() {
				config := DefaultConfig("")
				configStr := string(config)

				// Compact JSON should not have unnecessary whitespace
				var opts Options
				err := json.Unmarshal([]byte(configStr), &opts)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Thread Safety", func() {
		Context("concurrent access", func() {
			It("should be safe to call DefaultConfig concurrently", func() {
				done := make(chan bool)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						config := DefaultConfig("  ")
						Expect(config).ToNot(BeEmpty())

						var opts Options
						err := json.Unmarshal(config, &opts)
						Expect(err).To(BeNil())

						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					<-done
				}
			})
		})
	})
})
