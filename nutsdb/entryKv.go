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

import (
	"encoding/json"

	"github.com/xujiajun/nutsdb"

	liberr "github.com/nabbar/golib/errors"
)

type DataKV struct {
	Key []byte `mapstructure:"key" json:"key" yaml:"key" toml:"key" cbor:"key"`
	Val []byte `mapstructure:"val" json:"val" yaml:"val" toml:"val" cbor:"val"`
}

func DataKVFromJson(p []byte) (*DataKV, liberr.Error) {
	d := DataKV{}

	if e := json.Unmarshal(p, &d); e != nil {
		return nil, ErrorLogEntryUnmarshal.ErrorParent(e)
	}

	return &d, nil
}

func (d *DataKV) SetToTx(Tx *nutsdb.Tx, bucket string) liberr.Error {
	if Tx == nil {
		return ErrorTransactionClosed.Error(nil)
	}

	if e := Tx.SAdd(bucket, d.Key, d.Val); e != nil {
		return ErrorLogEntryAdd.ErrorParent(e)
	}

	return nil
}
