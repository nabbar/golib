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

package smtp_test

import (
	"crypto/tls"
	"fmt"
	"time"

	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("SMTP Benchmarks", Label("benchmark"), func() {

	Describe("Client Creation Performance", func() {
		It("should measure client creation time", func() {
			experiment := NewExperiment("Client Creation")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("creation_time", func() {
					client := newTestSMTPClient(cfg)
					client.Close()
				})
			}, SamplingConfig{N: 100, Duration: 10 * time.Second})

			AddReportEntry("Average creation time", experiment.GetStats("creation_time").DurationFor(StatMean))
		})

		It("should measure client creation with different TLS modes", func() {
			experiment := NewExperiment("Client Creation by TLS Mode")
			AddReportEntry(experiment.Name, experiment)

			modes := map[string]smtptp.TLSMode{
				"none":      smtptp.TLSNone,
				"starttls":  smtptp.TLSStartTLS,
				"stricttls": smtptp.TLSStrictTLS,
			}

			for name, mode := range modes {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, mode)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(name, func() {
						client := newTestSMTPClient(cfg)
						client.Close()
					})
				}, SamplingConfig{N: 50})
				AddReportEntry("Average "+name, experiment.GetStats(name).DurationFor(StatMean))
			}
		})
	})

	Describe("Clone Performance", func() {
		It("should measure clone operation time", func() {
			experiment := NewExperiment("Clone Operation")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			original := newTestSMTPClient(cfg)
			defer original.Close()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone_time", func() {
					cloned := original.Clone()
					cloned.Close()
				})
			}, SamplingConfig{N: 100, Duration: 10 * time.Second})

			AddReportEntry("Average clone time", experiment.GetStats("clone_time").DurationFor(StatMean))
		})

		It("should measure memory overhead of clones", func() {
			experiment := NewExperiment("Clone Memory Overhead")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			original := newTestSMTPClient(cfg)
			defer original.Close()

			clones := make([]interface{}, 0, 100)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone_create", func() {
					cloned := original.Clone()
					clones = append(clones, cloned)
				})
			}, SamplingConfig{N: 100})

			AddReportEntry("Clones created", len(clones))

			// Cleanup
			for _, c := range clones {
				if client, ok := c.(interface{ Close() }); ok {
					client.Close()
				}
			}
		})
	})

	Describe("Configuration Operations", func() {
		It("should measure config parsing performance", func() {
			experiment := NewExperiment("DSN Parsing")
			AddReportEntry(experiment.Name, experiment)

			dsns := []string{
				"tcp(localhost:25)/",
				"user:pass@tcp(smtp.example.com:587)/starttls",
				"tcp(mail.example.com:465)/tls?ServerName=smtp.example.com&SkipVerify=true",
			}

			for i, dsn := range dsns {
				name := fmt.Sprintf("dsn_%d", i)
				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(name, func() {
						model := smtpcfg.ConfigModel{DSN: dsn}
						_, err := model.Config()
						Expect(err).ToNot(HaveOccurred())
					})
				}, SamplingConfig{N: 100})
				AddReportEntry("Average "+name, experiment.GetStats(name).DurationFor(StatMean))
			}
		})

		It("should measure config update performance", func() {
			experiment := NewExperiment("Config Update")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("update_time", func() {
					newCfg := newTestConfig(testSMTPHost, testSMTPPort+idx%10, smtptp.TLSNone)
					tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
					client.UpdConfig(newCfg, tlsConfig)
				})
			}, SamplingConfig{N: 100})

			AddReportEntry("Average update time", experiment.GetStats("update_time").DurationFor(StatMean))
		})

		It("should measure DSN generation performance", func() {
			experiment := NewExperiment("DSN Generation")
			AddReportEntry(experiment.Name, experiment)

			model := smtpcfg.ConfigModel{
				DSN: "user:pass@tcp(smtp.example.com:587)/starttls?ServerName=mail.example.com",
			}
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("dsn_gen", func() {
					dsn := cfg.GetDsn()
					Expect(dsn).ToNot(BeEmpty())
				})
			}, SamplingConfig{N: 1000, Duration: 10 * time.Second})

			AddReportEntry("Average DSN generation", experiment.GetStats("dsn_gen").DurationFor(StatMean))
		})
	})

	Describe("Connection Operations", func() {
		It("should measure connection timeout performance", func() {
			experiment := NewExperiment("Connection Timeout")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig("240.0.0.1", 25, smtptp.TLSNone) // Non-routable IP

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("timeout_duration", func() {
					client := newTestSMTPClient(cfg)
					defer client.Close()

					ctx, cancel := contextWithTimeout(100 * time.Millisecond)
					defer cancel()

					_ = client.Check(ctx)
				})
			}, SamplingConfig{N: 20})

			AddReportEntry("Average timeout duration", experiment.GetStats("timeout_duration").DurationFor(StatMean))
		})

		It("should measure concurrent connection attempts", func() {
			experiment := NewExperiment("Concurrent Connections")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent_duration", func() {
					done := make(chan bool, 10)

					for i := 0; i < 10; i++ {
						go func() {
							client := newTestSMTPClient(cfg)
							defer client.Close()

							ctx, cancel := contextWithTimeout(1 * time.Second)
							defer cancel()

							_ = client.Check(ctx)
							done <- true
						}()
					}

					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 10})

			AddReportEntry("Average concurrent duration", experiment.GetStats("concurrent_duration").DurationFor(StatMean))
		})
	})

	Describe("Email Operations", func() {
		It("should measure email construction performance", func() {
			experiment := NewExperiment("Email Construction")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("email_construct", func() {
					from := "sender@example.com"
					to := "recipient@example.com"
					subject := "Test Subject"
					body := "Test Body Content"

					email := newTestEmail(from, to, subject, body)
					Expect(email).ToNot(BeNil())
				})
			}, SamplingConfig{N: 1000})

			AddReportEntry("Average email construct", experiment.GetStats("email_construct").DurationFor(StatMean))
		})

		It("should measure email construction with varying sizes", func() {
			experiment := NewExperiment("Email Construction by Size")
			AddReportEntry(experiment.Name, experiment)

			sizes := map[string]int{
				"small":  100,
				"medium": 10000,
				"large":  100000,
			}

			for name, size := range sizes {
				body := string(make([]byte, size))

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(name, func() {
						email := newTestEmail("sender@example.com", "recipient@example.com", "Test", body)
						Expect(email).ToNot(BeNil())
					})
				}, SamplingConfig{N: 100})
				AddReportEntry("Average "+name, experiment.GetStats(name).DurationFor(StatMean))
			}
		})
	})

	Describe("Concurrent Operations Performance", func() {
		It("should measure concurrent clone performance", func() {
			experiment := NewExperiment("Concurrent Clone")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			original := newTestSMTPClient(cfg)
			defer original.Close()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent_clone", func() {
					done := make(chan bool, 20)

					for i := 0; i < 20; i++ {
						go func() {
							cloned := original.Clone()
							cloned.Close()
							done <- true
						}()
					}

					for i := 0; i < 20; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 20})

			AddReportEntry("Average concurrent clone", experiment.GetStats("concurrent_clone").DurationFor(StatMean))
		})

		It("should measure concurrent close performance", func() {
			experiment := NewExperiment("Concurrent Close")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent_close", func() {
					cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
					client := newTestSMTPClient(cfg)

					done := make(chan bool, 10)

					for i := 0; i < 10; i++ {
						go func() {
							client.Close()
							done <- true
						}()
					}

					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 20})

			AddReportEntry("Average concurrent close", experiment.GetStats("concurrent_close").DurationFor(StatMean))
		})
	})

	Describe("Memory Efficiency", func() {
		It("should measure client memory footprint", func() {
			experiment := NewExperiment("Client Memory Footprint")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("client_memory", func() {
					clients := make([]interface{}, 0, 100)

					for i := 0; i < 100; i++ {
						client := newTestSMTPClient(cfg)
						clients = append(clients, client)
					}

					// Cleanup
					for _, c := range clients {
						if client, ok := c.(interface{ Close() }); ok {
							client.Close()
						}
					}
				})
			}, SamplingConfig{N: 10})

			AddReportEntry("Average client memory test", experiment.GetStats("client_memory").DurationFor(StatMean))
		})

		It("should measure config memory efficiency", func() {
			experiment := NewExperiment("Config Memory Efficiency")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("config_memory", func() {
					configs := make([]smtpcfg.Config, 0, 1000)

					for i := 0; i < 1000; i++ {
						dsn := fmt.Sprintf("tcp(localhost:%d)/", 25+i%100)
						model := smtpcfg.ConfigModel{DSN: dsn}
						cfg, err := model.Config()
						Expect(err).ToNot(HaveOccurred())
						configs = append(configs, cfg)
					}
				})
			}, SamplingConfig{N: 5})

			AddReportEntry("Average config memory test", experiment.GetStats("config_memory").DurationFor(StatMean))
		})
	})

	Describe("Stress Tests", func() {
		It("should handle rapid client creation and destruction", func() {
			experiment := NewExperiment("Rapid Client Lifecycle")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("lifecycle_50", func() {
					for i := 0; i < 50; i++ {
						client := newTestSMTPClient(cfg)
						client.Close()
					}
				})
			}, SamplingConfig{N: 20})

			AddReportEntry("Average 50 lifecycles", experiment.GetStats("lifecycle_50").DurationFor(StatMean))
		})

		It("should handle rapid config updates", func() {
			experiment := NewExperiment("Rapid Config Updates")
			AddReportEntry(experiment.Name, experiment)

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("updates_100", func() {
					for i := 0; i < 100; i++ {
						newCfg := newTestConfig(testSMTPHost, testSMTPPort+i%10, smtptp.TLSNone)
						tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
						client.UpdConfig(newCfg, tlsConfig)
					}
				})
			}, SamplingConfig{N: 10})

			AddReportEntry("Average 100 updates", experiment.GetStats("updates_100").DurationFor(StatMean))
		})
	})

	Describe("Comparison Tests", func() {
		It("should compare TLS modes performance", func() {
			experiment := NewExperiment("TLS Modes Comparison")
			AddReportEntry(experiment.Name, experiment)

			modes := []struct {
				name string
				mode smtptp.TLSMode
			}{
				{"none", smtptp.TLSNone},
				{"starttls", smtptp.TLSStartTLS},
				{"strict", smtptp.TLSStrictTLS},
			}

			for _, m := range modes {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, m.mode)
				metricName := m.name + "_create"

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(metricName, func() {
						client := newTestSMTPClient(cfg)
						client.Close()
					})
				}, SamplingConfig{N: 50})

				AddReportEntry("Average "+metricName, experiment.GetStats(metricName).DurationFor(StatMean))
			}
		})
	})
})
