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
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	liberr "github.com/nabbar/golib/errors"
	"github.com/xujiajun/nutsdb"
)

type CommandRequest struct {
	Cmd    CmdCode       `mapstructure:"cmd" json:"cmd" yaml:"cmd" toml:"cmd" cbor:"cmd"`
	Params []interface{} `mapstructure:"params" json:"params" yaml:"params" toml:"params" cbor:"params"`
}

type CommandResponse struct {
	Error error         `mapstructure:"error" json:"error" yaml:"error" toml:"error" cbor:"error"`
	Value []interface{} `mapstructure:"value" json:"value" yaml:"value" toml:"value" cbor:"value"`
}

func NewCommand() *CommandRequest {
	return &CommandRequest{}
}

func NewCommandByDecode(p []byte) (*CommandRequest, liberr.Error) {
	d := CommandRequest{}

	if e := cbor.Unmarshal(p, &d); e != nil {
		return nil, ErrorCommandUnmarshal.ErrorParent(e)
	}

	return &d, nil
}

func (c *CommandRequest) EncodeRequest() ([]byte, liberr.Error) {
	if p, e := cbor.Marshal(c); e != nil {
		return nil, ErrorCommandMarshal.ErrorParent(e)
	} else {
		return p, nil
	}
}

func (c *CommandRequest) DecodeResult(p []byte) (*CommandResponse, liberr.Error) {
	res := CommandResponse{}

	if e := cbor.Unmarshal(p, &res); e != nil {
		return nil, ErrorCommandResultUnmarshal.ErrorParent(e)
	} else {
		return &res, nil
	}
}

func (c *CommandRequest) RunLocal(tx *nutsdb.Tx) (*CommandResponse, liberr.Error) {
	if tx == nil {
		return nil, ErrorTransactionClosed.Error(nil)
	}

	if c.Cmd == CmdUnknown {
		return nil, ErrorClientCommandInvalid.Error(nil)
	}

	method := reflect.ValueOf(tx).MethodByName(c.Cmd.Name())
	nbPrm := method.Type().NumIn()

	if len(c.Params) != nbPrm {
		return nil, ErrorClientCommandParamsBadNumber.ErrorParent(fmt.Errorf("%s need %d parameters", c.Cmd.Name(), nbPrm))
	}

	params := make([]reflect.Value, nbPrm)
	for i := 0; i < nbPrm; i++ {
		params[i] = reflect.ValueOf(c.Params[i])
	}

	resp := method.Call(params)
	ret := CommandResponse{
		Error: nil,
		Value: make([]interface{}, 0),
	}

	for i := 0; i < len(resp); i++ {
		v := resp[i].Interface()
		if e, ok := v.(error); ok {
			ret.Error = e
		} else {
			ret.Value = append(ret.Value, v)
		}
	}

	if ret.Error == nil && len(ret.Value) < 1 {
		return nil, nil
	}

	return &ret, nil
}

func (c *CommandRequest) Run(tx *nutsdb.Tx) ([]byte, liberr.Error) {
	if r, err := c.RunLocal(tx); err != nil {
		return nil, err
	} else if p, e := cbor.Marshal(r); e != nil {
		return nil, ErrorCommandResultMarshal.ErrorParent(e)
	} else {
		return p, nil
	}
}
