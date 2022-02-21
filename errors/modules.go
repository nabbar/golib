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
	MinPkgArchive     = 100
	MinPkgArtifact    = 200
	MinPkgCertificate = 300
	MinPkgCluster     = 400
	MinPkgConfig      = 500
	MinPkgConsole     = 600
	MinPkgCrypt       = 700
	MinPkgHttpCli     = 800
	MinPkgHttpServer  = 900
	MinPkgIOUtils     = 1000
	MinPkgLDAP        = 1100
	MinPkgLogger      = 1200
	MinPkgMail        = 1300
	MinPkgMailer      = 1400
	MinPkgMailPooler  = 1500
	MinPkgNetwork     = 1600
	MinPkgNats        = 1700
	MinPkgNutsDB      = 1800
	MinPkgOAuth       = 1900
	MinPkgAws         = 2000
	MinPkgRouter      = 2100
	MinPkgSemaphore   = 2200
	MinPkgSMTP        = 2300
	MinPkgStatic      = 2400
	MinPkgVersion     = 2500
	MinPkgViper       = 2600

	MinAvailable = 4000

	// MIN_AVAILABLE @Deprecated use MinAvailable constant
	MIN_AVAILABLE = MinAvailable
)
