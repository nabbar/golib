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

package config_test

import (
	"testing"

	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("SMTP Config Benchmarks", func() {

	Describe("Config Creation Performance", func() {
		It("should benchmark simple DSN parsing", func() {
			experiment := gmeasure.NewExperiment("Simple DSN Parsing")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse", func() {
					model := newConfigModel("tcp(localhost:25)/")
					cfg, err := model.Config()
					Expect(err).ToNot(HaveOccurred())
					Expect(cfg).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			// Just record the stats, don't assert on absolute values which depend on hardware
			AddReportEntry("Average parse time", experiment.GetStats("parse").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark complex DSN parsing", func() {
			experiment := gmeasure.NewExperiment("Complex DSN Parsing")
			AddReportEntry(experiment.Name, experiment)

			dsn := "user:pass@tcp(smtp.example.com:587)/starttls?ServerName=mail.example.com&SkipVerify=true"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse", func() {
					model := newConfigModel(dsn)
					cfg, err := model.Config()
					Expect(err).ToNot(HaveOccurred())
					Expect(cfg).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Average parse time", experiment.GetStats("parse").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark validation", func() {
			experiment := gmeasure.NewExperiment("Config Validation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				model := newConfigModel("tcp(localhost:25)/")
				experiment.MeasureDuration("validate", func() {
					err := model.Validate()
					Expect(err).To(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Average validation time", experiment.GetStats("validate").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("DSN Generation Performance", func() {
		It("should benchmark GetDsn", func() {
			experiment := gmeasure.NewExperiment("DSN Generation")
			AddReportEntry(experiment.Name, experiment)

			cfg, err := createBasicConfig()
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("generate", func() {
					dsn := cfg.GetDsn()
					Expect(dsn).ToNot(BeEmpty())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Average generation time", experiment.GetStats("generate").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark DSN roundtrip", func() {
			experiment := gmeasure.NewExperiment("DSN Roundtrip")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("roundtrip", func() {
					model1 := newConfigModel("user:pass@tcp(smtp.example.com:587)/starttls")
					cfg1, err := model1.Config()
					Expect(err).ToNot(HaveOccurred())

					dsn := cfg1.GetDsn()
					model2 := newConfigModel(dsn)
					cfg2, err := model2.Config()
					Expect(err).ToNot(HaveOccurred())
					Expect(cfg2).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 500})

			AddReportEntry("Average roundtrip time", experiment.GetStats("roundtrip").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Getter/Setter Performance", func() {
		var cfg smtpcfg.Config

		BeforeEach(func() {
			var err error
			cfg, err = createBasicConfig()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should benchmark getter operations", func() {
			experiment := gmeasure.NewExperiment("Config Getters")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("getters", func() {
					_ = cfg.GetHost()
					_ = cfg.GetPort()
					_ = cfg.GetUser()
					_ = cfg.GetPass()
					_ = cfg.GetNet()
					_ = cfg.GetTlsMode()
					_ = cfg.IsTLSSkipVerify()
					_ = cfg.GetTlSServerName()
				})
			}, gmeasure.SamplingConfig{N: 10000})

			AddReportEntry("Average getters time", experiment.GetStats("getters").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark setter operations", func() {
			experiment := gmeasure.NewExperiment("Config Setters")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("setters", func() {
					cfg.SetHost("newhost.example.com")
					cfg.SetPort(587)
					cfg.SetUser("newuser")
					cfg.SetPass("newpass")
					cfg.SetTlsMode(smtptp.TLSStartTLS)
					cfg.ForceTLSSkipVerify(true)
					cfg.SetTLSServerName("tls.example.com")
				})
			}, gmeasure.SamplingConfig{N: 10000})

			AddReportEntry("Average setters time", experiment.GetStats("setters").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Concurrent Operations Performance", func() {
		It("should benchmark concurrent config creation", func() {
			experiment := gmeasure.NewExperiment("Concurrent Config Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent", func() {
					done := make(chan bool, 10)
					for i := 0; i < 10; i++ {
						go func() {
							model := newConfigModel("tcp(localhost:25)/")
							cfg, err := model.Config()
							Expect(err).ToNot(HaveOccurred())
							Expect(cfg).ToNot(BeNil())
							done <- true
						}()
					}
					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, gmeasure.SamplingConfig{N: 100})

			AddReportEntry("Average concurrent time", experiment.GetStats("concurrent").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark concurrent reads", func() {
			experiment := gmeasure.NewExperiment("Concurrent Reads")
			AddReportEntry(experiment.Name, experiment)

			cfg, err := createBasicConfig()
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("reads", func() {
					done := make(chan bool, 20)
					for i := 0; i < 20; i++ {
						go func() {
							_ = cfg.GetHost()
							_ = cfg.GetPort()
							_ = cfg.GetDsn()
							done <- true
						}()
					}
					for i := 0; i < 20; i++ {
						<-done
					}
				})
			}, gmeasure.SamplingConfig{N: 100})

			AddReportEntry("Average concurrent reads time", experiment.GetStats("reads").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Memory Allocation Performance", func() {
		It("should benchmark allocation for simple config", func() {
			experiment := gmeasure.NewExperiment("Simple Config Allocation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.RecordValue("allocations", float64(testing.AllocsPerRun(100, func() {
					model := newConfigModel("tcp(localhost:25)/")
					cfg, _ := model.Config()
					_ = cfg
				})))
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("allocations")
			AddReportEntry("Allocations per operation", stats.FloatFor(gmeasure.StatMean))
		})

		It("should benchmark allocation for complex config", func() {
			experiment := gmeasure.NewExperiment("Complex Config Allocation")
			AddReportEntry(experiment.Name, experiment)

			dsn := "user:pass@tcp(smtp.example.com:587)/starttls?ServerName=mail.example.com&SkipVerify=true"

			experiment.Sample(func(idx int) {
				experiment.RecordValue("allocations", float64(testing.AllocsPerRun(100, func() {
					model := newConfigModel(dsn)
					cfg, _ := model.Config()
					_ = cfg.GetDsn()
				})))
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("allocations")
			AddReportEntry("Allocations per operation", stats.FloatFor(gmeasure.StatMean))
		})

		It("should benchmark DSN generation allocations", func() {
			experiment := gmeasure.NewExperiment("DSN Generation Allocation")
			AddReportEntry(experiment.Name, experiment)

			cfg, err := createBasicConfig()
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.RecordValue("allocations", float64(testing.AllocsPerRun(100, func() {
					_ = cfg.GetDsn()
				})))
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("allocations")
			AddReportEntry("Allocations per GetDsn", stats.FloatFor(gmeasure.StatMean))
		})
	})

	Describe("Comparison Benchmarks", func() {
		It("should compare simple vs complex DSN parsing", func() {
			experiment := gmeasure.NewExperiment("Simple vs Complex DSN")
			AddReportEntry(experiment.Name, experiment)

			simpleDSN := "tcp(localhost:25)/"
			complexDSN := "user:pass@tcp(smtp.example.com:587)/starttls?ServerName=mail.example.com&SkipVerify=true"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("simple", func() {
					model := newConfigModel(simpleDSN)
					cfg, _ := model.Config()
					_ = cfg
				})

				experiment.MeasureDuration("complex", func() {
					model := newConfigModel(complexDSN)
					cfg, _ := model.Config()
					_ = cfg
				})
			}, gmeasure.SamplingConfig{N: 500})

			simpleStats := experiment.GetStats("simple")
			complexStats := experiment.GetStats("complex")

			AddReportEntry("Simple DSN mean", simpleStats.DurationFor(gmeasure.StatMean))
			AddReportEntry("Complex DSN mean", complexStats.DurationFor(gmeasure.StatMean))
		})

		It("should compare different TLS modes", func() {
			experiment := gmeasure.NewExperiment("TLS Mode Comparison")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("none", func() {
					model := newConfigModel("tcp(localhost:25)/")
					cfg, _ := model.Config()
					_ = cfg
				})

				experiment.MeasureDuration("starttls", func() {
					model := newConfigModel("tcp(localhost:587)/starttls")
					cfg, _ := model.Config()
					_ = cfg
				})

				experiment.MeasureDuration("tls", func() {
					model := newConfigModel("tcp(localhost:465)/tls")
					cfg, _ := model.Config()
					_ = cfg
				})
			}, gmeasure.SamplingConfig{N: 500})

			AddReportEntry("TLS None mean", experiment.GetStats("none").DurationFor(gmeasure.StatMean))
			AddReportEntry("TLS StartTLS mean", experiment.GetStats("starttls").DurationFor(gmeasure.StatMean))
			AddReportEntry("TLS Strict mean", experiment.GetStats("tls").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Stress Tests", func() {
		It("should handle rapid config creations", func() {
			experiment := gmeasure.NewExperiment("Rapid Config Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("burst", func() {
					for i := 0; i < 100; i++ {
						model := newConfigModel("tcp(localhost:25)/")
						cfg, err := model.Config()
						Expect(err).ToNot(HaveOccurred())
						Expect(cfg).ToNot(BeNil())
					}
				})
			}, gmeasure.SamplingConfig{N: 50})

			AddReportEntry("Average burst time for 100 configs", experiment.GetStats("burst").DurationFor(gmeasure.StatMean))
		})

		It("should handle rapid DSN regenerations", func() {
			experiment := gmeasure.NewExperiment("Rapid DSN Regeneration")
			AddReportEntry(experiment.Name, experiment)

			cfg, err := createBasicConfig()
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("burst", func() {
					for i := 0; i < 100; i++ {
						dsn := cfg.GetDsn()
						Expect(dsn).ToNot(BeEmpty())
					}
				})
			}, gmeasure.SamplingConfig{N: 50})

			AddReportEntry("Average burst time for 100 DSNs", experiment.GetStats("burst").DurationFor(gmeasure.StatMean))
		})
	})
})
