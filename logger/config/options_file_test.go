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
	libprm "github.com/nabbar/golib/file/perm"
	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionsFile", func() {
	Describe("Clone", func() {
		Context("with empty options file", func() {
			It("should return a valid clone", func() {
				original := OptionsFile{}
				clone := original.Clone()

				Expect(clone.Filepath).To(BeEmpty())
				Expect(clone.Create).To(BeFalse())
				Expect(clone.LogLevel).To(BeNil())
			})
		})

		Context("with full options file", func() {
			It("should clone all fields correctly", func() {
				fileMode, err := libprm.Parse("0644")
				Expect(err).To(BeNil())
				pathMode, err := libprm.Parse("0755")
				Expect(err).To(BeNil())

				original := OptionsFile{
					LogLevel:         []string{"Debug", "Info", "Warning"},
					Filepath:         "/var/log/application.log",
					Create:           true,
					CreatePath:       true,
					FileMode:         fileMode,
					PathMode:         pathMode,
					DisableStack:     true,
					DisableTimestamp: false,
					EnableTrace:      true,
					EnableAccessLog:  false,
				}

				clone := original.Clone()

				// Verify all fields are copied
				Expect(clone.LogLevel).To(Equal(original.LogLevel))
				Expect(clone.Filepath).To(Equal(original.Filepath))
				Expect(clone.Create).To(Equal(original.Create))
				Expect(clone.CreatePath).To(Equal(original.CreatePath))
				Expect(clone.FileMode).To(Equal(original.FileMode))
				Expect(clone.PathMode).To(Equal(original.PathMode))
				Expect(clone.DisableStack).To(Equal(original.DisableStack))
				Expect(clone.DisableTimestamp).To(Equal(original.DisableTimestamp))
				Expect(clone.EnableTrace).To(Equal(original.EnableTrace))
				Expect(clone.EnableAccessLog).To(Equal(original.EnableAccessLog))

				// Verify it's a deep copy - modify clone and check original
				clone.Filepath = "/tmp/modified.log"
				Expect(original.Filepath).To(Equal("/var/log/application.log"))

				clone.LogLevel[0] = "Modified"
				Expect(original.LogLevel[0]).To(Equal("Debug"))
			})
		})

		Context("with specific flags", func() {
			It("should correctly clone boolean flags", func() {
				original := OptionsFile{
					Create:           true,
					CreatePath:       false,
					DisableStack:     true,
					DisableTimestamp: true,
					EnableTrace:      false,
					EnableAccessLog:  true,
				}

				clone := original.Clone()

				Expect(clone.Create).To(BeTrue())
				Expect(clone.CreatePath).To(BeFalse())
				Expect(clone.DisableStack).To(BeTrue())
				Expect(clone.DisableTimestamp).To(BeTrue())
				Expect(clone.EnableTrace).To(BeFalse())
				Expect(clone.EnableAccessLog).To(BeTrue())
			})
		})
	})

	Describe("Field Types", func() {
		Context("FileMode and PathMode", func() {
			It("should accept valid permissions", func() {
				fileMode, err := libprm.Parse("0644")
				Expect(err).To(BeNil())

				optFile := OptionsFile{
					FileMode: fileMode,
				}

				Expect(optFile.FileMode.String()).To(Equal("0644"))
			})

			It("should handle different permission formats", func() {
				pathMode, err := libprm.Parse("0755")
				Expect(err).To(BeNil())

				optFile := OptionsFile{
					PathMode: pathMode,
				}

				Expect(optFile.PathMode.String()).To(Equal("0755"))
			})
		})

		Context("LogLevel", func() {
			It("should accept multiple log levels", func() {
				optFile := OptionsFile{
					LogLevel: []string{"Debug", "Info", "Warning", "Error", "Fatal", "Critical"},
				}

				Expect(optFile.LogLevel).To(HaveLen(6))
				Expect(optFile.LogLevel).To(ContainElement("Debug"))
				Expect(optFile.LogLevel).To(ContainElement("Fatal"))
			})

			It("should allow empty log level array", func() {
				optFile := OptionsFile{
					LogLevel: []string{},
				}

				Expect(optFile.LogLevel).To(BeEmpty())
			})
		})
	})
})

