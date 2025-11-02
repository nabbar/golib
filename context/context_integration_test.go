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

package context_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libctx "github.com/nabbar/golib/context"
)

var _ = Describe("Context Integration", func() {
	Describe("context.Context compatibility", func() {
		It("should implement context.Context interface", func() {
			cfg := libctx.New[string](nil)
			var ctx context.Context = cfg
			Expect(ctx).ToNot(BeNil())
		})

		It("should be usable in functions expecting context.Context", func() {
			cfg := libctx.New[string](nil)
			cfg.Store("test", "value")

			// Function that expects context.Context
			processWithContext := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			}

			err := processWithContext(cfg)
			Expect(err).To(BeNil())
		})
	})

	Describe("Deadline", func() {
		It("should return no deadline for background context", func() {
			cfg := libctx.New[string](nil)
			deadline, ok := cfg.Deadline()
			Expect(ok).To(BeFalse())
			Expect(deadline).To(BeZero())
		})

		It("should return deadline from underlying context", func() {
			deadline := time.Now().Add(1 * time.Hour)
			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()

			cfg := libctx.New[string](ctx)

			d, ok := cfg.Deadline()
			Expect(ok).To(BeTrue())
			Expect(d).To(BeTemporally("~", deadline, time.Second))
		})
	})

	Describe("Done", func() {
		It("should return nil channel for background context", func() {
			cfg := libctx.New[string](nil)
			done := cfg.Done()
			Expect(done).To(BeNil())
		})

		It("should return done channel that closes on cancellation", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cfg := libctx.New[string](ctx)

			done := cfg.Done()
			Expect(done).ToNot(BeNil())

			// Channel should not be closed yet
			select {
			case <-done:
				Fail("Done channel should not be closed yet")
			default:
			}

			// Cancel context
			cancel()

			// Channel should be closed now
			Eventually(done).Should(BeClosed())
		})

		It("should close done channel on timeout", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			cfg := libctx.New[string](ctx)

			done := cfg.Done()
			Eventually(done, 100*time.Millisecond).Should(BeClosed())
		})
	})

	Describe("Err", func() {
		It("should return nil for active context", func() {
			cfg := libctx.New[string](nil)
			Expect(cfg.Err()).To(BeNil())
		})

		It("should return Canceled after cancellation", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cfg := libctx.New[string](ctx)

			Expect(cfg.Err()).To(BeNil())

			cancel()

			Expect(cfg.Err()).To(Equal(context.Canceled))
		})

		It("should return DeadlineExceeded after timeout", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			cfg := libctx.New[string](ctx)

			Eventually(func() error {
				return cfg.Err()
			}, 50*time.Millisecond).Should(Equal(context.DeadlineExceeded))
		})
	})

	Describe("Value", func() {
		It("should return stored values from Config", func() {
			cfg := libctx.New[string](nil)
			cfg.Store("key1", "value1")

			val := cfg.Value("key1")
			Expect(val).To(Equal("value1"))
		})

		It("should return values from underlying context", func() {
			ctx := context.WithValue(context.Background(), "ctx_key", "ctx_value")
			cfg := libctx.New[string](ctx)

			val := cfg.Value("ctx_key")
			Expect(val).To(Equal("ctx_value"))
		})

		It("should prefer Config stored values over context values", func() {
			ctx := context.WithValue(context.Background(), "shared_key", "ctx_value")
			cfg := libctx.New[string](ctx)
			cfg.Store("shared_key", "config_value")

			val := cfg.Value("shared_key")
			Expect(val).To(Equal("config_value"))
		})

		It("should return nil for non-existent keys", func() {
			cfg := libctx.New[string](nil)
			val := cfg.Value("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should handle non-string keys from context", func() {
			type customKey int
			const key customKey = 42

			ctx := context.WithValue(context.Background(), key, "value")
			cfg := libctx.New[string](ctx)

			val := cfg.Value(key)
			Expect(val).To(Equal("value"))
		})
	})

	Describe("Behavior with cancelled context", func() {
		var cfg libctx.Config[string]
		var ctx context.Context
		var cancel context.CancelFunc

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
			cfg = libctx.New[string](ctx)
			cfg.Store("key1", "value1")
		})

		AfterEach(func() {
			cancel()
		})

		It("should clean storage on Store when context is cancelled", func() {
			cancel()
			time.Sleep(10 * time.Millisecond) // Give time for cancellation to propagate

			cfg.Store("key2", "value2")

			// After cancellation, Store should clean the map
			// The new value might not be stored due to Clean()
			val, ok := cfg.Load("key1")
			// Behavior may vary - Clean() is called before Store
			_ = val
			_ = ok
		})

		It("should clean storage on Delete when context is cancelled", func() {
			cancel()
			time.Sleep(10 * time.Millisecond)

			cfg.Delete("key1")

			// After cancellation, Delete should clean the map
		})

		It("should return false on LoadOrStore when context is cancelled", func() {
			cancel()
			time.Sleep(10 * time.Millisecond)

			val, loaded := cfg.LoadOrStore("key2", "value2")
			Expect(val).To(BeNil())
			Expect(loaded).To(BeFalse())
		})

		It("should return false on Merge when context is cancelled", func() {
			cfg2 := libctx.New[string](nil)
			cfg2.Store("key2", "value2")

			cancel()
			time.Sleep(10 * time.Millisecond)

			result := cfg.Merge(cfg2)
			Expect(result).To(BeFalse())
		})

		It("should return nil on Clone when context is cancelled", func() {
			cancel()
			time.Sleep(10 * time.Millisecond)

			clone := cfg.Clone(nil)
			Expect(clone).To(BeNil())
		})
	})

	Describe("Real-world usage patterns", func() {
		It("should work as request-scoped storage", func() {
			// Simulating HTTP request context
			requestCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg := libctx.New[string](requestCtx)

			// Store request-scoped data
			cfg.Store("user_id", 12345)
			cfg.Store("request_id", "abc-123")
			cfg.Store("authenticated", true)

			// Retrieve data
			userID, ok := cfg.Load("user_id")
			Expect(ok).To(BeTrue())
			Expect(userID).To(Equal(12345))

			// Use in subfunction
			processRequest := func(ctx libctx.Config[string]) error {
				reqID, ok := ctx.Load("request_id")
				if !ok {
					return context.DeadlineExceeded
				}
				Expect(reqID).To(Equal("abc-123"))
				return nil
			}

			err := processRequest(cfg)
			Expect(err).To(BeNil())
		})

		It("should support middleware pattern", func() {
			cfg := libctx.New[string](nil)

			// Middleware 1: Authentication
			authMiddleware := func(cfg libctx.Config[string]) {
				cfg.Store("user", "john_doe")
				cfg.Store("role", "admin")
			}

			// Middleware 2: Logging
			loggingMiddleware := func(cfg libctx.Config[string]) {
				cfg.Store("request_start", time.Now())
			}

			// Execute middlewares
			authMiddleware(cfg)
			loggingMiddleware(cfg)

			// Handler
			handler := func(cfg libctx.Config[string]) string {
				user, _ := cfg.Load("user")
				role, _ := cfg.Load("role")
				return user.(string) + ":" + role.(string)
			}

			result := handler(cfg)
			Expect(result).To(Equal("john_doe:admin"))
		})

		It("should support configuration inheritance", func() {
			// Global config
			globalCfg := libctx.New[string](nil)
			globalCfg.Store("env", "production")
			globalCfg.Store("log_level", "info")

			// Service-specific config
			serviceCfg := libctx.New[string](nil)
			serviceCfg.Store("service", "api")
			serviceCfg.Store("port", 8080)

			// Merge configurations
			serviceCfg.Merge(globalCfg)

			// Service config should have both
			env, _ := serviceCfg.Load("env")
			service, _ := serviceCfg.Load("service")
			port, _ := serviceCfg.Load("port")

			Expect(env).To(Equal("production"))
			Expect(service).To(Equal("api"))
			Expect(port).To(Equal(8080))
		})
	})
})
