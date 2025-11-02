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
	"strings"

	"github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("License Functions", func() {
	var (
		testPackage     = "TestApp"
		testDescription = "Test Application"
		testBuild       = "abc123def"
		testRelease     = "v1.2.3"
		testAuthor      = "Test Author"
		testPrefix      = "test"
	)

	Describe("GetLicenseName", func() {
		Context("for each license type", func() {
			It("should return correct name for MIT", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(Equal("MIT License"))
			})

			It("should return correct name for GNU GPL v3", func() {
				v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("GNU GENERAL PUBLIC LICENSE"))
				Expect(name).To(ContainSubstring("Version 3"))
			})

			It("should return correct name for GNU Affero GPL v3", func() {
				v := version.NewVersion(version.License_GNU_Affero_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("GNU AFFERO GENERAL PUBLIC LICENSE"))
				Expect(name).To(ContainSubstring("Version 3"))
			})

			It("should return correct name for GNU Lesser GPL v3", func() {
				v := version.NewVersion(version.License_GNU_Lesser_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("GNU LESSER GENERAL PUBLIC LICENSE"))
				Expect(name).To(ContainSubstring("Version 3"))
			})

			It("should return correct name for Mozilla PL v2", func() {
				v := version.NewVersion(version.License_Mozilla_PL_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("Mozilla Public License"))
				Expect(name).To(ContainSubstring("Version 2.0"))
			})

			It("should return correct name for Apache v2", func() {
				v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("Apache License"))
				Expect(name).To(ContainSubstring("Version 2.0"))
			})

			It("should return correct name for Unlicense", func() {
				v := version.NewVersion(version.License_Unlicense, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				Expect(v.GetLicenseName()).To(Equal("Free and unencumbered software"))
			})

			It("should return correct name for Creative Commons Zero v1", func() {
				v := version.NewVersion(version.License_Creative_Common_Zero_v1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("Creative Commons"))
				Expect(name).To(ContainSubstring("CC0 1.0 Universal"))
			})

			It("should return correct name for Creative Commons Attribution v4", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("Creative Commons"))
				Expect(name).To(ContainSubstring("Attribution 4.0 International"))
			})

			It("should return correct name for Creative Commons Attribution Share Alike v4", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_Share_Alike_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("Creative Commons"))
				Expect(name).To(ContainSubstring("Attribution-ShareAlike 4.0 International"))
			})

			It("should return correct name for SIL Open Font License 1.1", func() {
				v := version.NewVersion(version.License_SIL_Open_Font_1_1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				name := v.GetLicenseName()
				Expect(name).To(ContainSubstring("SIL OPEN FONT LICENSE"))
				Expect(name).To(ContainSubstring("Version 1.1"))
			})
		})
	})

	Describe("GetLicenseLegal", func() {
		Context("without additional licenses", func() {
			It("should return full legal text for MIT", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				legal := v.GetLicenseLegal()
				Expect(legal).ToNot(BeEmpty())
				Expect(legal).To(ContainSubstring("MIT License"))
				Expect(legal).To(ContainSubstring("Permission is hereby granted"))
			})

			It("should return full legal text for Apache v2", func() {
				v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				legal := v.GetLicenseLegal()
				Expect(legal).ToNot(BeEmpty())
				Expect(legal).To(ContainSubstring("Apache License"))
			})

			It("should return full legal text for GPL v3", func() {
				v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				legal := v.GetLicenseLegal()
				Expect(legal).ToNot(BeEmpty())
				Expect(legal).To(ContainSubstring("GNU GENERAL PUBLIC LICENSE"))
			})
		})

		Context("with additional licenses", func() {
			It("should concatenate multiple licenses with separators", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				legal := v.GetLicenseLegal(version.License_Apache_v2)

				Expect(legal).To(ContainSubstring("MIT License"))
				Expect(legal).To(ContainSubstring("Apache License"))
				// Should have separator between licenses
				Expect(strings.Count(legal, strings.Repeat("*", 80))).To(Equal(2))
			})

			It("should handle multiple additional licenses", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				legal := v.GetLicenseLegal(
					version.License_Apache_v2,
					version.License_Mozilla_PL_v2,
				)

				Expect(legal).To(ContainSubstring("MIT License"))
				Expect(legal).To(ContainSubstring("Apache License"))
				Expect(legal).To(ContainSubstring("Mozilla Public License"))
				// Should have separators (2 for each additional license)
				Expect(strings.Count(legal, strings.Repeat("*", 80))).To(Equal(4))
			})
		})
	})

	Describe("GetLicenseBoiler", func() {
		Context("without additional licenses", func() {
			It("should return boilerplate with package info for MIT", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("MIT License"))
				Expect(boiler).To(ContainSubstring("2024"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})

			It("should return boilerplate with package info for Apache v2", func() {
				v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("Apache License"))
				Expect(boiler).To(ContainSubstring("2024"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})

			It("should return boilerplate with package and description for GPL v3", func() {
				v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring(testPackage))
				Expect(boiler).To(ContainSubstring(testDescription))
				Expect(boiler).To(ContainSubstring("2024"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})

			It("should return boilerplate for AGPL v3", func() {
				v := version.NewVersion(version.License_GNU_Affero_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring(testPackage))
				Expect(boiler).To(ContainSubstring(testDescription))
				Expect(boiler).To(ContainSubstring("GNU Affero General Public License"))
			})

			It("should return boilerplate for LGPL v3", func() {
				v := version.NewVersion(version.License_GNU_Lesser_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("GNU Lesser General Public License"))
			})

			It("should return boilerplate for Mozilla PL v2", func() {
				v := version.NewVersion(version.License_Mozilla_PL_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("Mozilla Public"))
				Expect(boiler).To(ContainSubstring(testPackage))
			})

			It("should return boilerplate for Unlicense", func() {
				v := version.NewVersion(version.License_Unlicense, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("free and unencumbered software"))
			})

			It("should return boilerplate for CC0 v1", func() {
				v := version.NewVersion(version.License_Creative_Common_Zero_v1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("2024"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})

			It("should return boilerplate for CC BY 4", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("Creative Commons Attribution"))
			})

			It("should return boilerplate for CC SA 4", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_Share_Alike_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("Creative Commons"))
				Expect(boiler).To(ContainSubstring("Share Alike"))
			})

			It("should return boilerplate for SIL OFL 1.1", func() {
				v := version.NewVersion(version.License_SIL_Open_Font_1_1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()

				Expect(boiler).ToNot(BeEmpty())
				Expect(boiler).To(ContainSubstring("SIL Open Font License"))
			})
		})

		Context("with additional licenses", func() {
			It("should concatenate multiple boilerplates", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler(version.License_Apache_v2)

				Expect(boiler).To(ContainSubstring("MIT License"))
				Expect(boiler).To(ContainSubstring("Apache License"))
				Expect(boiler).To(ContainSubstring("2024"))
			})

			It("should use same year for all licenses", func() {
				v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler(
					version.License_Apache_v2,
					version.License_GNU_GPL_v3,
				)

				// All should have the same year (2024 from testTime)
				yearCount := strings.Count(boiler, "2024")
				Expect(yearCount).To(BeNumerically(">=", 3))
			})
		})
	})

	Describe("GetLicenseFull", func() {
		It("should include both boilerplate and legal text", func() {
			v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			full := v.GetLicenseFull()

			// Should contain boilerplate elements
			Expect(full).To(ContainSubstring("2024"))
			Expect(full).To(ContainSubstring(testAuthor))

			// Should contain legal text
			Expect(full).To(ContainSubstring("Permission is hereby granted"))

			// Should have separator between boilerplate and legal
			Expect(full).To(ContainSubstring(strings.Repeat("*", 80)))
		})

		It("should handle additional licenses", func() {
			v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			full := v.GetLicenseFull(version.License_Apache_v2)

			Expect(full).To(ContainSubstring("MIT License"))
			Expect(full).To(ContainSubstring("Apache License"))

			// Should have multiple separators
			separatorCount := strings.Count(full, strings.Repeat("*", 80))
			Expect(separatorCount).To(BeNumerically(">", 2))
		})

		It("should be comprehensive for MIT", func() {
			v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			full := v.GetLicenseFull()
			Expect(full).ToNot(BeEmpty())
			Expect(full).To(ContainSubstring(testAuthor))
		})

		It("should be comprehensive for Apache v2", func() {
			v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			full := v.GetLicenseFull()
			Expect(full).ToNot(BeEmpty())
			Expect(full).To(ContainSubstring(testAuthor))
		})

		It("should be comprehensive for GPL v3", func() {
			v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			full := v.GetLicenseFull()
			Expect(full).ToNot(BeEmpty())
			Expect(full).To(ContainSubstring(testAuthor))
		})
	})

	Describe("License Content Validation", func() {
		It("should have non-empty content for MIT", func() {
			v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			Expect(v.GetLicenseName()).ToNot(BeEmpty())
			Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
			Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
			Expect(v.GetLicenseFull()).ToNot(BeEmpty())
		})

		It("should have non-empty content for Apache v2", func() {
			v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			Expect(v.GetLicenseName()).ToNot(BeEmpty())
			Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
			Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
			Expect(v.GetLicenseFull()).ToNot(BeEmpty())
		})

		It("should have non-empty content for Unlicense", func() {
			v := version.NewVersion(version.License_Unlicense, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			Expect(v.GetLicenseName()).ToNot(BeEmpty())
			Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
			Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
			Expect(v.GetLicenseFull()).ToNot(BeEmpty())
		})
	})

	Describe("License Consistency", func() {
		It("should have boilerplate length <= legal text length for detailed licenses", func() {
			v := version.NewVersion(version.License_Apache_v2, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			boiler := v.GetLicenseBoiler()
			legal := v.GetLicenseLegal()

			// MIT has shorter boilerplate than full legal text
			Expect(len(boiler)).To(BeNumerically("<=", len(legal)))
		})

		It("should have full license include all parts", func() {
			v := version.NewVersion(version.License_Creative_Common_Zero_v1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			boiler := v.GetLicenseBoiler()
			legal := v.GetLicenseLegal()
			full := v.GetLicenseFull()

			// Full should be longer than both parts
			Expect(len(full)).To(BeNumerically(">", len(boiler)))
			Expect(len(full)).To(BeNumerically(">", len(legal)))
		})
	})

	Describe("Special License Cases", func() {
		Context("Unlicense", func() {
			It("should return same content for boiler and legal", func() {
				v := version.NewVersion(version.License_Unlicense, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()
				legal := v.GetLicenseLegal()

				Expect(boiler).To(Equal(legal))
			})
		})

		Context("Creative Commons licenses", func() {
			It("should include copyright information in CC0 boilerplate", func() {
				v := version.NewVersion(version.License_Creative_Common_Zero_v1, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()
				Expect(boiler).To(ContainSubstring("Copyright"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})

			It("should include copyright information in CC BY boilerplate", func() {
				v := version.NewVersion(version.License_Creative_Common_Attribution_v4_int, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()
				Expect(boiler).To(ContainSubstring("Copyright"))
				Expect(boiler).To(ContainSubstring(testAuthor))
			})
		})

		Context("GNU licenses", func() {
			It("should include package and description in GPL boilerplate", func() {
				v := version.NewVersion(version.License_GNU_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()
				Expect(boiler).To(ContainSubstring(testPackage))
				Expect(boiler).To(ContainSubstring(testDescription))
			})

			It("should include package and description in AGPL boilerplate", func() {
				v := version.NewVersion(version.License_GNU_Affero_GPL_v3, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
				boiler := v.GetLicenseBoiler()
				Expect(boiler).To(ContainSubstring(testPackage))
				Expect(boiler).To(ContainSubstring(testDescription))
			})
		})
	})

	Describe("Concurrent Access", func() {
		It("should safely retrieve license info from multiple goroutines", func() {
			v := version.NewVersion(version.License_MIT, testPackage, testDescription, testTime, testBuild, testRelease, testAuthor, testPrefix, testStruct{}, 0)
			done := make(chan bool, 20)

			for i := 0; i < 20; i++ {
				go func() {
					defer GinkgoRecover()
					Expect(v.GetLicenseName()).ToNot(BeEmpty())
					Expect(v.GetLicenseLegal()).ToNot(BeEmpty())
					Expect(v.GetLicenseBoiler()).ToNot(BeEmpty())
					Expect(v.GetLicenseFull()).ToNot(BeEmpty())
					done <- true
				}()
			}

			for i := 0; i < 20; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})