var _ = Describe("OptionsFiles", func() {
	Describe("Clone", func() {
		Context("with empty slice", func() {
			It("should return an empty slice", func() {
				original := OptionsFiles{}
				clone := original.Clone()

				Expect(clone).To(BeEmpty())
				Expect(clone).ToNot(BeNil())
			})
		})

		Context("with multiple files", func() {
			It("should clone all files correctly", func() {
				original := OptionsFiles{
					{
						Filepath:     "/var/log/app.log",
						Create:       true,
						LogLevel:     []string{"Info", "Warning"},
						EnableTrace:  true,
						DisableStack: false,
					},
					{
						Filepath:         "/var/log/error.log",
						Create:           false,
						LogLevel:         []string{"Error", "Fatal"},
						EnableTrace:      false,
						DisableTimestamp: true,
					},
					{
						Filepath:        "/var/log/debug.log",
						CreatePath:      true,
						LogLevel:        []string{"Debug"},
						EnableAccessLog: true,
					},
				}

				clone := original.Clone()

				// Verify length
				Expect(clone).To(HaveLen(3))

				// Verify first file
				Expect(clone[0].Filepath).To(Equal("/var/log/app.log"))
				Expect(clone[0].Create).To(BeTrue())
				Expect(clone[0].LogLevel).To(Equal([]string{"Info", "Warning"}))

				// Verify second file
				Expect(clone[1].Filepath).To(Equal("/var/log/error.log"))
				Expect(clone[1].Create).To(BeFalse())
				Expect(clone[1].DisableTimestamp).To(BeTrue())

				// Verify third file
				Expect(clone[2].Filepath).To(Equal("/var/log/debug.log"))
				Expect(clone[2].CreatePath).To(BeTrue())
				Expect(clone[2].EnableAccessLog).To(BeTrue())

				// Verify deep copy - modify clone and check original
				clone[0].Filepath = "/tmp/modified.log"
				Expect(original[0].Filepath).To(Equal("/var/log/app.log"))
			})
		})

		Context("with single file", func() {
			It("should clone single file correctly", func() {
				original := OptionsFiles{
					{
						Filepath: "/var/log/single.log",
						Create:   true,
					},
				}

				clone := original.Clone()

				Expect(clone).To(HaveLen(1))
				Expect(clone[0].Filepath).To(Equal("/var/log/single.log"))
				Expect(clone[0].Create).To(BeTrue())
			})
		})
	})

	Describe("Slice Operations", func() {
		Context("appending files", func() {
			It("should allow appending files", func() {
				files := OptionsFiles{
					{Filepath: "/var/log/first.log"},
				}

				files = append(files, OptionsFile{
					Filepath: "/var/log/second.log",
				})

				Expect(files).To(HaveLen(2))
				Expect(files[0].Filepath).To(Equal("/var/log/first.log"))
				Expect(files[1].Filepath).To(Equal("/var/log/second.log"))
			})
		})

		Context("merging slices", func() {
			It("should allow merging two OptionsFiles slices", func() {
				base := OptionsFiles{
					{Filepath: "/var/log/base.log"},
				}
				extension := OptionsFiles{
					{Filepath: "/var/log/ext1.log"},
					{Filepath: "/var/log/ext2.log"},
				}

				merged := append(base, extension...)

				Expect(merged).To(HaveLen(3))
				Expect(merged[0].Filepath).To(Equal("/var/log/base.log"))
				Expect(merged[1].Filepath).To(Equal("/var/log/ext1.log"))
				Expect(merged[2].Filepath).To(Equal("/var/log/ext2.log"))
			})
		})
	})
})
