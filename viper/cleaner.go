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
	"bytes"
	"encoding/json"
	"strings"

	spfvpr "github.com/spf13/viper"
)

func (v *viper) Unset(key ...string) error {
	if len(key) < 1 {
		return nil
	}

	configMap := v.v.AllSettings()

	for _, k := range key {
		configMap = unsetSub(configMap, k)
	}

	_v := &viper{
		v:      spfvpr.New(),
		base:   v.base,
		prfx:   v.prfx,
		deft:   v.deft,
		remote: v.remote,
	}

	var (
		encodedConfig []byte
		err           error
	)

	if encodedConfig, err = json.MarshalIndent(configMap, "", " "); err != nil {
		return err
	}

	_v.v.SetConfigType("json")

	if err = _v.v.ReadConfig(bytes.NewReader(encodedConfig)); err != nil {
		return err
	}

	v.v = _v.v

	return nil
}

func unsetSub(configMap map[string]interface{}, key string) map[string]interface{} {
	kkey := strings.Split(key, ".")

	if len(kkey) == 1 {
		delete(configMap, key)
		return configMap
	}

	for k, v := range configMap {
		if k != kkey[0] {
			continue
		}

		configMap[k] = unsetSub(v.(map[string]interface{}), strings.Join(kkey[1:], "."))
		break
	}

	return configMap
}
