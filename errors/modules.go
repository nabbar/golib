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

// The constants below define the base values for error codes in various internal modules.
// Each module has a reserved range of codes, starting at its respective base value.
// This prevents code collisions across different parts of the library.

const baseSub = 10
const baseInc = baseSub * baseSub
const moreInc = 2 * baseInc

const (
	// MinPkgArchive defines the starting error code for the Archive package.
	MinPkgArchive     = baseInc + iota

	// MinPkgArtifact defines the starting error code for the Artifact package.
	MinPkgArtifact    = baseInc + MinPkgArchive

	// MinPkgCertificate defines the starting error code for the Certificate package.
	MinPkgCertificate = baseInc + MinPkgArtifact

	// MinPkgCluster defines the starting error code for the Cluster package.
	MinPkgCluster     = baseInc + MinPkgCertificate

	// MinPkgConfig defines the starting error code for the Config package.
	MinPkgConfig      = baseInc + MinPkgCluster

	// MinPkgConsole defines the starting error code for the Console package.
	MinPkgConsole     = moreInc + MinPkgConfig

	// MinPkgCrypt defines the starting error code for the Crypt package.
	MinPkgCrypt       = baseInc + MinPkgConsole

	// MinPkgDatabaseGorm defines the starting error code for the Database GORM driver.
	MinPkgDatabaseGorm  = baseInc + MinPkgCrypt

	// MinPkgDatabaseKVDrv defines the starting error code for the Database Key-Value driver.
	MinPkgDatabaseKVDrv = baseSub + MinPkgDatabaseGorm

	// MinPkgDatabaseKVMap defines the starting error code for the Database Key-Value map implementation.
	MinPkgDatabaseKVMap = baseSub + MinPkgDatabaseKVDrv

	// MinPkgDatabaseKVTbl defines the starting error code for the Database Key-Value table implementation.
	MinPkgDatabaseKVTbl = baseSub + MinPkgDatabaseKVMap

	// MinPkgDatabaseKVItm defines the starting error code for the Database Key-Value item implementation.
	MinPkgDatabaseKVItm = baseSub + MinPkgDatabaseKVTbl

	// MinPkgFileProgress defines the starting error code for the File Progress package.
	MinPkgFileProgress     = baseInc + MinPkgDatabaseGorm

	// MinPkgFTPClient defines the starting error code for the FTP Client package.
	MinPkgFTPClient        = baseInc + MinPkgFileProgress

	// MinPkgHttpCli defines the starting error code for the HTTP Client package.
	MinPkgHttpCli          = baseInc + MinPkgFTPClient

	// MinPkgHttpCliDNSMapper defines the starting error code for the HTTP Client DNS Mapper.
	MinPkgHttpCliDNSMapper = baseSub + MinPkgHttpCli

	// MinPkgHttpServer defines the starting error code for the HTTP Server package.
	MinPkgHttpServer     = baseInc + MinPkgHttpCliDNSMapper

	// MinPkgHttpServerPool defines the starting error code for the HTTP Server Pool.
	MinPkgHttpServerPool = baseSub + MinPkgHttpServer

	// MinPkgIOUtils defines the starting error code for the IO Utilities package.
	MinPkgIOUtils    = baseInc + MinPkgHttpServer

	// MinPkgLDAP defines the starting error code for the LDAP package.
	MinPkgLDAP       = baseInc + MinPkgIOUtils

	// MinPkgLogger defines the starting error code for the Logger package.
	MinPkgLogger     = baseInc + MinPkgLDAP

	// MinPkgMail defines the starting error code for the Mail package.
	MinPkgMail       = baseInc + MinPkgLogger

	// MinPkgMailer defines the starting error code for the Mailer package.
	MinPkgMailer     = baseInc + MinPkgMail

	// MinPkgMailPooler defines the starting error code for the Mail Pooler package.
	MinPkgMailPooler = baseInc + MinPkgMailer

	// MinPkgMonitor defines the starting error code for the Monitor package.
	MinPkgMonitor     = baseInc + MinPkgMailPooler

	// MinPkgMonitorCfg defines the starting error code for the Monitor Config.
	MinPkgMonitorCfg  = baseSub + MinPkgMonitor

	// MinPkgMonitorPool defines the starting error code for the Monitor Pool.
	MinPkgMonitorPool = baseSub + MinPkgMonitorCfg

	// MinPkgNetwork defines the starting error code for the Network package.
	MinPkgNetwork   = baseInc + MinPkgMonitor

	// MinPkgNats defines the starting error code for the NATS package.
	MinPkgNats      = baseInc + MinPkgNetwork

	// MinPkgNutsDB defines the starting error code for the NutsDB package.
	MinPkgNutsDB    = baseInc + MinPkgNats

	// MinPkgOAuth defines the starting error code for the OAuth package.
	MinPkgOAuth     = baseInc + MinPkgNutsDB

	// MinPkgAws defines the starting error code for the AWS package.
	MinPkgAws       = baseInc + MinPkgOAuth

	// MinPkgRequest defines the starting error code for the Request package.
	MinPkgRequest   = baseInc + MinPkgAws

	// MinPkgRouter defines the starting error code for the Router package.
	MinPkgRouter    = baseInc + MinPkgRequest

	// MinPkgSemaphore defines the starting error code for the Semaphore package.
	MinPkgSemaphore = baseInc + MinPkgRouter

	// MinPkgSMTP defines the starting error code for the SMTP package.
	MinPkgSMTP       = baseInc + MinPkgSemaphore

	// MinPkgSMTPConfig defines the starting error code for the SMTP Config package.
	MinPkgSMTPConfig = baseInc + MinPkgSMTP

	// MinPkgStatic defines the starting error code for the Static package.
	MinPkgStatic  = baseInc + MinPkgSMTPConfig

	// MinPkgStatus defines the starting error code for the Status package.
	MinPkgStatus  = baseInc + MinPkgStatic

	// MinPkgSocket defines the starting error code for the Socket package.
	MinPkgSocket  = baseInc + MinPkgStatus

	// MinPkgVersion defines the starting error code for the Version package.
	MinPkgVersion = baseInc + MinPkgSocket

	// MinPkgViper defines the starting error code for the Viper package.
	MinPkgViper   = baseInc + MinPkgVersion

	// MinAvailable defines the starting point for custom user-defined error codes.
	MinAvailable = baseInc + MinPkgViper
)
