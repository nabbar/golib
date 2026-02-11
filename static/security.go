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
 */

package static

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libatm "github.com/nabbar/golib/atomic"
	libhtc "github.com/nabbar/golib/httpcli"
	loglvl "github.com/nabbar/golib/logger/level"
)

// secEvt represents a security event to be reported to external systems.
// This is a private type used internally for security event notifications.
// Events are sent to webhooks or callbacks configured in SecurityConfig.
type secEvt struct {
	Timestamp     time.Time         `json:"timestamp"`
	EventType     SecurityEventType `json:"event_type"`
	Severity      string            `json:"severity"` // low, medium, high, critical
	IP            string            `json:"ip"`
	Path          string            `json:"path"`
	Method        string            `json:"method"`
	StatusCode    int               `json:"status_code"`
	UserAgent     string            `json:"user_agent"`
	Referer       string            `json:"referer"`
	Details       map[string]string `json:"details,omitempty"`
	Blocked       bool              `json:"blocked"`
	RemoteAddr    string            `json:"remote_addr"`
	XForwardedFor string            `json:"x_forwarded_for,omitempty"`
}

// SetSecurityBackend configures the security backend integration.
// This method is thread-safe and uses atomic operations.
//
// If batch processing is enabled (BatchSize > 0), it initializes
// the batch processing system.
func (s *staticHandler) SetSecurityBackend(cfg SecurityConfig) {
	s.sec.Store(&cfg)

	// Initialize batch system if necessary
	if cfg.Enabled && cfg.BatchSize > 0 {
		s.initSecuBatch()
	}
}

// GetSecurityBackend returns the current security backend configuration.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) GetSecurityBackend() SecurityConfig {
	if cfg := s.sec.Load(); cfg != nil {
		return *cfg
	}
	return SecurityConfig{}
}

// AddSecurityCallback registers a Go callback function for security events.
// The callback will be invoked asynchronously when security events occur.
// This method is thread-safe.
func (s *staticHandler) AddSecurityCallback(callback SecuEvtCallback) {
	cfg := s.GetSecurityBackend()
	cfg.Callbacks = append(cfg.Callbacks, callback)
	s.SetSecurityBackend(cfg)
}

// notifySecuEvt notifies external systems of a security event.
// This method handles:
//   - Severity filtering
//   - Go callbacks (async)
//   - Webhooks (sync/async)
//   - Batch processing
//
// The method is non-blocking if WebhookAsync is true or callbacks are used.
func (s *staticHandler) notifySecuEvt(event secEvt) {
	cfg := s.GetSecurityBackend()

	if !cfg.Enabled {
		return
	}

	// Check minimum severity
	if !s.shouldNotifySeverity(event.Severity, cfg.MinSeverity) {
		return
	}

	// Go callbacks
	for _, callback := range cfg.Callbacks {
		if callback != nil {
			go callback(event) // Async to avoid blocking
		}
	}

	// Webhook
	if cfg.WebhookURL != "" {
		if cfg.BatchSize > 0 {
			s.qEvtForBatch(event)
		} else {
			if cfg.WebhookAsync {
				go s.sendWebhook(event, cfg)
			} else {
				s.sendWebhook(event, cfg)
			}
		}
	}
}

