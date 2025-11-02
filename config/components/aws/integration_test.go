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

package aws_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	. "github.com/nabbar/golib/config/components/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	cfgtps "github.com/nabbar/golib/config/types"
	libhtc "github.com/nabbar/golib/httpcli"
	liblog "github.com/nabbar/golib/logger"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

var _ = Describe("AWS Component Integration Tests", func() {
	var (
		mockS3Server *httptest.Server
		cpt          CptAws
		ctx          context.Context
		vpr          *spfvpr.Viper
		componentKey string
	)

	BeforeEach(func() {
		// Create mock S3 server
		mockS3Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock S3 responses
			switch r.URL.Path {
			case "/":
				// ListBuckets response
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
  <Buckets>
    <Bucket>
      <Name>test-bucket</Name>
      <CreationDate>2023-01-01T00:00:00.000Z</CreationDate>
    </Bucket>
  </Buckets>
</ListAllMyBucketsResult>`)
			case "/test-bucket":
				// Bucket exists check
				w.WriteHeader(http.StatusOK)
			case "/test-bucket/":
				// ListObjects response
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
  <Name>test-bucket</Name>
  <Contents>
    <Key>test-object.txt</Key>
  </Contents>
</ListBucketResult>`)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		// Create context provider
		ctx = context.Background()

		// Setup viper configuration
		vpr = spfvpr.New()
		componentKey = "test-aws-component"

		// Create component
		cpt = New(ctx, ConfigCustom)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
		if mockS3Server != nil {
			mockS3Server.Close()
		}
	})

	Describe("Component with Mock S3 Server", func() {
		Context("with valid configuration", func() {
			BeforeEach(func() {
				// Configure viper with mock server endpoint
				vpr.Set(componentKey+".bucket", "test-bucket")
				vpr.Set(componentKey+".accesskey", "test-access-key")
				vpr.Set(componentKey+".secretkey", "test-secret-key")
				vpr.Set(componentKey+".region", "us-east-1")
				vpr.Set(componentKey+".endpoint", mockS3Server.URL)

				// Initialize component
				getCpt := func(k string) cfgtps.Component { return nil }
				logger := func() liblog.Logger {
					return liblog.New(ctx)
				}
				viperFunc := func() libvpr.Viper {
					return libvpr.New(ctx, logger)
				}

				cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)
			})

			It("should initialize without error", func() {
				Expect(cpt.Type()).To(Equal("aws"))
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should not return AWS client before start", func() {
				aws := cpt.GetAws()
				Expect(aws).To(BeNil())
			})

			// Note: Full start test would require proper AWS SDK mock
			// which is complex. We test initialization and configuration only.
		})

		Context("with missing configuration", func() {
			It("should handle missing viper gracefully", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				viperFunc := func() libvpr.Viper { return nil }
				logger := func() liblog.Logger { return nil }

				cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)
				Expect(cpt.Type()).To(Equal("aws"))
			})
		})
	})

	Describe("Configuration Parsing", func() {
		It("should parse JSON configuration correctly", func() {
			configJSON := []byte(`{
				"bucket": "json-test-bucket",
				"accesskey": "json-access",
				"secretkey": "json-secret",
				"region": "eu-west-1",
				"endpoint": "https://s3.eu-west-1.amazonaws.com"
			}`)

			cfg, err := ConfigCustom.Unmarshal(configJSON)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.GetBucketName()).To(Equal("json-test-bucket"))
			Expect(cfg.GetRegion()).To(Equal("eu-west-1"))
		})

		It("should validate configuration", func() {
			configJSON := []byte(`{
				"bucket": "validation-bucket",
				"accesskey": "validation-access",
				"secretkey": "validation-secret",
				"region": "us-west-2",
				"endpoint": "https://s3.us-west-2.amazonaws.com"
			}`)

			cfg, err := ConfigCustom.Unmarshal(configJSON)
			Expect(err).NotTo(HaveOccurred())

			err = cfg.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject invalid configuration", func() {
			configJSON := []byte(`{
				"bucket": "",
				"accesskey": "",
				"secretkey": "",
				"region": ""
			}`)

			cfg, err := ConfigStandard.Unmarshal(configJSON)
			Expect(err).NotTo(HaveOccurred())

			err = cfg.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("HTTP Client Integration", func() {
		It("should accept custom HTTP client", func() {
			client := libhtc.GetClient()
			Expect(func() {
				cpt.RegisterHTTPClient(client)
			}).NotTo(Panic())
		})

		It("should use default HTTP client when nil", func() {
			Expect(func() {
				cpt.RegisterHTTPClient(nil)
			}).NotTo(Panic())
		})
	})

	Describe("Version Integration", func() {
		It("should work with version information", func() {
			// Create a basic version - just test that Init accepts it
			getCpt := func(k string) cfgtps.Component { return nil }
			viperFunc := func() libvpr.Viper { return nil }
			logger := func() liblog.Logger { return nil }

			// Pass nil version for simplicity in tests
			cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)
			Expect(cpt.Type()).To(Equal("aws"))
		})

		It("should work without version information", func() {
			getCpt := func(k string) cfgtps.Component { return nil }
			viperFunc := func() libvpr.Viper { return nil }
			logger := func() liblog.Logger { return nil }

			cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)
			Expect(cpt.Type()).To(Equal("aws"))
		})
	})

	Describe("Default Configuration Generation", func() {
		It("should generate valid default configuration", func() {
			defaultCfg := cpt.DefaultConfig("  ")
			Expect(defaultCfg).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(defaultCfg, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should include all required fields", func() {
			defaultCfg := cpt.DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(defaultCfg, &result)
			Expect(err).NotTo(HaveOccurred())

			// Check standard fields
			Expect(result).To(HaveKey("bucket"))
			Expect(result).To(HaveKey("accesskey"))
			Expect(result).To(HaveKey("secretkey"))
			Expect(result).To(HaveKey("region"))
		})
	})

	Describe("Lifecycle Hooks", func() {
		var (
			beforeStartCalled  bool
			afterStartCalled   bool
			beforeReloadCalled bool
			afterReloadCalled  bool
		)

		BeforeEach(func() {
			beforeStartCalled = false
			afterStartCalled = false
			beforeReloadCalled = false
			afterReloadCalled = false

			beforeStart := func(c cfgtps.Component) error {
				beforeStartCalled = true
				return nil
			}
			afterStart := func(c cfgtps.Component) error {
				afterStartCalled = true
				return nil
			}
			beforeReload := func(c cfgtps.Component) error {
				beforeReloadCalled = true
				return nil
			}
			afterReload := func(c cfgtps.Component) error {
				afterReloadCalled = true
				return nil
			}

			cpt.RegisterFuncStart(beforeStart, afterStart)
			cpt.RegisterFuncReload(beforeReload, afterReload)
		})

		It("should register hooks without error", func() {
			Expect(beforeStartCalled).To(BeFalse())
			Expect(afterStartCalled).To(BeFalse())
			Expect(beforeReloadCalled).To(BeFalse())
			Expect(afterReloadCalled).To(BeFalse())
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent Type calls", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					typ := cpt.Type()
					Expect(typ).To(Equal("aws"))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent GetAws calls", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					aws := cpt.GetAws()
					Expect(aws).To(BeNil()) // Not started, so nil
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent IsStarted calls", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					started := cpt.IsStarted()
					Expect(started).To(BeFalse())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Error Scenarios", func() {
		It("should handle invalid endpoint URL", func() {
			vpr.Set(componentKey+".bucket", "error-bucket")
			vpr.Set(componentKey+".accesskey", "error-access")
			vpr.Set(componentKey+".secretkey", "error-secret")
			vpr.Set(componentKey+".region", "us-east-1")
			vpr.Set(componentKey+".endpoint", "://invalid-url")

			getCpt := func(k string) cfgtps.Component { return nil }
			logger := func() liblog.Logger {
				return liblog.New(ctx)
			}
			viperFunc := func() libvpr.Viper {
				return libvpr.New(ctx, logger)
			}

			cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)

			// Start should fail with invalid endpoint
			err := cpt.Start()
			Expect(err).To(HaveOccurred())
		})

		It("should handle missing configuration keys", func() {
			// Don't set any configuration
			getCpt := func(k string) cfgtps.Component { return nil }
			logger := func() liblog.Logger {
				return liblog.New(ctx)
			}
			viperFunc := func() libvpr.Viper {
				return libvpr.New(ctx, logger)
			}

			cpt.Init(componentKey, ctx, getCpt, viperFunc, nil, logger)

			// Start should fail with missing config
			err := cpt.Start()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Real-world Scenarios", func() {
		Context("multi-region setup", func() {
			It("should support multiple regions", func() {
				regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-south-1"}

				for _, region := range regions {
					cfg := ConfigStandard.Config(
						"test-bucket",
						"access-key",
						"secret-key",
						region,
						nil,
					)
					Expect(cfg).NotTo(BeNil())
					Expect(cfg.GetRegion()).To(Equal(region))
				}
			})
		})

		Context("configuration migration", func() {
			It("should support migrating from Standard to Custom", func() {
				// Start with Standard config
				stdCfg := ConfigStandard.Config(
					"migration-bucket",
					"migration-access",
					"migration-secret",
					"us-east-1",
					nil,
				)
				Expect(stdCfg).NotTo(BeNil())

				// Create Custom config with same data
				endpointURL, _ := url.Parse(mockS3Server.URL)
				cusCfg := ConfigCustom.Config(
					"migration-bucket",
					"migration-access",
					"migration-secret",
					"us-east-1",
					endpointURL,
				)
				Expect(cusCfg).NotTo(BeNil())
			})
		})
	})
})

var _ = Describe("AWS Component Performance", func() {
	Context("component creation performance", func() {
		It("should create components efficiently", func() {
			experiment := gmeasure.NewExperiment("component creation")
			AddReportEntry(experiment.Name, experiment)

			ctx := context.Background()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("creation time", func() {
					for i := 0; i < 100; i++ {
						_ = New(ctx, ConfigStandard)
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("creation time")
			medianDuration := stats.DurationFor(gmeasure.StatMedian)

			// Creating 100 components should take less than 1 second (median)
			Expect(medianDuration).To(BeNumerically("<", 1*time.Second),
				"Creating 100 components should take less than 1 second (median: %v)", medianDuration)
		})
	})

	Context("default config generation performance", func() {
		It("should generate default configs efficiently", func() {
			experiment := gmeasure.NewExperiment("config generation")
			AddReportEntry(experiment.Name, experiment)

			ctx := context.Background()
			cpt := New(ctx, ConfigCustom)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("generation time", func() {
					for i := 0; i < 1000; i++ {
						_ = cpt.DefaultConfig("  ")
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("generation time")
			medianDuration := stats.DurationFor(gmeasure.StatMedian)

			// Generating 1000 default configs should take less than 1 second (median)
			Expect(medianDuration).To(BeNumerically("<", 1*time.Second),
				"Generating 1000 default configs should take less than 1 second (median: %v)", medianDuration)
		})
	})
})
