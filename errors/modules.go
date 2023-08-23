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

const baseSub = 10
const baseInc = baseSub * baseSub
const moreInc = 2 * baseInc

const (
	MinPkgArchive     = baseInc + iota
	MinPkgArtifact    = baseInc + MinPkgArchive
	MinPkgCertificate = baseInc + MinPkgArtifact
	MinPkgCluster     = baseInc + MinPkgCertificate
	MinPkgConfig      = baseInc + MinPkgCluster
	MinPkgConsole     = moreInc + MinPkgConfig
	MinPkgCrypt       = baseInc + MinPkgConsole

	MinPkgDatabaseGorm  = baseInc + MinPkgCrypt
	MinPkgDatabaseKVDrv = baseSub + MinPkgDatabaseGorm
	MinPkgDatabaseKVMap = baseSub + MinPkgDatabaseKVDrv
	MinPkgDatabaseKVTbl = baseSub + MinPkgDatabaseKVMap
	MinPkgDatabaseKVItm = baseSub + MinPkgDatabaseKVTbl

	MinPkgFileProgress = baseInc + MinPkgDatabaseGorm
	MinPkgFTPClient    = baseInc + MinPkgFileProgress
	MinPkgHttpCli      = baseInc + MinPkgFTPClient

	MinPkgHttpServer     = baseInc + MinPkgHttpCli
	MinPkgHttpServerPool = baseSub + MinPkgHttpServer

	MinPkgIOUtils    = baseInc + MinPkgHttpServer
	MinPkgLDAP       = baseInc + MinPkgIOUtils
	MinPkgLogger     = baseInc + MinPkgLDAP
	MinPkgMail       = baseInc + MinPkgLogger
	MinPkgMailer     = baseInc + MinPkgMail
	MinPkgMailPooler = baseInc + MinPkgMailer

	MinPkgMonitor     = baseInc + MinPkgMailPooler
	MinPkgMonitorCfg  = baseSub + MinPkgMonitor
	MinPkgMonitorPool = baseSub + MinPkgMonitorCfg

	MinPkgNetwork   = baseInc + MinPkgMonitor
	MinPkgNats      = baseInc + MinPkgNetwork
	MinPkgNutsDB    = baseInc + MinPkgNats
	MinPkgOAuth     = baseInc + MinPkgNutsDB
	MinPkgAws       = baseInc + MinPkgOAuth
	MinPkgRequest   = baseInc + MinPkgAws
	MinPkgRouter    = baseInc + MinPkgRequest
	MinPkgSemaphore = baseInc + MinPkgRouter

	MinPkgSMTP       = baseInc + MinPkgSemaphore
	MinPkgSMTPConfig = baseInc + MinPkgSMTP

	MinPkgStatic  = baseInc + MinPkgSMTPConfig
	MinPkgStatus  = baseInc + MinPkgStatic
	MinPkgSocket  = baseInc + MinPkgStatus
	MinPkgVersion = baseInc + MinPkgSocket
	MinPkgViper   = baseInc + MinPkgVersion

	MinAvailable = baseInc + MinPkgViper
)
