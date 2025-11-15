/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package gin_test

import (
	"context"
	"syscall"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libgin "github.com/nabbar/golib/context/gin"
)

var _ = Describe("GinTonic Context Integration", func() {
	var (
		ginCtx *ginsdk.Context
		gtx    libgin.GinTonic
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		ginCtx, _ = ginsdk.CreateTestContext(nil)
		gtx = libgin.New(ginCtx, nil)
	})

	Describe("context.Context compatibility", func() {
		It("should implement context.Context interface", func() {
			var ctx context.Context = gtx
			Expect(ctx).ToNot(BeNil())
		})

		It("should be usable in functions expecting context.Context", func() {
			processWithContext := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			}

			err := processWithContext(gtx)
			Expect(err).To(BeNil())
		})
	})

	Describe("Deadline", func() {
		It("should return no deadline for background context", func() {
			deadline, ok := gtx.Deadline()
			Expect(ok).To(BeFalse())
			Expect(deadline).To(BeZero())
		})
	})

	Describe("Done", func() {
		It("should return a non-nil done channel", func() {
			done := gtx.Done()
			Expect(done).ToNot(BeNil())
		})

		It("should not be closed for active context", func() {
			done := gtx.Done()
			select {
			case <-done:
				Fail("Done channel should not be closed for active context")
			default:
				// Expected: context is still active
			}
		})
	})

	Describe("Err", func() {
		It("should return nil for active context", func() {
			err := gtx.Err()
			Expect(err).To(BeNil())
		})
	})

	Describe("Value", func() {
		It("should return stored values", func() {
			gtx.Set("user_id", 12345)
			val := gtx.Value("user_id")
			Expect(val).To(Equal(12345))
		})

		It("should return nil for non-existent keys", func() {
			val := gtx.Value("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should support string keys", func() {
			gtx.Set("request_id", "abc-123")
			val := gtx.Value("request_id")
			Expect(val).To(Equal("abc-123"))
		})

		It("should support various key types", func() {
			// Gin context converts keys internally, so we test with string keys
			gtx.Set("key1", "value1")
			gtx.Set("key2", 123)

			val1 := gtx.Value("key1")
			val2 := gtx.Value("key2")

			Expect(val1).To(Equal("value1"))
			Expect(val2).To(Equal(123))
		})
	})

	Describe("CancelOnSignal", func() {
		It("should register signal handler without panicking", func() {
			Expect(func() {
				gtx.CancelOnSignal(syscall.SIGTERM)
			}).ToNot(Panic())
		})

		It("should register multiple signals", func() {
			Expect(func() {
				gtx.CancelOnSignal(syscall.SIGTERM, syscall.SIGINT)
			}).ToNot(Panic())
		})

		It("should not block when registering signals", func() {
			done := make(chan bool)
			go func() {
				gtx.CancelOnSignal(syscall.SIGTERM)
				done <- true
			}()

			select {
			case <-done:
				// Expected: signal registration completed
			case <-time.After(100 * time.Millisecond):
				Fail("CancelOnSignal should not block")
			}
		})
	})

	Describe("Real-world usage patterns", func() {
		It("should work as request-scoped storage", func() {
			// Simulating HTTP request context
			gtx.Set("user_id", 12345)
			gtx.Set("request_id", "abc-123")
			gtx.Set("authenticated", true)

			// Retrieve data
			userID := gtx.Value("user_id")
			Expect(userID).To(Equal(12345))

			// Use in subfunction
			processRequest := func(ctx libgin.GinTonic) error {
				reqID := ctx.GetString("request_id")
				if reqID == "" {
					return context.DeadlineExceeded
				}
				Expect(reqID).To(Equal("abc-123"))
				return nil
			}

			err := processRequest(gtx)
			Expect(err).To(BeNil())
		})

		It("should support middleware pattern", func() {
			// Middleware 1: Authentication
			authMiddleware := func(ctx libgin.GinTonic) {
				ctx.Set("user", "john_doe")
				ctx.Set("role", "admin")
			}

			// Middleware 2: Logging
			loggingMiddleware := func(ctx libgin.GinTonic) {
				ctx.Set("request_start", time.Now())
			}

			// Execute middlewares
			authMiddleware(gtx)
			loggingMiddleware(gtx)

			// Handler
			handler := func(ctx libgin.GinTonic) string {
				user := ctx.GetString("user")
				role := ctx.GetString("role")
				return user + ":" + role
			}

			result := handler(gtx)
			Expect(result).To(Equal("john_doe:admin"))
		})

		It("should handle complex data structures", func() {
			type User struct {
				ID    int
				Name  string
				Email string
			}

			user := User{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			}

			gtx.Set("current_user", user)

			retrievedUser := gtx.Value("current_user")
			Expect(retrievedUser).To(Equal(user))

			// Type assertion
			if u, ok := retrievedUser.(User); ok {
				Expect(u.Name).To(Equal("Alice"))
				Expect(u.Email).To(Equal("alice@example.com"))
			} else {
				Fail("Failed to assert user type")
			}
		})

		It("should support nested contexts", func() {
			gtx.Set("level1", "value1")

			// Simulate nested handler
			nestedHandler := func(ctx libgin.GinTonic) {
				ctx.Set("level2", "value2")

				// Both values should be accessible
				val1 := ctx.GetString("level1")
				val2 := ctx.GetString("level2")

				Expect(val1).To(Equal("value1"))
				Expect(val2).To(Equal("value2"))
			}

			nestedHandler(gtx)
		})

		It("should handle request metadata", func() {
			// Simulate HTTP request metadata
			gtx.Set("method", "POST")
			gtx.Set("path", "/api/users")
			gtx.Set("content_type", "application/json")
			gtx.Set("user_agent", "Mozilla/5.0")

			method := gtx.GetString("method")
			path := gtx.GetString("path")
			contentType := gtx.GetString("content_type")

			Expect(method).To(Equal("POST"))
			Expect(path).To(Equal("/api/users"))
			Expect(contentType).To(Equal("application/json"))
		})
	})
})
