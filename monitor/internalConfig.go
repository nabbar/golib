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

// runCfg holds the internal runtime configuration for the monitor.
// It stores normalized duration and count values used during health check execution.
type runCfg struct {
	checkTimeout  time.Duration // Maximum duration for a health check to complete
	intervalCheck time.Duration // Interval between normal health checks
	intervalFall  time.Duration // Interval when status is falling
	intervalRise  time.Duration // Interval when status is rising
	fallCountKO   uint8         // Number of failures needed to transition from Warn to KO
	fallCountWarn uint8         // Number of failures needed to transition from OK to Warn
	riseCountKO   uint8         // Number of successes needed to transition from KO to Warn
	riseCountWarn uint8         // Number of successes needed to transition from Warn to OK
}

// defConfig creates and stores a default configuration with minimum safe values.
// This is used when no configuration has been explicitly set.
func (o *mon) defConfig() *runCfg {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/defConfig", r)
		}
	}()

	cfg := &runCfg{}

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

// RegisterLoggerDefault registers a default logger function provider.
// This logger is used as a fallback when creating new loggers.
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

// getLoggerDefault retrieves the default logger if one has been registered.
// Returns nil if no default logger has been set.
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

// SetConfig updates the monitor configuration.
// It validates the configuration, normalizes values to safe minimums,
// and initializes a logger with the provided options.
// Returns an error if validation fails or logger initialization fails.
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

	if len(cfg.Name) < 1 {
		o.x.Store(keyName, defaultMonitorName)
	} else {
		o.x.Store(keyName, cfg.Name)
	}

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

	var n, e = liblog.NewFrom(ctx, &cfg.Logger, o.getLoggerDefault)

	f := n.GetFields()
	f = f.Add(LogFieldProcess, LogValueProcess)
	f = f.Add(LogFieldName, cfg.Name)
	n.SetFields(f)

	o.x.Store(keyLogger, n)
	return e
}

// GetConfig returns the current monitor configuration.
// It builds a Config from the internal runCfg and logger options.
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

// getName retrieves the configured monitor name.
// Returns the default name if none has been set.
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

// getCfg retrieves the current runtime configuration.
// Returns the default configuration if none has been set.
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

// getLog retrieves the configured logger.
// Returns nil if no logger has been configured.
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

// getLogger retrieves the configured logger or creates a new one.
// This always returns a valid logger instance.
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

// getFct retrieves the registered health check function.
// Returns nil if no health check has been registered.
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

// getLastCheck retrieves the last check results.
// Returns a new lastRun instance if none exists.
func (o *mon) getLastCheck() *lastRun {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/getLastCheck", r)
		}
	}()

	if i, l := o.x.Load(keyLastRun); !l {
		return newLastRun()
	} else if v, k := i.(*lastRun); !k {
		return newLastRun()
	} else {
		return v
	}
}

// setLastCheck stores the results from the last health check execution.
func (o *mon) setLastCheck(l *lastRun) {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/monitor/setLastCheck", r)
		}
	}()

	o.x.Store(keyLastRun, l)
}
