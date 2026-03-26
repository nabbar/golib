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

package monitor

import (
	"context"
	"time"

	libdur "github.com/nabbar/golib/duration"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	montps "github.com/nabbar/golib/monitor/types"
	"github.com/nabbar/golib/runner"
)

// runCfg is an internal structure that holds the normalized runtime configuration for the monitor.
//
// Fields Detail:
//   - checkTimeout: maximum execution time for a diagnostic function.
//   - intervalCheck: regular frequency of diagnostics when status is stable.
//   - intervalFall: frequency adjustment during status transitions.
//   - intervalRise: frequency adjustment during status transitions.
//   - fallCountKO: failure count thresholds for health degradation.
//   - fallCountWarn: failure count thresholds for health degradation.
//   - riseCountKO: success count thresholds for health improvement.
//   - riseCountWarn: success count thresholds for health improvement.
type runCfg struct {
	checkTimeout  time.Duration
	intervalCheck time.Duration
	intervalFall  time.Duration
	intervalRise  time.Duration
	fallCountKO   uint8
	fallCountWarn uint8
	riseCountKO   uint8
	riseCountWarn uint8
}

// defConfig initializes the monitor with a set of default safe configuration values.
//
// Logic:
// Ensures that intervals and timeouts are at least 1 or 5 seconds and that counters are at least 1.
// The resulting configuration is stored in the monitor's internal context (keyConfig).
func (o *mon) defConfig() *runCfg {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/defConfig", r)
		}
	}()

	cfg := &runCfg{}

	// Normalize minimum safety thresholds.
	if cfg.checkTimeout < 5*time.Second {
		cfg.checkTimeout = 5 * time.Second
	}

	if cfg.intervalCheck < time.Second {
		cfg.intervalCheck = time.Second
	}

	if cfg.intervalFall < time.Second {
		cfg.intervalFall = cfg.intervalCheck
	}

	if cfg.intervalRise < time.Second {
		cfg.intervalRise = cfg.intervalCheck
	}

	if cfg.fallCountKO < 1 {
		cfg.fallCountKO = 1
	}

	if cfg.fallCountWarn < 1 {
		cfg.fallCountWarn = 1
	}

	if cfg.riseCountKO < 1 {
		cfg.riseCountKO = 1
	}

	if cfg.riseCountWarn < 1 {
		cfg.riseCountWarn = 1
	}

	o.x.Store(keyConfig, cfg)
	return cfg
}

// RegisterLoggerDefault stores a provider function for a default logger.
//
// Provider Pattern:
// This provider is used as a fallback if no specific logger is configured for the monitor.
// If the provider function (liblog.FuncLog) is nil, a dummy function returning nil is registered.
func (o *mon) RegisterLoggerDefault(fct liblog.FuncLog) {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/RegisterLoggerDefault", r)
		}
	}()

	if fct == nil {
		fct = func() liblog.Logger {
			return nil
		}
	}

	o.x.Store(keyLoggerDef, fct)
}

// getLoggerDefault retrieves the active logger from the registered default provider.
// Returns nil if no provider is registered or if the provider returns nil.
func (o *mon) getLoggerDefault() liblog.Logger {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getLoggerDefault", r)
		}
	}()

	if i, l := o.x.Load(keyLoggerDef); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k || v == nil {
		return nil
	} else {
		return v()
	}
}

