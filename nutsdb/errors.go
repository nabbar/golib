/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nutsdb

import liberr "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty liberr.CodeError = iota + liberr.MinPkgNutsDB
	ErrorParamsMissing
	ErrorParamsMismatching
	ErrorValidateConfig
	ErrorValidateNutsDB
	ErrorClusterInit
	ErrorFolderCheck
	ErrorFolderCreate
	ErrorFolderCopy
	ErrorFolderDelete
	ErrorFolderExtract
	ErrorFolderArchive
	ErrorFolderCompress
	ErrorDatabaseClosed
	ErrorDatabaseKeyInvalid
	ErrorDatabaseBackup
	ErrorDatabaseSnapshot
	ErrorTransactionInit
	ErrorTransactionClosed
	ErrorTransactionCommit
	ErrorTransactionPutKey
	ErrorCommandInvalid
	ErrorCommandUnmarshal
	ErrorCommandMarshal
	ErrorCommandResultUnmarshal
	ErrorCommandResultMarshal
	ErrorLogEntryAdd
	ErrorClientCommandInvalid
	ErrorClientCommandParamsBadNumber
	ErrorClientCommandParamsMismatching
	ErrorClientCommandCall
	ErrorClientCommandCommit
	ErrorClientCommandResponseInvalid
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = liberr.ExistInMapMessage(ErrorParamsEmpty)
	liberr.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case liberr.UNK_ERROR:
		return ""
	case ErrorParamsEmpty:
		return "at least one given parameter is empty"
	case ErrorParamsMissing:
		return "at least one given parameter is missing"
	case ErrorParamsMismatching:
		return "at least one given parameter does not match the awaiting type"
	case ErrorValidateConfig:
		return "config seems to be invalid"
	case ErrorValidateNutsDB:
		return "database config seems to be invalid"
	case ErrorClusterInit:
		return "cannot start or join cluster"
	case ErrorFolderCheck:
		return "error while trying to check or stat folder"
	case ErrorFolderCreate:
		return "error while trying to create folder"
	case ErrorFolderCopy:
		return "error while trying to copy folder"
	case ErrorFolderArchive:
		return "error while trying to archive folder"
	case ErrorFolderCompress:
		return "error while trying to compress folder"
	case ErrorFolderExtract:
		return "error while trying to extract snapshot archive"
	case ErrorDatabaseClosed:
		return "database is closed"
	case ErrorDatabaseKeyInvalid:
		return "database key seems to be invalid"
	case ErrorDatabaseBackup:
		return "error occured while trying to backup database folder"
	case ErrorDatabaseSnapshot:
		return "error occured while trying to backup database to cluster members"
	case ErrorTransactionInit:
		return "cannot initialize new transaction from database"
	case ErrorTransactionClosed:
		return "transaction is closed"
	case ErrorTransactionCommit:
		return "cannot commit transaction writable into database"
	case ErrorTransactionPutKey:
		return "cannot send Put command into database transaction"
	case ErrorCommandInvalid:
		return "given query is not a valid DB command"
	case ErrorCommandUnmarshal:
		return "cannot unmarshall DB command"
	case ErrorCommandMarshal:
		return "cannot marshall DB command"
	case ErrorCommandResultUnmarshal:
		return "cannot unmarshall DB command result"
	case ErrorCommandResultMarshal:
		return "cannot marshall DB command result"
	case ErrorLogEntryAdd:
		return "cannot add key/value to database"
	case ErrorClientCommandInvalid:
		return "invalid command"
	case ErrorClientCommandParamsBadNumber:
		return "invalid number of parameters for client command"
	case ErrorClientCommandParamsMismatching:
		return "invalid type of parameter for client command"
	case ErrorClientCommandCall:
		return "error occured while running client command"
	case ErrorClientCommandCommit:
		return "error occured while commit client command"
	case ErrorClientCommandResponseInvalid:
		return "response of requested client command seems to be invalid"
	}

	return ""
}
