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
	MinPkgConsole     = 500
	MinPkgCrypt       = 600
	MinPkgHttpCli     = 700
	MinPkgHttpServer  = 800
	MinPkgIOUtils     = 900
	MinPkgLDAP        = 1000
	MinPkgMail        = 1100
	MinPkgMailer      = 1200
	MinPkgMailPooler  = 1300
	MinPkgNetwork     = 1400
	MinPkgNats        = 1500
	MinPkgNutsDB      = 1600
	MinPkgOAuth       = 1700
	MinPkgAws         = 1800
	MinPkgRouter      = 1900
	MinPkgSemaphore   = 2000
	MinPkgSMTP        = 2100
	MinPkgStatic      = 2200
	MinPkgVersion     = 2300

	MinAvailable = 4000

	// MIN_AVAILABLE @Deprecated use MinAvailable constant
	MIN_AVAILABLE = MinAvailable
)
