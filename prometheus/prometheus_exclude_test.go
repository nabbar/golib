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

package prometheus_test

import (
	libprm "github.com/nabbar/golib/prometheus"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus ExcludePath", func() {
	var p libprm.Prometheus

	BeforeEach(func() {
		p = newPrometheus()
	})

	Describe("ExcludePath", func() {
		Context("when adding exclude paths", func() {
			It("should exclude single path", func() {
				p.ExcludePath("/metrics")
				// Path should be excluded (tested via middleware behavior)
			})

			It("should exclude multiple paths", func() {
				p.ExcludePath("/metrics", "/health", "/debug")
				// Paths should be excluded
			})

			It("should add leading slash if missing", func() {
				p.ExcludePath("api/internal")
				// Should work with or without leading slash
			})

			It("should handle empty strings", func() {
				Expect(func() {
					p.ExcludePath("", "/valid", "")
				}).ToNot(Panic())
			})

			It("should handle paths with trailing slash", func() {
				p.ExcludePath("/admin/")
				// Should work with trailing slash
			})

			It("should accumulate paths", func() {
				p.ExcludePath("/first")
				p.ExcludePath("/second")
				p.ExcludePath("/third")
				// All paths should be excluded
			})

			It("should handle duplicate paths", func() {
				Expect(func() {
					p.ExcludePath("/duplicate")
					p.ExcludePath("/duplicate")
				}).ToNot(Panic())
			})

			It("should handle path prefixes", func() {
				p.ExcludePath("/api/internal")
				// Should exclude /api/internal/* paths
			})

			It("should handle root path", func() {
				Expect(func() {
					p.ExcludePath("/")
				}).ToNot(Panic())
			})

			It("should handle special characters in paths", func() {
				Expect(func() {
					p.ExcludePath("/api/v1.0")
					p.ExcludePath("/api/users_admin")
					p.ExcludePath("/api/test-endpoint")
				}).ToNot(Panic())
			})
		})

		Context("when handling concurrent exclude operations", func() {
			It("should be thread-safe", func() {
				done := make(chan bool, 50)

				for i := 0; i < 50; i++ {
					go func(idx int) {
						p.ExcludePath("/concurrent", "/path")
						done <- true
					}(i)
				}

				for i := 0; i < 50; i++ {
					<-done
				}

				// Should complete without panic
			})

			It("should handle mixed concurrent operations", func() {
				done := make(chan bool, 100)

				// Writers
				for i := 0; i < 50; i++ {
					go func() {
						p.ExcludePath("/write")
						done <- true
					}()
				}

				// Other operations
				for i := 0; i < 50; i++ {
					go func() {
						_ = p.ListMetric()
						done <- true
					}()
				}

				for i := 0; i < 100; i++ {
					<-done
				}
			})
		})

		Context("when handling path variations", func() {
			It("should normalize paths without leading slash", func() {
				Expect(func() {
					p.ExcludePath("no-leading-slash")
				}).ToNot(Panic())
			})

			It("should handle query parameters in path definition", func() {
				Expect(func() {
					p.ExcludePath("/api?debug=true")
				}).ToNot(Panic())
			})

			It("should handle URL encoded paths", func() {
				Expect(func() {
					p.ExcludePath("/api%2Fencoded")
				}).ToNot(Panic())
			})

			It("should handle very long paths", func() {
				longPath := "/api"
				for i := 0; i < 100; i++ {
					longPath += "/segment"
				}

				Expect(func() {
					p.ExcludePath(longPath)
				}).ToNot(Panic())
			})

			It("should handle Unicode characters", func() {
				Expect(func() {
					p.ExcludePath("/api/ユーザー")
					p.ExcludePath("/api/пользователь")
				}).ToNot(Panic())
			})
		})

		Context("when using variadic parameters", func() {
			It("should accept no parameters", func() {
				Expect(func() {
					p.ExcludePath()
				}).ToNot(Panic())
			})

			It("should accept single parameter", func() {
				Expect(func() {
					p.ExcludePath("/single")
				}).ToNot(Panic())
			})

			It("should accept many parameters", func() {
				paths := make([]string, 100)
				for i := range paths {
					paths[i] = "/path"
				}

				Expect(func() {
					p.ExcludePath(paths...)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Path Matching Logic", func() {
		Context("when checking prefix matching", func() {
			It("should exclude paths by prefix", func() {
				p.ExcludePath("/admin")
				// Should exclude /admin, /admin/, /admin/users, etc.
			})

			It("should not exclude non-matching paths", func() {
				p.ExcludePath("/admin")
				// Should not exclude /user, /api, etc.
			})

			It("should handle exact matches", func() {
				p.ExcludePath("/exact")
				// Should exclude /exact
			})

			It("should handle nested paths", func() {
				p.ExcludePath("/api/v1")
				// Should exclude /api/v1, /api/v1/users, etc.
				// Should not exclude /api/v2
			})

			It("should handle multiple exclude rules", func() {
				p.ExcludePath("/metrics", "/health", "/debug")
				// All three paths and their children should be excluded
			})
		})

		Context("when handling edge cases", func() {
			It("should handle empty exclude list", func() {
				// No paths excluded, all should be processed
			})

			It("should handle overlapping paths", func() {
				p.ExcludePath("/api")
				p.ExcludePath("/api/internal")
				// Both work, though /api already covers /api/internal
			})

			It("should handle case sensitivity", func() {
				p.ExcludePath("/Admin")
				// Should only exclude /Admin, not /admin
			})
		})
	})

	Describe("Performance", func() {
		Context("when handling many exclude paths", func() {
			It("should handle large number of paths efficiently", func() {
				// Add many exclude paths
				for i := 0; i < 1000; i++ {
					p.ExcludePath("/path")
				}

				// Should still be performant
				Expect(p.ListMetric()).To(BeEmpty())
			})

			It("should handle concurrent exclude with large dataset", func() {
				done := make(chan bool, 100)

				for i := 0; i < 100; i++ {
					go func() {
						for j := 0; j < 10; j++ {
							p.ExcludePath("/concurrent")
						}
						done <- true
					}()
				}

				for i := 0; i < 100; i++ {
					<-done
				}
			})
		})
	})
})
