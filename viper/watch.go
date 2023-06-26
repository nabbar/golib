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

package viper

import (
	"time"

	libnot "github.com/fsnotify/fsnotify"
	loglvl "github.com/nabbar/golib/logger/level"
)

func (v *viper) initWatchRemote(logLevelRemoteKO, logLevelRemoteOK loglvl.Level) {
	// open a goroutine to watch remote changes forever
	go func() {
		// unstopped for loop
		for {
			// delay after each request
			time.Sleep(time.Second * 5)

			if v.remote.provider == RemoteETCD {
				if v.logEntry(logLevelRemoteKO, "Remote config watching").ErrorAdd(true, v.v.WatchRemoteConfig()).Check(logLevelRemoteOK) {
					// skip error and try next time
					continue
				}
			} else {
				// reading remote config

				if v.logEntry(logLevelRemoteKO, "Remote config loading").ErrorAdd(true, v.v.ReadRemoteConfig()).Check(logLevelRemoteOK) {
					// skip error and try next time
					continue
				}
			}

			// add config model
			if v.logEntry(logLevelRemoteKO, "Remote config parsing").ErrorAdd(true, v.v.Unmarshal(v.remote.model)).Check(logLevelRemoteOK) {
				// skip error and try next time
				continue
			}
		}
	}()
}

func (v *viper) WatchFS(logLevelFSInfo loglvl.Level) {
	v.v.WatchConfig()
	v.v.OnConfigChange(func(e libnot.Event) {
		if v.remote.fct != nil {
			v.logEntry(logLevelFSInfo, "Reloading local config file '%s'...", e.Name).Log()
			v.remote.fct()
			v.logEntry(logLevelFSInfo, "local config file '%s' has been reloaded.", e.Name).Log()
		}
	})
}
