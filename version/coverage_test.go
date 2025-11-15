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
	"github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Coverage tests focus on improving code coverage by testing edge cases
// and code paths that may not be covered by functional tests.
var _ = Describe("Coverage Improvements", func() {
	var (
		testPackage     = "CoverageTest"
		testDescription = "Coverage Test"
		testBuild       = "coverage123"
		testRelease     = "v0.0.1"
		testAuthor      = "Coverage Author"
		testPrefix      = "cov"
	)

	Describe("License Edge Cases", func() {
		// Note: license type is not exported, so we test through the Version interface
		// All license types are tested via Version.GetLicense* methods

		Context("all license types coverage", func() {
			It("should handle MIT license", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("MIT"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle Apache v2 license", func() {
				v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("Apache"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle GPL v3 license", func() {
				v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("GNU"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle AGPL v3 license", func() {
				v := version.NewVersion(version.License_GNU_Affero_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("AFFERO"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle LGPL v3 license", func() {
				v := version.NewVersion(version.License_GNU_Lesser_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("LESSER"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle Mozilla PL v2 license", func() {
				v := version.NewVersion(version.License_Mozilla_PL_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("Mozilla"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle Unlicense", func() {
				v := version.NewVersion(version.License_Unlicense, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("unencumbered"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle CC0 v1 license", func() {
				v := version.NewVersion(version.License_Creative_Common_Zero_v1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("CC0"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle CC BY 4.0 license", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("Attribution"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle CC SA 4.0 license", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_Share_Alike_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("Share"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})

			It("should handle SIL OFL 1.1 license", func() {
				v := version.NewVersion(version.License_SIL_Open_Font_1_1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(ContainSubstring("SIL"))
				Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
				Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
				Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			})
		})
	})

	Describe("CheckGo Edge Cases", func() {
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

		Context("error handling", func() {
			It("should handle empty version string", func() {
				err := v.CheckGo("", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
			})

			It("should handle empty constraint", func() {
				err := v.CheckGo("1.18", "")
				Expect(err).ToNot(BeNil())
				// Empty constraint "" + "1.18" = "1.18" is treated as exact version match
				// This will likely fail with ErrorGoVersionConstraint unless runtime is exactly 1.18
			})

			It("should handle both empty", func() {
				err := v.CheckGo("", "")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
			})

			It("should handle malformed version", func() {
				err := v.CheckGo("not.a.version", ">=")
				Expect(err).ToNot(BeNil())
			})

			It("should handle malformed constraint", func() {
				err := v.CheckGo("1.18", "invalid-constraint")
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Version Model Edge Cases", func() {
		Context("extreme values", func() {
			It("should handle very long strings", func() {
				longString := string(make([]byte, 10000))
				for i := range longString {
					longString = longString[:i] + "x" + longString[i+1:]
				}

				v := version.NewVersion(
					version.License_MIT,
					longString,
					longString,
					testTime,
					longString,
					longString,
					longString,
					longString,
					testStruct{},
					0,
				)

				Expect(v).ToNot(BeNil())
				Expect(v.GetPackage()).ToNot(BeEmpty())
			})

			It("should handle unicode characters", func() {
				v := version.NewVersion(
					version.License_MIT,
					"测试包",
					"テストパッケージ",
					testTime,
					"сборка",
					"v1.0.0-αβγ",
					"作者 👨‍💻",
					"préfixe",
					testStruct{},
					0,
				)

				Expect(v).ToNot(BeNil())
				Expect(v.GetPackage()).To(ContainSubstring("测试包"))
				Expect(v.GetDescription()).To(ContainSubstring("テストパッケージ"))
			})

			It("should handle negative numSubPackage", func() {
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
					-1,
				)

				Expect(v).ToNot(BeNil())
				path := v.GetRootPackagePath()
				Expect(path).ToNot(BeEmpty())
			})
		})

		Context("date parsing edge cases", func() {
			It("should handle various invalid date formats", func() {
				invalidDates := []string{
					"not-a-date",
					"2024-13-01", // Invalid month
					"2024-01-32", // Invalid day
					"",
					"null",
					"undefined",
				}

				for _, date := range invalidDates {
					v := version.NewVersion(
						version.License_MIT,
						testPackage,
						testDescription,
						date,
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
	})

	Describe("Output Methods Coverage", func() {
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

		It("should generate consistent output across multiple calls", func() {
			// Call each method multiple times to ensure consistency
			for i := 0; i < 5; i++ {
				Expect(v.GetPackage()).To(Equal(testPackage))
				Expect(v.GetDescription()).To(Equal(testDescription))
				Expect(v.GetBuild()).To(Equal(testBuild))
				Expect(v.GetRelease()).To(Equal(testRelease))
			}
		})

		It("should have all output methods return non-empty strings", func() {
			Expect(v.GetAppId()).ToNot(BeEmpty())
			Expect(v.GetAuthor()).ToNot(BeEmpty())
			Expect(v.GetBuild()).ToNot(BeEmpty())
			Expect(v.GetDate()).ToNot(BeEmpty())
			Expect(v.GetDescription()).ToNot(BeEmpty())
			Expect(v.GetHeader()).ToNot(BeEmpty())
			Expect(v.GetInfo()).ToNot(BeEmpty())
			Expect(v.GetPackage()).ToNot(BeEmpty())
			Expect(v.GetRootPackagePath()).ToNot(BeEmpty())
			Expect(v.GetPrefix()).ToNot(BeEmpty())
			Expect(v.GetRelease()).ToNot(BeEmpty())
			Expect(v.GetLicenseName()).ToNot(BeEmpty())
			Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
			Expect(v.GetLicenseFull()).ToNot(BeEmpty())
			Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
		})
	})

	Describe("Multiple License Combinations", func() {
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

		It("should handle all possible license combinations", func() {
			// Test with multiple additional licenses
			legal := v.GetLicenseLegal(
				version.License_Apache_v2,
				version.License_GNU_GPL_v3,
				version.License_Mozilla_PL_v2,
			)
			Expect(legal).To(ContainSubstring("MIT"))
			Expect(legal).To(ContainSubstring("Apache"))
			Expect(legal).To(ContainSubstring("GNU"))
			Expect(legal).To(ContainSubstring("Mozilla"))

			boiler := v.GetLicenseBoiler(
				version.License_Apache_v2,
				version.License_GNU_GPL_v3,
			)
			Expect(boiler).To(ContainSubstring("MIT"))
			Expect(boiler).To(ContainSubstring("Apache"))

			full := v.GetLicenseFull(
				version.License_Apache_v2,
			)
			Expect(full).To(ContainSubstring("MIT"))
			Expect(full).To(ContainSubstring("Apache"))
		})

		It("should handle empty additional licenses", func() {
			legal := v.GetLicenseLegal()
			Expect(legal).To(ContainSubstring("MIT"))

			boiler := v.GetLicenseBoiler()
			Expect(boiler).To(ContainSubstring("MIT"))

			full := v.GetLicenseFull()
			Expect(full).To(ContainSubstring("MIT"))
		})
	})
})
