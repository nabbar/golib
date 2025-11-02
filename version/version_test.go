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

package version_test

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Creation and Getter Methods", func() {
	var (
		testPackage     = "TestApp"
		testDescription = "Test Application"
		testBuild       = "abc123def"
		testRelease     = "v1.2.3"
		testAuthor      = "Test Author"
		testPrefix      = "test"
	)

	Describe("NewVersion", func() {
		Context("with valid parameters", func() {
			It("should create a version instance successfully", func() {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v).ToNot(BeNil())
			})

			It("should parse date correctly", func() {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v.GetTime()).To(Equal(testTimeParsed))
				Expect(v.GetDate()).To(ContainSubstring("2024"))
			})

			It("should use current time for invalid date", func() {
				before := time.Now()
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					"invalid-date",
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)
				after := time.Now()

				parsedTime := v.GetTime()
				Expect(parsedTime).To(BeTemporally(">=", before))
				Expect(parsedTime).To(BeTemporally("<=", after))
			})

			It("should extract package path from reflection", func() {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				path := v.GetRootPackagePath()
				Expect(path).To(ContainSubstring("github.com/nabbar/golib/version"))
			})

			It("should handle numSubPackage correctly", func() {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					1, // Go up one directory
				)

				path := v.GetRootPackagePath()
				Expect(path).To(ContainSubstring("github.com/nabbar/golib"))
				Expect(path).ToNot(ContainSubstring("github.com/nabbar/golib/version"))
			})

			It("should use package from path if empty or noname", func() {
				v := version.NewVersion(
					version.License_MIT,
					"",
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				// Package is extracted from reflection, which includes _test suffix in test context
				Expect(v.GetPackage()).To(Equal("version_test"))
			})

			It("should use noname as empty and derive from path", func() {
				v := version.NewVersion(
					version.License_MIT,
					"noname",
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				// Package is extracted from reflection, which includes _test suffix in test context
				Expect(v.GetPackage()).To(Equal("version_test"))
			})
		})

		Context("with different license types", func() {
			It("should accept MIT license", func() {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v.GetLicenseName()).To(Equal("MIT License"))
			})

			It("should accept GPL v3 license", func() {
				v := version.NewVersion(
					version.License_GNU_GPL_v3,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v.GetLicenseName()).To(ContainSubstring("GNU GENERAL PUBLIC LICENSE"))
			})

			It("should accept Apache v2 license", func() {
				v := version.NewVersion(
					version.License_Apache_v2,
					testPackage,
					testDescription,
					testTime,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v.GetLicenseName()).To(ContainSubstring("Apache License"))
			})
		})
	})

	Describe("Getter Methods", func() {
		var v version.Version

		BeforeEach(func() {
			v = version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)
		})

		It("should return correct package name", func() {
			Expect(v.GetPackage()).To(Equal(testPackage))
		})

		It("should return correct description", func() {
			Expect(v.GetDescription()).To(Equal(testDescription))
		})

		It("should return correct build", func() {
			Expect(v.GetBuild()).To(Equal(testBuild))
		})

		It("should return correct release", func() {
			Expect(v.GetRelease()).To(Equal(testRelease))
		})

		It("should return correct author", func() {
			author := v.GetAuthor()
			Expect(author).To(ContainSubstring(testAuthor))
			Expect(author).To(ContainSubstring("source"))
		})

		It("should return uppercase prefix", func() {
			Expect(v.GetPrefix()).To(Equal(strings.ToUpper(testPrefix)))
		})

		It("should return formatted date", func() {
			date := v.GetDate()
			Expect(date).ToNot(BeEmpty())
			Expect(date).To(ContainSubstring("2024"))
		})

		It("should return time object", func() {
			t := v.GetTime()
			Expect(t).To(Equal(testTimeParsed))
		})

		It("should return correct app ID", func() {
			appId := v.GetAppId()
			Expect(appId).To(ContainSubstring(testRelease))
			Expect(appId).To(ContainSubstring(runtime.GOOS))
			Expect(appId).To(ContainSubstring(runtime.GOARCH))
			Expect(appId).To(ContainSubstring("Runtime"))
		})

		It("should return correct header", func() {
			header := v.GetHeader()
			Expect(header).To(ContainSubstring(testPackage))
			Expect(header).To(ContainSubstring(testRelease))
			Expect(header).To(ContainSubstring(testBuild))
		})

		It("should return correct info", func() {
			info := v.GetInfo()
			Expect(info).To(ContainSubstring("Release"))
			Expect(info).To(ContainSubstring(testRelease))
			Expect(info).To(ContainSubstring("Build"))
			Expect(info).To(ContainSubstring(testBuild))
			Expect(info).To(ContainSubstring("Date"))
		})

		It("should return root package path", func() {
			path := v.GetRootPackagePath()
			Expect(path).ToNot(BeEmpty())
			Expect(path).To(ContainSubstring("github.com"))
		})
	})

	Describe("Print Methods", func() {
		var v version.Version

		BeforeEach(func() {
			v = version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)
		})

		// Note: PrintInfo and PrintLicense methods write to stderr using println().
		// We test them indirectly by verifying the underlying Get methods work correctly.
		// Direct testing would pollute test output.

		It("should have valid data for PrintInfo", func() {
			// PrintInfo uses GetHeader internally
			header := v.GetHeader()
			Expect(header).ToNot(BeEmpty())
			Expect(header).To(ContainSubstring(testPackage))
			Expect(header).To(ContainSubstring(testRelease))
		})

		It("should have valid data for PrintLicense", func() {
			// PrintLicense uses GetLicenseBoiler internally
			boiler := v.GetLicenseBoiler()
			Expect(boiler).ToNot(BeEmpty())
			Expect(boiler).To(ContainSubstring("MIT License"))
		})

		It("should have valid data for PrintLicense with additional licenses", func() {
			// PrintLicense with args uses GetLicenseBoiler with additional licenses
			boiler := v.GetLicenseBoiler(version.License_Apache_v2)
			Expect(boiler).ToNot(BeEmpty())
			Expect(boiler).To(ContainSubstring("MIT License"))
			Expect(boiler).To(ContainSubstring("Apache License"))
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty string values gracefully", func() {
			v := version.NewVersion(
				version.License_MIT,
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				testStruct{},
				0,
			)

			Expect(v).ToNot(BeNil())
			Expect(v.GetPackage()).ToNot(BeEmpty()) // Should derive from path
			Expect(v.GetTime()).ToNot(BeZero())     // Should use current time
		})

		It("should handle very large numSubPackage values", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				100, // Very large value
			)

			Expect(v).ToNot(BeNil())
			path := v.GetRootPackagePath()
			Expect(path).ToNot(BeEmpty())
		})

		It("should handle special characters in fields", func() {
			v := version.NewVersion(
				version.License_MIT,
				"Test™Package©",
				"Description with émojis 🚀",
				testTime,
				"build-123.456",
				"v1.2.3-beta+meta",
				"Author Name <email@example.com>",
				"prefix_with_underscore",
				testStruct{},
				0,
			)

			Expect(v).ToNot(BeNil())
			Expect(v.GetPackage()).To(ContainSubstring("Test"))
			Expect(v.GetDescription()).To(ContainSubstring("émojis"))
			Expect(v.GetAuthor()).To(ContainSubstring("email@example.com"))
		})

		It("should handle different date formats", func() {
			formats := []string{
				"2024-01-15T10:30:00Z",
				"2024-01-15T10:30:00+01:00",
				"2024-01-15T10:30:00.123Z",
			}

			for _, format := range formats {
				v := version.NewVersion(
					version.License_MIT,
					testPackage,
					testDescription,
					format,
					testBuild,
					testRelease,
					testAuthor,
					testPrefix,
					testStruct{},
					0,
				)

				Expect(v).ToNot(BeNil())
				Expect(v.GetTime()).ToNot(BeZero())
			}
		})
	})

	Describe("Concurrency Safety", func() {
		It("should be safe to read from multiple goroutines", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					Expect(v.GetPackage()).To(Equal(testPackage))
					Expect(v.GetRelease()).To(Equal(testRelease))
					Expect(v.GetBuild()).To(Equal(testBuild))
					Expect(v.GetInfo()).ToNot(BeEmpty())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Integration Tests", func() {
		It("should provide complete version information", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)

			// Verify all information is present and formatted correctly
			info := v.GetInfo()
			header := v.GetHeader()
			appId := v.GetAppId()
			author := v.GetAuthor()

			Expect(info).To(ContainSubstring(testRelease))
			Expect(info).To(ContainSubstring(testBuild))
			Expect(header).To(ContainSubstring(testPackage))
			Expect(appId).To(ContainSubstring(testRelease))
			Expect(author).To(ContainSubstring(testAuthor))

			// Verify license information
			Expect(v.GetLicenseName()).ToNot(BeEmpty())
			Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
			Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
			Expect(v.GetLicenseFull()).ToNot(BeEmpty())
		})
	})

	Describe("Package Path Extraction", func() {
		It("should correctly extract package path with numSubPackage=0", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)

			path := v.GetRootPackagePath()
			// In test context, package includes _test suffix
			Expect(path).To(Equal("github.com/nabbar/golib/version_test"))
		})

		It("should correctly extract package path with numSubPackage=1", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				1,
			)

			path := v.GetRootPackagePath()
			Expect(path).To(Equal("github.com/nabbar/golib"))
		})

		It("should correctly extract package path with numSubPackage=2", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				2,
			)

			path := v.GetRootPackagePath()
			Expect(path).To(Equal("github.com/nabbar"))
		})
	})

	Describe("Year Extraction", func() {
		It("should correctly extract year from date", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				testTime,
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)

			// Year should be reflected in boilerplate
			boiler := v.GetLicenseBoiler()
			Expect(boiler).To(ContainSubstring("2024"))
		})

		It("should use current year for invalid date", func() {
			v := version.NewVersion(
				version.License_MIT,
				testPackage,
				testDescription,
				"invalid",
				testBuild,
				testRelease,
				testAuthor,
				testPrefix,
				testStruct{},
				0,
			)

			currentYear := fmt.Sprintf("%d", time.Now().Year())
			boiler := v.GetLicenseBoiler()
			Expect(boiler).To(ContainSubstring(currentYear))
		})
	})
})
