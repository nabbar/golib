/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package errors

const (
	MinPkgArchive        = 100
	MinPkgArtifact       = 200
	MinPkgCertificate    = 300
	MinPkgCluster        = 400
	MinPkgConfig         = 500
	MinPkgConsole        = 800
	MinPkgCrypt          = 900
	MinPkgDatabase       = 1000
	MinPkgFTPClient      = 1100
	MinPkgHttpCli        = 1200
	MinPkgHttpServer     = 1300
	MinPkgHttpServerPool = 1320
	MinPkgIOUtils        = 1400
	MinPkgLDAP           = 1500
	MinPkgLogger         = 1600
	MinPkgMail           = 1700
	MinPkgMailer         = 1800
	MinPkgMailPooler     = 1900
	MinPkgMonitor        = 2000
	MinPkgMonitorCfg     = 2020
	MinPkgMonitorPool    = 2100
	MinPkgNetwork        = 2200
	MinPkgNats           = 2300
	MinPkgNutsDB         = 2400
	MinPkgOAuth          = 2500
	MinPkgAws            = 2600
	MinPkgRequest        = 2700
	MinPkgRouter         = 2800
	MinPkgSemaphore      = 2900
	MinPkgSMTP           = 3000
	MinPkgSMTPConfig     = 3050
	MinPkgStatic         = 3100
	MinPkgStatus         = 3200
	MinPkgVersion        = 3300
	MinPkgViper          = 3400

	MinAvailable = 4000

	// MIN_AVAILABLE @Deprecated use MinAvailable constant
	MIN_AVAILABLE = MinAvailable
)
