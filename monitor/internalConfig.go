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
	"time"

	libctx "github.com/nabbar/golib/context"

	"github.com/nabbar/golib/monitor/types"

	monsts "github.com/nabbar/golib/monitor/status"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

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

func (o *mon) defConfig() *runCfg {
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

func (o *mon) RegisterLoggerDefault(fct liblog.FuncLog) {
	o.x.Store(keyLoggerDef, fct)
}

func (o *mon) getLoggerDefault() liblog.Logger {
	if i, l := o.x.Load(keyLoggerDef); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else {
		return v()
	}
}

func (o *mon) SetConfig(ctx libctx.FuncContext, cfg types.Config) liberr.Error {
	if ctx == nil {
		ctx = o.x.GetContext
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
		checkTimeout:  cfg.CheckTimeout,
		intervalCheck: cfg.IntervalCheck,
		intervalFall:  cfg.IntervalFall,
		intervalRise:  cfg.IntervalRise,
		fallCountKO:   cfg.FallCountKO,
		fallCountWarn: cfg.FallCountWarn,
		riseCountKO:   cfg.RiseCountKO,
		riseCountWarn: cfg.RiseCountWarn,
	}

	if cnf.checkTimeout < 5*time.Second {
		cnf.checkTimeout = 5 * time.Second
	}

	if cnf.intervalCheck < time.Second {
		cnf.intervalCheck = time.Second
	}

	if cnf.intervalFall < time.Second {
		cnf.intervalFall = cnf.intervalCheck
	}

	if cnf.intervalRise < time.Second {
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

	var n liblog.Logger

	if l := o.getLoggerDefault(); l == nil {
		n = liblog.New(ctx)
	} else {
		n = l
	}

	if e := n.SetOptions(&cfg.Logger); e != nil {
		return ErrorLoggerError.ErrorParent(e)
	}

	f := n.GetFields()
	n.SetFields(f.Add(LogFieldProcess, LogValueProcess).Add(LogFieldName, cfg.Name))
	n.SetLevel(liblog.GetCurrentLevel())

	if l := o.getLog(); l != nil {
		_ = l.Close()
	}

	o.x.Store(keyLogger, n)
	return nil
}

func (o *mon) GetConfig() types.Config {
	cfg := o.getCfg()
	if cfg == nil {
		cfg = &runCfg{}
	}

	opt := o.getLogger().GetOptions()
	if opt == nil {
		opt = &liblog.Options{}
	}

	return types.Config{
		Name:          o.getName(),
		CheckTimeout:  cfg.checkTimeout,
		IntervalCheck: cfg.intervalCheck,
		IntervalFall:  cfg.intervalFall,
		IntervalRise:  cfg.intervalRise,
		FallCountKO:   cfg.fallCountKO,
		FallCountWarn: cfg.fallCountWarn,
		RiseCountKO:   cfg.riseCountKO,
		RiseCountWarn: cfg.riseCountWarn,
		Logger:        *opt,
	}
}

func (o *mon) getName() string {
	if i, l := o.x.Load(keyName); !l {
		return defaultMonitorName
	} else if v, k := i.(string); !k {
		return defaultMonitorName
	} else {
		return v
	}
}

func (o *mon) getCfg() *runCfg {
	if i, l := o.x.Load(keyConfig); !l {
		return o.defConfig()
	} else if v, k := i.(*runCfg); !k {
		return o.defConfig()
	} else {
		return v
	}
}

func (o *mon) getLog() liblog.Logger {
	if i, l := o.x.Load(keyLogger); !l {
		return nil
	} else if v, k := i.(liblog.Logger); !k {
		return nil
	} else {
		return v
	}
}

func (o *mon) getLogger() liblog.Logger {
	i := o.getLog()

	if i == nil {
		return liblog.GetDefault()
	} else {
		return i
	}
}

func (o *mon) getFct() types.HealthCheck {
	if i, l := o.x.Load(keyHealthCheck); !l {
		return nil
	} else if v, k := i.(types.HealthCheck); !k {
		return nil
	} else {
		return v
	}
}

func (o *mon) getLastCheck() *lastRun {
	if i, l := o.x.Load(keyLastRun); !l {
		return &lastRun{
			status:  monsts.KO,
			runtime: time.Now(),
			isRise:  false,
			isFall:  false,
		}
	} else if v, k := i.(*lastRun); !k {
		return &lastRun{
			status:  monsts.KO,
			runtime: time.Now(),
			isRise:  false,
			isFall:  false,
		}
	} else {
		return v
	}
}

func (o *mon) setLastCheck(m middleWare) error {
	e := m.Next()
	l := &lastRun{
		status:  o.Status(),
		runtime: time.Now(),
		isRise:  o.IsRise(),
		isFall:  o.IsFall(),
	}
	o.x.Store(keyLastRun, l)
	return e
}
