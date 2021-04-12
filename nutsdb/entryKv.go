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
	"runtime"

	"github.com/fxamacker/cbor/v2"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	"github.com/xujiajun/nutsdb"
)

const (
	_MinSkipCaller = 2
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

func NewCommandByCaller(params ...interface{}) *CommandRequest {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(_MinSkipCaller, pc)
	f := runtime.FuncForPC(pc[0])

	d := &CommandRequest{}
	d.Cmd = CmdCodeFromName(f.Name())

	if len(params) > 0 {
		d.Params = params
	}

	return d
}

func (c *CommandRequest) InitParams(num int) {
	c.Params = make([]interface{}, num)
}

func (c *CommandRequest) InitParamsCounter(num int) int {
	c.InitParams(num)
	return 0
}

func (c *CommandRequest) SetParams(num int, val interface{}) {
	if num < len(c.Params) {
		c.Params[num] = val
		return
	}

	tmp := c.Params
	c.Params = make([]interface{}, len(c.Params)+1)

	if len(tmp) > 0 {
		for i := 0; i < len(tmp); i++ {
			c.Params[i] = tmp[i]
		}
	}

	c.Params[num] = val
}

func (c *CommandRequest) SetParamsInc(num int, val interface{}) int {
	c.SetParams(num, val)
	num++
	return num
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

	valTx := reflect.ValueOf(tx)
	mtName := c.Cmd.Name()
	method := valTx.MethodByName(mtName)
	nbPrm := method.Type().NumIn()

	if len(c.Params) != nbPrm {
		//nolint #goerr113
		return nil, ErrorClientCommandParamsBadNumber.ErrorParent(fmt.Errorf("%s need %d parameters", c.Cmd.Name(), nbPrm))
	}

	params := make([]reflect.Value, nbPrm)
	for i := 0; i < nbPrm; i++ {
		v := reflect.ValueOf(c.Params[i])
		liblog.DebugLevel.Logf("Param %d : type %s - Val %v", i, v.Type().Name(), v.Interface())

		if v.Type().Kind() == method.Type().In(i).Kind() {
			params[i] = v
			continue
		}

		if !v.Type().ConvertibleTo(method.Type().In(i)) {
			//nolint #goerr113
			return nil, ErrorClientCommandParamsMismatching.ErrorParent(fmt.Errorf("cmd: %s", mtName), fmt.Errorf("param num: %d", i), fmt.Errorf("param type: %s, avaitting type: %s", v.Type().Kind(), method.Type().In(i).Kind()))
		}

		//nolint #exhaustive
		switch method.Type().In(i).Kind() {
		case reflect.Bool:
			params[i] = reflect.ValueOf(v.Bool())
		case reflect.Int:
			params[i] = reflect.ValueOf(int(v.Int()))
		case reflect.Int8:
			params[i] = reflect.ValueOf(int8(v.Int()))
		case reflect.Int16:
			params[i] = reflect.ValueOf(int8(v.Int()))
		case reflect.Int32:
			params[i] = reflect.ValueOf(int16(v.Int()))
		case reflect.Int64:
			params[i] = reflect.ValueOf(v.Int())
		case reflect.Uintptr:
			params[i] = reflect.ValueOf(v.UnsafeAddr())
		case reflect.Uint:
			params[i] = reflect.ValueOf(uint(v.Uint()))
		case reflect.Uint8:
			params[i] = reflect.ValueOf(uint8(v.Uint()))
		case reflect.Uint16:
			params[i] = reflect.ValueOf(uint16(v.Uint()))
		case reflect.Uint32:
			params[i] = reflect.ValueOf(uint32(v.Uint()))
		case reflect.Uint64:
			params[i] = reflect.ValueOf(v.Uint())
		case reflect.Float32:
			params[i] = reflect.ValueOf(float32(v.Float()))
		case reflect.Float64:
			params[i] = reflect.ValueOf(v.Float())
		case reflect.Complex64:
			params[i] = reflect.ValueOf(complex64(v.Complex()))
		case reflect.Complex128:
			params[i] = reflect.ValueOf(v.Complex())
		case reflect.Interface:
			params[i] = reflect.ValueOf(v.Interface())
		case reflect.String:
			params[i] = reflect.ValueOf(v.String())
		}

		liblog.DebugLevel.Logf("Change Param %d : type %s to %v", i, v.Type().Name(), params[i].Type().Name())
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