// sendWebhook sends a single security event to the configured webhook URL.
// The event is sent as JSON or CEF format depending on configuration.
// Errors are logged but do not interrupt the request handling.
func (s *staticHandler) sendWebhook(event secEvt, cfg SecurityConfig) {
	var (
		err error
		buf *bytes.Buffer
		cnt []byte
		cli = libhtc.GetClient()
		req *http.Request
		rsp *http.Response
	)

	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
		if buf != nil {
			buf.Reset()
		}
		if len(cnt) > 0 {
			cnt = cnt[:0]
		}
	}()

	cli.Timeout = cfg.WebhookTimeout

	if cfg.EnableCEFFormat {
		buf = bytes.NewBuffer([]byte(s.formatCEF(event)))
	} else if cnt, err = json.Marshal(event); err != nil {
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "failed to marshal security event")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	} else {
		buf = bytes.NewBuffer(cnt)
	}

	if req, err = http.NewRequest(http.MethodPost, cfg.WebhookURL, buf); err != nil { // #nosec nolint
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "failed to create webhook request")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	}

	// Custom headers
	if cfg.EnableCEFFormat {
		req.Header.Set("Content-Type", "text/plain")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range cfg.WebhookHeaders {
		req.Header.Set(key, value)
	}

	if rsp, err = cli.Do(req); err != nil { // #nosec nolint
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "webhook request failed")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	}

	if rsp.StatusCode >= 400 {
		ent := s.getLogger().Entry(loglvl.WarnLevel, "webhook returned error status")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.FieldAdd("status", rsp.StatusCode)
		ent.Log()
	}
}

// formatCEF formats an event in CEF (Common Event Format).
// CEF is a standard format supported by many SIEM systems including
// Splunk, ArcSight, and QRadar.
//
// Format: CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
func (s *staticHandler) formatCEF(event secEvt) string {
	return fmt.Sprintf(
		"CEF:0|golib|static|1.0|%s|%s|%s|src=%s spt=%s request=%s cs1Label=UserAgent cs1=%s cs2Label=Referer cs2=%s outcome=%s",
		event.EventType,
		event.EventType,
		s.severityToCEF(event.Severity),
		event.IP,
		event.Method,
		event.Path,
		event.UserAgent,
		event.Referer,
		s.blockedToOutcome(event.Blocked),
	)
}

// severityToCEF converts severity level to CEF numeric value (0-10).
// Mapping:
//   - low: 3
//   - medium: 5
//   - high: 8
//   - critical: 10
func (s *staticHandler) severityToCEF(severity string) string {
	switch severity {
	case "low":
		return "3"
	case "medium":
		return "5"
	case "high":
		return "8"
	case "critical":
		return "10"
	default:
		return "5"
	}
}

// blockedToOutcome converts blocked status to CEF outcome field.
func (s *staticHandler) blockedToOutcome(blocked bool) string {
	if blocked {
		return "blocked"
	}
	return "allowed"
}

// shouldNotifySeverity checks if an event's severity meets the minimum threshold.
// Returns true if the event should be notified.
func (s *staticHandler) shouldNotifySeverity(eventSeverity, minSeverity string) bool {
	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	eventLevel := severityLevels[eventSeverity]
	minLevel := severityLevels[minSeverity]

	return eventLevel >= minLevel
}

// newSecuEvt creates a security event from Gin context.
// It extracts relevant information from the HTTP request including
// IP address, headers, and request details.
func (s *staticHandler) newSecuEvt(c *ginsdk.Context, eventType SecurityEventType, severity string, blocked bool, details map[string]string) secEvt {
	evt := secEvt{
		Timestamp:  time.Now(),
		EventType:  eventType,
		Severity:   severity,
		IP:         c.ClientIP(),
		Path:       c.Request.URL.Path,
		Method:     c.Request.Method,
		StatusCode: c.Writer.Status(),
		UserAgent:  c.GetHeader("User-Agent"),
		Referer:    c.GetHeader("Referer"),
		Details:    details,
		Blocked:    blocked,
		RemoteAddr: c.Request.RemoteAddr,
	}

	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		evt.XForwardedFor = xff
	}

	return evt
}

// evtBatch manages batched security events using atomic operations.
// This structure is thread-safe without mutexes:
//   - seq: Atomic counter for event sequencing
//   - evt: Atomic map storing events
//   - tms: Atomic map storing flush timer
type evtBatch struct {
	seq *atomic.Uint64
	evt libatm.MapTyped[uint64, secEvt]
	tms libatm.MapTyped[uint8, *time.Timer]
}