// SetConfig applies a new configuration to the monitor instance.
//
// Parameters:
//   - ctx: Context for the configuration process.
//   - cfg: The public Config structure containing new parameters.
//
// Normalization Process:
//  1. Validates the provided configuration.
//  2. Updates the monitor identifier (name).
//  3. Updates metadata (Info) if custom data is provided.
//  4. Normalizes durations to prevent invalid (zero or negative) intervals.
//  5. Rebuilds the structured logger (o.x.Store(keyLogger)) with monitor-specific fields.
//
// Note:
// This method is thread-safe and safe to call while the monitor is running.
func (o *mon) SetConfig(ctx context.Context, cfg montps.Config) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/SetConfig", r)
		}
	}()

	if ctx == nil {
		ctx = o.x
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// Identifier management.
	if len(cfg.Name) < 1 {
		o.x.Store(keyName, defaultMonitorName)
	} else {
		o.x.Store(keyName, cfg.Name)
	}

	// Update metadata implementations that support SetData.
	if len(cfg.Data) > 0 {
		if i := o.i.Load(); i != nil {
			if v, k := i.(montps.InfoSet); k {
				v.SetData(cfg.Data)
				o.i.Store(v)
			}
		}
	}

	// Normalization of execution parameters.
	cnf := &runCfg{
		checkTimeout:  cfg.CheckTimeout.Time(),
		intervalCheck: cfg.IntervalCheck.Time(),
		intervalFall:  cfg.IntervalFall.Time(),
		intervalRise:  cfg.IntervalRise.Time(),
		fallCountKO:   cfg.FallCountKO,
		fallCountWarn: cfg.FallCountWarn,
		riseCountKO:   cfg.RiseCountKO,
		riseCountWarn: cfg.RiseCountWarn,
	}

	// Safety thresholds for microsecond precision.
	if cnf.checkTimeout < time.Microsecond {
		cnf.checkTimeout = 5 * time.Second
	}

	if cnf.intervalCheck < time.Microsecond {
		cnf.intervalCheck = time.Second
	}

	if cnf.intervalFall < time.Microsecond {
		cnf.intervalFall = cnf.intervalCheck
	}

	if cnf.intervalRise < time.Microsecond {
		cnf.intervalRise = cnf.intervalCheck
	}

	if cnf.fallCountKO < 1 {
		cnf.fallCountKO = 1
	}

	if cnf.fallCountWarn < 1 {
		cnf.fallCountWarn = 1
	}

	if cnf.riseCountKO < 1 {
		cnf.riseCountKO = 1
	}

	if cnf.riseCountWarn < 1 {
		cnf.riseCountWarn = 1
	}

	o.x.Store(keyConfig, cnf)

	// Logger Rebuild Workflow:
	// Creates a new logger with specific context fields (process, name).
	var (
		def  liblog.FuncLog = o.getLoggerDefault
		n, e                = liblog.NewFrom(ctx, &cfg.Logger, def)
	)

	f := n.GetFields()
	f = f.Add(LogFieldProcess, LogValueProcess)
	f = f.Add(LogFieldName, cfg.Name)
	n.SetFields(f)

	o.x.Store(keyLogger, n)
	return e
}

// GetConfig reconstructs and returns a public Config structure from the monitor's internal state.
// This is used to inspect the current effective configuration (including normalized defaults).
func (o *mon) GetConfig() montps.Config {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/GetConfig", r)
		}
	}()

	cfg := o.getCfg()
	if cfg == nil {
		cfg = &runCfg{}
	}

	opt := o.getLogger().GetOptions()
	if opt == nil {
		opt = &logcfg.Options{}
	}

	return montps.Config{
		Name:          o.getName(),
		CheckTimeout:  libdur.ParseDuration(cfg.checkTimeout),
		IntervalCheck: libdur.ParseDuration(cfg.intervalCheck),
		IntervalFall:  libdur.ParseDuration(cfg.intervalFall),
		IntervalRise:  libdur.ParseDuration(cfg.intervalRise),
		FallCountKO:   cfg.fallCountKO,
		FallCountWarn: cfg.fallCountWarn,
		RiseCountKO:   cfg.riseCountKO,
		RiseCountWarn: cfg.riseCountWarn,
		Logger:        *opt,
	}
}

// getName is an internal helper that retrieves the monitor's display name from its internal context.
func (o *mon) getName() string {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getName", r)
		}
	}()

	if i, l := o.x.Load(keyName); !l {
		return defaultMonitorName
	} else if v, k := i.(string); !k {
		return defaultMonitorName
	} else {
		return v
	}
}

// getCfg is an internal helper that retrieves the normalized runtime configuration (runCfg).
func (o *mon) getCfg() *runCfg {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getCfg", r)
		}
	}()

	if i, l := o.x.Load(keyConfig); !l {
		return o.defConfig()
	} else if v, k := i.(*runCfg); !k {
		return o.defConfig()
	} else {
		return v
	}
}

// getLog is an internal helper that retrieves the current structured logger.
func (o *mon) getLog() liblog.Logger {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getLog", r)
		}
	}()

	if i, l := o.x.Load(keyLogger); !l {
		return nil
	} else if v, k := i.(liblog.Logger); !k || v == nil {
		return nil
	} else {
		return v
	}
}

// getLogger is an internal helper that ensures a valid logger instance is always available.
// If no logger is configured, it creates a basic one based on the monitor's context.
func (o *mon) getLogger() liblog.Logger {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getLogger", r)
		}
	}()

	i := o.getLog()

	if i == nil {
		return liblog.New(o.x)
	} else {
		return i
	}
}

// getFct is an internal helper that retrieves the registered diagnostic function.
func (o *mon) getFct() montps.HealthCheck {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getFct", r)
		}
	}()

	if i, l := o.x.Load(keyHealthCheck); !l {
		return nil
	} else if v, k := i.(montps.HealthCheck); !k {
		return nil
	} else {
		return v
	}
}