// initSecuBatch initializes the batch processing system.
// This is called automatically when batch processing is enabled.
func (s *staticHandler) initSecuBatch() {
	s.seb.Store(&evtBatch{
		seq: new(atomic.Uint64),
		evt: libatm.NewMapTyped[uint64, secEvt](),
		tms: libatm.NewMapTyped[uint8, *time.Timer](),
	})
}

// qEvtForBatch adds an event to the batch queue.
// When the batch is full (reaches BatchSize), it's sent immediately.
// Otherwise, a timer is started to flush after BatchTimeout.
func (s *staticHandler) qEvtForBatch(event secEvt) {
	b := s.seb.Load()
	if b == nil {
		return
	}

	b.seq.Add(1)
	b.evt.Store(b.seq.Load(), event)

	cfg := s.GetSecurityBackend()

	// Send immediately if batch is full
	if s.batchLen(b) >= cfg.BatchSize {
		s.flushBatch(b, cfg)
		return
	}

	// Start or reset timer
	if _, l := b.tms.Load(0); !l {
		b.tms.Store(0, time.AfterFunc(cfg.BatchTimeout, func() {
			s.flushBatchTimeout()
		}))
	}
}

// flushBatchTimeout flushes the batch when timeout occurs.
// This is called by the timer created in qEvtForBatch.
func (s *staticHandler) flushBatchTimeout() {
	b := s.seb.Load()
	if b == nil {
		return
	}

	cfg := s.GetSecurityBackend()
	s.flushBatch(b, cfg)
}

func (s *staticHandler) batchLen(b *evtBatch) int {
	nbe := 0
	b.evt.Range(func(_ uint64, _ secEvt) bool {
		nbe++
		return true
	})
	return nbe
}

// flushBatch sends all batched events to the webhook.
// This method is thread-safe and clears the batch after sending.
func (s *staticHandler) flushBatch(b *evtBatch, cfg SecurityConfig) {
	if s.batchLen(b) == 0 {
		return
	}

	// Copy events
	evt := make([]secEvt, 0, b.seq.Load())
	b.seq.Store(0)
	b.evt.Range(func(k uint64, v secEvt) bool {
		evt = append(evt, v)
		b.evt.Delete(k)
		return true
	})

	// Clear the batch
	if t, l := b.tms.Load(0); l && t != nil {
		t.Stop()
		b.tms.Delete(0)
	} else if l {
		b.tms.Delete(0)
	}

	// Send as batch
	go s.sendBatchWebhook(evt, cfg)
}

// sendBatchWebhook sends multiple events in a single webhook call.
// The events are sent as JSON array with event count.
// This reduces network overhead compared to individual requests.
func (s *staticHandler) sendBatchWebhook(events []secEvt, cfg SecurityConfig) {
	var (
		err error
		cli = libhtc.GetClient()
		cnt []byte
		buf *bytes.Buffer
		req *http.Request
		rsp *http.Response
	)

	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
		if buf != nil {
			buf.Reset()
		}
		if len(cnt) > 0 {
			cnt = cnt[:0]
		}
	}()

	cli.Timeout = cfg.WebhookTimeout
	cnt, err = json.Marshal(map[string]interface{}{
		"evt":   events,
		"count": len(events),
	})

	if err != nil {
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "failed to marshal batch event to webhook")
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	} else {
		buf = bytes.NewBuffer(cnt)
	}

	if req, err = http.NewRequest(http.MethodPost, cfg.WebhookURL, buf); err != nil { // #nosec nolint
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "failed to create request to send events to webhook")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.FieldAdd("eventCount", len(events))
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range cfg.WebhookHeaders {
		req.Header.Set(key, value)
	}

	if rsp, err = cli.Do(req); err != nil { // #nosec nolint
		ent := s.getLogger().Entry(loglvl.ErrorLevel, "batch webhook request failed")
		ent.FieldAdd("url", cfg.WebhookURL)
		ent.FieldAdd("eventCount", len(events))
		ent.ErrorAdd(true, err)
		ent.Log()
		return
	}
}
